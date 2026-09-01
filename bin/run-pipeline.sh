#!/usr/bin/env sh
# Runs the full monitoring pipeline once: ingest both sources, persist,
# detect relevance, and generate the report. This is the same sequence
# run manually across Sprints 2-4; daily-monitor.yml just calls this
# script on a schedule.
#
# RUN_WINDOW_START/END are optional: if unset, they're computed
# automatically (today - 2 days .. today - 1 day, in America/Sao_Paulo) via
# `monitor run-window`. Set them explicitly to reprocess a specific window.
set -eu

main() {
  if [ -z "${RUN_WINDOW_START:-}" ] || [ -z "${RUN_WINDOW_END:-}" ]; then
    eval "$(docker compose run --rm orchestrator run-window)"
    export RUN_WINDOW_START RUN_WINDOW_END
  fi
  echo "Run window: ${RUN_WINDOW_START} to ${RUN_WINDOW_END}"

  docker compose run --rm orchestrator migrate

  docker compose run --rm scripts ingest-camara --output /workdir/camara.json
  docker compose run --rm orchestrator load-bills --input /workdir/camara.json

  docker compose run --rm scripts ingest-senado --output /workdir/senado.json
  docker compose run --rm orchestrator load-bills --input /workdir/senado.json

  docker compose run --rm scripts detect-relevance --input /workdir/camara.json --output /workdir/camara_relevance.json
  docker compose run --rm scripts detect-relevance --input /workdir/senado.json --output /workdir/senado_relevance.json

  docker compose run --rm orchestrator generate-report \
    --input /workdir/camara_relevance.json \
    --input /workdir/senado_relevance.json
}

# Persist a plain-text record of every run outside any container's
# filesystem, so a rebuild/redeploy never wipes it. The app itself keeps
# writing only to stdout/stderr (Factor XI) - this script is the external
# agent doing the capture. `dash` (the /bin/sh on GitHub Actions runners)
# has no `pipefail` and no `>()` process substitution, so piping straight
# into `tee` would silently replace the pipeline's real exit code with
# tee's (always 0) and mask failures. The status-file dance below keeps
# the real exit code across the pipe.
mkdir -p logs
status_file=$(mktemp)
{
  echo "=== run started at $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="
  if main; then echo 0 >"$status_file"; else echo $? >"$status_file"; fi
  echo "=== run finished at $(date -u +%Y-%m-%dT%H:%M:%SZ) (exit $(cat "$status_file")) ==="
} 2>&1 | tee -a logs/pipeline.log
status=$(cat "$status_file")
rm -f "$status_file"
exit "$status"
