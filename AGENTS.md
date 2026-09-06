# Project

POS (point-of-sale) backend — Goravel (Go, Laravel-style) + PostgreSQL. API-only. Single-tenant, multi-outlet-ready schema.

- Module: `github.com/ridhoauliama97/pos-server` (Go 1.25, Goravel v1.18)
- Drivers: `goravel/gin` (HTTP), `goravel/postgres` (DB), `goravel/openai` (AI)

## Commands

```bash
# Database (Postgres via Docker)
docker compose up -d          # starts postgres (pos) + goravel app

# Run in dev (hot reload)
air

# Run without reload
go run .

# Tests
go test ./...

# Build / static checks
go build ./...
go vet ./...
gofmt -w .
```

## Structure

- `app/facades/` — facade accessors, one thin `return App().MakeX()` per service
- `app/http/controllers/` — HTTP controllers (api/v1 etc.)
- `app/models/` — ORM models
- `app/http/middleware/` — middlewares (role/auth)
- `app/services/` — business logic (TransactionService, StockService, ReportService)
- `app/policies/` — per-role authorization
- `bootstrap/` — app wiring: `Boot()`, `Migrations`, `Providers`
- `config/` — env-driven config; each file's `init()` calls `facades.Config().Add(...)`
- `database/migrations/` — Go migrations, registered in `bootstrap/migrations.go`
- `routes/` — route registration via `facades.Route()` (`web.go`, `grpc.go`); bootstrap calls these inside `WithRouting`
- `tests/` — Go tests (testify suite); boot the app via `bootstrap.Boot()` (see `tests/test_case.go`)

## Conventions

- Migrations implement `Signature()` / `Up()` / `Down()` and must be registered in `bootstrap/migrations.go`
- Routes are registered in `routes/` packages, enabled from `bootstrap/app.go` `WithRouting`
- Config files self-register via package `init()`; `config.Boot()` stays empty (anchor only)
- Controllers take `http.Context` and return `http.Response`

## Fixed Architecture Decisions (from PRD)

- **Online-first server.** Offline handling is client's job (local storage + sync queue), not backend.
- **Idempotent transactions** — client sends `client_uuid`; server returns existing transaction on replay (no double insert). Prepares future `POST /transactions/bulk-sync`.
- **Payment methods = table** (`payment_methods`), seeded with `cash`; extensible for gateways later.
- **Multi-outlet-ready schema** from day one — `outlet_id` on relevant tables, though only one default outlet exists now (single-tenant, not multi-tenant SaaS).
- **Structured receipt endpoint** returns JSON; thermal/ESC-POS printing is client-side.
- Keep audit `stock_movements` ledger from the start — table `transaction_items` stores `_snapshot` prices so historical receipts never change.

## Task Tracking

Implementation is broken down in `.tmp/tasks/` (features `phase-0-setup`, `phase-1-mvp-core`, `phase-2-enhancements`). Manage them with the task-management skill CLI:

```bash
node "C:\Users\Hexoda\.opencode\skills\task-management\scripts\task-cli.ts" status
node "C:\Users\Hexoda\.opencode\skills\task-management\scripts\task-cli.ts" next
node "C:\Users\Hexoda\.opencode\skills\task-management\scripts\task-cli.ts" complete <feature> <seq> "summary"
```

Note: use `node` directly (Node ≥ 23 strips types); `npx ts-node` hangs in this environment.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:

- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
