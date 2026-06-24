#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "=== Smoke + unit tests (campanhas / obras-plus) ==="
go test ./server/repositories/... ./server/services/... ./server/templates/core/... ./server/... -count=1 -cover

echo
echo "=== OK ==="
