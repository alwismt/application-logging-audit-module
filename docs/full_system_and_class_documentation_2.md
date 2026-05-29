# Application Logging & Audit Trail Module - Full System and Class Documentation

This document describes the current implementation of the standalone Go project **Application Logging & Audit Trail Module**. It is written for the CBSE report and is based on the actual repository contents.

Important accuracy note: the file `labs_report_en (3).docx` was not found in the repository. The CBSE mapping below follows the requirement list from the task and the existing project documents in `docs/`.

## 1. Project Overview

The project name is **Application Logging & Audit Trail Module**.

The purpose of the component is to give Go developers a reusable module for recording technical application logs and business/security audit trail events. It can be plugged into another application to support debugging, security monitoring, compliance reporting, and operational visibility.

The business problem solved by the component is that many applications need to know both what went wrong technically and what users did inside the system. A normal application log can explain technical failures such as database errors, service timeouts, and HTTP failures. An audit trail can explain user activity such as logins, failed logins, data updates, downloads, and permission failures.

Logging and audit trails are important because they help teams:

- Debug errors after they happen.
- Trace requests by request ID.
- Investigate suspicious user behavior.
- Show compliance evidence about who changed data and when.
- Monitor application health and failures.

The module can be reused in other Go projects by wiring the service interfaces and repositories into an existing application. The core packages are separated from HTTP handlers, so another application can call `LoggerService` and `AuditService` directly without using the demo API.

Main technologies used:

- Go 1.22 module.
- `net/http` with `github.com/go-chi/chi/v5` router.
- PostgreSQL using `github.com/jackc/pgx/v5/pgxpool`.
- SQLite using `modernc.org/sqlite` for local/default persistence.
- SQL migrations embedded into Go with `go:embed`.
- JWT authentication using `github.com/golang-jwt/jwt/v5`.
- Bcrypt password hashing using `golang.org/x/crypto/bcrypt`.
- React 18 and Vite for the admin frontend.
- Docker and Docker Compose.
- PlantUML diagrams.
- Go unit tests and integration tests.

Main users of the component:

- Backend developers who need logging and audit functionality.
- Administrators who use the admin frontend/API to view logs and audit events.
- Security or compliance reviewers who need a trace of user activity.
- Lecturers or evaluators reviewing the CBSE component design.

Difference between application logs and audit trail events:

| Type | Main Question Answered | Example | Typical Audience |
|---|---|---|---|
| Application logs | What happened technically inside the application? | `ERROR: payment service timeout` | Developers, operators |
| Audit trail events | Who did what action to which resource and was it successful? | `alice UPDATE_RECORD invoice/1001 SUCCESS` | Security, compliance, administrators |

## 2. CBSE Requirement Mapping

| CBSE Requirement | Where It Is Implemented | Explanation |
|---|---|---|
| Component description | `README.md`, `docs/component_description.md`, this document | Describes the component name, purpose, features, users, and business value. |
| Provided functions | `internal/logger/interface.go`, `internal/audit/interface.go`, `internal/handler/*` | Logger provides `Info`, `Warn`, `Error`, `Debug`. Audit provides `Record` and `Find`. Admin/demo APIs expose the component through HTTP. |
| Purpose | `README.md`, `docs/component_description.md` | The project records technical errors and user actions for debugging, monitoring, security, and compliance. |
| Business problems solved | `README.md`, `docs/component_description.md`, this document | Solves missing production diagnostics, missing user action traceability, and unsafe logging of sensitive data. |
| Intended use | `examples/basic_usage.go`, `docs/user_guide.md`, this document | Intended for Go applications that need embedded logging and audit storage. |
| Restrictions | `README.md`, `docs/component_description.md`, section 24 | Restrictions include synchronous writes, simple admin auth, PostgreSQL/SQLite persistence, and internal package visibility. |
| Internal architecture | `internal/app`, `internal/logger`, `internal/audit`, `internal/database`, `docs/architecture.md` | Layered structure: router, middleware, handlers, services, repositories, database. |
| Class diagram support | `docs/diagrams/class_diagram.puml`, generated PNG | Shows services, repositories, interfaces, sanitizer, exporters, and core data objects. |
| Sequence diagram support | `docs/diagrams/sequence_diagram.puml`, generated PNG | Shows request flow for recording an audit update event. |
| State diagram support | `docs/diagrams/state_diagram.puml`, generated PNG | Shows lifecycle from created to validated to stored, including validation/storage failure states. |
| Deployment diagram support | `docs/diagrams/deployment_diagram.puml`, generated PNG | Shows browser/admin UI, Go HTTP server, module, migration logic, and PostgreSQL deployment. |
| User documentation | `README.md`, `docs/user_guide.md`, `frontend/README.md` | Explains setup, running, Docker, frontend, and demo credentials. |
| API documentation | `docs/api_documentation.md`, `docs/swagger.yaml`, `internal/swagger/openapi.yaml` | Documents HTTP routes, request/response patterns, and optional Swagger UI. |
| Usage examples | `examples/basic_usage.go`, README curl examples, section 22 | Shows direct service usage and REST API usage. |
| Installation instructions | `README.md`, `Makefile`, `frontend/README.md` | Includes Docker, local Go, frontend install, and migration commands. |
| Dependencies | `go.mod`, `frontend/package.json`, Dockerfiles | Lists Go, React, Vite, pgx, chi, JWT, bcrypt, SQLite, Docker, PostgreSQL. |
| Deployment requirements | `Dockerfile`, `docker-compose.yml`, `docker-compose.postgres.yml`, section 23 | Backend container, frontend container, SQLite volume by default, PostgreSQL overlay. |
| Changelog | `docs/changelog.md` | Contains unreleased changes, v1.1.0 SQLite default, and v1.0.0 initial release. |
| License | `docs/license.md` | MIT license text is present. There is no root-level `LICENSE` file; recommended improvement: add one. |
| Competing component analysis | `docs/competing_components.md`, section 25 | Compares Logrus, Zap, Zerolog, Log4j, Winston, Audit.NET, and ELK Stack. |

## 3. System Architecture

The project follows a layered architecture.

| Layer | Main Files | Responsibility |
|---|---|---|
| Backend layer | `cmd/server/main.go`, `internal/app/app.go` | Starts the application, loads config, initializes dependencies, runs HTTP server. |
| Component core layer | `internal/logger`, `internal/audit` | Provides reusable services, interfaces, validation, sanitization, and domain types. |
| Database/persistence layer | `internal/database`, repository files | Connects to SQLite/PostgreSQL, ensures schema, inserts and queries records. |
| Middleware layer | `internal/middleware` | Adds CORS, request logging, request IDs, status capture, admin auth. |
| Handler/API layer | `internal/handler` | Parses HTTP input, calls services/repositories, writes JSON/CSV responses. |
| Frontend/admin UI layer | `frontend/src` | React UI for login, dashboard, logs, audit events, filters, pagination, export. |
| Authentication layer | `internal/adminauth`, `internal/middleware/admin_auth.go` | Seeds admin user, hashes passwords, issues JWTs, validates JWT/API key. |
| Testing layer | `internal/**/*_test.go` | Unit tests with mocks and SQLite smoke tests; integration tests with PostgreSQL build tag. |

Main request flow:

```text
Frontend/Admin Client -> Backend API -> Authentication Middleware -> Handler -> Logger/Audit Service -> Repository -> PostgreSQL or SQLite
```

API gateway style request logging:

1. Every HTTP request enters the Chi router.
2. `HTTPLogger.Middleware` checks for `X-Request-ID`; if missing, it creates a UUID request ID.
3. The middleware wraps the response writer with `responseWriter`.
4. The handler runs.
5. After the handler finishes, middleware captures status code, latency, method, path, IP address, user agent, and request ID.
6. It calls `LoggerService.LogHTTP`.
7. Logging errors are ignored with `_ = h.logger.LogHTTP(...)`, so a logging failure does not break the original HTTP request.

## 4. Folder Structure Explanation

Actual high-level folder structure:

