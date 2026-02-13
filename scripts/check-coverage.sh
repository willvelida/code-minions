#!/usr/bin/env bash
# check-coverage.sh — Shared coverage threshold check.
# Used by both CI and the pre-push git hook.
#
# Usage: ./scripts/check-coverage.sh [threshold]
#   threshold: minimum coverage percentage (default: 70)

set -euo pipefail

THRESHOLD="${1:-70}"
COVER_OUT="$(mktemp)"

trap 'rm -f "$COVER_OUT"' EXIT

echo "Running tests with coverage..."
go test -coverprofile="$COVER_OUT" ./...

TOTAL=$(go tool cover -func="$COVER_OUT" | grep '^total:' | awk '{print $NF}' | tr -d '%')

echo ""
echo "Total coverage: ${TOTAL}%"
echo "Threshold:      ${THRESHOLD}%"

if awk "BEGIN {exit !(${TOTAL} < ${THRESHOLD})}"; then
  echo "❌ Coverage ${TOTAL}% is below the ${THRESHOLD}% threshold."
  exit 1
fi

echo "✅ Coverage ${TOTAL}% meets the ${THRESHOLD}% threshold."
