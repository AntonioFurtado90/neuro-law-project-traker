# Architecture

Single Git repository (monorepo), split into three independently containerized
components:

```
contracts/      JSON Schemas shared by scripts and orchestrator (the data contract)
scripts/        Python — data ingestion (Camara/Senado) and relevance detection
orchestrator/   Go — pipeline sequencing, persistence, reporting
```

Each component builds its own Docker image and runs in its own container —
`scripts`, `orchestrator`, and `db` (Postgres) never share a runtime
environment or a dependency tree. This keeps the language split
(Python for scripts, Go for orchestration/backend) clean and avoids
dependency hell between components.

## Responsibilities

- **scripts (Python)** — stateless. Reads input via `--input`/env vars, writes
  a JSON output file conforming to a schema in `contracts/schemas/`, exits.
  Never talks to the database directly.
- **orchestrator (Go)** — the only component with database credentials.
  Invokes the `scripts` container once per pipeline step (ingest-camara,
  ingest-senado, detect-relevance, ...), reads its JSON output, persists to
  Postgres, and writes the final report through a pluggable `Sink`.
- **db (Postgres)** — the only piece of the system that persists state across
  runs. In production it is a managed instance (Neon/Supabase/Railway), since
  GitHub Actions runners are ephemeral.

See [`contract.md`](contract.md) for how `orchestrator` and `scripts`
communicate, [`database.md`](database.md) for the schema, and
[`roadmap.md`](roadmap.md) for the sprint plan.