```text
cmd/server/                 Backend entry point
database/                   Embedded SQL migrations
database/migrations/        PostgreSQL and SQLite SQL migration files
docs/                       CBSE/user/API documentation
docs/diagrams/              PlantUML source diagrams and generated PNGs
examples/                   Direct component usage example
frontend/                   React + Vite admin frontend
internal/adminauth/         Admin user, bcrypt, JWT, auth repositories
internal/app/               Application wiring and route registration
internal/audit/             Audit event domain, service, repositories
internal/common/            Pagination, JSON responses, query time parsing
internal/config/            Environment configuration loading
internal/database/          SQLite/PostgreSQL connections and migrations
internal/exporter/          JSON and CSV export functions
internal/handler/           HTTP handlers
internal/logger/            Logging domain, service, repositories, sanitizer
internal/middleware/        CORS, request logging, response writer, admin auth
internal/swagger/           Embedded OpenAPI spec and optional Swagger UI
```

Folder purposes:

| Folder | Purpose |
|---|---|
| `cmd/server` | Contains `main.go`, the process entry point. |
| `internal/app` | Builds the app, chooses database driver, creates repositories/services/middleware, registers routes, starts/shuts down HTTP server. |
| `internal/config` | Loads `.env` and environment variables into `Config`. |
| `internal/database` | Connects to PostgreSQL or SQLite, pings DB, checks tables, runs embedded migrations. |
| `internal/logger` | Implements application logging model, service, repository interface, PostgreSQL repository, SQLite repository, and sanitizer. |
| `internal/audit` | Implements audit event model, service, repository interface, PostgreSQL repository, and SQLite repository. |
| `internal/middleware` | Implements request logging, response status capture, CORS, and admin route authentication. |
| `internal/handler` | Implements health, demo logging, demo audit, admin login, admin list/get/export APIs. |
| `internal/exporter` | Exports logs and audit events as formatted JSON or CSV. |
| `database/migrations` | Contains Flyway-style SQL files for PostgreSQL and SQLite variants. |
| `docs` | Contains CBSE documentation, API docs, architecture, changelog, license, competing analysis. |
| `docs/diagrams` | Contains PlantUML class, sequence, state, and deployment diagrams. |
| `examples` | Contains `basic_usage.go`, showing direct service integration. |
| `frontend` | Contains React/Vite admin UI, Dockerfile, Nginx config, and build output. |
| Test files | Co-located in `internal/<package>/*_test.go`; integration tests use `//go:build integration`. |

## 5. Backend Entry Point

Entry point: `cmd/server/main.go`.

Startup flow:

1. `config.Load()` reads `.env` and environment variables.
2. `app.New(cfg)` creates the application.
3. `application.Run()` starts the HTTP server.

Important functions:

| Function | Responsibility |
|---|---|
| `main()` | Loads config, initializes app, runs server, prints fatal errors to stderr and exits with code 1. |
| `app.New(cfg)` | Connects to DB, ensures schemas, seeds admin user, creates repositories/services/middleware/router/server. |
| `app.Run()` | Starts `ListenAndServe`, waits for `SIGINT` or `SIGTERM`, gracefully shuts down server. |
| `app.closeDB()` | Closes PostgreSQL pool or SQLite database handle. |
| `app.Router()` | Exposes HTTP handler for tests. |

Database initialization:

- If `DB_DRIVER=sqlite`, the app opens `SQLITE_PATH`, ensures core and admin SQLite schemas, and creates SQLite repositories.
- If `DB_DRIVER=postgres`, the app builds/uses `DatabaseURL`, creates a `pgxpool.Pool`, ensures core and admin PostgreSQL schemas, and creates PostgreSQL repositories.

Route registration happens in `internal/app/routes.go` using Chi. The server listens on `":" + cfg.AppPort` with 15 second read/write timeouts.

Shutdown is implemented: `app.Run()` listens for `SIGINT` and `SIGTERM`, then calls `http.Server.Shutdown` with a 10 second timeout and closes the database connection.

## 6. Configuration System

Configuration is implemented in `internal/config/config.go`.

The `Config` struct fields are:

```go
type Config struct {
    AppEnv string
    AppPort string
    DBDriver string
    SQLitePath string
    DatabaseURL string
    DBHost string
    DBPort string
    DBName string
    DBUser string
    DBPassword string
    DBSSLMode string
    DBAutoMigrate bool
    ServiceName string
    EnableConsoleLogging bool
    EnableDatabaseLogging bool
    LogLevel string
    AdminUsername string
    AdminPassword string
    AdminAPIKey string
    JWTSecret string
    JWTExpiryHours int
    CORSOrigins string
    EnableSwaggerUI bool
}
```

`config.Load()` calls `godotenv.Load()` and then reads environment variables. Defaults are used when variables are missing.

| Environment Variable | Purpose | Example |
|---|---|---|
| `APP_ENV` | Runtime environment label | `local` |
| `APP_PORT` | Backend HTTP port | `8080` |
| `DB_DRIVER` | Database driver, `sqlite` or `postgres` | `postgres` |
| `SQLITE_PATH` | SQLite database file path | `./data/logger.db` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | PostgreSQL database name | `loggerdb` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `DB_AUTO_MIGRATE` | Enables Go startup schema creation when tables are missing | `true` |
| `SERVICE_NAME` | Stored in `LogEntry.Source` | `application-logging-audit-module` |
| `ENABLE_CONSOLE_LOGGING` | Prints logs to stdout | `true` |
| `ENABLE_DATABASE_LOGGING` | Persists logs to DB | `true` |
| `LOG_LEVEL` | Config field exists, but current logger service does not filter by it | `INFO` |
| `ADMIN_USERNAME` | Default seeded admin username | `admin` |
| `ADMIN_PASSWORD` | Default seeded admin password before bcrypt hashing | `12345678` |
| `ADMIN_API_KEY` | API key accepted for admin routes | `super-secret-admin-key` |
| `JWT_SECRET` | HMAC secret for admin JWTs | `change-this-secret` |
| `JWT_EXPIRY_HOURS` | JWT lifetime in hours | `24` |
| `CORS_ORIGINS` | Allowed frontend origins, comma separated | `http://localhost:5173` |
| `ENABLE_SWAGGER_UI` | Enables `/swagger/` route | `false` |
| `TEST_DB_HOST` | Integration test PostgreSQL host | `localhost` |
| `TEST_DB_PORT` | Integration test PostgreSQL port | `5432` |
| `TEST_DB_NAME` | Integration test PostgreSQL DB | `loggerdb_test` |
| `TEST_DB_USER` | Integration test DB user | `postgres` |
| `TEST_DB_PASSWORD` | Integration test DB password | `postgres` |
| `TEST_DB_SSLMODE` | Integration test SSL mode | `disable` |
| `VITE_API_URL` | Frontend backend base URL | `http://localhost:8080` |

Default behavior:

- `DB_DRIVER=sqlite`
- `SQLITE_PATH=./data/logger.db`
- `APP_PORT=8080`
- `ADMIN_USERNAME=admin`
- `ADMIN_PASSWORD=12345678`
- `JWT_EXPIRY_HOURS=24`
- `ENABLE_SWAGGER_UI=false`

## 7. Database and Migration System

The project supports both PostgreSQL and SQLite. PostgreSQL is the production-like option; SQLite is the default local/demo option.

PostgreSQL connection logic:

- Implemented in `internal/database/postgres.go`.
- `Connect(ctx, databaseURL)` calls `ConnectWithRetry`.
- `ConnectWithRetry` parses the URL, creates a `pgxpool.Pool`, sets max/min connection values, pings the database, and retries transient failures.
- `Ping(ctx, pool)` wraps `pool.Ping`.

SQLite connection logic:

- Implemented in `internal/database/sqlite.go`.
- `ConnectSQLite(path)` creates the parent directory, opens a SQLite DSN, enables foreign keys and busy timeout, sets max open/idle connections to 1, and pings the database.

Migration file location:

```text
database/migrations/
```

Migration files:

| File | Purpose |
|---|---|
| `init.sql` | Baseline Flyway placeholder. |
| `V1_1__create_logging_audit_tables.sql` | PostgreSQL `application_logs` and `audit_events` tables and indexes. |
| `V1_2__create_admin_users.sql` | PostgreSQL `admin_users` table and username index. |
| `V1_1_sqlite__create_logging_audit_tables.sql` | SQLite version of logging/audit schema. |
| `V1_2_sqlite__create_admin_users.sql` | SQLite version of admin users schema. |

How migrations are run:

