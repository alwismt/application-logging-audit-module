# Application Logging & Audit Trail Module

A reusable Go component for **application logging** and **audit trail** tracking with **SQLite** (default) or **PostgreSQL** persistence, HTTP middleware, and admin APIs.

## Business problem

Modern applications need reliable technical logs for debugging and audit trails for security/compliance (who did what, when). This module provides both in one embeddable component with sanitization, REST APIs, and CBSE-style documentation.

## Features

- Log levels: INFO, WARNING, ERROR, DEBUG
- Audit actions: LOGIN, LOGOUT, CREATE/UPDATE/DELETE_RECORD, DOWNLOAD_FILE, FAILED_LOGIN, PERMISSION_DENIED
- Sensitive data masking (passwords, tokens, cards, secrets)
- HTTP middleware (request ID, latency, status)
- Admin API with JWT authentication and JSON/CSV export
- React + Vite admin frontend (login, logs, audit, export)
- **SQLite by default** for local dev and Docker (no database server)
- **PostgreSQL** via `DB_DRIVER=postgres` and `DB_*` env vars
- Flyway-style SQL migrations (optional CLI, Postgres only) + **Go auto-migrate on startup** when tables are missing
- Docker Compose: SQLite stack by default; optional Postgres overlay
- OpenAPI spec: [docs/swagger.yaml](docs/swagger.yaml)
- Optional Swagger UI at `/swagger/` when `ENABLE_SWAGGER_UI=true` (disabled by default)

## Project structure

```
cmd/server/          HTTP entry point
internal/
  app/               Wiring and routes
  common/            Pagination, HTTP helpers
  config/            Environment configuration
  database/          SQLite + PostgreSQL connectivity
  logger/            Logging component
  audit/             Audit trail component
  middleware/        HTTP logging middleware
  handler/           REST handlers
  exporter/          JSON/CSV export
database/migrations/ SQL migrations (Postgres + SQLite variants)
docs/                CBSE documentation + diagrams
frontend/            React + Vite admin UI
examples/            Integration example
```

Tests are co-located: `internal/<package>/*_test.go` (no top-level `tests/` folder).

## Installation

