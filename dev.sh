#!/usr/bin/env bash
set -euo pipefail

BINARY="valence"

echo "==> Running tests..."
go test ./... -v

echo ""
echo "==> Building $BINARY..."
go build -o "$BINARY" .

echo ""
echo "==> Done. Run ./$BINARY -h to get started."
