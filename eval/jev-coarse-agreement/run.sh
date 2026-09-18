#!/usr/bin/env bash
# Run the Jev coarse filter over every coarse_chunk.N.json in the given
# detector run folders and stage the worker decisions beside the output.
#
# usage: eval/jev-coarse-agreement/run.sh <run-dir> [<run-dir> ...]
# env:   TYPESAFE_API_KEY (or a saved key), NEWSJACK_BIN (default: bin/newsjack)
#        JEV_QUESTIONS (optional question-set override passed as --questions)
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "usage: $0 <run-dir> [<run-dir> ...]" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
newsjack_bin="${NEWSJACK_BIN:-$repo_root/bin/newsjack}"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
out_dir="$repo_root/eval/jev-coarse-agreement/runs/$stamp"
mkdir -p "$out_dir"

for run_dir in "$@"; do
  run_name="$(basename "$run_dir")"
  shopt -s nullglob
  chunks=("$run_dir"/coarse_chunk.*.json)
  shopt -u nullglob
  if [ "${#chunks[@]}" -eq 0 ]; then
    echo "skip $run_name: no coarse_chunk.N.json" >&2
    continue
  fi
  for chunk in "${chunks[@]}"; do
    n="$(basename "$chunk" .json)"; n="${n#coarse_chunk.}"
    worker="$run_dir/coarse_decisions.$n.json"
    if [ ! -f "$worker" ]; then
      echo "skip $run_name chunk $n: no worker decisions" >&2
      continue
    fi
    cp "$chunk" "$out_dir/chunk.$run_name.$n.json"
    cp "$worker" "$out_dir/worker.$run_name.$n.json"
    args=(coarse-filter --engine jev --candidates "$chunk" --output "$out_dir/jev.$run_name.$n.json")
    if [ -n "${JEV_QUESTIONS:-}" ]; then
      args+=(--questions "$JEV_QUESTIONS")
    fi
    echo "jev  $run_name chunk $n" >&2
    "$newsjack_bin" "${args[@]}" || echo "warn $run_name chunk $n: coarse-filter exited $?" >&2
  done
done

echo "$out_dir"
