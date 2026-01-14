# TCP Message Processor

TCP-based message processing system with persistent connections, authentication, job distribution, and cryptographic validation.

## Overview

This project implements a client-server TCP system that handles:
- Client authentication
- Task distribution (jobs)
- Result submission and validation
- Per-session rate limiting
- Statistics persistence
- Asynchronous processing via message broker

## Architecture

### Go Workspaces Monorepo

The project uses **Go Workspaces** (`go.work`) to organize multiple related modules:

```
tcp-message-processor/
├── server/          # TCP Server
├── client/          # TCP Client
└── common/          # Shared code
```

**Why Go Workspaces?**

1. **Integrated Development** - Changes to `common` module are reflected immediately in `server` and `client` without publishing versions
2. **Shared Code** - `common/pkg` contains utilities reused by both (logger, hash, tcp protocol, etc.)
3. **Simplified Testing** - Integration tests can import and use code from multiple modules
4. **Unified Versioning** - Easier to maintain compatibility between components
5. **Developer Experience** - IDEs and Go tools automatically recognize local dependencies

## Technical Decisions

### Database: Raw SQL vs GORM

**Decision: Use `database/sql` with PostgreSQL driver directly**

While GORM is commonly used in conventional REST APIs and I have experience with it, this project doesn't benefit from an ORM due to its simple data model and write-only operations.

**Why NOT use GORM here:**

1. **Single Write Operation** - Only INSERT statements, no complex queries or relationships
2. **No Domain Model** - Statistics are fire-and-forget aggregations, not domain entities
3. **Asynchronous Only** - Consumer writes to DB without business logic or validation
4. **No Migrations Needed** - Single static table, schema defined in SQL file
5. **Unnecessary Abstraction** - GORM features (associations, hooks, query builder) unused

**When GORM makes sense:**
- REST APIs with full CRUD operations
- Complex domain models with relationships (1:N, N:M)
- Dynamic query building based on filters
- Need for database abstraction (multi-DB support)
- Migration management and schema evolution

### Message Broker: RabbitMQ

Asynchronous submission processing:

- Publisher sends submission events
- Consumer processes and persists to database
- Server doesn't block on database writes
- Messages survive restarts

### Rate Limiting

Uses **`golang.org/x/time/rate`** (stdlib):

- Per-session limiter (1 req/second per client)
- Thread-safe with `sync.Map`
- In-memory storage
- Official Go library

### Integration Tests

Client includes automated integration tests using **Testcontainers**:

- Location: `client/test/integration/ratelimit_test.go`
- Coverage: Rate limiting validation (1 req/second)
- Infrastructure: PostgreSQL, RabbitMQ, TCP server in containers
- Execution: `cd client && make test-integration`

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (optional)
- golangci-lint (for linting)


## Project Structure

```
.
├── go.work                 # Go workspace configuration
├── docker-compose.yml      # Infrastructure (Postgres + RabbitMQ)
├── README.md              # This file
│
├── common/                # Shared code
│   └── pkg/
│       ├── closer/        # Resource cleanup utility
│       ├── env/           # Environment variable helpers
│       ├── errlog/        # Conditional error logging
│       ├── hash/          # SHA256
│       ├── logger/        # Structured logger (zap)
│       ├── noncegen/      # Nonce generation
│       ├── random/        # Random delay
│       └── tcp/           # TCP message protocol
│
├── server/                # TCP Server
│   ├── cmd/server/        # Entry point
│   ├── internal/          # Server logic
│   └── README.md          # Server-specific documentation
│
└── client/                # TCP Client
    ├── cmd/client/        # Entry point
    ├── internal/          # Client logic
    ├── test/integration/  # Integration tests with Testcontainers
    └── README.md          # Client-specific documentation
```

## Quick Start

### 1. Start Infrastructure

```bash
# PostgreSQL + RabbitMQ
docker compose up -d
```

### 2. Run Server

```bash
cd server
make run-local
```

### 3. Run Client(s)

```bash
# Separate terminal - Single client
cd client
make run-local

# Or multiple clients
make run-multi
```

### 4. Check Results

```bash
# Connect to PostgreSQL
docker exec -it tcp-message-processor-postgres-1 psql -U tcpuser -d tcpprocessor

# Query
SELECT * FROM submissions ORDER BY timestamp DESC LIMIT 10;
```

## Configuration

Both modules support configuration via `env.yaml` file or environment variables (overrides file).

See module-specific documentation:
- [Server Configuration](server/README.md#configuration)
- [Client Configuration](client/README.md#configuration)

## Additional Resources

- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **PostgreSQL**: `localhost:5432` (tcpuser/tcppass/tcpprocessor)

## Future Improvements

- **Observability**: Metrics (Prometheus), distributed tracing (OpenTelemetry), dashboards (Grafana)
- **Resilience**: Circuit breaker, retry policies, health checks
- **Scalability**: Load balancer, multiple server instances, Redis for distributed rate limiting
- **Security**: TLS for TCP connections, real authentication (JWT, OAuth), action auditing

