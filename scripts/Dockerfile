# Build context is the repo root (needs access to ../contracts for schema tests).
# docker build -f scripts/Dockerfile --target runtime .

FROM python:3.12-slim AS base
WORKDIR /app
COPY scripts/requirements.txt scripts/requirements.txt
RUN pip install --no-cache-dir -r scripts/requirements.txt

FROM base AS test
COPY scripts/requirements-dev.txt scripts/requirements-dev.txt
RUN pip install --no-cache-dir -r scripts/requirements-dev.txt
COPY contracts ./contracts
COPY scripts ./scripts
RUN pip install --no-cache-dir -e ./scripts
RUN ruff check scripts/src scripts/tests
RUN pytest scripts/tests

FROM base AS runtime
COPY scripts/pyproject.toml scripts/pyproject.toml
COPY scripts/src ./scripts/src
RUN pip install --no-cache-dir -e ./scripts --no-deps
ENTRYPOINT ["neurolaw"]
