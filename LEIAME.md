# Neuro Law Project Tracker

Um bot que monitora diariamente Projetos de Lei apresentados na Câmara dos
Deputados e no Congresso Nacional (Senado Federal) que possam impactar os
**Fundos Constitucionais** — FCO, FNE, FNO, os fundos constitucionais de
financiamento regional previstos no Art. 159 da Constituição Federal.

## Como funciona (MVP)

1. **Ingestão** (Python) coleta dados de PLs a partir das APIs oficiais de
   dados abertos da Câmara dos Deputados e do Senado Federal.
2. **Detecção de relevância** (Python) sinaliza os PLs cuja ementa/texto
   corresponde a um conjunto curado de termos relacionados aos fundos
   constitucionais.
3. **Orquestração e persistência** (Go) sequencia a pipeline, grava os
   resultados no Postgres e gera um relatório diário.

Cada parte — os scripts em Python, o orquestrador em Go e o banco Postgres —
roda em seu próprio container, sem compartilhar árvore de dependências. Veja
[`docs/architecture.md`](docs/architecture.md) para o desenho completo,
[`docs/contract.md`](docs/contract.md) para o contrato entre os containers,
[`docs/database.md`](docs/database.md) para o esquema do banco e
[`docs/roadmap.md`](docs/roadmap.md) para o plano de sprints.

## Requisitos

- Docker e Docker Compose (tudo roda containerizado, inclusive o
  desenvolvimento local, para manter a paridade entre dev e produção).

## Começando

```bash
cp .env.example .env
docker compose build
docker compose run --rm scripts version
docker compose run --rm orchestrator version
docker compose up -d db
docker compose run --rm orchestrator migrate
```

## Ingerindo PLs da Câmara

```bash
RUN_WINDOW_START=2026-08-01 RUN_WINDOW_END=2026-08-05 \
  docker compose run --rm scripts ingest-camara --output /workdir/camara.json
docker compose run --rm orchestrator load-bills --input /workdir/camara.json
```

`ingest-camara` escreve um envelope `ingestion_result.json`; `load-bills`
persiste isso no Postgres (idempotente — rodar de novo com o mesmo arquivo
não cria duplicatas). Esse fluxo manual de dois passos é o que a pipeline
automatizada (Sprint 4) vai rodar sozinha.

## Rodando os testes

Os testes rodam como parte do build das imagens Docker (estágios
`test`/`build`), da mesma forma que o CI executa:

```bash
docker build --target test -f scripts/Dockerfile .
docker build --target build -f orchestrator/Dockerfile .
```

Os testes que dependem do banco precisam de um Postgres rodando e são
executados à parte (é assim também que o job `orchestrator-db-tests` do CI
roda):

```bash
docker compose up -d db
docker compose run --rm orchestrator migrate
docker compose run --rm orchestrator-integration-tests
```

## Práticas de engenharia

Este projeto segue a metodologia Twelve-Factor App e um framework mínimo de
engenharia: toda mudança vem com testes, o CI roda a cada push, a branch
`main` está sempre verde, e os lançamentos são pequenos e incrementais. Veja
o `CLAUDE.md` para o conjunto completo de diretrizes seguidas pelo projeto.

## English

An English version of this document is available at
[`README.md`](README.md).
