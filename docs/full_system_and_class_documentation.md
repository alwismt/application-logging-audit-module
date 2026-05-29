# Application Logging & Audit Trail Module — Full System and Class Documentation

This document describes the complete implementation of the **Application Logging & Audit Trail Module**, a standalone Go component for CBSE (Component Based Software Engineering) coursework. All details below are derived from the actual codebase unless marked as **Not implemented yet / recommended improvement**.

---

## 1. Project Overview

### Project name

**Application Logging & Audit Trail Module** (Go module: `application-logging-audit-module`)

### Purpose of the component

A reusable component that developers can plug into their applications to:

- Securely log technical errors and operational events (application logs)
- Track user actions for security, debugging, monitoring, and compliance (audit trail)

### Business problem solved

| Problem | How this component helps |
|---------|--------------------------|
| Production errors are hard to diagnose | Centralized ERROR logs with stack traces, request IDs, and metadata |
| User actions cannot be traced | Audit events tied to user, resource, IP, and timestamp |
| Sensitive data may leak into logs | Built-in `SensitiveDataSanitizer` masks passwords, tokens, and card numbers |
| No admin visibility into stored data | REST admin APIs plus a React admin UI for search, filter, and export |

### Why logging and audit trail are important

- **Application logs** answer: *What went wrong in the system?* (exceptions, service failures, HTTP errors)
- **Audit trail events** answer: *Who did what, when, and was it allowed?* (login, record updates, permission denials)

Together they support incident response, compliance audits, and debugging without mixing technical noise with business accountability records.

### Reuse in other projects

The component is designed as layered Go packages (`internal/logger`, `internal/audit`) with interfaces (`Logger`, `Auditor`, repositories). Another Go application can wire repositories and services directly (see `examples/basic_usage.go`) or run this project as a standalone HTTP service with admin APIs.

### Main technologies

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22, chi router, pgx (PostgreSQL), modernc.org/sqlite |
| Auth | golang-jwt/jwt, bcrypt (`golang.org/x/crypto`) |
| Database | SQLite (default) or PostgreSQL |
| Frontend | React 18, Vite 5, react-router-dom |
| API docs | OpenAPI (`docs/swagger.yaml`), optional Swagger UI |
| Diagrams | PlantUML (`.puml` in `docs/diagrams/`) |

### Main users

- **Application developers** — embed logging/audit services in Go code
- **Operators / SRE** — use HTTP middleware and health checks
- **Administrators** — use admin UI or REST APIs to query and export logs and audit events
- **CBSE evaluators** — review architecture, interfaces, tests, and documentation

### Application logs vs audit trail events

| Aspect | Application logs | Audit trail events |
|--------|------------------|-------------------|
| Purpose | Technical/debug information | Business/security accountability |
| Typical content | Level, message, stack trace, service metadata | Action, resource, old/new values, user, status |
| Levels / types | INFO, WARNING, ERROR, DEBUG | Actions: LOGIN, UPDATE_RECORD, etc.; status SUCCESS/FAILURE |
| Table | `application_logs` | `audit_events` |
| Primary API | `Logger` interface (`Info`, `Warn`, `Error`, `Debug`) | `Auditor` / `AuditService.Record` |

---

## 2. CBSE Requirement Mapping

| CBSE Requirement | Where It Is Implemented | Explanation |
|------------------|-------------------------|-------------|
| 1. Component description | `docs/component_description.md`, §1 of this document | Name, capabilities, purpose |
| 2. Provided functions | `internal/logger`, `internal/audit`, `internal/middleware` | Logging levels, audit actions, HTTP logging |
| 3. Purpose | §1 above, `docs/component_description.md` | Debug + compliance use cases |
| 4. Business problems solved | §1 table, `docs/component_description.md` | Error tracing, audit, sanitization, admin visibility |
| 5. Intended use | `docs/user_guide.md`, `examples/basic_usage.go` | Go apps, SQLite/Postgres, demo via Docker |
| 6. Restrictions | §24 of this document | Demo auth, sync writes, `internal/` import limits |
| 7. Internal architecture | `docs/architecture.md`, §3–4 here | Layered design: handlers → services → repositories |
| 8. Class diagram support | `docs/diagrams/class_diagram.puml` | Interfaces and concrete types |
| 9. Sequence diagram support | `docs/diagrams/sequence_diagram.puml` | Demo audit update flow |
| 10. State diagram support | `docs/diagrams/state_diagram.puml` | Log/audit entry lifecycle |
| 11. Deployment diagram support | `docs/diagrams/deployment_diagram.puml` | Browser, app, DB, Docker |
| 12. User documentation | `docs/user_guide.md`, `frontend/README.md` | Install, run, admin UI |
| 13. API documentation | `docs/api_documentation.md`, `docs/swagger.yaml` | REST routes and types |
| 14. Usage examples | `examples/basic_usage.go`, §22 curl examples | Programmatic and HTTP usage |
| 15. Installation instructions | `README.md`, `docs/user_guide.md`, §23 | Go, Node, Docker, env files |
| 16. Dependencies | `go.mod`, `frontend/package.json` | Listed with versions |
| 17. Deployment requirements | §23, `docker-compose.yml` | Ports, env vars, migrations |
| 18. Changelog | `docs/changelog.md` | v1.0.0, v1.1.0, Unreleased |
| 19. License | `docs/license.md` (MIT) | No `LICENSE` at repo root |
| 20. Competing component analysis | `docs/competing_components.md`, §25 | Logrus, Zap, ELK, etc. |

---

## 3. System Architecture

### High-level layers

1. **Backend / HTTP layer** — `cmd/server`, `internal/app` (chi router, server lifecycle)
2. **Component core** — `internal/logger`, `internal/audit` (business logic, validation, sanitization)
3. **Persistence** — `internal/logger/repository*.go`, `internal/audit/repository*.go`, `internal/adminauth`
4. **Middleware** — HTTP logging, CORS, admin authentication
5. **Handler / API** — `internal/handler` (REST endpoints)
6. **Frontend / admin UI** — `frontend/` (React)
7. **Authentication** — `internal/adminauth`, `internal/middleware/admin_auth.go`
8. **Testing** — unit tests (`go test ./...`) and integration tests (`-tags=integration`)

