# TCP Message Processor

A stateful TCP message processing server with authentication, job distribution, cryptographic validation, and asynchronous statistics aggregation.

## Features

- Long-lived TCP connections with concurrent client support
- Session-based authentication with username tracking
- Periodic job distribution with rotating server nonces (configurable interval, default 30 seconds)
- Cryptographic validation using SHA256 hashing
- Rate limiting and duplicate detection
- Asynchronous event processing with RabbitMQ
- Statistics aggregation in PostgreSQL

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Make

## Quick Start

```bash
# Run server locally (starts dependencies and server)
make run-local

# Stop dependencies
make down
```

The server will be available at `localhost:8888`

See [implementation.md](docs/implementation.md) for complete requirements validation and architecture overview.

## Configuration

Configuration is managed through `env.yaml` file or environment variables. Default values:

```yaml
db_host: localhost
db_port: "5433"
db_user: tcpuser
db_password: tcppass
db_name: tcpprocessor

server_port: "8888"
server_host: 0.0.0.0
broadcast_interval_seconds: 30

rabbitmq_url: amqp://guest:guest@localhost:5672/
```

### Environment Variables

- `BROADCAST_INTERVAL_SECONDS`: Job broadcast interval in seconds (default: 30)

## Protocol

The server uses a JSON-RPC 2.0 style TCP protocol on port `8888`.

### Message Flow

1. **Authorize**: `{"id":1, "method":"authorize", "params":{"username":"user"}}`
2. **Job**: Server broadcasts `{"id":null, "method":"job", "params":{"job_id":1, "server_nonce":"123"}}`
3. **Submit**: `{"id":2, "method":"submit", "params":{"job_id":1, "client_nonce":"456", "result":"SHA256(123456)"}}`

### Validation Rules

- Job must exist in session history
- SHA256 hash must match `SHA256(server_nonce + client_nonce)`
- Max 1 submission per second per user
- Client nonce cannot be reused

## Development


### Available Commands

```bash
make help            # Show all commands
make run-local       # Run server locally (with dependencies)
make down            # Stop dependencies
make test            # Run all tests
make lint            # Run linter
make lint-fix        # Run linter with auto-fix
make mocks           # Generate mocks
```

### Monitoring

- **RabbitMQ UI**: http://localhost:15672 (guest/guest)
- **Database**: `docker-compose exec postgres psql -U tcpuser -d tcpprocessor`

### Testing

Run unit tests:
```bash
make test
```

Tests cover:
- Session domain logic and validation
- SHA256 hash calculations
- Rate limiting
- Duplicate detection



