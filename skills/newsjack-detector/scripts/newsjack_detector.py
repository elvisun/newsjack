#!/usr/bin/env python3
"""v0 newsjack monitoring engine.

This command ingests public signals, clusters related evidence, and computes
mechanical monitor scores. It does not decide whether to pitch. The skill owns
that PR judgment.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import re
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

SCRIPT_DIR = Path(__file__).parent.resolve()
sys.path.insert(0, str(SCRIPT_DIR))

import brand_safety_flags
import newsjack_store
import source_registry
from lib import dates
from monitor_profile import MonitorProfile


STOP_WORDS = {
    "about", "after", "again", "against", "also", "and", "are", "because",
    "been", "being", "but", "can", "could", "did", "does", "doing", "for",
    "from", "had", "has", "have", "her", "here", "him", "his", "how",
    "into", "its", "just", "more", "new", "not", "now", "our", "out",
    "over", "per", "said", "say", "says", "she", "should", "than", "that",
    "the", "their", "them", "then", "there", "these", "they", "this",
    "those", "through", "too", "under", "was", "what", "when", "where",
    "which", "while", "who", "why", "will", "with", "would", "you", "your",
}

SOURCE_QUALITY = {
    "news_search": 0.95,
    "x": 0.70,
    "reddit": 0.62,
    "hackernews": 0.72,
}

ENGAGEMENT_FIELDS = (
    "score", "num_comments", "comments", "likes", "reposts", "replies",
    "quotes", "views", "points",
)


@dataclass
class EvidenceItem:
    source: str
    title: str
    url: str
    excerpt: str = ""
    author: str | None = None
    container: str | None = None
    published_at: str | None = None
    engagement: dict[str, Any] = field(default_factory=dict)
    metadata: dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "EvidenceItem":
        return cls(
            source=str(payload.get("source") or "unknown"),
            title=str(payload.get("title") or payload.get("excerpt") or "").strip(),
            url=str(payload.get("url") or "").strip(),
            excerpt=str(payload.get("excerpt") or "").strip(),
            author=payload.get("author"),
            container=payload.get("container"),
            published_at=payload.get("published_at"),
            engagement=dict(payload.get("engagement") or {}),
            metadata=dict(payload.get("metadata") or {}),
        )

    def text(self) -> str:
        return " ".join(part for part in [self.title, self.excerpt, self.author or "", self.container or ""] if part)

    def public_dict(self) -> dict[str, Any]:
        return {
            "source": self.source,
            "title": self.title,
            "url": self.url,
            "author": self.author,
            "container": self.container,
            "published_at": self.published_at,
            "excerpt": self.excerpt[:500],
            "engagement": self.engagement,
        }


@dataclass
class SignalCluster:
    evidence: list[EvidenceItem] = field(default_factory=list)

    def add(self, item: EvidenceItem) -> None:
        self.evidence.append(item)

    @property
    def title(self) -> str:
        for item in self.evidence:
            if item.source == "news_search" and item.title:
                return item.title
        return self.evidence[0].title if self.evidence else ""

    def text(self) -> str:
        return " ".join(item.text() for item in self.evidence)

    def sources(self) -> list[str]:
        return sorted({item.source for item in self.evidence})

    def urls(self) -> list[str]:
        return [item.url for item in self.evidence if item.url]


def _tokens(text: str) -> set[str]:
    return {
        token
        for token in re.findall(r"[a-z0-9][a-z0-9+._-]{2,}", text.lower())
        if token not in STOP_WORDS
    }


def _jaccard(left: str, right: str) -> float:
    left_tokens = _tokens(left)
    right_tokens = _tokens(right)
    if not left_tokens or not right_tokens:
        return 0.0
    return len(left_tokens & right_tokens) / len(left_tokens | right_tokens)


def _profile_matches(profile: MonitorProfile, text: str) -> list[str]:
    lower = text.lower()
    terms = [
        *(profile.topics or []),
        *(profile.competitors or []),
        *(profile.standing or []),
        *(profile.proof_assets or []),
    ]
    matches = []
    for term in terms:
        term = term.strip()
        if term and term.lower() in lower and term not in matches:
            matches.append(term)
    return matches[:12]


def _profile_match_score(profile: MonitorProfile, text: str) -> float:
    context = profile.match_text()
    if not context:
        return 0.4
    overlap = _jaccard(context, text)
    phrase_bonus = min(0.4, 0.08 * len(_profile_matches(profile, text)))
    return min(1.0, overlap + phrase_bonus)


def _parse_time(value: str | None) -> datetime | None:
    if not value:
        return None
    raw = value.strip()
    if not raw:
        return None
    if raw.endswith("Z"):
        raw = f"{raw[:-1]}+00:00"
    try:
        parsed = datetime.fromisoformat(raw)
    except ValueError:
        try:
            parsed = datetime.fromisoformat(f"{raw[:10]}T00:00:00+00:00")
        except ValueError:
            return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def _min_age_hours(cluster: SignalCluster, now: datetime) -> float | None:
    ages = []
    for item in cluster.evidence:
        parsed = _parse_time(item.published_at)
        if parsed:
            ages.append(max(0.0, (now - parsed).total_seconds() / 3600.0))
    return min(ages) if ages else None


def _freshness_score(age_hours: float | None, lookback_days: int) -> float:
    if age_hours is None:
        return 0.35
    if age_hours <= 4:
        return 1.0
    if age_hours <= 24:
        return 0.86
    window_hours = max(24, lookback_days * 24)
    return max(0.05, 1.0 - (age_hours / window_hours))


def _decay_bucket(age_hours: float | None) -> str:
    if age_hours is None:
        return "unknown"
    if age_hours <= 1:
        return "30min"
    if age_hours <= 4:
        return "4hr"
    if age_hours <= 24:
        return "24hr"
    if age_hours <= 168:
        return "week"
    return "month"


def _source_agreement_score(sources: list[str]) -> float:
    if len(sources) >= 3:
        return 1.0
    if len(sources) == 2:
        return 0.78
    if sources and sources[0] == "news_search":
        return 0.55
    return 0.32


def _source_quality_score(cluster: SignalCluster) -> float:
    if not cluster.evidence:
        return 0.0
    return sum(SOURCE_QUALITY.get(item.source, 0.5) for item in cluster.evidence) / len(cluster.evidence)


def _engagement_score(cluster: SignalCluster) -> float:
    total = 0.0
    for item in cluster.evidence:
        for field in ENGAGEMENT_FIELDS:
            try:
                value = float(item.engagement.get(field) or 0)
            except (TypeError, ValueError):
                value = 0.0
            if value > 0:
                total += math.log1p(value)
    return min(1.0, total / 24.0)


def _novelty_score(urls: list[str], seen: dict[str, dict[str, Any]]) -> float:
    if not urls:
        return 0.50
    unseen = [url for url in urls if url not in seen]
    if not unseen:
        return 0.10
    return len(unseen) / len(urls)


def _rank_signal(
    cluster: SignalCluster,
    *,
    profile: MonitorProfile,
    seen: dict[str, dict[str, Any]],
    now: datetime,
    lookback_days: int,
) -> dict[str, Any]:
    text = cluster.text()
    sources = cluster.sources()
    urls = cluster.urls()
    age_hours = _min_age_hours(cluster, now)
    freshness = _freshness_score(age_hours, lookback_days)
    novelty = _novelty_score(urls, seen)
    source_agreement = _source_agreement_score(sources)
    source_quality = _source_quality_score(cluster)
    engagement = _engagement_score(cluster)
    profile_match = _profile_match_score(profile, text)
    rank = round(
        100
        * (
            0.24 * freshness
            + 0.20 * source_agreement
            + 0.18 * novelty
            + 0.16 * profile_match
            + 0.12 * source_quality
            + 0.10 * engagement
        ),
        1,
    )
    evidence = [item.public_dict() for item in cluster.evidence[:8]]
    signal_id = _signal_id(cluster.title, urls, text)
    return {
        "id": signal_id,
        "title": cluster.title,
        "sources": sources,
        "evidence": evidence,
        "features": {
            "age_hours": None if age_hours is None else round(age_hours, 2),
            "decay_bucket": _decay_bucket(age_hours),
            "source_count": len(sources),
            "evidence_count": len(cluster.evidence),
            "seen_before": bool(urls and all(url in seen for url in urls)),
            "seen_urls": {url: seen[url] for url in urls if url in seen},
            "profile_matches": _profile_matches(profile, text),
            "safety_flags": brand_safety_flags.flag_text(text, exclusions=profile.exclusions),
        },
        "scores": {
            "rank": rank,
            "freshness": round(freshness, 3),
            "source_agreement": round(source_agreement, 3),
            "novelty": round(novelty, 3),
            "profile_match": round(profile_match, 3),
            "source_quality": round(source_quality, 3),
            "momentum": round(engagement, 3),
        },
    }


def _signal_id(title: str, urls: list[str], text: str) -> str:
    basis = "|".join([title, *urls[:5]]) or text[:400]
    return hashlib.sha1(basis.encode("utf-8")).hexdigest()[:16]


def _cluster_items(items: list[EvidenceItem]) -> list[SignalCluster]:
    clusters: list[SignalCluster] = []
    sorted_items = sorted(
        items,
        key=lambda item: (
            0 if item.source == "news_search" else 1,
            item.published_at or "",
        ),
    )
    for item in sorted_items:
        if not item.title and not item.excerpt:
            continue
        placed = False
        for cluster in clusters:
            if item.url and item.url in cluster.urls():
                cluster.add(item)
                placed = True
                break
            if _jaccard(item.text(), cluster.text()) >= 0.32:
                cluster.add(item)
                placed = True
                break
        if not placed:
            clusters.append(SignalCluster([item]))
    return clusters


def _build_queries(args: argparse.Namespace, profile: MonitorProfile) -> list[str]:
    queries: list[str] = []
    queries.extend(args.topic or [])
    if args.query:
        queries.append(" ".join(args.query).strip())
    queries.extend(profile.query_terms())
    return _dedupe([query for query in queries if query])


def _dedupe(values: list[str]) -> list[str]:
    seen = set()
    output = []
    for value in values:
        key = value.lower().strip()
        if not key or key in seen:
            continue
        seen.add(key)
        output.append(value.strip())
    return output


def _load_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip()
        if not key:
            continue
        if value and value[0] in {"'", '"'} and value[-1:] == value[0]:
            value = value[1:-1]
        values[key] = value
    return values


def _env_file_values() -> dict[str, str]:
    repo_root = SCRIPT_DIR.parents[2]
    paths = [
        repo_root / ".env",
        Path.cwd() / ".env",
    ]
    merged: dict[str, str] = {}
    for path in paths:
        merged.update(_load_env_file(path))
    return merged


def _config_from_env() -> dict[str, str | None]:
    file_env = _env_file_values()

    def get(key: str) -> str | None:
        return os.environ.get(key) or file_env.get(key)

    return {
        "MEDIALYST_API_KEY": get("MEDIALYST_API_KEY"),
        "MEDIALYST_API_BASE": get("MEDIALYST_API_BASE"),
        "MEDIALYST_NEWS_PATH": get("MEDIALYST_NEWS_PATH"),
    }


def _mock_items(query: str, now: datetime) -> list[EvidenceItem]:
    today = now.date().isoformat()
    return [
        EvidenceItem(
            source="news_search",
            title=f"Regulators open inquiry tied to {query}",
            url=f"https://example.com/news/{hashlib.sha1(query.encode()).hexdigest()[:8]}",
            container="Example News",
            published_at=today,
            excerpt=f"Officials are examining claims and compliance practices around {query}.",
        ),
        EvidenceItem(
            source="x",
            title=f"Experts are reacting to {query}",
            url=f"https://x.com/example/status/{hashlib.sha1((query + 'x').encode()).hexdigest()[:8]}",
            author="example",
            container="x.com",
            published_at=today,
            excerpt=f"Thread: the {query} inquiry is moving faster than vendors expected.",
            engagement={"likes": 120, "reposts": 22, "replies": 9},
        ),
    ]


def collect_query(
    query: str,
    *,
    sources: list[str],
    config: dict[str, Any],
    depth: str,
    lookback_days: int,
    mock: bool,
    now: datetime,
) -> tuple[list[EvidenceItem], dict[str, str]]:
    if mock:
        return _mock_items(query, now), {}
    from_date, to_date = dates.get_date_range(lookback_days)
    errors: dict[str, str] = {}
    items: list[EvidenceItem] = []
    with ThreadPoolExecutor(max_workers=min(4, max(1, len(sources)))) as executor:
        futures = {
            executor.submit(
                source_registry.collect_source,
                source,
                query,
                from_date=from_date,
                to_date=to_date,
                depth=depth,
                config=config,
            ): source
            for source in sources
        }
        for future in as_completed(futures):
            source = futures[future]
            raw_items, error = future.result()
            if error:
                errors[source] = error
            for raw in raw_items:
                item = EvidenceItem.from_dict(raw)
                if item.source != "news_search" and _jaccard(query, item.text()) < 0.08:
                    continue
                if item.title or item.excerpt:
                    items.append(item)
    return items, errors


def run(args: argparse.Namespace) -> int:
    profile = MonitorProfile.from_file(args.profile) if args.profile else MonitorProfile()
    queries = _build_queries(args, profile)
    if not queries:
        raise SystemExit("Provide a query, --topic, or --profile with topics/competitors.")

    config = _config_from_env()
    requested_sources = source_registry.parse_sources(args.sources)
    sources = requested_sources if args.mock else source_registry.available_sources(config, requested_sources)
    if not sources:
        raise SystemExit(
            "No requested sources are available. Configure MEDIALYST_API_KEY and xurl auth, "
            "or rerun with --mock."
        )

    now = datetime.now(timezone.utc)
    db_path = Path(args.store).expanduser() if args.store else None
    all_signals: list[dict[str, Any]] = []
    source_errors: dict[str, dict[str, str]] = {}

    for query in queries:
        items, errors = collect_query(
            query,
            sources=sources,
            config=config,
            depth=args.depth,
            lookback_days=args.lookback_days,
            mock=args.mock,
            now=now,
        )
        if errors:
            source_errors[query] = errors
        clusters = _cluster_items(items)
        urls = [url for cluster in clusters for url in cluster.urls()]
        seen = newsjack_store.seen_status(urls, db_path=db_path)
        for cluster in clusters:
            signal = _rank_signal(
                cluster,
                profile=profile,
                seen=seen,
                now=now,
                lookback_days=args.lookback_days,
            )
            signal["query"] = query
            all_signals.append(signal)

    all_signals.sort(key=lambda signal: signal["scores"]["rank"], reverse=True)
    signals = all_signals[: args.limit]
    run_id = None
    if args.save:
        run_id = newsjack_store.record_run(
            monitor_name=args.monitor_name,
            profile=profile.public_dict(),
            queries=queries,
            signals=signals,
            db_path=db_path,
        )

    payload = {
        "monitor": {
            "name": args.monitor_name,
            "generated_at": now.isoformat(),
            "profile": profile.public_dict(),
            "queries": queries,
            "sources_requested": requested_sources,
            "sources_used": sources,
            "lookback_days": args.lookback_days,
            "depth": args.depth,
            "mock": args.mock,
        },
        "signals": signals,
        "source_errors": source_errors,
        "store": {
            "saved": args.save,
            "run_id": run_id,
            "path": str(db_path or newsjack_store.db_path_from_env()) if args.save else None,
        },
    }
    if args.emit == "brief":
        print(_brief(payload))
    else:
        print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def diagnose(args: argparse.Namespace) -> int:
    config = _config_from_env()
    requested = source_registry.parse_sources(args.sources)
    payload = {
        "sources_requested": requested,
        "sources_available": source_registry.available_sources(config, requested),
        "news_search_configured": bool(config.get("MEDIALYST_API_KEY")),
        "xurl_available": "x" in source_registry.available_sources(config, ["x"]),
        "store_path": str(Path(args.store).expanduser()) if args.store else str(newsjack_store.db_path_from_env()),
    }
    print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def recent(args: argparse.Namespace) -> int:
    db_path = Path(args.store).expanduser() if args.store else None
    print(json.dumps(newsjack_store.recent_runs(args.limit, db_path=db_path), indent=2))
    return 0


def _brief(payload: dict[str, Any]) -> str:
    lines = ["newsjack monitor", ""]
    for index, signal in enumerate(payload.get("signals") or [], start=1):
        features = signal["features"]
        lines.append(
            f"{index}. {signal['title']} "
            f"({signal['scores']['rank']}, {features['decay_bucket']}, {', '.join(signal['sources'])})"
        )
        if signal["evidence"]:
            lines.append(f"   {signal['evidence'][0].get('url')}")
        if features.get("safety_flags"):
            lines.append(f"   safety_flags={len(features['safety_flags'])}")
    if not payload.get("signals"):
        lines.append("No signals returned.")
    return "\n".join(lines)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Collect and rank newsjack monitoring evidence.")
    sub = parser.add_subparsers(dest="command")

    run_parser = sub.add_parser("run", help="Run a monitor pass.")
    run_parser.add_argument("query", nargs="*", help="Ad hoc monitor query.")
    run_parser.add_argument("--topic", action="append", help="Topic to monitor. Repeatable.")
    run_parser.add_argument("--profile", help="Path to monitor profile JSON.")
    run_parser.add_argument("--sources", help="Comma-separated sources. Default: news_search,x.")
    run_parser.add_argument("--depth", choices=["quick", "default", "deep"], default="quick")
    run_parser.add_argument("--lookback-days", type=int, default=7)
    run_parser.add_argument("--limit", type=int, default=20)
    run_parser.add_argument("--mock", action="store_true")
    run_parser.add_argument("--save", action="store_true")
    run_parser.add_argument("--store", help="Override SQLite store path.")
    run_parser.add_argument("--monitor-name")
    run_parser.add_argument("--emit", choices=["json", "brief"], default="json")
    run_parser.set_defaults(func=run)

    diag_parser = sub.add_parser("diagnose", help="Show source availability.")
    diag_parser.add_argument("--sources")
    diag_parser.add_argument("--store")
    diag_parser.set_defaults(func=diagnose)

    recent_parser = sub.add_parser("recent", help="List recent stored monitor runs.")
    recent_parser.add_argument("--limit", type=int, default=10)
    recent_parser.add_argument("--store")
    recent_parser.set_defaults(func=recent)

    return parser


def main() -> int:
    parser = build_parser()
    commands = {"run", "diagnose", "recent"}
    if len(sys.argv) > 1 and sys.argv[1] not in commands and sys.argv[1] not in {"-h", "--help"}:
        argv = ["run", *sys.argv[1:]]
    else:
        argv = sys.argv[1:]
    args = parser.parse_args(argv)
    if not args.command:
        parser.print_help()
        return 1
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
