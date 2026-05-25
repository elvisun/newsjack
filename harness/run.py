#!/usr/bin/env python3
"""Container-first Newsjack agent runtime harness."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_IMAGE = "newsjack-agent-harness:local"
DEFAULT_INSTALLER_URL = "https://newsjack.sh/install.sh"
RUNTIMES = ("codex", "claude", "hermes", "openclaw")
EXPECTED_SKILLS = (
    "newsjack-setup",
    "newsjack-detector",
    "media-list-manager",
)
MODEL_ENV_KEYS = (
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
)
SECRET_MARKERS = (
    "API_KEY",
    "TOKEN",
    "SECRET",
    "PASSWORD",
)


@dataclass(frozen=True)
class RuntimeSpec:
    name: str
    binary: str
    version_cmd: tuple[str, ...]
    skills_dir: str
    native_smoke_cmd: tuple[str, ...]
    acp_candidates: tuple[tuple[str, ...], ...]


RUNTIME_SPECS = {
    "codex": RuntimeSpec(
        name="codex",
        binary="codex",
        version_cmd=("codex", "--version"),
        skills_dir=".agents/skills",
        native_smoke_cmd=(
            "codex",
            "exec",
            "--skip-git-repo-check",
            "--ephemeral",
            "--ignore-rules",
            "--sandbox",
            "danger-full-access",
            "--dangerously-bypass-approvals-and-sandbox",
            "Reply exactly READY and stop.",
        ),
        acp_candidates=(
            ("codex", "--acp", "--stdio"),
            ("codex-acp",),
            ("acpx", "codex"),
        ),
    ),
    "claude": RuntimeSpec(
        name="claude",
        binary="claude",
        version_cmd=("claude", "--version"),
        skills_dir=".claude/skills",
        native_smoke_cmd=(
            "claude",
            "--bare",
            "--max-budget-usd",
            "0.05",
            "-p",
            "Reply exactly READY and stop.",
        ),
        acp_candidates=(
            ("claude", "--acp", "--stdio"),
            ("claude-agent-acp",),
            ("acpx", "claude"),
        ),
    ),
    "hermes": RuntimeSpec(
        name="hermes",
        binary="hermes",
        version_cmd=("hermes", "--version"),
        skills_dir=".hermes/skills",
        native_smoke_cmd=(
            "hermes",
            "chat",
            "-Q",
            "--ignore-rules",
            "--provider",
            "anthropic",
            "--max-turns",
            "1",
            "-q",
            "Reply exactly READY and stop.",
        ),
        acp_candidates=(
            ("hermes", "--acp", "--stdio"),
            ("hermes", "acp", "--accept-hooks"),
        ),
    ),
    "openclaw": RuntimeSpec(
        name="openclaw",
        binary="openclaw",
        version_cmd=("openclaw", "--version"),
        skills_dir=".openclaw/skills",
        native_smoke_cmd=(
            "openclaw",
            "agent",
            "--local",
            "--agent",
            "main",
            "--message",
            "Reply exactly READY and stop.",
            "--json",
        ),
        acp_candidates=(
            ("openclaw", "--acp", "--stdio"),
            ("openclaw", "acp"),
        ),
    ),
}


class HarnessError(RuntimeError):
    pass


def log(message: str) -> None:
    print(f"harness: {message}", flush=True)


def redact(value: str) -> str:
    redacted = value
    for key, secret in os.environ.items():
        if not secret:
            continue
        if any(marker in key.upper() for marker in SECRET_MARKERS):
            redacted = redacted.replace(secret, "<redacted>")
    return redacted


def command_text(cmd: Sequence[str]) -> str:
    return " ".join(sh_quote(part) for part in cmd)


def sh_quote(value: str) -> str:
    if not value:
        return "''"
    safe = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+-=.,/:@%"
    if all(ch in safe for ch in value):
        return value
    return "'" + value.replace("'", "'\"'\"'") + "'"


def run_cmd(
    cmd: Sequence[str],
    *,
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
    timeout: int = 120,
    log_path: Path | None = None,
    display: str | None = None,
    check: bool = True,
) -> subprocess.CompletedProcess[str]:
    rendered = display or command_text(cmd)
    rendered = redact(rendered)
    if log_path:
        log_path.parent.mkdir(parents=True, exist_ok=True)
        with log_path.open("a", encoding="utf-8") as handle:
            handle.write(f"\n$ {rendered}\n")
    log(rendered)
    result = subprocess.run(
        list(cmd),
        cwd=str(cwd) if cwd else None,
        env=env,
        text=True,
        capture_output=True,
        timeout=timeout,
    )
    output = ""
    if result.stdout:
        output += result.stdout
    if result.stderr:
        output += result.stderr
    output = redact(output)
    if log_path:
        with log_path.open("a", encoding="utf-8") as handle:
            if output:
                handle.write(output)
                if not output.endswith("\n"):
                    handle.write("\n")
            handle.write(f"[exit {result.returncode}]\n")
    if check and result.returncode != 0:
        raise HarnessError(f"command failed ({result.returncode}): {rendered}\n{output[-4000:]}")
    return result


def run_shell(
    script: str,
    *,
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
    timeout: int = 120,
    log_path: Path | None = None,
    display: str | None = None,
    check: bool = True,
) -> subprocess.CompletedProcess[str]:
    return run_cmd(
        ("bash", "-lc", script),
        cwd=cwd,
        env=env,
        timeout=timeout,
        log_path=log_path,
        display=display or script,
        check=check,
    )


def parse_args(argv: Sequence[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runtime", choices=(*RUNTIMES, "all"), default="all")
    parser.add_argument(
        "--mode",
        choices=("ci-installer", "native-smoke", "acp-smoke", "setup-flow"),
        default="ci-installer",
    )
    source = parser.add_mutually_exclusive_group()
    source.add_argument("--local-source", dest="local_source", action="store_true", default=True)
    source.add_argument("--production-path", dest="local_source", action="store_false")
    parser.add_argument("--installer-url", default=DEFAULT_INSTALLER_URL)
    parser.add_argument("--image", default=DEFAULT_IMAGE)
    parser.add_argument("--env-file", default=None)
    parser.add_argument("--build-image", action="store_true")
    parser.add_argument("--container-runtime", choices=("auto", "docker", "podman"), default="auto")
    parser.add_argument("--inside-container", action="store_true")
    parser.add_argument("--repo", default=str(ROOT))
    parser.add_argument("--runs", default=str(ROOT / "harness" / "runs"))
    parser.add_argument("--run-id", default=None)
    parser.add_argument("--timeout", type=int, default=600)
    return parser.parse_args(argv)


def selected_runtimes(runtime: str) -> list[str]:
    return list(RUNTIMES) if runtime == "all" else [runtime]


def find_container_tool(preference: str) -> str:
    candidates = ("docker", "podman") if preference == "auto" else (preference,)
    for candidate in candidates:
        path = shutil.which(candidate)
        if path:
            return candidate
        if candidate == "docker":
            orbstack_docker = Path("/Applications/OrbStack.app/Contents/MacOS/xbin/docker")
            if orbstack_docker.exists():
                return str(orbstack_docker)
    raise HarnessError("no container runtime found on PATH; install Docker or Podman to run fresh-container tests")


def build_image(tool: str, image: str) -> None:
    run_cmd((tool, "build", "-f", "harness/Dockerfile", "-t", image, "."), cwd=ROOT, timeout=1800)


def host_main(args: argparse.Namespace) -> int:
    tool = find_container_tool(args.container_runtime)
    if args.build_image:
        build_image(tool, args.image)

    run_id = args.run_id or datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    repo = Path(args.repo).resolve()
    runs = Path(args.runs).resolve()
    runs.mkdir(parents=True, exist_ok=True)

    if args.mode == "ci-installer" and args.env_file:
        raise HarnessError("ci-installer must not receive --env-file; it is the no-token lane")

    failures: list[str] = []
    for runtime in selected_runtimes(args.runtime):
        name = f"newsjack-harness-{runtime}-{run_id.lower()}"
        container_cmd = [
            tool,
            "run",
            "--rm",
            "--name",
            name,
            "-v",
            f"{repo}:/repo:ro",
            "-v",
            f"{runs}:/runs",
        ]
        env_file_in_repo = None
        if args.env_file:
            env_file = Path(args.env_file).resolve()
            try:
                rel_env_file = env_file.relative_to(repo)
            except ValueError:
                container_cmd.extend(("--env-file", str(env_file)))
            else:
                env_file_in_repo = "/repo/" + str(rel_env_file)
                container_cmd.extend(("-e", f"NEWSJACK_HARNESS_ENV_FILE={env_file_in_repo}"))
        inner_cmd = [
            "python3",
            "/repo/harness/run.py",
            "--inside-container",
            "--runtime",
            runtime,
            "--mode",
            args.mode,
            "--repo",
            "/repo",
            "--runs",
            "/runs",
            "--timeout",
            str(args.timeout),
            "--installer-url",
            args.installer_url,
            "--local-source" if args.local_source else "--production-path",
        ]
        container_cmd.extend(("-e", f"NEWSJACK_HARNESS_RUN_ID={run_id}", args.image))
        if env_file_in_repo:
            container_cmd.extend(
                (
                    "sh",
                    "-lc",
                    f"set -a; . {sh_quote(env_file_in_repo)}; set +a; exec {command_text(inner_cmd)}",
                )
            )
        else:
            container_cmd.extend(inner_cmd)
        try:
            run_cmd(container_cmd, cwd=repo, timeout=args.timeout + 300)
        except (HarnessError, subprocess.TimeoutExpired) as exc:
            failures.append(f"{runtime}: {exc}")

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    log(f"all requested runtimes passed; run id: {run_id}")
    return 0


def base_container_env(runtime: str, run_dir: Path, repo: Path) -> dict[str, str]:
    env = os.environ.copy()
    run_id = run_dir.parent.name.replace("/", "_")
    home = Path("/home/newsjack-harness") / f"{run_id}-{runtime}"
    env.update(
        {
            "HOME": str(home),
            "XDG_CONFIG_HOME": str(home / ".config"),
            "XDG_CACHE_HOME": str(home / ".cache"),
            "XDG_DATA_HOME": str(home / ".local" / "share"),
            "PATH": f"{home}/.newsjack/bin:{env.get('PATH', '')}",
            "NEWSJACK_RUNTIMES": runtime,
            "NEWSJACK_STORE": str(run_dir / "newsjack.db"),
            "NEWSJACK_INSTALL_MCP": "1",
        }
    )
    if repo.exists():
        env.setdefault("NEWSJACK_SOURCE_DIR", str(repo))
    return env


def assert_no_model_keys(mode: str) -> None:
    present = [key for key in MODEL_ENV_KEYS if os.environ.get(key)]
    if mode == "ci-installer" and present:
        raise HarnessError(
            "ci-installer is the no-token lane and refuses model-provider env vars: "
            + ", ".join(present)
        )


def assert_model_allowed(mode: str) -> None:
    if mode == "ci-installer":
        return
    if os.environ.get("NEWSJACK_HARNESS_ALLOW_MODEL_CALLS") != "1":
        raise HarnessError(
            f"{mode} can call paid model APIs; set NEWSJACK_HARNESS_ALLOW_MODEL_CALLS=1 to run it"
        )


def write_env_json(runtime: str, mode: str, run_dir: Path, env: dict[str, str]) -> None:
    keys = [
        "HOME",
        "NEWSJACK_RUNTIMES",
        "NEWSJACK_SOURCE_DIR",
        "NEWSJACK_INSTALL_MCP",
        "NEWSJACK_STORE",
        "MEDIALYST_API_BASE",
        "MEDIALYST_NEWS_PATH",
        "NEWSJACK_HARNESS_ALLOW_MODEL_CALLS",
    ]
    payload = {
        "runtime": runtime,
        "mode": mode,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "env": {key: env.get(key) for key in keys if env.get(key) is not None},
        "secrets_present": {
            key: bool(env.get(key))
            for key in ("OPENAI_API_KEY", "ANTHROPIC_API_KEY", "MEDIALYST_API_KEY", "X_BEARER_TOKEN", "TWITTER_BEARER_TOKEN")
        },
    }
    (run_dir / "env.json").write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def install_newsjack(args: argparse.Namespace, runtime: str, repo: Path, env: dict[str, str], log_path: Path) -> None:
    env = env.copy()
    if args.local_source:
        env["NEWSJACK_SOURCE_DIR"] = str(repo)
        run_cmd(("sh", str(repo / "install.sh")), env=env, timeout=300, log_path=log_path, display="sh /repo/install.sh")
    else:
        env.pop("NEWSJACK_SOURCE_DIR", None)
        run_shell(
            f"curl -fsSL {sh_quote(args.installer_url)} | sh",
            env=env,
            timeout=300,
            log_path=log_path,
        )


def assert_path(path: Path, description: str) -> None:
    if not path.exists():
        raise HarnessError(f"missing {description}: {path}")


def assert_executable(path: Path, description: str) -> None:
    assert_path(path, description)
    if not os.access(path, os.X_OK):
        raise HarnessError(f"{description} is not executable: {path}")


def assert_installed(runtime: str, run_dir: Path, env: dict[str, str]) -> None:
    home = Path(env["HOME"])
    spec = RUNTIME_SPECS[runtime]
    assert_path(home / ".newsjack" / "newsjack", "installed Newsjack repo")
    assert_executable(home / ".newsjack" / "bin" / "newsjack", "Newsjack CLI")
    skills_root = home / spec.skills_dir
    assert_path(skills_root, f"{runtime} skills directory")
    for skill in EXPECTED_SKILLS:
        assert_path(skills_root / skill / "SKILL.md", f"{runtime} skill {skill}")

    result = run_cmd(
        (str(home / ".newsjack" / "bin" / "newsjack"), "skills"),
        env=env,
        timeout=60,
        log_path=run_dir / "assertions.log",
    )
    listed = set(result.stdout.split())
    missing = [skill for skill in EXPECTED_SKILLS if skill not in listed]
    if missing:
        raise HarnessError(f"newsjack skills missing expected entries: {', '.join(missing)}")

    if runtime == "hermes":
        hermes_config = home / ".hermes" / "config.yaml"
        assert_path(hermes_config, "Hermes config")
        text = hermes_config.read_text(encoding="utf-8")
        if "mcp_servers:" not in text or "medialyst:" not in text:
            raise HarnessError("Hermes MCP config does not contain mcp_servers.medialyst")


def collect_versions(runtime: str, run_dir: Path, env: dict[str, str]) -> None:
    spec = RUNTIME_SPECS[runtime]
    payload: dict[str, dict[str, object]] = {}
    version_commands = [
        spec.version_cmd,
        ("node", "--version"),
        ("npm", "--version"),
        ("python3", "--version"),
    ]
    for cmd in version_commands:
        result = run_cmd(cmd, env=env, timeout=45, log_path=run_dir / "versions.log", check=False)
        payload[" ".join(cmd)] = {
            "returncode": result.returncode,
            "stdout": redact(result.stdout).strip(),
            "stderr": redact(result.stderr).strip(),
        }
    (run_dir / "versions.json").write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def run_direct_detector(repo: Path, run_dir: Path, env: dict[str, str]) -> None:
    home = Path(env["HOME"])
    output = run_dir / "artifacts" / "detector-mock.json"
    output.parent.mkdir(parents=True, exist_ok=True)
    result = run_cmd(
        (
            str(home / ".newsjack" / "bin" / "newsjack"),
            "detector",
            "run",
            "specialty coffee",
            "--profile",
            str(repo / "fixtures" / "newsjack-detector-agent" / "profile.bluebottle.json"),
            "--mock",
            "--emit",
            "json",
        ),
        cwd=repo,
        env=env,
        timeout=180,
        log_path=run_dir / "detector.log",
    )
    output.write_text(result.stdout, encoding="utf-8")
    try:
        parsed = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise HarnessError(f"detector mock did not emit valid JSON: {exc}") from exc
    if not isinstance(parsed, dict):
        raise HarnessError("detector mock JSON was not an object")


def maybe_medialyst_login(run_dir: Path, env: dict[str, str]) -> None:
    api_key = env.get("MEDIALYST_API_KEY")
    if not api_key:
        return
    home = Path(env["HOME"])
    run_cmd(
        (str(home / ".newsjack" / "bin" / "newsjack"), "login", "--key", api_key),
        env=env,
        timeout=60,
        log_path=run_dir / "medialyst-auth.log",
        display="~/.newsjack/bin/newsjack login --key <redacted>",
    )
    run_cmd(
        (str(home / ".newsjack" / "bin" / "newsjack"), "auth", "status"),
        env=env,
        timeout=60,
        log_path=run_dir / "medialyst-auth.log",
    )


def maybe_runtime_login(runtime: str, run_dir: Path, env: dict[str, str]) -> None:
    if runtime != "codex":
        return
    api_key = env.get("OPENAI_API_KEY")
    if not api_key:
        raise HarnessError("codex integration modes require OPENAI_API_KEY")
    run_shell(
        "printf %s \"$OPENAI_API_KEY\" | codex login --with-api-key",
        env=env,
        timeout=60,
        log_path=run_dir / "runtime-auth.log",
        display="printf <redacted> | codex login --with-api-key",
    )
    run_cmd(
        ("codex", "login", "status"),
        env=env,
        timeout=60,
        log_path=run_dir / "runtime-auth.log",
        check=False,
    )


def run_native_smoke(runtime: str, run_dir: Path, env: dict[str, str]) -> None:
    spec = RUNTIME_SPECS[runtime]
    result = run_cmd(
        spec.native_smoke_cmd,
        env=env,
        cwd=run_dir / "artifacts",
        timeout=240,
        log_path=run_dir / "native-smoke.log",
        check=True,
    )
    (run_dir / "final-response.md").write_text(result.stdout or result.stderr, encoding="utf-8")


def run_acp_probe(runtime: str, run_dir: Path, env: dict[str, str]) -> tuple[str, int]:
    prompt = "Reply exactly READY and stop."
    result = run_acp_prompt(
        runtime,
        prompt,
        run_dir,
        env,
        timeout=240,
        max_turns=1,
        log_name="acp-smoke.log",
    )
    (run_dir / "final-response.md").write_text(result.stdout or result.stderr, encoding="utf-8")
    return "acpx", result.returncode


def acpx_command(runtime: str, prompt: str, cwd: Path, timeout: int, max_turns: int) -> tuple[str, ...]:
    base = (
        "acpx",
        "--cwd",
        str(cwd),
        "--auth-policy",
        "fail",
        "--approve-all",
        "--non-interactive-permissions",
        "fail",
        "--timeout",
        str(timeout),
        "--max-turns",
        str(max_turns),
        "--format",
        "text",
    )
    if runtime == "codex":
        return (*base, "--agent", "codex-acp", "exec", prompt)
    if runtime in ("claude", "openclaw"):
        return (*base, runtime, "exec", prompt)
    if runtime == "hermes":
        return (*base, "--agent", "hermes acp --accept-hooks", "exec", prompt)
    raise HarnessError(f"unsupported ACP runtime: {runtime}")


def run_acp_prompt(
    runtime: str,
    prompt: str,
    run_dir: Path,
    env: dict[str, str],
    *,
    timeout: int,
    max_turns: int,
    log_name: str,
    check: bool = True,
) -> subprocess.CompletedProcess[str]:
    if not shutil.which("acpx", path=env.get("PATH")):
        raise HarnessError("acpx is not installed in the harness container")
    env = env.copy()
    if runtime == "codex" and env.get("OPENAI_API_KEY"):
        env["ACPX_AUTH_OPENAI_API_KEY"] = env["OPENAI_API_KEY"]
    elif runtime == "claude" and env.get("ANTHROPIC_API_KEY"):
        env["ACPX_AUTH_API_KEY"] = env["ANTHROPIC_API_KEY"]
        env["ACPX_AUTH_ANTHROPIC_API_KEY"] = env["ANTHROPIC_API_KEY"]
    elif runtime == "hermes":
        openrouter_key = env.get("OPENROUTER_API_KEY") or env.get("ACPX_AUTH_OPENROUTER")
        if not openrouter_key:
            raise HarnessError(
                "Hermes ACP requires OPENROUTER_API_KEY or ACPX_AUTH_OPENROUTER; "
                "use native-smoke with ANTHROPIC_API_KEY for the current Hermes path"
            )
        env["ACPX_AUTH_OPENROUTER"] = openrouter_key
    artifact_dir = run_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=True)
    cmd = acpx_command(runtime, prompt, artifact_dir, timeout, max_turns)
    (run_dir / "selected-acp.json").write_text(
        json.dumps({"controller": "acpx", "command": command_text(cmd[: min(len(cmd), 4)])}, indent=2) + "\n",
        encoding="utf-8",
    )
    gateway = None
    if runtime == "openclaw":
        gateway = start_openclaw_gateway(run_dir, env)
    try:
        return run_cmd(
            cmd,
            env=env,
            cwd=artifact_dir,
            timeout=timeout + 30,
            log_path=run_dir / log_name,
            check=check,
        )
    finally:
        if gateway:
            gateway.terminate()
            try:
                gateway.wait(timeout=10)
            except subprocess.TimeoutExpired:
                gateway.kill()
                gateway.wait(timeout=10)


def start_openclaw_gateway(run_dir: Path, env: dict[str, str]) -> subprocess.Popen[str]:
    log_path = run_dir / "openclaw-gateway.log"
    handle = log_path.open("a", encoding="utf-8")
    cmd = (
        "openclaw",
        "gateway",
        "run",
        "--dev",
        "--auth",
        "none",
        "--bind",
        "loopback",
        "--port",
        "18789",
        "--allow-unconfigured",
        "--compact",
    )
    handle.write(f"\n$ {command_text(cmd)}\n")
    handle.flush()
    process = subprocess.Popen(
        cmd,
        env=env,
        text=True,
        stdout=handle,
        stderr=subprocess.STDOUT,
    )
    handle.close()
    for _ in range(45):
        if process.poll() is not None:
            raise HarnessError(f"OpenClaw gateway exited before becoming ready; see {log_path}")
        try:
            gateway_log = log_path.read_text(encoding="utf-8", errors="replace")
        except OSError:
            gateway_log = ""
        if "[gateway] ready" in gateway_log:
            return process
        time_sleep(1)
    process.terminate()
    try:
        process.wait(timeout=10)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=10)
    handle.close()
    raise HarnessError(f"OpenClaw gateway did not become ready; see {log_path}")


def time_sleep(seconds: float) -> None:
    import time

    time.sleep(seconds)


def run_setup_flow(runtime: str, repo: Path, run_dir: Path, env: dict[str, str]) -> None:
    prompt = (repo / "harness" / "prompts" / "setup-flow.md").read_text(encoding="utf-8")
    artifact_dir = run_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=True)
    prompt = prompt.replace("ARTIFACT_DIR", str(artifact_dir))

    if runtime == "hermes" and not (env.get("OPENROUTER_API_KEY") or env.get("ACPX_AUTH_OPENROUTER")):
        result = run_cmd(
            (
                "hermes",
                "chat",
                "-Q",
                "--ignore-rules",
                "--provider",
                "anthropic",
                "--max-turns",
                "12",
                "--yolo",
                "-q",
                prompt,
            ),
            env=env,
            cwd=artifact_dir,
            timeout=900,
            log_path=run_dir / "setup-flow.log",
            check=False,
        )
    else:
        result = run_acp_prompt(
            runtime,
            prompt,
            run_dir,
            env,
            timeout=900,
            max_turns=15,
            log_name="setup-flow.log",
            check=False,
        )
    (run_dir / "final-response.md").write_text(result.stdout or result.stderr, encoding="utf-8")

    profile = artifact_dir / "profile.json"
    result_md = artifact_dir / "result.md"
    assert_path(profile, "setup-flow profile.json")
    assert_path(result_md, "setup-flow result.md")
    try:
        profile_data = json.loads(profile.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise HarnessError(f"profile.json is invalid JSON: {exc}") from exc
    if not (profile_data.get("company") or profile_data.get("name")):
        raise HarnessError("profile.json missing required field: company")
    if not profile_data.get("description"):
        raise HarnessError("profile.json missing required field: description")
    if result.returncode != 0:
        (run_dir / "setup-flow.warning.txt").write_text(
            "ACP controller exited nonzero after artifacts were produced and validated. "
            f"Exit code: {result.returncode}\n",
            encoding="utf-8",
        )


def inside_container_main(args: argparse.Namespace) -> int:
    if args.runtime == "all":
        raise HarnessError("--inside-container requires one concrete --runtime")
    runtime = args.runtime
    spec = RUNTIME_SPECS[runtime]
    assert_no_model_keys(args.mode)
    assert_model_allowed(args.mode)

    repo = Path(args.repo)
    run_id = args.run_id or os.environ.get("NEWSJACK_HARNESS_RUN_ID") or datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    run_dir = Path(args.runs) / run_id / runtime
    run_dir.mkdir(parents=True, exist_ok=True)
    (run_dir / "artifacts").mkdir(parents=True, exist_ok=True)

    env = base_container_env(runtime, run_dir, repo)
    home = Path(env["HOME"])
    if str(home).startswith("/home/newsjack-harness/") and home.exists():
        shutil.rmtree(home)
    if not args.local_source:
        env.pop("NEWSJACK_SOURCE_DIR", None)
    write_env_json(runtime, args.mode, run_dir, env)

    if not shutil.which(spec.binary, path=env.get("PATH")):
        raise HarnessError(f"runtime binary not found in container: {spec.binary}")

    collect_versions(runtime, run_dir, env)
    install_newsjack(args, runtime, repo, env, run_dir / "installer.log")
    assert_installed(runtime, run_dir, env)
    run_direct_detector(repo, run_dir, env)

    if args.mode in ("native-smoke", "acp-smoke", "setup-flow"):
        maybe_medialyst_login(run_dir, env)
        maybe_runtime_login(runtime, run_dir, env)

    if args.mode == "native-smoke":
        run_native_smoke(runtime, run_dir, env)
    elif args.mode == "acp-smoke":
        run_acp_probe(runtime, run_dir, env)
    elif args.mode == "setup-flow":
        run_setup_flow(runtime, repo, run_dir, env)

    (run_dir / "PASS").write_text("ok\n", encoding="utf-8")
    log(f"{runtime} {args.mode} passed; artifacts: {run_dir}")
    return 0


def main(argv: Sequence[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    try:
        if args.inside_container:
            return inside_container_main(args)
        return host_main(args)
    except subprocess.TimeoutExpired as exc:
        print(f"harness: timeout: {exc}", file=sys.stderr)
        return 124
    except HarnessError as exc:
        print(f"harness: error: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
