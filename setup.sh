#!/usr/bin/env bash
# Bootstraps the avacado dev environment: installs missing dev tools, downloads
# Go module dependencies, and wires up the repo's git hooks.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

echo "🚀 Setting up avacado dev environment..."
echo ""

if ! command -v go >/dev/null 2>&1; then
  echo "❌ Go is not installed. Install it from https://go.dev/dl/ and re-run this script."
  exit 1
fi
echo "✅ go: $(go version)"

GOBIN="$(go env GOPATH)/bin"
export PATH="$GOBIN:$PATH"

# Persist GOBIN on PATH for later steps in a GitHub Actions job (each `run:` step
# is a fresh shell, so the `export` above only applies within this script).
if [ -n "${GITHUB_PATH:-}" ]; then
  echo "$GOBIN" >>"$GITHUB_PATH"
fi

echo ""
echo "📦 Downloading Go module dependencies..."
go mod download

echo ""
if command -v mockgen >/dev/null 2>&1; then
  echo "✅ mockgen already installed: $(command -v mockgen)"
else
  echo "🔧 Installing mockgen..."
  MOCK_VERSION="$(go list -m -f '{{.Version}}' go.uber.org/mock)"
  go install "go.uber.org/mock/mockgen@${MOCK_VERSION}"
fi

echo ""
if command -v golangci-lint >/dev/null 2>&1; then
  echo "✅ golangci-lint already installed: $(golangci-lint version 2>&1 | head -n1)"
else
  echo "🔧 Installing golangci-lint..."
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN"
fi

echo ""
if command -v govulncheck >/dev/null 2>&1; then
  echo "✅ govulncheck already installed: $(command -v govulncheck)"
else
  echo "🔧 Installing govulncheck..."
  go install golang.org/x/vuln/cmd/govulncheck@latest
fi

echo ""
echo "🪝 Configuring git hooks (githooks/)..."
chmod +x githooks/*
git config core.hooksPath githooks
echo "✅ git hooks configured: pre-commit will run make test-short, make lint, make vulncheck"

echo ""
echo "✨ Setup complete!"

if ! command -v mockgen >/dev/null 2>&1 || ! command -v golangci-lint >/dev/null 2>&1 || ! command -v govulncheck >/dev/null 2>&1; then
  echo "⚠️  $GOBIN is not on your PATH — add it to your shell profile so the installed tools can be found."
fi
