# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Avacado is a from-scratch Redis-server reimplementation in Go: a RESP protocol parser/serializer, a single-threaded
command executor, and in-memory storage engines for strings, lists, and hashmaps. It speaks the real RESP wire
protocol, so any `redis-cli` or `go-redis` client can talk to it (see `commands.md` for the supported command list
and `todo_commands.md` for what's still unimplemented).

## Commands

Run `./setup.sh` (or `make setup`) once to install missing dev tools (`mockgen`, `golangci-lint`,
`govulncheck`), download module dependencies, and configure git hooks — a `pre-commit` hook
(`githooks/pre-commit`, wired via `git config core.hooksPath githooks`) runs `make test-short`, `make lint`,
and `make vulncheck` on every commit.

```bash
make test            # go generate ./... (mocks) + go test -v ./...
make test-short       # same, without -v
make test-coverage    # generates coverage.out / coverage.html
make mocks             # regenerate mocks only (go generate ./...)
make lint               # run golangci-lint
make vulncheck           # run govulncheck
make clean              # remove coverage artifacts
```

**Always run `make lint` and `make vulncheck` after finishing an edit task** (before considering the task done)
and fix any issues they report.

Run a single test:
```bash
go test ./internal/command/kv/... -run TestGet -v
go test ./integration/command/hashmap/... -run TestHSet_SetSingleField -v
```

Run the server locally:
```bash
go run cmd/server/main.go --port 6379
redis-cli -p 6379 PING
```

**Always regenerate mocks after touching any interface with a `//go:generate mockgen` directive** (storage
interfaces, `command.Command`/`Parser`, `protocol.Protocol`) — `make test` does this automatically via the `mocks`
prerequisite, but running `go test` directly will not.

## Architecture

Request flow: `net.Conn` → `protocol.Protocol` (RESP parse/serialize) → `command.ParserRegistry` (raw args →
`Command`) → `executor.Executor` (single goroutine, serialized execution) → `storage.Storage` (in-memory engines).

- **`internal/protocol/resp/`** — RESP wire format parsing/serialization. `protocol.Value`/`Message`/`Response` in
  `internal/protocol/protocol.go` are the transport-agnostic types every other layer works with.
- **`internal/command/<group>/`** (`kv`, `kv/expiry`, `list`, `hashmap`, `connection`) — one file per Redis command,
  each defining a `Command` struct (holds parsed args, implements `Execute(ctx, storage) *protocol.Response`) and a
  matching `Parser` struct (`Parse(*protocol.Message) (Command, error)` + `Name() string`). `internal/command/kv/get.go`
  is the canonical minimal example.
- **`internal/command/registry/registry.go`** — `DefaultParserRegistry` maps uppercased command name → `Parser`. Every
  new command must be registered here or the server returns "unknown command".
- **`internal/executor/`** — `Executor.Run` drains a single channel in one goroutine, so `storage.Storage` needs no
  internal locking. It also owns the BLPOP/BRPOP blocked-client queue: a `Command` that pushes to a list can
  implement `PushedKey() string` to trigger `tryUnblockClient`, and blocking commands register themselves via the
  `command.BlockRegistry` interface (injected into `ctx`) to get a channel + cancel func back.
- **`internal/storage/storage.go`** — top-level `Storage` interface aggregating `KV()`, `Lists()`, `Maps()`. Each data
  type lives in its own package (`internal/storage/kv`, `internal/storage/lists`, `internal/storage/hashmaps`) with
  a `Store`-like interface plus a `memory/` subpackage holding the actual implementation (`DefaultStorage` in
  `storage.go` wires the memory implementations together). Lists use a quicklist/listpack encoding
  (`internal/storage/listpack/`, `internal/storage/lists/memory/quicklist.go`) once past a size threshold
  (`MAX_LIST_PACK_SIZE` env var, default 8192).
- **`internal/server/server.go`** — per-connection loop: parse → route through registry → submit to executor →
  serialize response. Handles the `BlockCh` case on `*protocol.Response` for blocking commands.
- Every storage/command/protocol interface carries a `//go:generate mockgen ...` comment; mocks live in a sibling
  `mock/` package (e.g. `internal/storage/kv/mock/store.go`).

### Adding a new command

Follow the existing pattern rather than improvising: storage interface method (if needed) → memory implementation →
regenerate mocks → command file (struct + `Execute` + `Parser` + `Name()`) → register in
`internal/command/registry/registry.go` → integration test under `integration/command/<group>/`.

This repo has slash-command skills that automate exactly this workflow — `/redis-command-planner`,
`/implementation-task-planner`, and `/implement-redis-command` (which chains the first two, then executes the
generated task list, running `make test` after every task). Prefer these when implementing a new command end-to-end.

## Testing layout

- Unit tests sit next to the code they test (`*_test.go` per command/storage file), using mocks from the sibling
  `mock/` packages for isolation (e.g. command tests mock `storage.Storage`).
- **`integration/`** spins up a real server on a loopback TCP port (`integration.StartNewServer(port)`) and drives it
  with an actual `go-redis` client — this is what exercises the full RESP wire protocol. Each command-group package
  under `integration/command/<group>/` has its own `TestMain` starting the server on a fixed port (check existing
  files for the port already in use before picking a new one, e.g. hashmap tests use `6003`).
