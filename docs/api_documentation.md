# API Documentation

> Machine-readable spec: [swagger.yaml](swagger.yaml) (OpenAPI 3.0). Import into Swagger UI or Postman.

## Swagger UI (optional, dev)

Interactive docs are served by the Go server when enabled:

1. Set `ENABLE_SWAGGER_UI=true` in `.env` (default is `false`).
2. Run `make run` (or `make build` then start the binary).
3. Open [http://localhost:8080/swagger/](http://localhost:8080/swagger/) in a browser.

After editing [swagger.yaml](swagger.yaml), run `make sync-swagger` so the embedded copy under `internal/swagger/openapi.yaml` stays in sync (also runs automatically for `make build` and Docker builds).

Do not enable Swagger UI in production deployments.

## Public types

### LogEntry

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| level | string | INFO, WARNING, ERROR, DEBUG |
| message | string | Log message |
| source | string | Service/component name |
| request_id | string | Correlation ID |
| user_id | UUID? | Optional user |
| metadata | object | JSON metadata (sanitized) |
| created_at | datetime | UTC timestamp |

### AuditEvent

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID? | Acting user |
| username | string | Display name |
| action | string | LOGIN, UPDATE_RECORD, etc. |
| resource_type | string | e.g. invoice |
| resource_id | string | Resource identifier |
| old_value, new_value | object | Change snapshots |
| status | string | SUCCESS or FAILURE |
| created_at | datetime | UTC timestamp |

## Service interfaces

### Logger

```go
Info(ctx context.Context, message string, metadata map[string]any) error
Warn(ctx context.Context, message string, metadata map[string]any) error
Error(ctx context.Context, message string, err error, metadata map[string]any) error
Debug(ctx context.Context, message string, metadata map[string]any) error
```

### Auditor

```go
Record(ctx context.Context, event AuditEvent) error
Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
```

## REST routes

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health + DB status |
| POST | /logs/log-info | Create INFO log |
| POST | /logs/log-error | Create ERROR log |
| POST | /logs/audit-login | Record LOGIN audit |
| POST | /logs/audit-update | Record UPDATE_RECORD audit |
| POST | /admin/login | Admin login (returns JWT) |
| GET | /admin/logs | List logs (requires auth; filters: level, source, user_id, request_id, from, to, page, limit) |
| GET | /admin/logs/{id} | Get log by ID |
| GET | /admin/logs/export | Export (?format=json\|csv) |
| GET | /admin/audit-events | List audit events (filters: user_id, username, action, resource_type, status, request_id, from, to, page, limit) |
| GET | /admin/audit-events/{id} | Get audit event |
| GET | /admin/audit-events/export | Export audit events |

## Admin authentication

Admin routes (except `POST /admin/login`) require authentication:

- **JWT:** `Authorization: Bearer <token>` from login response
- **API key:** `X-API-Key: <ADMIN_API_KEY>` or `Authorization: ApiKey <key>`

Demo defaults (when env vars are unset): username `admin`, password `12345678`.

### Login

```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}'
```

### List ERROR logs

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"12345678"}' | jq -r .token)

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/admin/logs?level=ERROR&page=1&limit=20"
```

### Record demo login audit

```bash
curl -X POST http://localhost:8080/logs/audit-login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","ip_address":"192.168.1.10"}'
```

### Export audit events as CSV

```bash
curl "http://localhost:8080/admin/audit-events/export?format=csv&action=LOGIN" -o audit.csv
```
