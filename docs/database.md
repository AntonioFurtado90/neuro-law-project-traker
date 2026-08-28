# Database

Postgres, owned exclusively by `orchestrator` (scripts never connect to it
directly). Schema is managed by plain SQL migrations in
`orchestrator/migrations/` (introduced in Sprint 1), applied as a one-off
administrative process (Factor XII) — never automatically at container
startup.

Planned tables (see the project plan for full column lists):

- **`bills`** — canonical bill records, `UNIQUE(source, external_id)` for
  idempotent re-ingestion.
- **`ingestion_runs`** — one row per ingestion attempt, for debugging API
  failures.
- **`relevance_results`** — relevance verdicts per bill; `method` is
  `keyword` today, `ml` once Phase B ships, without a schema change.
- **`reports`** — history of what was actually reported/notified.
- **`labels`** (reserved) — supervised-learning labels for the future ML
  sprint; not created until the labeling strategy is decided.

In production, Postgres is a managed instance (Neon/Supabase/Railway free
tier) reachable via `DATABASE_URL`, since GitHub Actions runners are
ephemeral and cannot host a persistent database container between scheduled
runs. In development, `docker-compose.yml` runs a local `postgres:16-alpine`
container instead — same schema, same migrations, same `DATABASE_URL`
mechanism (Factor X: dev/prod parity).
