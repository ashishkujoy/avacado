#!/usr/bin/env zsh

set -e  # Exit on error

echo "🔧 Generating mocks..."
go generate ./...

echo ""
echo "✅ Mocks generated successfully!"
echo ""

echo "🧪 Running tests..."
go test -v ./...

echo ""
echo "✨ All tests completed successfully!"