Repository: [https://github.com/alwismt/application-logging-audit-module](https://github.com/alwismt/application-logging-audit-module.git)

The module uses **SQLite by default**—no separate database server. You only need a `.env` file (from [`.env.example`](.env.example)) if you want PostgreSQL or custom settings.

| Path | Best for | Go required? |
|------|----------|----------------|
| [Go project (source)](#option-1-go-project-source) | Developers cloning this repo | Yes (1.22+) |
| [Pre-built binary](#option-2-pre-built-binary-no-go-required) | Any stack calling the REST API only | No |
| [Docker](#option-3-docker-recommended) | Fastest full stack (API + optional UI) | No |
| [Existing Go app (`go get`)](#option-4-use-in-an-existing-go-project-go-get) | Import as a library in your module | Yes (1.22+) |

### Option 1: Go project (source)

```bash
git clone https://github.com/alwismt/application-logging-audit-module.git
cd application-logging-audit-module
go mod tidy
make run
```

Server listens on `http://localhost:8080`. Tables are created automatically in `./data/logger.db` (SQLite).

### Option 2: Pre-built binary (no Go required)

1. Download the server binary for your OS/arch from [GitHub Releases](https://github.com/alwismt/application-logging-audit-module/releases).
2. Make it executable and run it (place [`.env.example`](.env.example) beside it only if you need overrides):

```bash
chmod +x server
./server
```

3. Call the REST API (see [API examples](#api-examples)). Default demo admin: `admin` / `12345678`.

No `.env` is required for SQLite—the binary creates `./data/logger.db` on first start. If Releases are not published yet, build locally: `make build` then run `./bin/server`.

### Option 3: Docker (recommended)

```bash
docker compose up --build
# or: make docker-up
```

SQLite data is stored in a Docker volume; no host database install. See [Quick start](#quick-start) below.

### Option 4: Use in an existing Go project (`go get`)

Add the module to your application:

```bash
go get github.com/alwismt/application-logging-audit-module@latest
```

Import the public package (do **not** import `internal/...` from outside this module):

```go
import "github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"
```

**Standalone server** (same as `make run` / `cmd/server`):

```go
mod, err := loggingaudit.NewFromEnv()
if err != nil { log.Fatal(err) }
log.Fatal(mod.Run())
```

**Mount HTTP routes** in an existing router (`/health`, `/logs`, `/admin`):

```go
mod, err := loggingaudit.NewFromEnv()
if err != nil { log.Fatal(err) }
r.Mount("/", mod.Handler())
```

See [examples/embed_router/main.go](examples/embed_router/main.go).

**Library only** (logger + audit, no HTTP server):

```go
mod, _ := loggingaudit.NewFromEnv()
_ = mod.Logger().Info(ctx, "hello", nil)
_ = mod.Auditor().Record(ctx, loggingaudit.AuditEvent{Action: "LOGIN", Status: "SUCCESS"})
```

See [examples/basic_usage.go](examples/basic_usage.go).

| Command | Purpose |
|---------|---------|
| `go get github.com/alwismt/application-logging-audit-module@latest` | Add library dependency to your `go.mod` |
| `go install github.com/alwismt/application-logging-audit-module/cmd/server@latest` | Install the standalone server binary to `$GOPATH/bin` |

Until a Git tag is published, use `@main` or a commit pseudo-version instead of `@latest`.

SQLite is the default: no `.env` required unless you switch to PostgreSQL or override settings.

### Database configuration

| Mode | When to use | What to configure |
|------|-------------|-------------------|
| **SQLite (default)** | Local dev, demos, binary-only use, default Docker | Nothing. Optional: `SQLITE_PATH` in `.env` |
| **PostgreSQL** | External / physical database server | `.env`: `DB_DRIVER=postgres` plus `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` |

For PostgreSQL locally or in Docker with a real DB server, see [Local development (PostgreSQL)](#local-development-postgresql) and [Docker with PostgreSQL](#docker-with-postgresql).

## Quick start

After [installation](#installation), verify the API (Docker or local/binary on port `8080`):

No `.env`, Postgres, or Flyway required for the default SQLite stack:

```bash
docker compose up --build
# or: make docker-up
```

```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/logs/log-error \
  -H "Content-Type: application/json" \
  -d '{"message":"Payment service timeout"}'
# Admin API requires login (demo: admin / 12345678)
TOKEN=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}' | jq -r .token)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/logs?level=ERROR"
```

### Docker with PostgreSQL {#docker-with-postgresql}

```bash
make docker-up-postgres
# docker compose -f docker-compose.yml -f docker-compose.postgres.yml up --build
```

## Local development (SQLite default)

```bash
cp .env.example .env   # optional; defaults work without .env
make run               # uses ./data/logger.db, auto-creates tables
```

To browse the API interactively, set `ENABLE_SWAGGER_UI=true` in `.env` and open http://localhost:8080/swagger/.

### Admin frontend

Requires Node.js 18+:

```bash
make frontend-install
make frontend-dev      # http://localhost:5173
```

In another terminal, run `make run` (backend on port 8080).

| Setting | Default |
|---------|---------|
| `VITE_API_URL` | `http://localhost:8080` (see `frontend/.env.example`) |
| Demo login | `admin` / `12345678` |

Override credentials via backend `.env`: `ADMIN_USERNAME`, `ADMIN_PASSWORD`.

## Frontend screenshots

<!-- Add screenshots of login, logs, and audit pages here -->

### Local development (PostgreSQL) {#local-development-postgresql}

```bash
cp .env.example .env
# Set DB_DRIVER=postgres and DB_* in .env
createdb loggerdb
make run
```

Optional: `make migrate-up` (Flyway, PostgreSQL only) if you prefer external schema management.

## API examples

```bash
# Demo audit login
curl -X POST http://localhost:8080/logs/audit-login \
  -H "Content-Type: application/json" \
  -d '{"username":"user25"}'

# Filter audit events (with token from login above)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/audit-events?action=LOGIN&status=SUCCESS&page=1&limit=20"

# Export logs as CSV
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/logs/export?format=csv" -o logs.csv
```

## Usage examples

```bash
# Public pkg/loggingaudit API (recommended for embedding)
go run ./examples/basic_usage.go

# Mount routes in Chi
go run ./examples/embed_router/main.go
```

## Testing

| Command | Description |
|---------|-------------|
| `make test` | Unit tests (includes SQLite repository smoke tests) |
| `make test-integration` | Requires PostgreSQL; optional `.env.dev` with `TEST_DB_*` |
| `make test-all` | Both |
| `make coverage` | HTML coverage report |

## Makefile targets

- `make run`, `build`, `tidy`
- `make migrate-up` — Flyway local migrate (PostgreSQL only)
- `make diagrams` — Export PlantUML to PNG
- `make docker-up` — SQLite Docker stack (+ frontend on port 5173)
- `make docker-up-postgres` — Docker with PostgreSQL
- `make frontend-install`, `frontend-dev`, `frontend-build` — Admin UI

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/component_description.md](docs/component_description.md) | Component overview (CBSE §1) |
| [docs/architecture.md](docs/architecture.md) | Internal design + diagrams (CBSE §2) |
| [docs/api_documentation.md](docs/api_documentation.md) | API reference (CBSE §3) |
| [docs/user_guide.md](docs/user_guide.md) | Installation & quick start |
| [docs/competing_components.md](docs/competing_components.md) | Competitor analysis (CBSE §4) |
| [docs/changelog.md](docs/changelog.md) | Version history |
| [docs/license.md](docs/license.md) | MIT License |

## Docker database access

**SQLite (default Compose):** data lives in the `sqlite_data` volume at `/data/logger.db` inside the app container.

**PostgreSQL overlay:** Postgres is internal-only (no host port 5432):

```bash
docker compose -f docker-compose.yml -f docker-compose.postgres.yml exec postgres psql -U postgres -d loggerdb
```

## Security note

Production systems should use HTTPS, secure password policies, RBAC, refresh tokens, CSRF protection, and stronger session management. Passwords are bcrypt-hashed in `admin_users`.

## Limitations

- Synchronous database writes only
- SQLite is suitable for local/demo; use PostgreSQL for production-like workloads
- External Go apps should import `pkg/loggingaudit` (not `internal/...`), or run this module as a separate HTTP service
- Integration tests require running PostgreSQL

## License

MIT — see [docs/license.md](docs/license.md)
