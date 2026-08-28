# Roadmap

- **Sprint 0 — Scaffolding, CI and Docker** *(current)*: directory skeleton,
  multi-stage Dockerfiles for `scripts`/`orchestrator`, `docker-compose.yml`
  bringing up all three containers, `ci.yml`, `.env.example`.
- **Sprint 1 — Database**: `0001_init` migration, `internal/db` repositories,
  Postgres wired into `docker-compose.yml` and CI.
- **Sprint 2 — Camara ingestion**: `camara_client.py`, `ingest-camara`
  subcommand, `Bill` model, recorded fixtures, idempotent persistence.
- **Sprint 3 — Senado ingestion + unified schema**: same for Senado, plus
  merge/dedupe across sources in the orchestrator.
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
