# TCP Message Processor - Client

TCP client for job processing and result submission.

> See [main README](../README.md) for overview, technical decisions, and prerequisites.

## Quick Start

```bash
# Run with default username ($USER)
make run-local

# Run with custom username
CLIENT_USERNAME=miner1 make run-local

# Run multiple clients
make run-multi
```

## Configuration

Via `env.yaml` or environment variables (uppercase with underscores):

```yaml
server_host: localhost
server_port: "8888"
client_username: client1
submission_min_seconds: 1
submission_max_seconds: 60
```

Key variables:
- `CLIENT_USERNAME` - Client username (default: $USER)
- `SUBMISSION_MIN_SECONDS` / `SUBMISSION_MAX_SECONDS` - Delay range

## Usage

```bash
# Single client
make run-local

# Multiple clients (default: 3)
make run-multi

# Custom count
CLIENT_COUNT=5 make run-multi

# Stop all
make stop-clients
```

Logs: `logs/client*.log`

## Commands

```bash
make run-local       # Run single client
make run-multi       # Run multiple clients
make stop-clients    # Stop all running clients
make test            # Run unit tests
make test-integration # Run integration tests (with Testcontainers)
make lint            # Run linter
make clean           # Clean artifacts
```

## Testing

```bash
# Unit tests
make test

# Integration tests (Testcontainers)
make test-integration
```

Coverage: Job manager, nonce generation, random delay, hash calculation, rate limiting.