### Request flow

```text
Frontend/Admin Client → Backend API (chi) → CORS → HTTP Logger Middleware
  → Handler → [Admin Auth Middleware on /admin/* except login]
  → LoggerService / AuditService → Repository → SQLite or PostgreSQL
```

### API gateway style request logging

`HTTPLogger` middleware (`internal/middleware/http_logger.go`) wraps each request:

- Assigns or propagates `X-Request-ID`
- Wraps `ResponseWriter` to capture HTTP status code
- Measures latency
- Captures method, path, IP (`X-Forwarded-For` / `X-Real-IP` / `RemoteAddr`), user agent
- Writes an INFO log via `LoggerService.LogHTTP` (errors from logging are ignored with `_ =` so requests always complete)

This mimics API gateway access logs without a separate gateway product.

---

## 4. Folder Structure Explanation

```text
cbse/
├── cmd/server/              # Application entry point (main.go)
├── internal/
│   ├── app/                 # App wiring, routes, server Run/Shutdown
│   ├── adminauth/           # Admin users, JWT, bcrypt, seed
│   ├── audit/               # Audit domain (service, repos, types)
│   ├── common/              # JSON helpers, pagination, query time parsing
│   ├── config/              # Env-based configuration
│   ├── database/            # Connect, migrate, ping (Postgres + SQLite)
│   ├── exporter/            # JSON/CSV export helpers
│   ├── handler/             # HTTP handlers (health, logs, audit, auth)
│   ├── logger/              # Logging domain (service, sanitizer, repos)
│   ├── middleware/          # HTTP logger, CORS, admin auth
│   └── swagger/             # Embedded OpenAPI + Swagger UI routes
├── database/
│   ├── embed.go             # Embeds SQL migration files for Go auto-migrate
│   └── migrations/          # Flyway-style SQL (Postgres + SQLite variants)
├── docs/                    # CBSE and user documentation, swagger.yaml, diagrams
├── docs/diagrams/           # PlantUML source (.puml)
├── examples/                # basic_usage.go integration example
├── frontend/                # React admin UI
├── Dockerfile               # Backend container image
├── docker-compose.yml       # SQLite stack (app + frontend)
├── docker-compose.postgres.yml  # Overlay for PostgreSQL
├── Makefile                 # run, test, docker, frontend, diagrams
├── go.mod / go.sum
└── README.md
```

| Folder | Purpose |
|--------|---------|
| `cmd/server` | Minimal `main`: load config, `app.New`, `app.Run` |
| `internal/app` | Composes DB, services, middleware, routes, HTTP server |
| `internal/config` | Environment variables and database URL builders |
| `internal/database` | Connection pools, schema ensure, embedded migrations |
| `internal/logger` | Application logging component |
| `internal/audit` | Audit trail component |
| `internal/middleware` | Cross-cutting HTTP concerns |
| `internal/handler` | REST API adapters |
| `internal/exporter` | Serialize logs/audit for download |
| `database/migrations` | Versioned SQL schema (V1_1 logs/audit, V1_2 admin) |
| `docs` | User guide, API docs, changelog, license, diagrams |
| `frontend` | Admin login, dashboard, logs, audit pages |
| `*_test.go` | Unit and integration tests throughout `internal/` |

---

## 5. Backend Entry Point

### File: `cmd/server/main.go`

**Responsibilities:**

1. `config.Load()` — reads `.env` (via godotenv) and environment variables; exits on error
2. `app.New(cfg)` — connects database, ensures schema, seeds admin user, builds services and router
3. `application.Run()` — starts HTTP server, waits for SIGINT/SIGTERM, graceful shutdown

### `internal/app/app.go`

| Function | Purpose |
|----------|---------|
| `New(cfg)` | Selects SQLite or Postgres driver; runs `EnsureSchema*` and `EnsureAdminSchema*`; seeds admin; creates `LoggerService`, `AuditService`, middleware; registers routes; creates `http.Server` on `:{AppPort}` with 15s read/write timeouts |
| `Run()` | `ListenAndServe` in goroutine; on signal, `Shutdown` with 10s timeout; closes DB |
| `closeDB()` | Closes pgx pool or SQLite `*sql.DB` |
| `Router()` | Exposes handler for tests |

### Route registration

All routes are defined in `internal/app/routes.go` (see §11).

### Shutdown

Implemented: `signal.Notify` for SIGINT/SIGTERM, `server.Shutdown`, then `closeDB()`.

---

## 6. Configuration System

### Config struct (`internal/config/config.go`)

Fields: `AppEnv`, `AppPort`, `DBDriver`, `SQLitePath`, `DatabaseURL` (Postgres only), `DBHost`, `DBPort`, `DBName`, `DBUser`, `DBPassword`, `DBSSLMode`, `DBAutoMigrate`, `ServiceName`, `EnableConsoleLogging`, `EnableDatabaseLogging`, `LogLevel`, `AdminUsername`, `AdminPassword`, `AdminAPIKey`, `JWTSecret`, `JWTExpiryHours`, `CORSOrigins`, `EnableSwaggerUI`.

`Load()` calls `godotenv.Load()` (ignored if `.env` missing). For `DB_DRIVER=postgres`, `DatabaseURL` is built via `BuildDatabaseURL` from split `DB_*` variables (not read as a single `DATABASE_URL` env var).

### Environment variables

