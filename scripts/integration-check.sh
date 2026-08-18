#!/usr/bin/env bash
# End-to-end check against a running docker-compose stack (or any reachable
# deployment): waits for both APIs to report healthy, obtains a JWT from
# go-api, submits a sample matrix, and asserts the combined QR + statistics
# response has the expected shape. Requires curl and jq. Exits non-zero on
# any failure, so it can be used as a CI gate.
set -euo pipefail

GO_API_URL="${GO_API_URL:-http://localhost:8080}"
NODE_API_URL="${NODE_API_URL:-http://localhost:4000}"
GO_API_PRESHARED_KEY="${GO_API_PRESHARED_KEY:-change-me-preshared-key}"

wait_for_health() {
  local name="$1" url="$2" attempts=30
  echo "Waiting for $name at $url/health ..."
  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$url/health" > /dev/null 2>&1; then
      echo "  $name is healthy."
      return 0
    fi
    sleep 1
  done
  echo "  $name did not become healthy after ${attempts}s" >&2
  exit 1
}

wait_for_health "go-api" "$GO_API_URL"
wait_for_health "node-api" "$NODE_API_URL"

echo "Requesting a token from go-api ..."
TOKEN_RESPONSE=$(curl -fsS -X POST "$GO_API_URL/api/v1/auth/token" \
  -H "X-API-Key: $GO_API_PRESHARED_KEY")
TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "Failed to obtain a token. Response: $TOKEN_RESPONSE" >&2
  exit 1
fi
echo "  token obtained."

echo "Submitting sample matrix to go-api ..."
QR_RESPONSE=$(curl -fsS -X POST "$GO_API_URL/api/v1/matrix/qr" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"matrix": [[12, -51, 4], [6, 167, -68], [-4, 24, -41]]}')

echo "Validating response shape ..."
echo "$QR_RESPONSE" | jq -e '.matrices.Q and .matrices.R' > /dev/null \
  || { echo "Missing matrices.Q/R in response: $QR_RESPONSE" >&2; exit 1; }
echo "$QR_RESPONSE" | jq -e '.statistics.perMatrix.Q and .statistics.perMatrix.R' > /dev/null \
  || { echo "Missing statistics.perMatrix.Q/R in response: $QR_RESPONSE" >&2; exit 1; }
echo "$QR_RESPONSE" | jq -e '.statistics.overall.anyDiagonal != null' > /dev/null \
  || { echo "Missing statistics.overall.anyDiagonal in response: $QR_RESPONSE" >&2; exit 1; }

echo "Integration check passed."
echo "$QR_RESPONSE" | jq '.statistics'
