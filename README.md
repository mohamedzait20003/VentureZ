# VentureZ

An **agentic CFO / financial planner for startups** — an AI that answers cash-flow and
planning questions ("Can we afford two senior hires this quarter?") backed by deterministic,
auditable financial math.

> Portfolio project demonstrating a polyglot microservice architecture with an AI agent core.

## Architecture at a glance

```
          TanStack Start + TypeScript (frontend)
                          │ REST / JSON
                   ┌──────▼──────┐
                   │  Gateway / BFF │  Go
                   └──────┬──────┘
                   gRPC   │   gRPC
              ┌───────────┴───────────┐
              ▼                       ▼
        ┌──────────┐           ┌──────────────┐
        │  Agent    │  gRPC →   │ Financial    │
        │ (Python)  │ ◀───────  │ Engine (Go)  │
        │  Claude   │   tool    └──────┬───────┘
        └──────────┘                   │
                                   Postgres
```

**Core decision — the LLM never does arithmetic.** The Python agent reasons and orchestrates;
every actual number comes from the typed Go financial-engine, called as a tool over gRPC. This
keeps results correct, deterministic, and auditable. See [backend/docs/adr/](backend/docs/adr/).

## Tech stack

| Layer | Tech | Why |
|---|---|---|
| Frontend | TanStack Start, TypeScript | Full-stack React framework (file-based routing, SSR) |
| — routing/data | TanStack Router + React Query (TanStack Query) | Type-safe routing + server-state caching |
| — client state | Zustand | Lightweight global UI/client state |
| Gateway | Go | Low-latency REST edge + gRPC fan-out |
| Agent | Python, FastAPI, Anthropic SDK | Where the AI ecosystem lives |
| Financial Engine | Go | Deterministic, tested money math |
| Contracts | Protocol Buffers + gRPC | One typed contract across languages |
| Data | PostgreSQL | Scenarios & projections |
| Observability | OpenTelemetry + Jaeger | One trace across all services |

## Repository layout

```
VentureZ/
├── frontend/   TanStack Start app (TypeScript, React Query, Zustand)
└── backend/
    ├── proto/      Contract-first protobuf API definitions (source of truth)
    ├── gen/        Generated gRPC stubs (Go + Python)
    ├── services/   gateway (Go), financial-engine (Go), agent (Python)
    ├── deploy/     Infra-as-config (cloud deploy, OpenTelemetry collector)
    └── docs/       Architecture Decision Records (ADRs) + C4 diagrams
```

## Getting started

```sh
# Frontend
cd frontend && npm install && npm run dev

# Backend (all services + postgres + jaeger)
cd backend && make proto && make up
```

Requires an `ANTHROPIC_API_KEY` for the agent service (see `backend/README.md`).
