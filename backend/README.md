# VentureZ — Backend

Agentic CFO / financial planner for startups. Polyglot microservices.

## Layout

```
backend/
├── proto/      Contract-first API definitions (protobuf) — single source of truth
├── gen/        Generated gRPC stubs (Go + Python) produced from proto/
├── services/   The runnable services
│   ├── gateway/            Go   — REST BFF for the frontend, gRPC clients downstream
│   ├── financial-engine/   Go   — deterministic money math (runway, cashflow, scenarios)
│   └── agent/              Python — agentic CFO brain (Claude tool-use loop)
├── deploy/     Infra-as-config (cloud deploy, OpenTelemetry collector)
└── docs/       Architecture Decision Records (ADRs) + C4 diagrams
```

## Key architecture decision

**The LLM never does arithmetic.** The Python agent reasons and orchestrates; every actual
number comes from the typed Go financial-engine, called as a tool over gRPC. This keeps
results correct, deterministic, and auditable. See [docs/adr/](docs/adr/).

## Talking to each other

- **Frontend → gateway:** REST / JSON (browser-friendly)
- **service → service:** gRPC + protobuf (typed, cross-language, contract-first)

## Quick start

```sh
make proto     # generate Go + Python stubs from proto/
make up        # docker-compose: all services + postgres + jaeger (tracing)
make test      # run tests across services
```
