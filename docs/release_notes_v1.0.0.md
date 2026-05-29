# Application Logging & Audit Trail Module v1.0.0

Initial release of the Application Logging & Audit Trail Module.

## Included

- Go backend server
- React admin frontend
- Structured application logging
- Audit trail event recording
- SQLite and PostgreSQL support
- Embedded database migrations
- REST API endpoints
- JWT-based admin authentication
- API key support for admin routes
- Log and audit event filtering
- JSON and CSV export
- Sensitive data sanitization
- Docker Compose setup
- Documentation and examples

## Release assets

This release provides:

- Backend binaries for macOS, Linux, and Windows
- SHA256 checksums
- Full Docker Compose ZIP package for easy local/demo deployment

## How to run the full package

Download and extract:

application-logging-audit-module-v1.0.0.zip

Then run:

```bash
./start.sh
```

Or manually:

```bash
docker compose up --build -d
```

Health check:

```bash
curl http://localhost:8080/health
```

## Important notes

The standalone binaries run the Go backend server only.

For the complete backend + frontend experience, use the Docker Compose release ZIP package.

Before production use, update:

- ADMIN_USERNAME
- ADMIN_PASSWORD
- ADMIN_API_KEY
- JWT_SECRET
- CORS_ORIGINS
- Database credentials

## Known limitations

- Advanced role-based access control is not implemented yet.
- Asynchronous high-volume logging is not implemented yet.
- Multi-tenant separation is not implemented yet.
- Automatic log retention and archival policies are future work.
