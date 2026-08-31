#!/usr/bin/env sh
# Runs the full monitoring pipeline once: ingest both sources, persist,
# detect relevance, and generate the report. This is the same sequence
# run manually across Sprints 2-4; Sprint 5's daily-monitor.yml just calls
# this script on a schedule.
#
# Requires RUN_WINDOW_START and RUN_WINDOW_END to be set in the environment
# (or in .env, read by docker-compose).
set -eu

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
