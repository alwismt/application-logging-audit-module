# Component Description

## Component name

**Application Logging & Audit Trail Module**

## Provided functions

### Application logging

- `Info`, `Warn`, `Error`, `Debug` — write structured application logs
- Log levels: INFO, WARNING, ERROR, DEBUG
- Optional console output and database persistence (SQLite default, PostgreSQL optional)
- Metadata JSON with automatic sensitive-data masking

### Audit trail

- `Record` — persist user actions with resource context
- Supported actions: LOGIN, LOGOUT, CREATE_RECORD, UPDATE_RECORD, DELETE_RECORD, DOWNLOAD_FILE, FAILED_LOGIN, PERMISSION_DENIED
- Status: SUCCESS or FAILURE
- Old/new value snapshots (sanitized)

### HTTP integration

- Middleware: request ID, latency, method, path, status, IP, user agent
- Admin API: list, get by ID, export (JSON/CSV)
- Demo endpoints for quick evaluation

## Purpose

Provide a reusable Go component that developers can embed in applications to:

1. Debug production issues via reliable technical logs
2. Meet security and compliance needs via immutable-style audit trails

## Business problems solved

| Problem | How this component helps |
|---------|--------------------------|
| Unknown errors in production | Centralized ERROR logs with stack traces and request IDs |
| Cannot trace user actions | Audit events tied to user, resource, IP, and timestamp |
| Sensitive data in logs | Built-in sanitizer masks passwords, tokens, cards |
| No admin visibility | REST admin endpoints for UI integration |

## Intended use

- Microservices and monoliths written in Go
- Teams needing SQLite (local/demo) or PostgreSQL-backed log and audit storage
- CBSE coursework demonstrating component-based design

## Restrictions

- Go 1.22+ required for source integration; pre-built binaries and Docker do not require Go on the host
- Persistence via SQLite (default) or PostgreSQL (`DB_DRIVER=postgres`)
- Admin routes require JWT (from `POST /admin/login`) or `X-API-Key` / `Authorization: ApiKey` — demo credentials only; harden for production
- Synchronous writes only (async logging documented as future work)
- External Go modules must use the public package `pkg/loggingaudit` (`go get github.com/alwismt/application-logging-audit-module`); `internal/` is not importable from other repos

## Important information for users

- **Go library:** `import "github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"` — provides `NewFromEnv`, `Handler`, `Run`, `Logger`, and `Auditor`
- **SQLite (default):** run the binary, `make run`, or Docker with no `.env`; data goes to `SQLITE_PATH` (default `./data/logger.db`)
- **PostgreSQL:** copy `.env.example` to `.env`, set `DB_DRIVER=postgres`, and configure `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- On first start the app auto-creates tables when they are missing (`DB_AUTO_MIGRATE=true`)
- Optional: copy `.env.dev.example` to `.env.dev` for integration test database overrides
- Logging failures return errors from services but do not crash HTTP handlers
- See [user_guide.md](user_guide.md) for installation and [competing_components.md](competing_components.md) for comparisons