| Environment Variable | Purpose | Example |
|---------------------|---------|---------|
| `APP_ENV` | Runtime environment label | `local`, `docker` |
| `APP_PORT` | HTTP listen port | `8080` |
| `DB_DRIVER` | Database backend | `sqlite` (default), `postgres` |
| `SQLITE_PATH` | SQLite file path | `./data/logger.db` |
| `DB_HOST` | Postgres host | `localhost` |
| `DB_PORT` | Postgres port | `5432` |
| `DB_NAME` | Postgres database name | `loggerdb` |
| `DB_USER` | Postgres user | `postgres` |
| `DB_PASSWORD` | Postgres password | `postgres` |
| `DB_SSLMODE` | Postgres SSL mode | `disable` |
| `DB_AUTO_MIGRATE` | Run embedded migrations on startup | `true` |
| `SERVICE_NAME` | Log `source` field default | `application-logging-audit-module` |
| `ENABLE_CONSOLE_LOGGING` | Print logs to stdout | `true` |
| `ENABLE_DATABASE_LOGGING` | Persist logs to DB | `true` |
| `LOG_LEVEL` | Documented default level label | `INFO` |
| `ADMIN_USERNAME` | Seed admin username | `admin` |
| `ADMIN_PASSWORD` | Seed admin plain password | `12345678` |
| `ADMIN_API_KEY` | Static API key for admin routes | `super-secret-admin-key` |
| `JWT_SECRET` | HMAC secret for JWT | `change-this-secret` |
| `JWT_EXPIRY_HOURS` | JWT lifetime | `24` |
| `CORS_ORIGINS` | Comma-separated allowed origins | `http://localhost:5173` |
| `ENABLE_SWAGGER_UI` | Mount `/swagger/` routes | `false` |
| `TEST_DB_*` | Integration test DB overrides | See `BuildTestDatabaseURL()` |

Template: `.env.example` at repository root; frontend: `frontend/.env.example` (`VITE_API_URL`).

**Note:** `LOG_LEVEL` is loaded into config but not enforced as a filter in `LoggerService` at time of writing — service accepts all four levels when called.

---

## 7. Database and Migration System

### Connection logic

| Driver | Function | Details |
|--------|----------|---------|
| SQLite | `database.ConnectSQLite` | Creates parent dir; `modernc.org/sqlite`; `MaxOpenConns=1`; foreign keys + busy timeout |
| Postgres | `database.Connect` → `ConnectWithRetry` | `pgxpool`; 5 retries, 2s backoff; `MaxConns=10` |

### Migration approaches

1. **Go embedded (default on startup)** — SQL embedded in `database/embed.go`; `EnsureSchema` / `EnsureAdminSchema` (and SQLite variants) run when `DB_AUTO_MIGRATE=true` and tables are missing.
2. **Flyway CLI (optional)** — `make migrate-up` runs Flyway against local Postgres `loggerdb` using `database/migrations/`.

There is **no** separate Flyway schema history table managed by Go; Go migrations are idempotent checks via `TablesExist` / `AdminTableExists`.

### SQL file organization

| File | Purpose |
|------|---------|
| `V1_1__create_logging_audit_tables.sql` | Postgres: `application_logs`, `audit_events` |
| `V1_1_sqlite__create_logging_audit_tables.sql` | SQLite equivalents (TEXT/JSON as TEXT) |
| `V1_2__create_admin_users.sql` | Postgres: `admin_users` |
| `V1_2_sqlite__create_admin_users.sql` | SQLite: `admin_users` |
| `init.sql` | Flyway placeholder comment |

### Table: `application_logs` (PostgreSQL)

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID PRIMARY KEY | Unique log entry ID |
| `level` | VARCHAR(20) NOT NULL | INFO, WARNING, ERROR, DEBUG |
| `message` | TEXT NOT NULL | Human-readable message |
| `source` | VARCHAR(255) | Service/component name |
| `request_id` | VARCHAR(255) | Correlation ID for tracing |
| `user_id` | UUID NULL | Optional associated user |
| `error_code` | VARCHAR(100) | Optional error classification |
| `stack_trace` | TEXT | Error text / stack for ERROR level |
| `metadata` | JSONB | Arbitrary structured context (sanitized) |
| `created_at` | TIMESTAMPTZ | UTC timestamp |

**Indexes:** `level`, `created_at`, `user_id`, `request_id`.

SQLite stores UUIDs and metadata as TEXT; same logical columns.

### Table: `audit_events` (PostgreSQL)

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID PRIMARY KEY | Unique event ID |
| `user_id` | UUID NULL | Acting user |
| `username` | VARCHAR(255) | Username snapshot |
| `action` | VARCHAR(100) NOT NULL | LOGIN, UPDATE_RECORD, etc. |
| `resource_type` | VARCHAR(100) | Entity type (e.g. invoice) |
| `resource_id` | VARCHAR(255) | Entity identifier |
| `old_value` | JSONB | State before change (sanitized) |
| `new_value` | JSONB | State after change (sanitized) |
| `ip_address` | VARCHAR(100) | Client IP |
| `user_agent` | TEXT | Browser/client string |
| `request_id` | VARCHAR(255) | Correlation ID |
| `status` | VARCHAR(50) NOT NULL | SUCCESS or FAILURE |
| `metadata` | JSONB | Extra context |
| `created_at` | TIMESTAMPTZ | Event time |

**Indexes:** `user_id`, `action`, `resource_type`, `status`, `created_at`, `request_id`.

### Table: `admin_users`

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID PRIMARY KEY | Admin user ID |
| `username` | TEXT UNIQUE NOT NULL | Login name |
| `password_hash` | TEXT NOT NULL | bcrypt hash |
| `created_at` | TIMESTAMPTZ | Creation time |

### Design notes

- **JSONB (Postgres)** / **TEXT JSON (SQLite)** stores flexible metadata and audit snapshots.
- **Request IDs** link HTTP middleware logs, handlers, and audit events.
- **Parameterized SQL** (`$1`, `$2`, … in Postgres; `?` in SQLite) prevents SQL injection in repositories.

---

## 8. Logger Package Documentation

**Path:** `internal/logger/`

### LogEntry

**Purpose:** Domain model for one application log row.