- Go auto-migration is implemented through embedded SQL in `database/embed.go`.
- `EnsureSchema` and `EnsureAdminSchema` run PostgreSQL migrations when tables are missing and `DB_AUTO_MIGRATE=true`.
- `EnsureSchemaSQLite` and `EnsureAdminSchemaSQLite` do the same for SQLite.
- `make migrate-up` also exists for manual Flyway PostgreSQL migrations.
- No schema migration history table is implemented by Go. Flyway can manage its own history if `make migrate-up` is used.

### application_logs

PostgreSQL schema:

| Column | Type | Purpose |
|---|---|---|
| `id` | `UUID PRIMARY KEY` | Unique log entry identifier. |
| `level` | `VARCHAR(20) NOT NULL` | Log severity: INFO, WARNING, ERROR, DEBUG. |
| `message` | `TEXT NOT NULL` | Human-readable log message. |
| `source` | `VARCHAR(255)` | Service/component name from config. |
| `request_id` | `VARCHAR(255)` | Correlation ID for tracing HTTP requests. |
| `user_id` | `UUID NULL` | Optional user identifier from context. |
| `error_code` | `VARCHAR(100)` | Optional error code field in model; not currently set by service methods. |
| `stack_trace` | `TEXT` | In current code, stores `err.Error()` for error logs, not a real stack trace. |
| `metadata` | `JSONB` | Structured sanitized metadata. |
| `created_at` | `TIMESTAMP WITH TIME ZONE DEFAULT now()` | Creation timestamp. |

Indexes:

- `idx_application_logs_level`
- `idx_application_logs_created_at`
- `idx_application_logs_user_id`
- `idx_application_logs_request_id`

### audit_events

PostgreSQL schema:

| Column | Type | Purpose |
|---|---|---|
| `id` | `UUID PRIMARY KEY` | Unique audit event identifier. |
| `user_id` | `UUID NULL` | Optional acting user ID. |
| `username` | `VARCHAR(255)` | Acting username. |
| `action` | `VARCHAR(100) NOT NULL` | Audit action such as LOGIN or UPDATE_RECORD. |
| `resource_type` | `VARCHAR(100)` | Type of affected resource, for example invoice. |
| `resource_id` | `VARCHAR(255)` | Identifier of affected resource. |
| `old_value` | `JSONB` | Sanitized previous data snapshot. |
| `new_value` | `JSONB` | Sanitized new data snapshot. |
| `ip_address` | `VARCHAR(100)` | Client IP address supplied by handler. |
| `user_agent` | `TEXT` | HTTP user agent. |
| `request_id` | `VARCHAR(255)` | Request correlation ID. |
| `status` | `VARCHAR(50) NOT NULL` | SUCCESS or FAILURE. |
| `metadata` | `JSONB` | Extra sanitized structured metadata. |
| `created_at` | `TIMESTAMP WITH TIME ZONE DEFAULT now()` | Creation timestamp. |

Indexes:

- `idx_audit_events_user_id`
- `idx_audit_events_action`
- `idx_audit_events_resource_type`
- `idx_audit_events_status`
- `idx_audit_events_created_at`
- `idx_audit_events_request_id`

JSONB metadata allows flexible structured context without changing the schema for every new field. Request IDs connect HTTP request logs with business audit events. User IDs connect data to user accounts when the calling application provides them.

Parameterized SQL is important because it separates SQL code from user-supplied values. The repositories use `$1`, `$2`, etc. for PostgreSQL and `?` for SQLite to reduce SQL injection risk.

## 8. Logger Package Documentation

Package location:

```text
internal/logger/
```

Files:

| File | Purpose |
|---|---|
| `types.go` | Defines `LogEntry`, `LogFilter`, and valid log levels. |
| `interface.go` | Defines `Logger` and `LogRepository` interfaces. |
| `service.go` | Implements `LoggerService`, context helpers, validation, sanitization, console logging, DB logging. |
| `repository.go` | Implements PostgreSQL repository. |
| `repository_sqlite.go` | Implements SQLite repository. |
| `sanitizer.go` | Implements sensitive data masking. |
| `*_test.go` | Unit, sanitizer, SQLite, and PostgreSQL integration tests. |

### LogEntry

Purpose: Represents one application log record.

Fields:

- `ID uuid.UUID`: unique log ID.
- `Level string`: INFO, WARNING, ERROR, DEBUG.
- `Message string`: log message.
- `Source string`: configured service name.
- `RequestID string`: HTTP/request correlation ID.
- `UserID *uuid.UUID`: optional user ID.
- `ErrorCode string`: optional error code, currently not set by service methods.
- `StackTrace string`: current service stores the error text for error logs.
- `Metadata map[string]any`: structured sanitized data.
- `CreatedAt time.Time`: UTC creation time.

Methods: none directly.

Used by: `LoggerService`, `LogRepository`, `PostgresLogRepository`, `SQLiteLogRepository`, handlers, exporters, frontend API responses.

CBSE relevance: core data contract of the logging component.

### LogFilter

Purpose: Carries query filters for listing logs.

Fields:

- `Level`
- `RequestID`
- `UserID`
- `Source`
- `From`
- `To`
- `Pagination`

Methods: none directly.

Used by: `LogRepository.Find`, `LogHandler.ListLogs`, `LogHandler.ExportLogs`.

CBSE relevance: separates query criteria from handler and repository implementation.

### Logger interface

Purpose: Public logging behavior used by handlers and reusable applications.

Methods:

- `Info(ctx, message, metadata) error`
- `Warn(ctx, message, metadata) error`
- `Error(ctx, message, err, metadata) error`
- `Debug(ctx, message, metadata) error`

Used by: `LogHandler`, direct integration example.

CBSE relevance: provided interface of the component.

### LogRepository interface

Purpose: Persistence abstraction for log storage.

Methods:

- `Insert(ctx, entry) error`
- `Find(ctx, filter) ([]LogEntry, error)`
- `FindByID(ctx, id) (*LogEntry, error)`

Used by: `LoggerService`, `LogHandler`, PostgreSQL/SQLite implementations, tests.

CBSE relevance: makes persistence replaceable and testable.

### LoggerService

Purpose: Main logging service.

Fields:

- `repo LogRepository`: persistence target.
- `sanitizer *SensitiveDataSanitizer`: masks sensitive metadata.
- `source string`: service name stored in entries.
- `consoleEnabled bool`: controls stdout logging.
- `dbEnabled bool`: controls database persistence.

Methods:

| Method | Purpose |
|---|---|
| `NewService(repo, source, consoleEnabled, dbEnabled)` | Constructor. |
| `Info(ctx, message, metadata)` | Creates INFO log. |
| `Warn(ctx, message, metadata)` | Creates WARNING log. |
| `Debug(ctx, message, metadata)` | Creates DEBUG log. |
| `Error(ctx, message, err, metadata)` | Creates ERROR log and stores `err.Error()` in `StackTrace`. |
| `log(ctx, level, message, stack, metadata, err)` | Internal common logging flow. |
| `LogHTTP(ctx, message, metadata)` | Used by middleware; delegates to `Info`. |
| `WithRequestID(ctx, requestID)` | Stores request ID in context. |
| `WithUserID(ctx, userID)` | Stores user ID in context. |

Method details:

```text
Method: Info/Warn/Debug/Error
Purpose: Create a log at the selected level.
Parameters: context, message, metadata; Error also accepts err.
Return values: error from validation or repository insert.
Flow: delegate to internal log method.
Error handling: invalid levels and repository errors are returned.
```

```text
Method: log
Purpose: Build and optionally persist a LogEntry.
Parameters: context, level, message, stack, metadata, err.
Return values: error or nil.
Flow: normalize level -> validate against ValidLevels -> sanitize metadata -> create UUID/time -> read request/user ID from context -> optionally print to console -> optionally insert into repository.
Error handling: returns invalid level error, metadata marshal/repository errors through repository call.
```

Used by: handlers, middleware, example app.

CBSE relevance: service implementation behind the provided logging interface.

### PostgresLogRepository

Purpose: PostgreSQL implementation of `LogRepository`.

Fields:

- `pool *pgxpool.Pool`

Methods:

