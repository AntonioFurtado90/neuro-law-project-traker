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
  Persists ingestion/relevance results to Postgres (`load-bills`,
  `generate-report`) and writes the final report through a pluggable `Sink`.
  It does **not** invoke the `scripts` container itself — see below.
- **db (Postgres)** — the only piece of the system that persists state across
  runs. In production it is a managed instance (Neon/Supabase/Railway), since
  GitHub Actions runners are ephemeral.

## Who sequences the pipeline

An earlier sketch of this project had the orchestrator invoke `scripts` as a
container itself (`docker compose run` from inside Go). That was deliberately
dropped: it would require mounting the host's Docker socket into the
`orchestrator` container and installing the Docker CLI there, breaking the
minimal/distroless runtime image and granting root-equivalent host access
just to sequence a few commands. Instead, [`bin/run-pipeline.sh`](../bin/run-pipeline.sh)
sequences the `docker compose run` calls for both containers, in order, from
outside either of them. Sprint 5's `daily-monitor.yml` just calls this same
script on a schedule.

See [`contract.md`](contract.md) for how `orchestrator` and `scripts`
communicate, [`database.md`](database.md) for the schema, and
[`roadmap.md`](roadmap.md) for the sprint plan.