**Fields:** `ID`, `Level`, `Message`, `Source`, `RequestID`, `UserID`, `ErrorCode`, `StackTrace`, `Metadata`, `CreatedAt`.

**Methods:** None (data struct).

**Used by:** `LoggerService`, `LogRepository` implementations, handlers, exporter.

**CBSE relevance:** Core entity in class diagram; maps 1:1 to `application_logs`.

---

### LogFilter

**Purpose:** Query criteria for listing logs.

**Fields:** `Level`, `RequestID`, `UserID`, `Source`, `From`, `To`, `Pagination`.

**Used by:** `PostgresLogRepository.Find`, `SQLiteLogRepository.Find`, `LogHandler`.

---

### Logger (interface)

**Purpose:** Public contract for application logging.

**Methods:**

| Method | Purpose | Parameters | Returns |
|--------|---------|------------|---------|
| `Info` | Write INFO log | `ctx`, `message`, `metadata` | `error` |
| `Warn` | Write WARNING log | same | `error` |
| `Error` | Write ERROR log | `ctx`, `message`, `err`, `metadata` | `error` |
| `Debug` | Write DEBUG log | `ctx`, `message`, `metadata` | `error` |

**Flow:** Implemented by `LoggerService` → validation → sanitize → optional console → repository insert.

**Error handling:** Invalid level returns `fmt.Errorf`; repository errors propagate.

---

### LogRepository (interface)

**Purpose:** Persistence abstraction.

**Methods:** `Insert`, `Find`, `FindByID`.

**Implementations:** `PostgresLogRepository`, `SQLiteLogRepository`.

---

### LoggerService

**Purpose:** Default `Logger` implementation with sanitization and toggles.

**Fields:** `repo`, `sanitizer`, `source`, `consoleEnabled`, `dbEnabled`.

**Methods:**

| Method | Purpose |
|--------|---------|
| `NewService(repo, source, consoleEnabled, dbEnabled)` | Constructor |
| `Info`, `Warn`, `Debug`, `Error` | Delegate to `log()` |
| `LogHTTP` | Middleware helper; calls `Info` |
| `WithRequestID`, `WithUserID` | Context helpers for correlation |

**Method: `log` (internal)**

- **Purpose:** Validate level, sanitize metadata, build `LogEntry`, optional stdout, conditional DB insert.
- **Parameters:** level, message, stack, metadata, err.
- **Return:** `error` from repository or nil if DB disabled.
- **Error handling:** Rejects unknown levels; does not panic on console failures.

**Used by:** `internal/app`, `HTTPLogger` middleware, demo handlers.

**CBSE relevance:** Service layer in layered architecture; implements interface segregation.

---

### PostgresLogRepository / SQLiteLogRepository

**Purpose:** Driver-specific SQL for logs.

**Methods:** `Insert` (parameterized INSERT), `Find` (dynamic WHERE + pagination), `FindByID`.

**Error handling:** Wrap errors with context (`insert log`, `marshal metadata`).

---

### SensitiveDataSanitizer

**Purpose:** Remove secrets from metadata and audit JSON maps before storage.

**Fields:** None (stateless).

**Methods:**

| Method | Purpose |
|--------|---------|
| `NewSanitizer()` | Constructor |
| `SanitizeMap(data)` | Recursively mask sensitive keys and card-like strings |

**Sensitive keys (substring match, case-insensitive):** `password`, `passwd`, `token`, `authorization`, `auth`, `secret`, `api_key`, `apikey`, `cvv`, `card`, `credit_card`, `credit card`, `pan`.

**Card numbers:** Regex `\b(?:\d[ -]*?){13,19}\b` in string values → `***MASKED***`.

**Used by:** `LoggerService`, `AuditService`.

---

### Log creation flow

```text
Application calls LoggerService.Info/Error/...
  → level validated (ValidLevels)
  → metadata sanitized (SensitiveDataSanitizer)
  → LogEntry created (UUID, timestamps, optional ctx request_id/user_id)
  → optional console printf
  → if dbEnabled: LogRepository.Insert → application_logs
```

---

## 9. Audit Package Documentation

**Path:** `internal/audit/`

### AuditEvent

**Purpose:** One auditable user action record.

**Fields:** `ID`, `UserID`, `Username`, `Action`, `ResourceType`, `ResourceID`, `OldValue`, `NewValue`, `IPAddress`, `UserAgent`, `RequestID`, `Status`, `Metadata`, `CreatedAt`.

**Valid actions:** `LOGIN`, `LOGOUT`, `CREATE_RECORD`, `UPDATE_RECORD`, `DELETE_RECORD`, `DOWNLOAD_FILE`, `FAILED_LOGIN`, `PERMISSION_DENIED`.

**Valid statuses:** `SUCCESS`, `FAILURE`.

---

### AuditFilter

**Purpose:** Query criteria for audit listing.

**Fields:** `UserID`, `Username`, `Action`, `ResourceType`, `Status`, `RequestID`, `From`, `To`, `Pagination`.

---

### Auditor (interface)

**Methods:** `Record(ctx, event)`, `Find(ctx, filter)`.

**Note:** `AuditService` implements `Record`; `Find` delegates to repository. Handler uses `Auditor` for writes and `AuditRepository` for reads/exports.

---

### AuditRepository (interface)

**Methods:** `Insert`, `Find`, `FindByID`.

**Implementations:** `PostgresAuditRepository`, `SQLiteAuditRepository`.

---

### AuditService

**Purpose:** Validate and sanitize audit events before insert.

**Method: `Record`**

- **Purpose:** Enforce required action/status and whitelist values; assign ID/timestamp; sanitize `OldValue`, `NewValue`, `Metadata`.
- **Parameters:** `ctx`, `AuditEvent`.
- **Returns:** `error` on validation or DB failure.
- **Validation:** Empty action/status rejected; invalid action/status rejected.
- **Flow:** Normalize → sanitize maps → `repo.Insert`.

**Method: `Find`**