| Method | Purpose | Parameters | Return Values | Error Handling |
|---|---|---|---|---|
| `NewPostgresRepository(pool)` | Constructor | `*pgxpool.Pool` | `*PostgresLogRepository` | No explicit validation. |
| `Insert(ctx, entry)` | Inserts into `application_logs` | context, `LogEntry` | `error` | Wraps JSON marshal and insert errors. |
| `Find(ctx, filter)` | Lists logs by filters | context, `LogFilter` | `[]LogEntry`, `error` | Wraps query errors; row scan errors returned. |
| `FindByID(ctx, id)` | Gets one log | context, UUID | `*LogEntry`, `error` | Returns `nil, nil` if no row. |

Flow: JSON metadata is marshaled, nil metadata becomes `{}`, SQL uses positional parameters, results are scanned and metadata is unmarshaled.

CBSE relevance: concrete persistence component.

### SQLiteLogRepository

Purpose: SQLite implementation of `LogRepository`.

Fields:

- `db *sql.DB`

Methods mirror `PostgresLogRepository` but use `database/sql`, `?` placeholders, UUIDs as strings, and timestamps formatted as RFC3339Nano strings.

CBSE relevance: demonstrates replaceable repository implementation through the same interface.

### SensitiveDataSanitizer

Purpose: Masks sensitive keys and card-like numbers before logs/audit data are stored.

Fields: none.

Methods:

| Method | Purpose |
|---|---|
| `NewSanitizer()` | Constructor. |
| `SanitizeMap(data)` | Returns a sanitized copy of a map. |
| `sanitizeValue(key, value)` | Internal recursive sanitizer. |
| `isSensitiveKey(key)` | Internal helper that checks key names. |

Sensitive keys currently masked when the key contains:

```text
password, passwd, token, authorization, auth, secret, api_key, apikey,
cvv, card, credit_card, credit card, pan
```

String values that match a 13 to 19 digit card-like pattern are also replaced with `***MASKED***`.

Log creation flow:

```text
Application calls LoggerService -> service validates level -> metadata sanitized -> LogEntry created -> repository inserts into PostgreSQL or SQLite
```

## 9. Audit Package Documentation

Package location:

```text
internal/audit/
```

Files:

| File | Purpose |
|---|---|
| `types.go` | Defines audit event model, filter, valid actions, valid statuses. |
| `interface.go` | Defines `Auditor` and `AuditRepository`. |
| `service.go` | Implements validation, default ID/time, sanitization, repository insert. |
| `repository.go` | PostgreSQL repository. |
| `repository_sqlite.go` | SQLite repository. |
| `*_test.go` | Service, SQLite, and PostgreSQL integration tests. |

### AuditEvent

Purpose: Represents one recorded user/system action.

Fields:

- `ID uuid.UUID`: unique event ID.
- `UserID *uuid.UUID`: optional user ID.
- `Username string`: acting username.
- `Action string`: valid action such as LOGIN or UPDATE_RECORD.
- `ResourceType string`: affected resource type.
- `ResourceID string`: affected resource ID.
- `OldValue map[string]any`: sanitized old snapshot.
- `NewValue map[string]any`: sanitized new snapshot.
- `IPAddress string`: client IP address.
- `UserAgent string`: client user agent.
- `RequestID string`: request correlation ID.
- `Status string`: SUCCESS or FAILURE.
- `Metadata map[string]any`: sanitized extra metadata.
- `CreatedAt time.Time`: UTC creation time.

Valid actions:

```text
LOGIN, LOGOUT, CREATE_RECORD, UPDATE_RECORD, DELETE_RECORD,
DOWNLOAD_FILE, FAILED_LOGIN, PERMISSION_DENIED
```

Valid statuses:

```text
SUCCESS, FAILURE
```

Methods: none directly.

Used by: `AuditService`, repositories, handlers, exporters.

CBSE relevance: core data contract of the audit component.

### AuditFilter

Purpose: Carries query filters for listing audit events.

Fields:

- `UserID`
- `Username`
- `Action`
- `ResourceType`
- `Status`
- `RequestID`
- `From`
- `To`
- `Pagination`

Used by: `AuditRepository.Find`, `AuditHandler.ListAuditEvents`, `AuditHandler.ExportAuditEvents`.

### Auditor interface

Purpose: Public audit behavior used by HTTP handlers or another application.

Methods:

- `Record(ctx, event) error`
- `Find(ctx, filter) ([]AuditEvent, error)`

### AuditRepository interface

Purpose: Persistence abstraction for audit events.

Methods:

- `Insert(ctx, event) error`
- `Find(ctx, filter) ([]AuditEvent, error)`
- `FindByID(ctx, id) (*AuditEvent, error)`

### AuditService

Purpose: Validates and records audit events.

Fields:

- `repo AuditRepository`
- `sanitizer *logger.SensitiveDataSanitizer`

Methods:

| Method | Purpose | Parameters | Return Values | Error Handling |
|---|---|---|---|---|
| `NewService(repo)` | Constructor | `AuditRepository` | `*AuditService` | No explicit validation. |
| `Record(ctx, event)` | Validates, normalizes, sanitizes, inserts event | context, `AuditEvent` | `error` | Returns missing/invalid action/status errors and repository errors. |
| `Find(ctx, filter)` | Delegates query to repository | context, `AuditFilter` | `[]AuditEvent`, `error` | Returns repository error. |

Validation logic:

- `Action` is trimmed, uppercased, required, and must be in `ValidActions`.
- `Status` is trimmed, uppercased, required, and must be in `ValidStatuses`.
- If `ID` is empty, a new UUID is assigned.
- If `CreatedAt` is zero, current UTC time is assigned.
- `OldValue`, `NewValue`, and `Metadata` are sanitized.

### PostgresAuditRepository

Purpose: PostgreSQL implementation of `AuditRepository`.

Fields:

- `pool *pgxpool.Pool`

Methods:

| Method | Purpose | Parameters | Return Values | Error Handling |
|---|---|---|---|---|
| `NewPostgresRepository(pool)` | Constructor | `*pgxpool.Pool` | `*PostgresAuditRepository` | No explicit validation. |
| `Insert(ctx, event)` | Inserts into `audit_events` | context, `AuditEvent` | `error` | Wraps JSON marshal and insert errors. |
| `Find(ctx, filter)` | Lists events by filters | context, `AuditFilter` | `[]AuditEvent`, `error` | Wraps query errors; row scan errors returned. |
| `FindByID(ctx, id)` | Gets one event | context, UUID | `*AuditEvent`, `error` | Returns `nil, nil` when no row exists. |

### SQLiteAuditRepository

Purpose: SQLite implementation of `AuditRepository`.

Fields:

- `db *sql.DB`

Methods mirror `PostgresAuditRepository` but use SQLite placeholders and string storage for UUID/time/JSON.

Audit flow:

```text
User action happens -> application calls AuditService -> event validated -> old/new values sanitized -> event stored in audit_events table
```

## 10. Middleware Documentation

Package location:

```text
internal/middleware/
```

Files:

| File | Purpose |
|---|---|
| `http_logger.go` | HTTP request logging middleware. |
| `response_writer.go` | Wraps response writer to capture status code. |
| `admin_auth.go` | Protects admin routes using JWT or API key. |
| `cors.go` | Handles allowed origins and preflight requests. |
| `*_test.go` | Middleware tests. |

### HTTP logging middleware

Struct: `HTTPLogger`

Fields:

- `logger *logger.LoggerService`

Methods:

- `NewHTTPLogger(svc)`
- `Middleware(next http.Handler) http.Handler`

Responsibilities:

- Read or create `X-Request-ID`.
- Set `X-Request-ID` response header.
- Wrap response writer.
- Measure latency.
- Capture method, path, status code, latency, IP address, user agent, and request ID.
- Write an INFO log through `LoggerService.LogHTTP`.

Request ID handling:

- If the request already has `X-Request-ID`, it is reused.
- Otherwise, a new UUID string is generated.

Status code capture:

- `responseWriter` defaults status to `200 OK`.
- When a handler calls `WriteHeader`, the wrapper stores the actual code.

Latency calculation:

- Uses `time.Now()` before handler execution and `time.Since(start)` after execution.
- Stores `latency_ms`.

IP address capture:

- Uses first value from `X-Forwarded-For` if present.
- Otherwise uses `X-Real-IP`.
- Otherwise splits `RemoteAddr`.

User agent capture:

- Uses `r.UserAgent()`.

Behavior when logging fails:

- The middleware ignores logging errors using `_ = h.logger.LogHTTP(...)`.
- This is intentional because logging failure should not break the business request.

