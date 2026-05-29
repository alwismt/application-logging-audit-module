# Analysis of Similar, Competing Components

## Summary comparison

| Component | Focus | Audit trail | Persistence | HTTP/API | Documentation style |
|-----------|-------|-------------|-------------|----------|---------------------|
| **This module** | Logs + audit | Yes | PostgreSQL | Admin REST + middleware | CBSE + OpenAPI + UML |
| Logrus | App logging | No | None (stdout/files) | No | README, godoc |
| Zap | High-perf logging | No | Optional sinks | No | README, benchmarks |
| Zerolog | JSON logging | No | Optional | No | Minimal API docs |
| Log4j (Java) | App logging | No | Appenders | No | XML config, wiki |
| Winston (Node) | App logging | No | Transports | No | npm README |
| Audit.NET (.NET) | Audit trail | Yes | DB/NoSQL | Limited | GitHub wiki |
| ELK Stack | Log aggregation | Search only | Elasticsearch | Kibana UI | Elastic docs |

## Logrus (Go)

- **Use:** Structured logging to stdout or hooks
- **Audit:** No first-class audit events
- **Install:** `go get github.com/sirupsen/logrus`
- **Docs:** README + examples on GitHub
- **Difference:** No built-in PostgreSQL schema, sanitization, or admin API

## Zap (Go)

- **Use:** High-performance production logging
- **Audit:** Not designed for compliance audit trails
- **Install:** `go get go.uber.org/zap`
- **Docs:** Excellent performance notes; minimal audit guidance
- **Difference:** Optimized for speed, not dual log+audit domain model

## Zerolog (Go)

- **Use:** Zero-allocation JSON logs
- **Audit:** No
- **Install:** `go get github.com/rs/zerolog`
- **Docs:** Compact API reference
- **Difference:** Logger-only; no HTTP middleware or export endpoints

## Log4j (Java)

- **Use:** Enterprise Java logging with appenders
- **Audit:** Separate concern (often custom)
- **Install:** Maven/Gradle dependency
- **Docs:** Extensive XML configuration guides
- **Difference:** Java ecosystem; not a Go reusable component

## Winston (Node.js)

- **Use:** Flexible transports (file, console, cloud)
- **Audit:** Requires custom implementation
- **Install:** `npm install winston`
- **Docs:** npm package page + recipes
- **Difference:** Node-specific; no unified audit schema

## Audit.NET (.NET)

- **Use:** Change tracking and audit events in .NET apps
- **Audit:** Primary focus
- **Install:** NuGet package
- **Docs:** Wiki with provider configuration
- **Difference:** .NET only; less emphasis on application debug logging

## ELK Stack (Elastic)

- **Use:** Centralized log search and dashboards
- **Audit:** Can index audit-like events but not a library component
- **Install:** Docker/Kubernetes stack
- **Docs:** Official Elastic documentation
- **Difference:** Infrastructure product, not an embeddable Go module

## How this project improves on research findings

1. **Combines** application logging and audit trails in one component
2. **Documents** with CBSE structure, UML (source + PNG), and OpenAPI
3. **Ships** Docker Compose for one-command demo (`docker compose up`)
4. **Exposes** admin APIs for UI integration without requiring Kibana
5. **Implements** security sanitization at the service layer by default
