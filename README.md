# Neuro Law Project Tracker

A bot that daily monitors bills (Projetos de Lei) presented in Brazil's
Câmara dos Deputados and Congresso Nacional (Senado Federal) that may impact
the **Fundos Constitucionais** — FCO, FNE, FNO, the constitutional regional
financing funds established under Art. 159 of the Brazilian Constitution.

## How it works (MVP)

1. **Ingestion** (Python) pulls bill data from the official open-data APIs of
   the Câmara dos Deputados and the Senado Federal.
2. **Relevance detection** (Python) flags bills whose ementa/text matches a
   curated set of terms related to the constitutional funds.
3. **Orchestration and persistence** (Go) sequences the pipeline, stores
   results in Postgres, and produces a daily report.

Each piece — the Python scripts, the Go orchestrator, and the Postgres
database — runs in its own container, so they never share a dependency tree.
See [`docs/architecture.md`](docs/architecture.md) for the full design,
[`docs/contract.md`](docs/contract.md) for how the containers talk to each
other, [`docs/database.md`](docs/database.md) for the schema, and
[`docs/roadmap.md`](docs/roadmap.md) for the sprint plan.

## Requirements

- Docker and Docker Compose (everything runs containerized, including local
  development, to keep dev/prod parity).

## Getting started

```bash
cp .env.example .env
docker compose build
docker compose run --rm scripts version
docker compose run --rm orchestrator version
docker compose up -d db
docker compose run --rm orchestrator migrate
```

## Running the full pipeline

```bash
./bin/run-pipeline.sh
```

With no `RUN_WINDOW_START`/`RUN_WINDOW_END` set, the window is computed
automatically (`monitor run-window`: the 2 days ending yesterday, in
`America/Sao_Paulo`) — this is what the daily cron runs unattended. Set
both explicitly to reprocess a specific window instead:

```bash
RUN_WINDOW_START=2026-08-01 RUN_WINDOW_END=2026-08-05 ./bin/run-pipeline.sh
```

Either way it ingests both sources, persists the bills, runs keyword-based
relevance detection, and writes a Markdown report to
`./output/report-<date>.md` — idempotent end to end (safe to re-run for the
same window; it updates instead of duplicating). Each step is also runnable
on its own:

```bash
docker compose run --rm scripts ingest-camara --output /workdir/camara.json
docker compose run --rm orchestrator load-bills --input /workdir/camara.json
docker compose run --rm scripts detect-relevance --input /workdir/camara.json --output /workdir/camara_relevance.json
docker compose run --rm orchestrator generate-report --input /workdir/camara_relevance.json
```

`ingest-*` and `detect-relevance` write JSON envelopes (see
[`docs/contract.md`](docs/contract.md)); `load-bills` and `generate-report`
are source-agnostic, so the same commands work for `ingest-senado`'s output
too. See [`docs/architecture.md`](docs/architecture.md) for why sequencing
lives in a shell script rather than inside the Go orchestrator.

## Setting up the daily cron

[`.github/workflows/daily-monitor.yml`](.github/workflows/daily-monitor.yml)
runs `bin/run-pipeline.sh` daily (and on-demand via `workflow_dispatch`,
with optional date overrides), uploading the report as a build artifact.
It expects a `DATABASE_URL` repository secret pointing at a managed Postgres
instance — GitHub Actions runners are ephemeral, so production Postgres
can't be the local `db` container. See
[`docs/database.md`](docs/database.md) → "Provisioning the production
database" for how to create one (Neon/Supabase/Railway all work) and add
the secret.

## Running tests

Tests run as part of the Docker build (the `test`/`build` image stages), the
same way CI runs them:

```bash
docker build --target test -f scripts/Dockerfile .
docker build --target build -f orchestrator/Dockerfile .
```

Database-backed tests need a running Postgres and run separately (also how
CI's `orchestrator-db-tests` job runs them):

```bash
docker compose up -d db
docker compose run --rm orchestrator migrate
docker compose run --rm orchestrator-integration-tests
```

## Engineering practices

This project follows the Twelve-Factor App methodology and a minimal
engineering framework: every change ships with tests, CI runs on every push,
`main` is always green, and releases are small and incremental. See
`CLAUDE.md` for the full set of guidelines this project adheres to.

## Português

Uma versão deste documento em português está disponível em
[`LEIAME.md`](LEIAME.md).
