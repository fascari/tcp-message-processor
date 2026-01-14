#!/usr/bin/env bash
set -euo pipefail

CLIENT_COUNT=${CLIENT_COUNT:-3}
SERVER_HOST=${SERVER_HOST:-localhost}
SERVER_PORT=${SERVER_PORT:-8888}
SUBMISSION_MIN_SECONDS=${SUBMISSION_MIN_SECONDS:-1}
SUBMISSION_MAX_SECONDS=${SUBMISSION_MAX_SECONDS:-1}
LOG_DIR=${LOG_DIR:-logs}

mkdir -p "$LOG_DIR"

start_client() {
  local username=$1
  LOG_FILE="$LOG_DIR/${username}.log"
  echo "Starting client ${username} -> ${LOG_FILE}"
  CLIENT_USERNAME="$username" \
  SERVER_HOST="$SERVER_HOST" \
  SERVER_PORT="$SERVER_PORT" \
  SUBMISSION_MIN_SECONDS="$SUBMISSION_MIN_SECONDS" \
  SUBMISSION_MAX_SECONDS="$SUBMISSION_MAX_SECONDS" \
  go run cmd/client/main.go > "$LOG_FILE" 2>&1 &
}

for idx in $(seq 1 "$CLIENT_COUNT"); do
  username="client${idx}"
  start_client "$username"
done

echo "Spawned ${CLIENT_COUNT} clients. Logs: ${LOG_DIR}/client*.log"

