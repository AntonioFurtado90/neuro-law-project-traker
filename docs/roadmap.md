# Roadmap

- **Sprint 0 — Scaffolding, CI and Docker** *(done)*: directory skeleton,
  multi-stage Dockerfiles for `scripts`/`orchestrator`, `docker-compose.yml`
  bringing up all three containers, `ci.yml`, `.env.example`.
- **Sprint 1 — Database** *(done)*: `0001_init` migration (`bills`,
  `ingestion_runs`, `relevance_results`, `reports`), a hand-rolled migration
  runner (`monitor migrate`), `BillsRepo` with idempotent upserts, Postgres
  wired into `docker-compose.yml` and CI (`orchestrator-db-tests`).
- **Sprint 2 — Camara ingestion** *(done)*: `camara_client.py` (stdlib
  `urllib`, no new runtime dependency), `Bill` model, `ingest-camara`
  subcommand, recorded fixtures, and a `load-bills` orchestrator subcommand
  that persists an `ingestion_result.json` via `BillsRepo` — the manual
  stand-in for the automated pipeline `containerexec`/`pipeline.go` will
  build in Sprint 4.
- **Sprint 3 — Senado ingestion** *(done)*: `senado_client.py` against the
  `/processo` endpoint (the older `/materia/pesquisa/lista` is officially
  deprecated) — filters reliably by year/type only, so date-window filtering
  happens client-side; `Bill.from_senado_item`, `ingest-senado` subcommand.
  No orchestrator changes needed: `bills`/`bill.schema.json` already model
  `source` as `camara`/`senado` since Sprint 0/1, so `load-bills` persists
  either source unchanged, with no merge/dedupe step required.
- **Sprint 4 — Keyword relevance detection + report**: curated FCO/FNE/FNO
  keyword list, `detect-relevance` subcommand, full pipeline wiring through
  containers, `FileSink`.
- **Sprint 5 — Daily cron on GitHub Actions**: `daily-monitor.yml`, run-window
  computation, managed Postgres configured, artifact upload,
  `workflow_dispatch` for one-off runs.
- **Sprint 6+ (future) — Notification**: a real `Sink` implementation
  (email/Telegram/Slack) once the channel is decided.
- **Sprint 7+ (future, explicitly separate) — ML (Phase B)**: decide the
  labeling strategy, then feature engineering and model comparison (decision
  tree, k-NN, logistic regression), behind the same `detect-relevance`
  contract so the orchestrator never changes.

Full architectural context and rationale: see the project plan referenced in
the repository's change history.
