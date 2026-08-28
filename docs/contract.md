# Go ↔ Python Contract

`orchestrator` and `scripts` never share a process or a filesystem beyond an
explicit shared work directory. The contract between them is:

1. **Invocation**: the orchestrator runs the `scripts` image as a container
   per pipeline step (`docker compose run --rm scripts <subcommand> ...` in
   dev, an equivalent container step in CI), passing configuration only via
   environment variables and `--input`/`--output` file paths.
2. **Data**: input/output are JSON files validated against the schemas in
   [`contracts/schemas/`](../contracts/schemas). Payloads are never passed
   over stdout — stdout/stderr are reserved for structured logs (Factor XI).
3. **Exit codes**:
   - `0` — success, output file is guaranteed schema-valid.
   - `1` — recoverable/partial failure (see the `status` field in
     `ingestion_result.schema.json`); the orchestrator decides whether to
     continue with partial data.
   - `2` — fatal/config error; the orchestrator aborts the pipeline.
4. **Work directory**: `RUN_WORKDIR` holds the intermediate JSON files for a
   single run. It is never assumed to persist between runs (Factor VI).

Schemas currently defined:

- `bill.schema.json` — canonical representation of a Projeto de Lei.
- `ingestion_result.schema.json` — envelope returned by ingestion subcommands.
- `relevance_report.schema.json` — envelope returned by relevance-detection
  subcommands (keyword-based today, ML-based in a future phase — both produce
  the same shape, so the orchestrator never needs to change).