### Admin authentication middleware

Struct: `AdminAuth`

Fields:

- `tokens *adminauth.TokenService`
- `apiKey string`

Methods:

- `NewAdminAuth(tokens, apiKey)`
- `Middleware(next)`
- `authenticate(r)`

Authorization header validation:

- Accepts `Authorization: Bearer <jwt>`.
- Accepts `Authorization: ApiKey <key>` when configured.
- Accepts `X-API-Key: <key>` when configured.
- Returns HTTP `401` JSON `{"error":"unauthorized"}` when authentication fails.

Middleware is important for API gateway style logging because it captures cross-cutting request data outside individual handlers. This keeps handlers focused on business logic while still recording every request consistently.

## 11. Handler/API Documentation

Package location:

```text
internal/handler/
```

Handlers:

| Handler | File | Purpose |
|---|---|---|
| `HealthHandler` | `health_handler.go` | Reports app and database status. |
| `LogHandler` | `log_handler.go` | Demo log creation, admin log list/get/export. |
| `AuditHandler` | `audit_handler.go` | Demo audit creation, admin audit list/get/export. |
| `AuthHandler` | `auth_handler.go` | Admin login and JWT issuing. |

API route table:

| Method | Route | Purpose | Auth Required |
|---|---|---|---|
| GET | `/health` | Health check and DB status | No |
| POST | `/demo/log-info` | Create demo INFO log | No |
| POST | `/demo/log-error` | Create demo ERROR log | No |
| POST | `/demo/audit-login` | Record demo LOGIN audit event | No |
| POST | `/demo/audit-update` | Record demo UPDATE_RECORD audit event | No |
| POST | `/admin/login` | Admin login, returns JWT | No |
| GET | `/admin/logs` | List logs with filters and pagination | Yes |
| GET | `/admin/logs/{id}` | Get one log by UUID | Yes |
| GET | `/admin/logs/export` | Export logs as JSON or CSV | Yes |
| GET | `/admin/audit-events` | List audit events with filters and pagination | Yes |
| GET | `/admin/audit-events/{id}` | Get one audit event by UUID | Yes |
| GET | `/admin/audit-events/export` | Export audit events as JSON or CSV | Yes |

### GET /health

Request body: none.

Response body:

```json
{
  "status": "ok",
  "database": "up"
}
```

If DB ping fails, response is still HTTP 200 with:

```json
{
  "status": "degraded",
  "database": "down"
}
```

### POST /demo/log-info

Request body:

```json
{
  "message": "User dashboard loaded",
  "metadata": {
    "component": "dashboard"
  }
}
```

If `message` is empty, the handler uses `Demo info log message`.

Response:

```json
{
  "status": "logged"
}
```

Service called: `Logger.Info`.

Error handling: invalid JSON returns 400; service error returns 500.

### POST /demo/log-error

Request body is the same as `/demo/log-info`. If message is empty, the handler uses `Demo error log message`.

The handler creates a demo error with message `simulated error for demo` and calls `Logger.Error`.

Response:

```json
{
  "status": "logged"
}
```

### POST /demo/audit-login

Request body:

```json
{
  "user_id": "optional-uuid",
  "username": "alice",
  "ip_address": "192.168.1.10"
}
```

If `username` is empty, the handler uses `demo_user`. Invalid `user_id` is ignored.

Service called: `AuditService.Record` through `Auditor`.

Response:

```json
{
  "status": "recorded",
  "id": ""
}
```

Implementation note: the handler returns `event.ID.String()`, but `AuditEvent` is passed by value to `AuditService.Record`, so the generated ID inside the service is not written back to the handler's local `event`. Therefore the returned `id` may be empty. Recommended improvement: make `Record` return the stored event/ID or pass a pointer.

### POST /demo/audit-update

Request body:

```json
{
  "user_id": "optional-uuid",
  "username": "alice",
  "resource_type": "invoice",
  "resource_id": "1001",
  "old_value": {
    "status": "draft"
  },
  "new_value": {
    "status": "paid"
  }
}
```

Defaults:

- `resource_type`: `invoice`
- `resource_id`: `1001`

Response:

```json
{
  "status": "recorded",
  "id": ""
}
```

Same ID return limitation as `/demo/audit-login`.

### POST /admin/login

Request body:

```json
{
  "username": "admin",
  "password": "12345678"
}
```

Response:

```json
{
  "token": "<jwt>"
}
```

Error handling:

- Invalid JSON: 400.
- Missing username/password: 400.
- Wrong credentials: 401.
- Repository/token errors: 500.

### GET /admin/logs

Query parameters:

- `level`
- `request_id`
- `source`
- `user_id`
- `from`
- `to`
- `page`
- `limit`

Response:

```json
{
  "data": [],
  "page": 1,
  "limit": 20
}
```

Invalid `user_id`, `from`, or `to` values are ignored by current parsing logic instead of returning 400.

### GET /admin/logs/{id}

Path parameter:

- `id`: UUID.

Responses:

- 200 with `LogEntry`.
- 400 for invalid UUID.
- 404 if not found.
- 500 on repository error.

### GET /admin/logs/export

Query parameters:

- Same filters as `/admin/logs`.
- `format=json` or `format=csv`; default is `json`.

Response:

- `application/json` with attachment filename `logs.json`.
- `text/csv` with attachment filename `logs.csv`.
- 400 if format is not `json` or `csv`.

### GET /admin/audit-events

Query parameters:

- `user_id`
- `username`
- `action`
- `resource_type`
- `status`
- `request_id`
- `from`
- `to`
- `page`
- `limit`

Response:

```json
{
  "data": [],
  "page": 1,
  "limit": 20
}
```

### GET /admin/audit-events/{id}

Path parameter:

- `id`: UUID.

Responses:

- 200 with `AuditEvent`.
- 400 for invalid UUID.
- 404 if not found.
- 500 on repository error.

### GET /admin/audit-events/export

Query parameters:

- Same filters as `/admin/audit-events`.
- `format=json` or `format=csv`; default is `json`.

## 12. Admin Authentication System

Admin authentication is implemented in:

- `internal/adminauth`
- `internal/middleware/admin_auth.go`
- `internal/handler/auth_handler.go`

How admin login works:

1. On startup, `adminauth.SeedDefaultAdmin` checks whether `ADMIN_USERNAME` exists.
2. If not, it hashes `ADMIN_PASSWORD` using bcrypt and inserts a row into `admin_users`.
3. `POST /admin/login` accepts username/password.
4. `AuthHandler.Login` loads the admin user by username.
5. `adminauth.CheckPassword` verifies the bcrypt hash.
6. `TokenService.IssueToken` creates a JWT signed with HS256.
7. The token is returned to the frontend.

Credentials come from environment variables:

- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`

Token/API key configuration:

- `JWT_SECRET` signs JWTs.
- `JWT_EXPIRY_HOURS` controls expiry.
- `ADMIN_API_KEY` enables API-key access.

How backend validates auth:

- Protected `/admin` routes use `AdminAuth.Middleware`.
- The middleware accepts:
  - `Authorization: Bearer <jwt>`
  - `X-API-Key: <ADMIN_API_KEY>`
  - `Authorization: ApiKey <ADMIN_API_KEY>`

How frontend stores token:

- `frontend/src/services/api.js` stores the token in `localStorage` under `admin_token`.
- `apiFetch` adds `Authorization: Bearer <token>` to requests.
- On HTTP 401, frontend clears the token and redirects to `/login`.

Protected routes:

- `/admin/logs`
- `/admin/logs/{id}`
- `/admin/logs/export`
- `/admin/audit-events`
- `/admin/audit-events/{id}`
- `/admin/audit-events/export`

Unprotected routes:

- `/health`
- `/demo/*`
- `/admin/login`
- `/swagger/` when enabled

Limitations:

- This is a simple demo authentication system.
- Production should use HTTPS, stronger secret management, password policies, RBAC, secure cookies/sessions or hardened token storage, refresh-token strategy, CSRF protection where relevant, login throttling, and audit of admin login failures.

## 13. Frontend Documentation

Frontend folder:

```text
frontend/
```

Technology:

- React 18.
- Vite 5.
- React Router DOM 6.
- Nginx container for production build.

Frontend structure:

```text
frontend/src/App.jsx
frontend/src/main.jsx
frontend/src/services/api.js
frontend/src/components/Navbar.jsx
frontend/src/components/ProtectedRoute.jsx
frontend/src/components/LogTable.jsx
frontend/src/components/AuditTable.jsx
frontend/src/pages/LoginPage.jsx
frontend/src/pages/DashboardPage.jsx
frontend/src/pages/LogsPage.jsx
frontend/src/pages/AuditPage.jsx
frontend/src/styles/app.css
```

Frontend component table:

| Frontend File/Component | Purpose |
|---|---|
| `src/main.jsx` | Mounts React app and wraps it with `BrowserRouter`. |
| `src/App.jsx` | Defines routes `/login`, `/dashboard`, `/logs`, `/audit`, and redirects. |
| `src/services/api.js` | Stores token, performs login, authenticated fetches, logs/audit loading, export downloads. |
| `src/components/ProtectedRoute.jsx` | Redirects to `/login` if no token exists in localStorage. |
| `src/components/Navbar.jsx` | Navigation links and logout button. |
| `src/components/LogTable.jsx` | Displays logs with level, message, source, request ID, and creation time. |
| `src/components/AuditTable.jsx` | Displays audit username, action, resource type, status, IP, and timestamp. |
| `src/pages/LoginPage.jsx` | Login form; posts credentials; stores token through API service. |
| `src/pages/DashboardPage.jsx` | Loads recent ERROR logs and recent audit events; links to logs/audit pages. |
| `src/pages/LogsPage.jsx` | Log filters, pagination, table, CSV/JSON export buttons. |
| `src/pages/AuditPage.jsx` | Audit filters, pagination, table, CSV/JSON export buttons. |
| `src/styles/app.css` | UI styling. |
| `Dockerfile` | Builds frontend with Node 20 and serves with Nginx. |
| `nginx.conf` | Serves SPA and falls back to `index.html`. |

Login page:

- Default username field is `admin`.
- Password field is empty.
- On success, navigates to `/dashboard`.
- Shows error banner on failure.

Dashboard page:

- Loads `/admin/logs?level=ERROR&page=1&limit=5`.
- Loads `/admin/audit-events?page=1&limit=5`.
- Displays recent errors and audit events.

Logs page:

- Filters by `level`, `from`, `to`.
- Uses `page` and `limit`.
- Provides Previous/Next pagination.
- Exports CSV or JSON.
- Not implemented yet / recommended improvement: frontend does not expose all backend log filters such as `source`, `request_id`, and `user_id`.

Audit page:

- Filters by `username`, `action`, `status`, `resource_type`, `from`, `to`.
- Uses `page` and `limit`.
- Provides Previous/Next pagination.
- Exports CSV or JSON.
- Not implemented yet / recommended improvement: frontend does not expose `request_id` or `user_id` audit filters.

Logout behavior:

- `Navbar` calls `clearToken()` and sets `window.location.href = '/login'`.

Frontend flow:

```text
Admin opens frontend -> login page -> sends credentials -> receives token -> stores token -> calls protected admin APIs -> displays logs/audit events
```

Frontend authentication limitation:

- Token is stored in localStorage. This is simple for the demo but is not the most secure production option.

## 14. Export System

Export package:

```text
internal/exporter/
```

Files:

- `json_exporter.go`
- `csv_exporter.go`
- `json_exporter_test.go`
- `csv_exporter_test.go`

JSON export:

- `ExportLogsJSON(entries)` returns indented JSON.
- `ExportAuditJSON(events)` returns indented JSON.

CSV export:

- `ExportLogsCSV(entries)` writes columns:
  - `id`, `level`, `message`, `source`, `request_id`, `user_id`, `error_code`, `created_at`, `metadata`
- `ExportAuditCSV(events)` writes columns:
  - `id`, `user_id`, `username`, `action`, `resource_type`, `resource_id`, `status`, `ip_address`, `request_id`, `created_at`, `metadata`

Export routes:

- `GET /admin/logs/export?format=json`
- `GET /admin/logs/export?format=csv`
- `GET /admin/audit-events/export?format=json`
- `GET /admin/audit-events/export?format=csv`

Data exported:

- Logs returned by the same filters as `/admin/logs`.
- Audit events returned by the same filters as `/admin/audit-events`.

Metadata handling:

- JSON export preserves metadata as JSON objects.
- CSV export marshals metadata maps into a JSON string column.

Frontend export behavior:

- `exportResource` fetches a blob, creates an object URL, creates a temporary anchor, triggers download, and revokes the object URL.
- Frontend filenames are `logs.csv`, `logs.json`, `audit_events.csv`, or `audit_events.json`.

Not implemented yet / recommended improvement:

- `FormatExportFilename` exists but is not used by handlers.
- Export loads filtered rows into memory; streaming export is not implemented.

## 15. Security Design

Implemented security decisions:

- Sensitive data sanitization before persistence.
- Passwords are bcrypt-hashed in `admin_users`.
- Admin routes are protected by JWT or API key middleware.
- SQL uses parameterized placeholders.
- HTTP middleware creates request IDs for traceability.
- Middleware does not panic or break requests when request logging fails.
- JSON response helper avoids manually constructing JSON strings.
- CORS allows configured origins only, not all origins.

Sensitive fields masked:

```text
password
passwd
token
authorization
auth
secret
api_key
apikey
cvv
card
credit_card
credit card
pan
```

Card-like string values with 13 to 19 digits are also masked.

Security limitations and improvements:

- Demo credentials have insecure defaults; production must override them.
- Default `JWT_SECRET=change-this-secret` is not secure for production.
- API key is static and stored in environment variables.
- No RBAC roles are implemented.
- No rate limiting or login throttling is implemented.
- No HTTPS enforcement is implemented in the app.
- No refresh token/session revocation system is implemented.
- Frontend stores JWT in localStorage.
- Demo endpoints are public and can create logs/audit events.
- Audit records are append-only by convention, but database-level immutability constraints/triggers are not implemented.

## 16. Error Handling Strategy

Service-level errors:

- `LoggerService` returns errors for invalid log levels and repository insert failures.
- `AuditService` returns errors for missing/invalid action/status and repository insert failures.

Repository-level errors:

- Repositories wrap database and JSON errors with context such as `insert log`, `find logs`, or `marshal json`.
- `FindByID` returns `nil, nil` when a row is not found.

Middleware logging failures:

- HTTP logging middleware ignores logging failure so normal request handling is not broken.

Handler error responses:

- `common.WriteError` returns JSON in this format:

```json
{
  "error": "message"
}
```

Common status codes:

- 400 for invalid JSON or invalid ID.
- 401 for unauthorized admin access or invalid credentials.
- 404 for missing log/audit event.
- 500 for service/repository/database errors.

Database connection errors:

- `app.New` fails startup if database connection, schema creation, admin schema, or admin seeding fails.
- `cmd/server/main.go` prints the error and exits with code 1.

Frontend error handling:

- Login page displays login failure messages.
- Dashboard/logs/audit pages show error banners.
- `apiFetch` redirects to login on 401.

Why logging failure should not crash business functionality:

- Logging is a support concern. If a request succeeds but logging storage is temporarily unavailable, the original user operation should not necessarily fail. The middleware follows this principle by ignoring logging errors. Direct service callers can still choose to handle returned logging errors.

## 17. Reusability and Component Independence

The module is reusable because it separates domain services from HTTP handlers.

Reusable parts:

- `logger.Logger` interface.
- `logger.LogRepository` interface.
- `logger.NewService`.
- `audit.Auditor` interface.
- `audit.AuditRepository` interface.
- `audit.NewService`.
- PostgreSQL and SQLite repository constructors.
- Config-based database choice.
- HTTP middleware for request logging.

Clean separation:

- Services do not depend on HTTP handlers.
- Repositories do not depend on handlers.
- Handlers depend on interfaces.
- Tests can use mock repositories without a real DB.

Important Go packaging limitation:

- Packages are under `internal/`, so they cannot be imported by a different Go module outside this repository tree. To reuse from another external project, a developer would need to move public packages to `pkg/`, copy/fork the module, or keep the consuming application inside the same parent module. This is a recommended improvement for production reusable library distribution.

### How another developer can use this component in 5 minutes

1. Configure `.env`.
2. Start the database or use default SQLite.
3. Create repositories.
4. Create services.
5. Call `Info`, `Error`, or `Record`.

Example from `examples/basic_usage.go`, shortened:

```go
cfg, _ := config.Load()
ctx := context.Background()

pool, _ := database.Connect(ctx, cfg.DatabaseURL)
defer pool.Close()
_ = database.EnsureSchema(ctx, pool, cfg.DBAutoMigrate)

logRepo := logger.NewPostgresRepository(pool)
auditRepo := audit.NewPostgresRepository(pool)

loggerSvc := logger.NewService(logRepo, cfg.ServiceName, true, true)
auditSvc := audit.NewService(auditRepo)

_ = loggerSvc.Info(ctx, "User dashboard loaded", map[string]any{
    "component": "dashboard",
})

_ = auditSvc.Record(ctx, audit.AuditEvent{
    Username:     "demo_user",
    Action:       "UPDATE_RECORD",
    ResourceType: "invoice",
    ResourceID:   "1001",
    Status:       "SUCCESS",
    NewValue:     map[string]any{"status": "paid"},
})
```

For default SQLite, use `database.ConnectSQLite`, `database.EnsureSchemaSQLite`, `logger.NewSQLiteRepository`, and `audit.NewSQLiteRepository`.

## 18. UML Diagram Explanation

Diagrams are located in:

```text
docs/diagrams/
```

Generated PNG files are located in:

```text
docs/diagrams/generated/
```

### Class diagram

Source: `docs/diagrams/class_diagram.puml`

What it shows:

- `Logger`, `LogRepository`, `Auditor`, `AuditRepository` interfaces.
- `LoggerService`, `AuditService`.
- `PostgresLogRepository`, `PostgresAuditRepository`.
- `SensitiveDataSanitizer`.
- Exporters.
- `Config`, `LogEntry`, `AuditEvent`.

Files/classes represented:

- `internal/logger/*`
- `internal/audit/*`
- `internal/exporter/*`
- `internal/config/config.go`
- `internal/middleware/http_logger.go`

CBSE usefulness: shows component interfaces and relationships.

Accuracy note: the current class diagram does not include SQLite repositories or admin authentication classes. Recommended improvement: update it to include `SQLiteLogRepository`, `SQLiteAuditRepository`, `AdminUser`, `TokenService`, and admin repositories.

### Sequence diagram

Source: `docs/diagrams/sequence_diagram.puml`

What it shows:

- User sends `POST /demo/audit-update`.
- Chi/API gateway passes to `AuditHandler`.
- `AuditService` validates and sanitizes.
- Repository inserts into PostgreSQL.
- API returns 201.

CBSE usefulness: explains runtime collaboration between layers.

Accuracy note: it names PostgreSQL only. The code can also use SQLite. Recommended improvement: mention "database repository" or show both database options.

### State diagram

Source: `docs/diagrams/state_diagram.puml`

What it shows:

- New log/audit entry starts in Created state.
- It can become Validated or ValidationFailed.
- Validated entries become Stored or StorageFailed.

CBSE usefulness: shows lifecycle and failure states for component records.

### Deployment diagram

Source: `docs/diagrams/deployment_diagram.puml`

What it shows:

- Client browser/admin UI.
- Docker Compose network.
- Go HTTP server on port 8080.
- Application Logging & Audit Trail Module.
- Go migration component.
- PostgreSQL internal database.

CBSE usefulness: shows deployment relationships and infrastructure.

Accuracy note: current default Docker Compose uses SQLite and a frontend container. The diagram focuses on PostgreSQL and should be updated to include the default SQLite volume and React/Nginx frontend container.

## 19. Unit Tests Documentation

Unit tests are co-located with packages under `internal/`.

Run unit tests:

```bash
go test ./...
```

Mock strategy:

- Service tests define small mock repositories in test files.
- Handler tests define mock services/repositories.
- Middleware tests use a capturing repository.
- Unit tests do not require a real PostgreSQL database.
- Some repository smoke tests use temporary SQLite databases, which are local and do not need an external DB server.

Unit test table:

| Test File | What It Tests |
|---|---|
| `internal/adminauth/password_test.go` | Bcrypt password hashing and password verification. |
| `internal/audit/service_test.go` | Audit validation, required fields, sanitization, repository error propagation. |
| `internal/audit/repository_sqlite_test.go` | SQLite audit insert and find behavior. |
| `internal/config/config_test.go` | Database URL building, special password encoding, validation, test DB fallback, default SQLite config, PostgreSQL URL config, unsupported driver, Swagger flag, default admin credentials. |
| `internal/exporter/csv_exporter_test.go` | CSV export for logs/audit records. |
| `internal/exporter/json_exporter_test.go` | JSON export for logs/audit records. |
| `internal/handler/audit_handler_test.go` | Demo audit login handler and get audit event handler. |
| `internal/handler/auth_handler_test.go` | Admin login success and invalid password response. |
| `internal/handler/log_handler_test.go` | Log listing handler response shape. |
| `internal/logger/repository_sqlite_test.go` | SQLite log insert and find behavior. |
| `internal/logger/sanitizer_test.go` | Sensitive key masking and card-number masking. |
| `internal/logger/service_test.go` | Log creation, invalid level rejection, metadata sanitization, repository error propagation. |
| `internal/middleware/admin_auth_test.go` | JWT and API key admin authentication. |
| `internal/middleware/http_logger_test.go` | Request ID, status, method, path, user agent, IP, metadata capture. |
| `internal/swagger/swagger_test.go` | Embedded OpenAPI spec, Swagger UI mounting, app Swagger enabled/disabled behavior. |

Why unit tests do not require a real database:

- Most tests use interfaces and in-memory mock structs.
- SQLite tests create local test databases.
- PostgreSQL tests are separated by the `integration` build tag.

## 20. Integration Tests Documentation

Run integration tests:

```bash
go test -tags=integration ./...
```

Integration tests use PostgreSQL. They do not use Testcontainers. They build the test database URL with `config.BuildTestDatabaseURL`, using `TEST_DB_*` variables and falling back to regular `DB_*` variables/defaults. If PostgreSQL is unavailable, `SetupTestPool` skips tests that cannot connect.

PostgreSQL requirement:

- A reachable PostgreSQL server is required.
- Default test database name is `loggerdb_test`.
- Migrations are applied by `database.EnsureSchema(ctx, pool, true)`.

Integration test table:

| Integration Test File | What It Tests |
|---|---|
| `internal/database/postgres_integration_test.go` | PostgreSQL connection/ping and creation of core tables. |
| `internal/database/migrate_integration_test.go` | `EnsureSchema` creates tables and is idempotent when run again. |
| `internal/logger/repository_integration_test.go` | PostgreSQL log repository insert, find by ID, and filtered find. |
| `internal/audit/repository_integration_test.go` | PostgreSQL audit repository insert and filtered find. |
| `internal/handler/admin_api_integration_test.go` | Admin API unauthorized behavior, login, demo log creation, protected log listing, demo audit creation, protected audit listing. |

Migration setup for tests:

- `SetupTestPool` calls `EnsureSchema`.
- `TruncateTables` truncates `application_logs` and `audit_events`.
- Admin API integration test uses `app.New`, which also ensures the `admin_users` table and seeds the admin user.

Not implemented yet / recommended improvement:

- Integration tests do not use Testcontainers, so the developer must provide PostgreSQL manually.
- `TruncateTables` only truncates `application_logs` and `audit_events`, not `admin_users`.

## 21. Makefile and Run Commands

Makefile commands:

| Command | Purpose |
|---|---|
| `make run` | Copies Swagger spec into embedded location and starts backend with `go run ./cmd/server`. |
| `make sync-swagger` | Copies `docs/swagger.yaml` to `internal/swagger/openapi.yaml`. |
| `make build` | Syncs Swagger and builds backend binary to `bin/server`. |
| `make tidy` | Runs `go mod tidy`. |
| `make test` | Runs unit tests with `go test ./...`. |
| `make test-integration` | Runs integration tests with `go test -tags=integration ./...`. |
| `make test-all` | Runs unit and integration tests. |
| `make coverage` | Generates `coverage.out` and `coverage.html`. |
| `make migrate-up` | Runs Flyway PostgreSQL migrations against localhost `loggerdb`. |
| `make diagrams` | Uses Docker PlantUML image to generate PNG diagrams. |
| `make docker-up` | Runs default Docker Compose stack with SQLite backend and frontend. |
| `make docker-up-postgres` | Runs Docker Compose with PostgreSQL overlay. |
| `make frontend-install` | Runs `npm install` in `frontend`. |
| `make frontend-dev` | Starts Vite dev server on port 5173. |
| `make frontend-build` | Builds frontend production files. |

Manual commands:

```bash
go run ./cmd/server
go test ./...
go test -tags=integration ./...
go run ./examples/basic_usage.go
cd frontend && npm install
cd frontend && npm run dev
cd frontend && npm run build
```

## 22. API Usage Examples

Assume backend is running at `http://localhost:8080`.

### Create an info log

```bash
curl -X POST http://localhost:8080/demo/log-info \
  -H "Content-Type: application/json" \
  -d '{
    "message": "User dashboard loaded",
    "metadata": {
      "component": "dashboard",
      "password": "will-be-masked"
    }
  }'
```

### Create an error log

```bash
curl -X POST http://localhost:8080/demo/log-error \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Payment service timeout",
    "metadata": {
      "service": "payment",
      "token": "will-be-masked"
    }
  }'
```

### Record audit login

```bash
curl -X POST http://localhost:8080/demo/audit-login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "ip_address": "192.168.1.10"
  }'
```

### Record audit update

```bash
curl -X POST http://localhost:8080/demo/audit-update \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "resource_type": "invoice",
    "resource_id": "1001",
    "old_value": {
      "status": "draft"
    },
    "new_value": {
      "status": "paid"
    }
  }'
```

### Admin login

```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}'
```

### Fetch logs

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}' | jq -r .token)

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/logs?level=ERROR&page=1&limit=20"
```

### Fetch audit events

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/audit-events?action=LOGIN&status=SUCCESS&page=1&limit=20"
```

### Export logs as CSV

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/logs/export?format=csv" \
  -o logs.csv
```

### Export audit events as JSON

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/audit-events/export?format=json" \
  -o audit_events.json
```

## 23. Deployment Requirements

Backend requirements:

- Go module declares Go `1.22`.
- Backend Dockerfile uses `golang:1.24-alpine` builder and `alpine:3.20` runtime.
- Backend listens on port `8080` by default.

Database requirements:

- SQLite is default and requires no separate DB server.
- PostgreSQL is supported through `DB_DRIVER=postgres`.
- Docker PostgreSQL overlay uses `postgres:16-alpine`.

Frontend requirements:

- Node.js 18+ is stated in `README.md`.
- Frontend Dockerfile uses `node:20-alpine`.
- Vite dev server uses port `5173`.
- Docker frontend serves on host port `5173` mapped to Nginx port 80.

Environment variables:

- Configure backend through `.env` or container environment.
- Configure frontend API base URL with `VITE_API_URL`.

Database setup:

- SQLite: default `./data/logger.db`; app creates tables when `DB_AUTO_MIGRATE=true`.
- PostgreSQL local: create database such as `loggerdb`, set `DB_DRIVER=postgres` and `DB_*` variables.
- Docker PostgreSQL: use `make docker-up-postgres`.

Migration setup:

- Go auto-migration runs on startup when tables are missing.
- Manual Flyway PostgreSQL migration is available with `make migrate-up`.

Running backend:

```bash
make run
# or
go run ./cmd/server
```

Running frontend:

```bash
make frontend-install
make frontend-dev
```

Docker:

```bash
make docker-up
make docker-up-postgres
```

Ports:

| Port | Purpose |
|---|---|
| `8080` | Backend HTTP API. |
| `5173` | Vite dev server or Docker frontend host port. |
| `5432` | PostgreSQL internal container port in overlay; not published to host by default. |

## 24. Restrictions and Limitations

Current restrictions/limitations:

- Admin authentication is simple and demo-oriented.
- No production-ready RBAC.
- No advanced log retention policy is implemented.
- No async logging queue is implemented.
- No distributed tracing implementation is included beyond request IDs.
- Frontend is a simple admin/demo UI.
- PostgreSQL is required for integration tests.
- SQLite is default for local/demo use.
- Migrations must exist/run before admin APIs can query tables.
- Go auto-migration does not implement a schema history table.
- `LOG_LEVEL` config exists but is not used for filtering log writes.
- Error logs store `err.Error()` in `StackTrace`; real stack traces are not collected.
- Demo audit handlers may return an empty `id` because events are passed by value to the service.
- Demo endpoints are public.
- `internal/` package layout limits reuse from external Go modules.
- No root-level `LICENSE` file exists; license text is in `docs/license.md`.

## 25. Similar and Competing Components

| Component | Language/Platform | Main Use | Audit Trail Support | Database Storage | Admin UI | Difference from This Project |
|---|---|---|---|---|---|---|
| Logrus | Go | Structured application logging | No first-class audit model | Not built in; hooks possible | No | Logger only; no audit schema, admin APIs, frontend, or CBSE docs. |
| Zap | Go | High-performance structured logging | No first-class audit model | Not built in; external sinks possible | No | Optimized for speed; this project focuses on logs plus audit persistence/admin access. |
| Zerolog | Go | Zero-allocation JSON logging | No first-class audit model | Not built in | No | Very lightweight logger; no audit workflow or database schema. |
| Log4j | Java | Enterprise Java logging | Not primary purpose | Appenders can write to many targets | No built-in admin UI | Java ecosystem logging library, not a Go component. |
| Winston | Node.js | Node application logging | Not primary purpose | Transports can be configured | No built-in admin UI | Node-specific logger; audit must be custom built. |
| Audit.NET | .NET | Audit trail recording | Yes | Supports providers | Limited/not the same | Strong audit focus but .NET-specific and not an app-log plus Go admin module. |
| ELK Stack / Elastic Stack | Infrastructure | Log aggregation, search, dashboards | Can store audit-like events | Elasticsearch | Kibana | Large external platform, not an embeddable Go component with service interfaces. |

How this project is different:

- Combines application logging and audit trail in one module.
- Includes PostgreSQL persistence and SQLite default persistence.
- Includes admin APIs for list/get/export.
- Includes React admin frontend.
- Includes admin authentication.
- Includes CBSE documentation and UML diagrams.
- Designed as a reusable educational component, not only a logging library or external observability platform.

## 26. Changelog and License

Changelog:

- Present in `docs/changelog.md`.
- Current entries include `Unreleased`, `v1.1.0 - SQLite default`, and `v1.0.0 - Initial release`.

Current version:

- The frontend package declares version `1.0.0`.
- The Go module does not declare an application version constant.
- Documentation changelog lists `v1.1.0` and unreleased changes.

Initial release details from changelog:

- Application logging with PostgreSQL persistence.
- Audit trail actions.
- Sensitive data sanitization.
- HTTP middleware.
- Admin REST API.
- JSON/CSV export.
- Flyway-style SQL files and Go auto-migration.
- Docker Compose.
- OpenAPI and CBSE documentation.

License:

- MIT License text exists in `docs/license.md`.
- Not implemented yet / recommended improvement: add a root-level `LICENSE` file for standard repository discovery.

## 27. Final CBSE Evaluation Summary

This project satisfies the main CBSE requirements because it is a reusable component with clear provided interfaces, concrete implementations, documentation, diagrams, tests, and deployment support.

CBSE strengths:

- It is a reusable logging and audit component.
- It has clear interfaces: `Logger`, `LogRepository`, `Auditor`, `AuditRepository`, and admin auth repository interfaces.
- It has internal architecture with separated config, database, service, repository, middleware, handler, exporter, and frontend layers.
- It has user documentation, API documentation, and usage examples.
- It has installation and run instructions through README, Makefile, Docker, and frontend docs.
- It has UML support with class, sequence, state, and deployment diagrams.
- It has unit tests and PostgreSQL integration tests.
- It has a competing component analysis.
- It solves a real business problem: secure technical logging and traceable user activity for debugging, monitoring, security, and compliance.

Recommended improvements before production use:

- Move reusable packages from `internal/` to a public `pkg/` path if external module import is required.
- Harden authentication with RBAC, secure token/session design, HTTPS, rate limiting, and secure secret management.
- Implement log retention, streaming export, and optional async logging.
- Update UML diagrams to include SQLite, frontend container, and admin authentication.
- Fix demo audit response IDs by returning the generated event ID from `AuditService.Record`.