- **Purpose:** Pass-through to repository.

**Used by:** Demo handlers, `examples/basic_usage.go`.

---

### Audit flow

```text
User action in app → AuditService.Record
  → action/status validated
  → old_value, new_value, metadata sanitized
  → AuditRepository.Insert → audit_events table
```

---

## 10. Middleware Documentation

**Path:** `internal/middleware/`

### HTTPLogger

**Purpose:** API-gateway-style access logging for every HTTP request.

**Behavior:**

1. Read or generate `X-Request-ID`; set response header.
2. Wrap writer with `responseWriter` to capture status (default 200 until `WriteHeader`).
3. Call next handler; measure latency.
4. Build metadata: method, path, status_code, latency_ms, ip_address, user_agent.
5. `LogHTTP` on logger — **errors ignored** (`_ =`) so logging never blocks or fails the request.

**CBSE relevance:** Cross-cutting concern; separates HTTP observability from handlers.

---

### responseWriter

**Purpose:** Embed `http.ResponseWriter` and record status code.

**Methods:** `WriteHeader` (stores code), `Status()` getter.

---

### AdminAuth

**Purpose:** Protect `/admin/*` routes except `POST /admin/login`.

**Authentication (any one succeeds):**

1. Header `X-API-Key: <ADMIN_API_KEY>`
2. Header `Authorization: Bearer <JWT>`
3. Header `Authorization: ApiKey <ADMIN_API_KEY>`

**Failure:** `401` JSON `{"error":"unauthorized"}` via `common.WriteError`.

---

### CORS

**Purpose:** Allow browser frontend to call API from configured origins (`CORS_ORIGINS`).

**Behavior:** Sets `Access-Control-Allow-Origin` when Origin matches; handles `OPTIONS` with 204.

---

### Request ID handling

- Header name: `X-Request-ID` (`RequestIDHeader` constant).
- Generated with `uuid.New()` if absent.

---

## 11. Handler/API Documentation

**Path:** `internal/handler/`

### HealthHandler

| Route | Method | Auth | Response |
|-------|--------|------|----------|
| `/health` | GET | No | `{"status":"ok","database":"up"}` or degraded if ping fails |

---

### LogHandler

| Route | Method | Body / Query | Service |
|-------|--------|--------------|---------|
| `/demo/log-info` | POST | `{message?, metadata?}` | `Logger.Info` |
| `/demo/log-error` | POST | same | `Logger.Error` (simulated error) |
| `/admin/logs` | GET | `level`, `request_id`, `user_id`, `source`, `from`, `to`, `page`, `limit` | `LogRepository.Find` |
| `/admin/logs/{id}` | GET | — | `LogRepository.FindByID` |
| `/admin/logs/export` | GET | same filters + `format=json\|csv` | Find + exporter |

---

### AuditHandler

| Route | Method | Body / Query | Service |
|-------|--------|--------------|---------|
| `/demo/audit-login` | POST | `user_id`, `username`, `ip_address` | `Auditor.Record` (LOGIN) |
| `/demo/audit-update` | POST | `user_id`, `username`, `resource_*`, `old_value`, `new_value` | `Auditor.Record` (UPDATE_RECORD) |
| `/admin/audit-events` | GET | `username`, `action`, `resource_type`, `status`, `request_id`, `user_id`, `from`, `to`, `page`, `limit` | `AuditRepository.Find` |
| `/admin/audit-events/{id}` | GET | — | `FindByID` |
| `/admin/audit-events/export` | GET | filters + `format` | Find + exporter |

---

### AuthHandler

| Route | Method | Body | Response |
|-------|--------|------|----------|
| `/admin/login` | POST | `{username, password}` | `{"token":"<JWT>"}` or 401 |

Uses `adminauth.Repository.FindByUsername`, `CheckPassword`, `TokenService.IssueToken`.

---

### API summary table

| Method | Route | Purpose | Auth Required |
|--------|-------|---------|---------------|
| GET | `/health` | Health + DB ping | No |
| POST | `/demo/log-info` | Demo INFO log | No |
| POST | `/demo/log-error` | Demo ERROR log | No |
| POST | `/demo/audit-login` | Demo LOGIN audit | No |
| POST | `/demo/audit-update` | Demo UPDATE audit | No |
| POST | `/admin/login` | Admin JWT login | No |
| GET | `/admin/logs` | List logs | Yes |
| GET | `/admin/logs/export` | Export logs | Yes |
| GET | `/admin/logs/{id}` | Get log by ID | Yes |
| GET | `/admin/audit-events` | List audit events | Yes |
| GET | `/admin/audit-events/export` | Export audit | Yes |
| GET | `/admin/audit-events/{id}` | Get audit by ID | Yes |
| GET | `/swagger/`, `/swagger/openapi.yaml` | API docs UI/spec | No (if `ENABLE_SWAGGER_UI=true`) |

**Pagination defaults:** page `1`, limit `20`, max limit `100` (`internal/common/pagination.go`).

**Date filters:** `from` / `to` query params parsed by `common.ParseQueryTime` (ISO-8601).

---

## 12. Admin Authentication System

### Login flow

1. Client `POST /admin/login` with username/password.
2. `AuthHandler` loads user from `admin_users` by username.
3. `adminauth.CheckPassword` compares bcrypt hash.
4. `TokenService.IssueToken` returns HS256 JWT (expiry from `JWT_EXPIRY_HOURS`).

### Credentials source

- On first startup, `adminauth.SeedDefaultAdmin` creates user from `ADMIN_USERNAME` / `ADMIN_PASSWORD` if not exists.
- Passwords stored as bcrypt hashes (`adminauth.HashPassword`).

### Token validation

`AdminAuth` middleware calls `TokenService.ValidateToken` for Bearer tokens.

### Frontend storage

- Key: `admin_token` in `localStorage` (`frontend/src/services/api.js`).
- Sent as `Authorization: Bearer <token>` on admin API calls.
- On `401`, token cleared and redirect to `/login`.

