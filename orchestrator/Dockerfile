# Build context is the repo root (needs access to ../contracts for the
# models/bill contract test, same reasoning as scripts/Dockerfile).
# docker build -f orchestrator/Dockerfile --target runtime .

FROM golang:1.25-alpine AS deps
WORKDIR /app
COPY orchestrator/go.mod orchestrator/go.sum ./orchestrator/
WORKDIR /app/orchestrator
RUN go mod download

FROM deps AS build
WORKDIR /app
COPY contracts ./contracts
COPY orchestrator ./orchestrator
WORKDIR /app/orchestrator
RUN go vet ./...
RUN go test ./...
RUN CGO_ENABLED=0 go build -o /out/monitor ./cmd/monitor

FROM deps AS test-integration
WORKDIR /app
COPY contracts ./contracts
COPY orchestrator ./orchestrator
WORKDIR /app/orchestrator
ENTRYPOINT ["go", "test", "-tags=integration", "./..."]

FROM gcr.io/distroless/static-debian12 AS runtime
COPY --from=build /out/monitor /usr/local/bin/monitor
ENTRYPOINT ["/usr/local/bin/monitor"]
