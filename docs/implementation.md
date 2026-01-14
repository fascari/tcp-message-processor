# Implementation Overview

How the TCP server meets all project requirements.

## Core Concepts

### Persistent Connections
Connections stay open after authentication. Each client runs in its own goroutine reading messages in a loop.

**Implementation:** `server/handler.go` - one goroutine per connection

### Session Management
Sessions track username, job history, submissions, and rate limits. In-memory storage with `sync.RWMutex` for concurrent access.

**Implementation:** `session/session.go` and `session/store.go`

### Job Broadcasting
Every 30 seconds, a new job with fresh nonce is broadcast to all connected clients. Each client's session stores job history.

**Implementation:** `server/broadcaster.go` - timer triggers broadcast every 30s

### Rate Limiting
Per-user rate limiting using `golang.org/x/time/rate` (Token Bucket algorithm). Each user limited to 1 submission/second with independent limiters.

**Why this library:**
- Industry standard Token Bucket algorithm
- Thread-safe, no race conditions
- Auto token regeneration, no memory leaks
- Battle-tested in production

**Implementation:** `pkg/ratelimit/ratelimit.go`

### Submission Validation
Validates job_id/nonce pair, SHA256 hash, rate limits, and duplicate nonces. Returns specific error messages.

**Errors:**
- `"task does not exist"` - invalid job_id or wrong nonce
- `"task expired"` - old job_id (newer job received)
- `"invalid result"` - SHA256 mismatch
- `"submission too frequent"` - rate limit hit
- `"duplicate submission"` - nonce reused

**Implementation:** `server/submit.go`, `pkg/hash/sha256.go`

### Async Stats Processing
Submissions publish events to RabbitMQ for immediate client response. Consumer writes to PostgreSQL asynchronously.

**Flow:** Client → Validation → RabbitMQ → Immediate Response  
**Consumer:** Picks up events → Writes to PostgreSQL

**Implementation:** `infra/publisher/` and `infra/consumer/`

## Requirements Checklist

### Authentication
- Persistent TCP connections  
- Track username per session  
- Concurrent sessions  
- Validate authentication requests

**Location:** `server/handler.go`, `server/authorize.go`, `session/`

### Job Distribution
- Generate job with nonce every 30s  
- Broadcast to all clients  
- Track job history per session

**Location:** `server/broadcaster.go`, `session/session.go`

### Submissions
- Validate job_id/server_nonce combination  
- Verify SHA256 hash  
- Enforce 1/second rate limit  
- Detect duplicate client_nonce  
- Expired job detection

**Location:** `server/submit.go`, `pkg/ratelimit/`, `pkg/hash/`

### Statistics
Count submissions per user  
Group by minute timestamp  
Store in PostgreSQL

**Location:** `stats/stats.go`, `db/migrations/001_initial_schema.sql`

### Message Queue
- RabbitMQ integration  
- Non-blocking responses  
- Async persistence  
- Fault tolerance

**Location:** `infra/publisher/`, `infra/consumer/`, `pkg/broker/`

## Concurrency Model

- **Per-client goroutine:** Handles connection lifecycle
- **Broadcaster goroutine:** Sends jobs every 30s
- **Consumer goroutine:** Processes RabbitMQ events
- **Session store:** Protected by `sync.RWMutex`
- **Rate limiters:** Thread-safe per-user