### Protected routes

All `/admin/*` except `POST /admin/login` (chi subgroup with `adminAuth.Middleware`).

### Alternative auth

Static `ADMIN_API_KEY` via `X-API-Key` or `Authorization: ApiKey <key>` — useful for scripts without login.

### Limitations (demo system)

- Single default admin user seed; no RBAC roles.
- JWT secret must be strong in production; default is insecure.
- No HTTPS enforcement in app; use reverse proxy in production.
- No refresh tokens or session revocation list.
- Frontend `ProtectedRoute` only checks token presence, not expiry until API returns 401.

**Production should use:** HTTPS, strong secrets, password policies, RBAC, optional OAuth2/OIDC, rate limiting on login.

---

## 13. Frontend Documentation

### Technology

React 18 + Vite 5 + react-router-dom 6. Dev server port **5173** (`vite.config.js`). Production build served by nginx in Docker (port **5173** mapped to container 80).

### Folder structure

| Path | Purpose |
|------|---------|
| `src/main.jsx` | React root + `BrowserRouter` |
| `src/App.jsx` | Route definitions |
| `src/pages/LoginPage.jsx` | Admin login form |
| `src/pages/DashboardPage.jsx` | Landing after login |
| `src/pages/LogsPage.jsx` | Log filters, table, pagination, export |
| `src/pages/AuditPage.jsx` | Audit filters, table, pagination, export |
| `src/components/Navbar.jsx` | Navigation + logout |
| `src/components/ProtectedRoute.jsx` | Redirect if no token |
| `src/components/LogTable.jsx` | Log rows display |
| `src/components/AuditTable.jsx` | Audit rows display |
| `src/services/api.js` | HTTP client, auth, export blobs |
| `src/styles/app.css` | Layout and forms |

### Routes

| Route | Page | Protected |
|-------|------|-----------|
| `/login` | LoginPage | No |
| `/dashboard` | DashboardPage | Yes |
| `/logs` | LogsPage | Yes |
| `/audit` | AuditPage | Yes |
| `/`, `*` | Redirect to `/dashboard` | — |

### Frontend flow

```text
Admin opens frontend → /login (or redirected if no token)
  → POST /admin/login → store JWT in localStorage
  → Protected pages call GET /admin/logs or /admin/audit-events
  → Export buttons call /admin/.../export with format=csv|json
  → Logout clears token → /login
```

### Filtering and pagination

- Logs: level, from, to; page/limit sent as query params.
- Audit: action, status, from, to; same pagination pattern.

### Frontend component table

| Frontend File/Component | Purpose |
|-------------------------|---------|
| `LoginPage.jsx` | Username/password form; prefills `admin` |
| `DashboardPage.jsx` | Links to logs and audit sections |
| `LogsPage.jsx` | Filter form, pagination, export JSON/CSV |
| `AuditPage.jsx` | Audit filters, pagination, export |
| `ProtectedRoute.jsx` | Client-side guard using `getToken()` |
| `Navbar.jsx` | Nav links; logout clears token |
| `LogTable.jsx` | Renders log list columns |
| `AuditTable.jsx` | Renders audit list columns |
| `api.js` | `login`, `getLogs`, `getAuditEvents`, `exportLogs`, `exportAuditEvents` |

---

## 14. Export System

### Backend (`internal/exporter/`)

| Function | Output |
|----------|--------|
| `ExportLogsJSON` | Indented JSON array of `LogEntry` |
| `ExportLogsCSV` | CSV with header row |
| `ExportAuditJSON` | Indented JSON array of `AuditEvent` |
| `ExportAuditCSV` | CSV with header row |

### Routes

- `GET /admin/logs/export?format=json|csv` (+ same filters as list)
- `GET /admin/audit-events/export?format=json|csv`

Default format: **json** if `format` omitted.

### Metadata handling

JSON/CSV serializers use stored (already sanitized) values from database.

### Frontend

`exportLogs` / `exportAuditEvents` download blob via temporary `<a>` element; default format **csv** in UI params.

---

## 15. Security Design

| Decision | Implementation |
|----------|----------------|
| Sensitive data sanitization | `SensitiveDataSanitizer` on all metadata and audit old/new values |
| Password storage | bcrypt hashes in `admin_users` |
| Admin route protection | `AdminAuth` middleware + JWT/API key |
| SQL injection prevention | Parameterized queries in repositories |
| Request traceability | `X-Request-ID` on requests and in log metadata |
| Logging must not break app | HTTP logger ignores `LogHTTP` errors; Recoverer middleware on panics |
| CORS | Restricts browser origins to configured list |

### Sensitive fields masked

Keys containing: password, passwd, token, authorization, auth, secret, api_key, apikey, cvv, card, credit_card, credit card, pan. Card-number patterns in strings.

### Limitations / future improvements

- No field-level encryption at rest
- No audit log immutability / WORM storage
- Demo credentials in default env
- **Not implemented yet:** rate limiting, MFA, OWASP security headers middleware

---

## 16. Error Handling Strategy

| Layer | Behavior |
|-------|----------|
| Service | Validation errors as `fmt.Errorf` (invalid level, action, status) |
| Repository | Wrapped DB errors (`insert log`, etc.) |
| Middleware logging | Errors discarded intentionally |
| Handler | `common.WriteError` → JSON `{"error":"message"}` with appropriate HTTP status |
| DB connection | `app.New` fails fast; health reports `database: down` |
| Frontend | try/catch; display `error` state; 401 → redirect login |

**Principle:** Logging/audit failures in middleware must not crash business handlers; demo handlers return 500 if `Record`/`Info` fails.

---

## 17. Reusability and Component Independence

### Public interfaces (within module)

- `logger.Logger`, `logger.LogRepository`
- `audit.Auditor`, `audit.AuditRepository`

### Constructors

- `logger.NewService(repo, source, console, db)`
- `audit.NewService(repo)`
- `logger.NewPostgresRepository` / `NewSQLiteRepository`
- `audit.NewPostgresRepository` / `NewSQLiteRepository`

