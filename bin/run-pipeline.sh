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
