# Changelog

## Unreleased

- Admin JWT authentication and `admin_users` table (default credentials `admin` / `12345678`)
- React + Vite admin frontend with login, logs, audit views, filters, and export
- CORS support, date-range query filters, Docker frontend service

## v1.1.0 — SQLite default

- SQLite as default database (`DB_DRIVER=sqlite`, `SQLITE_PATH`); PostgreSQL via `DB_DRIVER=postgres`
- SQLite and Postgres repository implementations; driver-specific embedded migrations
- Docker Compose: SQLite-only stack by default; `docker-compose.postgres.yml` overlay for PostgreSQL

## v1.0.0 — Initial release

- Application logging (INFO, WARNING, ERROR, DEBUG) with PostgreSQL persistence
- Audit trail recording (LOGIN, LOGOUT, CREATE_RECORD, UPDATE_RECORD, DELETE_RECORD, DOWNLOAD_FILE, FAILED_LOGIN, PERMISSION_DENIED)
- Sensitive data sanitization for passwords, tokens, authorization headers, secrets, and card numbers
- HTTP middleware for request logging with request ID, latency, and status capture
- Admin REST API for querying and exporting logs and audit events (JSON/CSV)
- Flyway-style SQL files (optional CLI) and Go auto-migrate on startup
- Docker Compose: internal PostgreSQL + Go app (host port 8080 only)
- OpenAPI specification (`docs/swagger.yaml`)
- CBSE documentation and PlantUML diagrams with PNG exports