### Configuration-based behavior

Console vs DB logging, driver selection, auto-migrate, service name as log source.

### Separation from business logic

Handlers depend on interfaces, not SQL. Sanitization centralized in services.

### Import note

Packages live under `internal/` — **not importable** by external Go modules per Go convention. Reuse options: copy module, monorepo, or fork.

### How another developer can use this component in 5 minutes

1. Copy `.env.example` to `.env` (SQLite default).
2. Run `go run ./examples/basic_usage.go`.
3. Inspect `application_logs` and `audit_events` in `./data/logger.db`.

Or run full stack: `make docker-up` → open `http://localhost:5173`, login `admin` / `12345678`.

**Code sample** (from `examples/basic_usage.go`):

```go
loggerSvc := logger.NewService(logRepo, cfg.ServiceName, true, true)
auditSvc := audit.NewService(auditRepo)

_ = loggerSvc.Info(ctx, "User dashboard loaded", map[string]any{"component": "dashboard"})
_ = auditSvc.Record(ctx, audit.AuditEvent{
    Username: "demo_user", Action: "UPDATE_RECORD",
    ResourceType: "invoice", ResourceID: "1001", Status: "SUCCESS",
})
```

---

## 18. UML Diagram Explanation

| Diagram | File | Shows | CBSE value |
|---------|------|-------|------------|
| Class | `docs/diagrams/class_diagram.puml` | Logger/Auditor interfaces, services, repositories, handlers | Static structure, interface segregation |
| Sequence | `docs/diagrams/sequence_diagram.puml` | Demo audit update: client → chi → handler → service → repository → DB | Dynamic interaction |
| State | `docs/diagrams/state_diagram.puml` | Entry lifecycle: Created → Validated → Stored | Data lifecycle rules |
| Deployment | `docs/diagrams/deployment_diagram.puml` | Browser, React UI, Go server, PostgreSQL/SQLite, Docker | Physical deployment |

**PNG exports:** Run `make diagrams` → output in `docs/diagrams/generated/`. PNGs may not be committed; source `.puml` files are authoritative.

**Note:** Sequence diagram may reference Postgres; implementation also supports SQLite — treat as logical persistence node.

---

## 19. Unit Tests Documentation

**Command:**

```bash
go test ./...
```

**Strategy:** Hand-written mocks implementing interfaces; in-memory SQLite for repository tests; `httptest` for handlers. **No real database required** for default unit test run.

| Test File | What It Tests |
|-----------|---------------|
| `internal/config/config_test.go` | URL builders, `Load` defaults, drivers, Swagger flag |
| `internal/logger/service_test.go` | Levels, sanitization, invalid level, repo errors |
| `internal/logger/sanitizer_test.go` | Key and card masking |
| `internal/logger/repository_sqlite_test.go` | SQLite repo insert/find |
| `internal/audit/service_test.go` | Validation, sanitization, repo errors |
| `internal/audit/repository_sqlite_test.go` | SQLite audit repo |
| `internal/handler/log_handler_test.go` | List logs with mocks |
| `internal/handler/audit_handler_test.go` | Demo login, get by ID |
| `internal/handler/auth_handler_test.go` | Login success/failure |
| `internal/middleware/http_logger_test.go` | Status, request ID, metadata capture |
| `internal/middleware/admin_auth_test.go` | Bearer, X-API-Key, ApiKey, unauthorized |
| `internal/exporter/json_exporter_test.go` | JSON round-trip |
| `internal/exporter/csv_exporter_test.go` | CSV header and row |
| `internal/adminauth/password_test.go` | bcrypt hash/check |
| `internal/swagger/swagger_test.go` | Embedded spec and UI routes |

---

## 20. Integration Tests Documentation

**Command:**

```bash
go test -tags=integration ./...
```

**Requirements:** PostgreSQL reachable at URL from `BuildTestDatabaseURL()` (`TEST_DB_*` or `DB_*`, default database `loggerdb_test`). Tests **skip** if connection fails. **No testcontainers.**

**Setup:** `database.SetupTestPool` connects, runs `EnsureSchema`, truncates tables between tests.

| Integration Test File | What It Tests |
|----------------------|---------------|
| `internal/database/postgres_integration_test.go` | Connect, ping |
| `internal/database/migrate_integration_test.go` | Tables exist; idempotent EnsureSchema |
| `internal/logger/repository_integration_test.go` | Postgres log insert/find |
| `internal/audit/repository_integration_test.go` | Postgres audit insert/find |
| `internal/handler/admin_api_integration_test.go` | Full app: 401 without auth, login, demo data, list endpoints |

---

## 21. Makefile and Run Commands

| Command | Purpose |
|---------|---------|
| `make run` | Sync swagger + `go run ./cmd/server` |
| `make build` | Build `bin/server` |
| `make tidy` | `go mod tidy` |
| `make test` | Unit tests |
| `make test-integration` | Integration tests (`-tags=integration`) |
| `make test-all` | Both |
| `make coverage` | `coverage.out` + `coverage.html` |
| `make migrate-up` | Flyway on local Postgres |
| `make diagrams` | PlantUML → PNG in `docs/diagrams/generated/` |
| `make docker-up` | Docker Compose (SQLite app + frontend) |
| `make docker-up-postgres` | Compose with Postgres overlay |
| `make frontend-install` | `npm install` in frontend |
| `make frontend-dev` | Vite dev server |
| `make frontend-build` | Production frontend build |
| `make sync-swagger` | Copy `docs/swagger.yaml` to embedded path |

**Manual:**

```bash
go run ./cmd/server
go test ./...
```

---

## 22. API Usage Examples

Assume server on `http://localhost:8080`.

### Create an info log

```bash
curl -s -X POST http://localhost:8080/demo/log-info \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello from curl","metadata":{"component":"demo"}}'
```

### Create an error log

