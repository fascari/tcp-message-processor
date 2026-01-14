# TCP Message Processor - Server

TCP server for authentication, job distribution, and result validation.

> See [main README](../README.md) for overview, technical decisions, and prerequisites.

## Quick Start

```bash
# Start server (includes dependencies)
make run-local

# Stop dependencies
make down
```

Server runs on `localhost:8888`

## Code Reference

### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| TCP Handler | `internal/handler/` | Connection management, message routing |
| Session Store | `internal/session/` | Session state, job history, rate limiting |
| Job Broadcaster | `internal/broadcaster/` | Periodic job distribution (30s) |
| Rate Limiter | `pkg/ratelimit/` | Per-user submission throttling |
| RabbitMQ Publisher | `internal/infra/publisher/` | Async event publishing |
| RabbitMQ Consumer | `internal/infra/consumer/` | Event processing, DB persistence |
| Database | `internal/infra/database/` | PostgreSQL connection |
| Broker Abstraction | `pkg/broker/` | RabbitMQ setup utilities |

### Concurrency Model

- **Per-client goroutine** - Handles connection lifecycle
- **Broadcaster goroutine** - Sends jobs every 30s
- **Consumer goroutine** - Processes RabbitMQ events
- **Session store** - Protected by `sync.RWMutex`
- **Rate limiters** - Thread-safe per-user (`golang.org/x/time/rate`)

## Configuration

Via `env.yaml` or environment variables (uppercase with underscores):

```yaml
db_host: localhost
db_port: "5432"
db_user: tcpuser
db_password: tcppass
db_name: tcpprocessor
db_sslmode: disable

server_host: 0.0.0.0
server_port: "8888"
broadcast_interval_seconds: 30

rabbitmq_url: amqp://guest:guest@localhost:5672/
```

Key variables:
- `BROADCAST_INTERVAL_SECONDS` - Job broadcast interval (default: 30)
- `DB_SSLMODE` - PostgreSQL SSL mode

## Protocol

### Authentication

```json
// Request
{"id": 1, "method": "authorize", "params": {"username": "admin"}}

// Response
{"id": 1, "result": true}
```

### Job Broadcast

```json
{"id": null, "method": "job", "params": {"job_id": 1, "server_nonce": "abc123"}}
```

### Result Submission

```json
// Request
{"id": 2, "method": "submit", "params": {
  "job_id": 1,
  "client_nonce": "def456",
  "result": "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"
}}

// Success
{"id": 2, "result": true}

// Error
{"id": 2, "result": false, "error": "Submission too frequent"}
```

### Validation

- Job must exist in session
- Hash: `SHA256(server_nonce + client_nonce)`
- Rate limit: 1 submission/second
- No duplicate nonces

### Errors

| Error | Description |
|-------|-------------|
| `Task does not exist` | Invalid job_id |
| `Invalid result` | Wrong SHA256 hash |
| `Submission too frequent` | Rate limit exceeded |
| `Duplicate submission` | Nonce reused |

## Commands

```bash
make run-local       # Run server with dependencies
make down            # Stop dependencies
make test            # Run unit tests
make lint            # Run linter
make clean           # Clean build artifacts
```

## Monitoring

**RabbitMQ Management**: http://localhost:15672 (guest/guest)

**PostgreSQL**:
```bash
docker exec -it tcp-message-processor-postgres-1 psql -U tcpuser -d tcpprocessor

# View recent submissions
SELECT * FROM submissions ORDER BY timestamp DESC LIMIT 10;
```

## Testing

```bash
make test
```

Coverage: Session management, SHA256, rate limiting, duplicate detection.

