# Database

Postgres, owned exclusively by `orchestrator` (scripts never connect to it
directly). Schema is managed by plain SQL migrations in
`orchestrator/migrations/`, applied via a small hand-rolled runner
(`orchestrator/internal/db.Migrate`, ~80 lines on top of `pgx`) rather than a
migration framework — the only requirement is applying `*.up.sql` files in
order, once, so a full library would be an extra dependency for no real
benefit. Applied filenames are tracked in a `schema_migrations` table;
`.down.sql` files exist for manual rollback via `psql` but are not executed
automatically.

Migrations run as a one-off administrative process (Factor XII) — never
automatically at container startup:

```bash
docker compose up -d db
docker compose run --rm orchestrator migrate
```

Tables created by `0001_init.up.sql`:

- **`bills`** — canonical bill records, `UNIQUE(source, external_id)` for
  idempotent re-ingestion.
- **`ingestion_runs`** — one row per ingestion attempt, for debugging API
  failures.
- **`relevance_results`** — relevance verdicts per bill; `method` is
  `keyword` today, `ml` once Phase B ships, without a schema change.
  `0002_relevance_unique.up.sql` adds an `evaluated_date` column and a
  `UNIQUE(bill_id, method, evaluated_date)` index so re-running detection
  the same day updates in place instead of duplicating (a plain column is
  used rather than an index on `evaluated_at::date`, since that cast is not
  IMMUTABLE and Postgres refuses to index it).
- **`reports`** — history of what was actually reported/notified.
- **`labels`** (reserved) — supervised-learning labels for the future ML
  sprint; not created until the labeling strategy is decided.

In production, Postgres is a managed instance (Neon/Supabase/Railway free
tier) reachable via `DATABASE_URL`, since GitHub Actions runners are
ephemeral and cannot host a persistent database container between scheduled
runs. In development, `docker-compose.yml` runs a local `postgres:16-alpine`
container instead — same schema, same migrations, same `DATABASE_URL`
mechanism (Factor X: dev/prod parity).
