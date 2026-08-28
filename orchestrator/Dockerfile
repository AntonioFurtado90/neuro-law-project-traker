# Build context is the repo root (kept consistent with scripts/Dockerfile;
# also gives future sprints access to ../contracts for contract checks).
# docker build -f orchestrator/Dockerfile --target runtime .

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY orchestrator/go.mod ./
COPY orchestrator/ ./
RUN go vet ./...
RUN go test ./...
RUN CGO_ENABLED=0 go build -o /out/monitor ./cmd/monitor

FROM gcr.io/distroless/static-debian12 AS runtime
COPY --from=build /out/monitor /usr/local/bin/monitor
ENTRYPOINT ["/usr/local/bin/monitor"]
