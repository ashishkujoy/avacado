# Avacado

Avacado is a from-scratch Redis-server reimplementation in Go: a RESP protocol parser/serializer, a
single-threaded command executor, and in-memory storage engines for strings, lists, and hashmaps. It
speaks the real RESP wire protocol, so any `redis-cli` or `go-redis` client can talk to it (see
[`commands.md`](commands.md) for the supported command list and [`todo_commands.md`](todo_commands.md)
for what's still unimplemented).

## Prerequisites

- Go (see [`go.mod`](go.mod) for the required version)

Everything else — `mockgen`, `golangci-lint`, `govulncheck` — is installed for you by the setup script below.

## Setup

After cloning, run:

```bash
./setup.sh
```

This installs any missing dev tools (`mockgen`, `golangci-lint`, `govulncheck`), downloads Go module
dependencies, and configures the repo's git hooks (`git config core.hooksPath githooks`) so that
`git commit` automatically runs `make test-short`, `make lint`, and `make vulncheck` — a commit is
blocked if any of them fail. The script is idempotent, so re-running it is safe (already-installed tools
are skipped). `make setup` runs the same script.

<details>
<summary>Installing tools manually</summary>

- [`golangci-lint`](https://golangci-lint.run/docs/welcome/install/local/) — required to run `make lint`.
- [`govulncheck`](https://go.dev/doc/tutorial/govulncheck) — required to run `make vulncheck`. Install with:
  ```bash
  go install golang.org/x/vuln/cmd/govulncheck@latest
  ```
- [`mockgen`](https://github.com/uber-go/mock) — required to regenerate mocks (`make mocks` / `make test`),
  invoked via `go generate ./...`. Install the version matching `go.uber.org/mock` in [`go.mod`](go.mod):
  ```bash
  go install go.uber.org/mock/mockgen@latest
  ```
- `redis-cli` — optional, for manually poking the server (see Getting Started below). Part of the
  [Redis](https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/) distribution
  (e.g. `brew install redis` on macOS).
- `redis-server` / `redis-benchmark` — optional, only needed to run the `/benchmark` skill, which compares
  avacado against real Redis. Also part of the Redis distribution above.

</details>

## Getting Started

Run the server locally:

```bash
go run cmd/server/main.go --port 6379
redis-cli -p 6379 PING
```

## Development

```bash
make test            # go generate ./... (mocks) + go test -v ./...
make test-short       # same, without -v
make test-coverage    # generates coverage.out / coverage.html
make mocks             # regenerate mocks only (go generate ./...)
make lint               # run golangci-lint
make vulncheck           # run govulncheck
make clean              # remove coverage artifacts
```

Run a single test:

```bash
go test ./internal/command/kv/... -run TestGet -v
go test ./integration/command/hashmap/... -run TestHSet_SetSingleField -v
```

Before opening a PR, make sure `make test`, `make lint`, and `make vulncheck` all pass with no issues.

See [`CLAUDE.md`](CLAUDE.md) for the architecture overview and conventions used throughout the codebase.
