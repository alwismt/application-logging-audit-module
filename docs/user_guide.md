# User Guide

## Dependencies

- Go 1.22+
- Node.js 18+ (for admin frontend)
- Docker & Docker Compose (recommended)
- PostgreSQL 14+ (optional — only when `DB_DRIVER=postgres`)
- Flyway (optional, PostgreSQL migrations via CLI)

Go modules: `pgx/v5`, `chi/v5`, `google/uuid`, `joho/godotenv`, `modernc.org/sqlite`

## Installation

Repository: [https://github.com/alwismt/application-logging-audit-module](https://github.com/alwismt/application-logging-audit-module.git)

**SQLite is the default database**—no PostgreSQL server required. Copy [`.env.example`](../.env.example) to `.env` only when you need PostgreSQL or custom ports, credentials, or secrets.

### Option 1: Go project (source)

Requires Go 1.22+.

```bash
git clone https://github.com/alwismt/application-logging-audit-module.git
cd application-logging-audit-module
go mod tidy
make run
```

Creates `./data/logger.db` and tables on first start (`DB_AUTO_MIGRATE=true`).

### Option 2: Pre-built binary (no Go required)

1. Download the server binary from [GitHub Releases](https://github.com/alwismt/application-logging-audit-module/releases) for your platform.
2. Run it from any directory (optional `.env` beside the binary):

```bash
chmod +x server
./server
```

3. Use the REST API at `http://localhost:8080` (see [API documentation](api_documentation.md)). Demo admin: `admin` / `12345678`.

If no release artifact is available, build from source: `make build` → `./bin/server`.

### Option 3: Docker

```bash
docker compose up --build
# or: make docker-up
```

Uses SQLite in a Docker volume; no `.env` required for the default stack.

### Database mode

| Mode | Setup |
|------|--------|
| **SQLite (default)** | No `.env` needed. Optional `SQLITE_PATH` (default `./data/logger.db`). |
| **PostgreSQL** | `.env` with `DB_DRIVER=postgres` and `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` |

See [Environment files](#environment-files) below for all variables.

## 5-minute quick start (Docker, SQLite)

One command — no `.env`, no Postgres, no Flyway:

```bash
docker compose up --build
```

The app stores data in a Docker volume (`sqlite_data`) and publishes **only** `localhost:8080`.

Wait until the app logs `Server listening on :8080`, then:

```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/logs/log-info \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello from quick start"}'
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}'
```

### Docker with PostgreSQL

```bash
make docker-up-postgres
```

Uses `docker-compose.yml` + `docker-compose.postgres.yml` (internal Postgres, no host port 5432).

## Environment files

| File | Purpose |
|------|---------|
| `.env.example` | `DB_DRIVER`, `SQLITE_PATH`, and optional `DB_*` for Postgres |
| `.env.dev.example` | Optional `TEST_DB_*` for integration tests (copy to `.env.dev`) |

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DRIVER` | `sqlite` | `sqlite` or `postgres` |
| `SQLITE_PATH` | `./data/logger.db` | SQLite file path (when driver is sqlite) |
| `DB_HOST`, `DB_PORT`, … | see `.env.example` | Used when `DB_DRIVER=postgres` |
| `ADMIN_USERNAME` | `admin` | Seeded admin user if missing |
| `ADMIN_PASSWORD` | `12345678` | Demo password (bcrypt-hashed in DB) |
| `JWT_SECRET` | `change-this-secret` | HS256 signing key |
| `CORS_ORIGINS` | `http://localhost:5173` | Allowed browser origins |

For PostgreSQL, the URL is built in Go from `DB_*` fields — not stored as a single string in `.env`.

## Admin frontend

```bash
make frontend-install
cp frontend/.env.example frontend/.env   # optional
make frontend-dev
```

Open http://localhost:5173 and sign in with `admin` / `12345678` (unless overridden in backend `.env`).

### Authentication flow

1. Browser posts credentials to `POST /admin/login`.
2. Server validates against `admin_users` (seeded from env on first start).
3. Server returns a JWT; the frontend stores it in `localStorage` as `admin_token`.
4. Subsequent API calls send `Authorization: Bearer <token>`.
5. On `401`, the frontend clears the token and redirects to `/login`.

Alternative for scripts: `X-API-Key` header matching `ADMIN_API_KEY`.

### Swagger UI (development only)

Set `ENABLE_SWAGGER_UI=true` in `.env`, restart the server, and open http://localhost:8080/swagger/. The UI loads Swagger UI from a CDN and uses the embedded OpenAPI spec. Leave this disabled in production (`ENABLE_SWAGGER_UI=false` is the default).

### Production security

This auth model is for demonstration. Do not expose admin credentials or JWT secrets in production without TLS, proper secret management, role-based access control, refresh tokens, and CSRF protection for cookie-based sessions.

## Local setup (SQLite, default)

Same as [Option 1: Go project](#option-1-go-project-source) or run the [pre-built binary](#option-2-pre-built-binary-no-go-required) without Go.

```bash
make run
```

Creates `./data/logger.db` and tables on first start when `DB_AUTO_MIGRATE=true`.

## Local setup (PostgreSQL)

1. `createdb loggerdb`
2. Copy `.env.example` to `.env`, set `DB_DRIVER=postgres` and adjust `DB_*`
3. `make run`

Set `DB_AUTO_MIGRATE=false` in production if you manage schema externally (app fails start if tables are missing).

## Using the component in your code

See [examples/basic_usage.go](../examples/basic_usage.go). With SQLite (default):

```go
db, _ := database.ConnectSQLite(cfg.SQLitePath)
_ = database.EnsureSchemaSQLite(ctx, db, true)
logRepo := logger.NewSQLiteRepository(db)
loggerSvc := logger.NewService(logRepo, "my-service", true, true)
```

With PostgreSQL, use `database.Connect`, `EnsureSchema`, and `NewPostgresRepository`.

## Testing

### Unit tests (no PostgreSQL)

```bash
make test
# or: go test ./...
```

Includes SQLite repository smoke tests (`:memory:`).

### Integration tests (requires PostgreSQL)

```bash
createdb loggerdb_test
cp .env.dev.example .env.dev   # optional: uncomment TEST_DB_* overrides
make test-integration
```

Uses `TEST_DB_*` with fallbacks to `DB_*` and default database name `loggerdb_test`. App wiring uses `DB_DRIVER=postgres` in test setup.

### Coverage

```bash
make coverage
```

## Deployment requirements

**Docker Compose (SQLite):** env in `docker-compose.yml`; no database container.

**Docker Compose (PostgreSQL):** `make docker-up-postgres`; app connects to `postgres:5432` on the internal network.

**Local / bare metal:** SQLite file path writable, or PostgreSQL reachable when `DB_DRIVER=postgres`; tables auto-create on start when `DB_AUTO_MIGRATE=true` (or optional `make migrate-up` with Flyway for Postgres).

- Terminate TLS at a reverse proxy; rotate `JWT_SECRET` and `ADMIN_API_KEY` per environment

## Troubleshooting

**Port 8080 in use:** change `APP_PORT` in `.env` or stop the other process.

**Postgres connection refused (postgres driver):** ensure Postgres is running and `DB_*` match your instance.

**SQLite permission errors:** ensure the directory for `SQLITE_PATH` exists and is writable (the app creates parent dirs).

## Inspecting data

**SQLite (local):**

```bash
sqlite3 ./data/logger.db "SELECT id, level, message FROM application_logs LIMIT 5;"
```

**PostgreSQL (Compose overlay):**

```bash
docker compose -f docker-compose.yml -f docker-compose.postgres.yml exec postgres psql -U postgres -d loggerdb
```