```bash
curl -s -X POST http://localhost:8080/demo/log-error \
  -H "Content-Type: application/json" \
  -d '{"message":"Something failed","metadata":{"service":"payment"}}'
```

### Record audit login

```bash
curl -s -X POST http://localhost:8080/demo/audit-login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","ip_address":"192.168.1.10"}'
```

### Admin login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}' | jq -r .token)
echo "$TOKEN"
```

### Fetch logs

```bash
curl -s "http://localhost:8080/admin/logs?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Fetch audit events

```bash
curl -s "http://localhost:8080/admin/audit-events?action=LOGIN&page=1" \
  -H "Authorization: Bearer $TOKEN"
```

### API key alternative

```bash
curl -s "http://localhost:8080/admin/logs" \
  -H "X-API-Key: super-secret-admin-key"
```

---

## 23. Deployment Requirements

| Requirement | Version / notes |
|-------------|-----------------|
| Go | 1.22+ (`go.mod`) |
| PostgreSQL | Optional; 14+ recommended when `DB_DRIVER=postgres` |
| SQLite | Default; no separate server |
| Node.js | 18+ for frontend build (Vite 5) |
| Flyway | Optional; only for `make migrate-up` |

### Environment

Copy `.env.example` → `.env`. For frontend: `frontend/.env` with `VITE_API_URL`.

### Database setup

- **Automatic:** `DB_AUTO_MIGRATE=true` on app start (default).
- **Manual Flyway:** `make migrate-up` (Postgres only).

### Running backend

```bash
make run
# or
go run ./cmd/server
```

### Running frontend

```bash
make frontend-dev   # http://localhost:5173
```

### Ports

| Service | Port |
|---------|------|
| Backend | 8080 (`APP_PORT`) |
| Frontend dev | 5173 |
| Frontend Docker | 5173 → nginx 80 |
| Postgres (compose overlay) | 5432 |

---

## 24. Restrictions and Limitations

- **Authentication:** Demo-grade JWT + single seeded admin; not enterprise IAM.
- **RBAC:** Not implemented — any authenticated admin sees all logs/events.
- **Log retention:** No automatic purge or archival policies.
- **Async logging:** Synchronous DB writes only; no buffered queue.
- **Distributed tracing:** No OpenTelemetry/Jaeger integration.
- **LOG_LEVEL env:** Loaded but not used to filter outgoing logs in service.
- **Frontend:** Demo UI; no user management screens.
- **Database:** Admin APIs require schema tables — run app with auto-migrate or apply SQL first.
- **Integration tests:** Require Postgres; SQLite not used in `-tags=integration` suite.
- **External import:** `internal/` packages cannot be imported by other modules without restructuring.

---

## 25. Similar and Competing Components

| Component | Language/Platform | Main Use | Audit Trail Support | Database Storage | Admin UI | Difference from This Project |
|-----------|-------------------|----------|---------------------|------------------|----------|-------------------------------|
| Logrus | Go | App logging | No | No (hooks/files) | No | Logger only; no audit schema or REST admin |
| Zap | Go | High-perf logging | No | Optional sinks | No | Speed-focused; no compliance audit model |
| Zerolog | Go | JSON logging | No | Optional | No | Zero-allocation logger; no HTTP/admin layer |
| Log4j | Java | Enterprise logging | No (separate) | Appenders | No | Java ecosystem; not embeddable Go module |
| Winston | Node.js | Flexible transports | Custom only | Via transports | No | Node-specific; no unified audit types |
| Audit.NET | .NET | Change tracking | Yes | DB/NoSQL | Limited | .NET only; less application debug logging |
| ELK Stack | Infra | Log aggregation/search | Search only | Elasticsearch | Kibana | Infrastructure stack; not a reusable Go library |

### How this project differs

- Combines **application logging** and **audit trail** in one component
- **SQLite or PostgreSQL** persistence with embedded migrations
- **Admin REST API** + **React UI** + optional Swagger
- **Security sanitization** built into services
- **CBSE documentation**, PlantUML diagrams, tests, and competing analysis included
- Designed as an **educational reusable component** with clear interfaces and layers

---

## 26. Changelog and License

### Current version

Documented releases in `docs/changelog.md`:

- **Unreleased:** JWT admin auth, React frontend, CORS, date filters, Docker frontend
- **v1.1.0:** SQLite default, dual repositories, Docker SQLite stack
- **v1.0.0:** Initial logging, audit, middleware, admin API, Flyway SQL, Docker Postgres, OpenAPI, CBSE docs

### License

**MIT License** — full text in [`docs/license.md`](./license.md).

**Not implemented at repo root:** `LICENSE` and `CHANGELOG` files — recommended to add root copies for GitHub conventions.

---

## 27. Final CBSE Evaluation Summary

This project satisfies CBSE component requirements because it:

1. **Defines a reusable component** with a clear name, purpose, and boundaries (logging + audit, not a full unrelated app).
2. **Exposes explicit interfaces** (`Logger`, `Auditor`, repositories) demonstrating interface-based design.
3. **Documents internal architecture** in layers: handlers → services → repositories → database.
4. **Provides user documentation** (`docs/user_guide.md`, frontend README, this document).
5. **Includes installation and deployment instructions** (Makefile, Docker Compose, env templates).
6. **Supports UML** with class, sequence, state, and deployment PlantUML diagrams.
7. **Implements automated tests** — unit tests without DB; integration tests against Postgres.
8. **Analyzes competing components** (Logrus, Zap, ELK, etc.) and states differentiation.
9. **Solves a real business problem** — operational debugging and accountability/compliance auditing.
10. **Ships working artifacts** — HTTP API, admin UI, migrations, export, sanitization, and examples.

The module is suitable for inclusion in a university CBSE report, technical README, and oral presentation to lecturers, with this file serving as the single comprehensive reference for classes, flows, APIs, and deployment.

---

*Generated from repository source at `application-logging-audit-module`. For API field details see `docs/swagger.yaml`.*
