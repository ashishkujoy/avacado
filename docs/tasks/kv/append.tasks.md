# APPEND Implementation Tasks

**Command**: `APPEND`
**Storage type**: `kv`
**Plan**: `docs/plans/kv/append.plan.md`

---

## Task 1 — Storage interface: add Append method signature

**File**: `internal/storage/kv/store.go`

**What to do**: Add the following method to the `Store` interface:

```go
Append(ctx context.Context, key string, value []byte) (int64, error)
```

The file already has the `//go:generate` directive:
```
//go:generate sh -c "rm -f mock/store.go && mockgen -source=store.go -destination=mock/store.go -package=mockkv"
```

**Done when**: File compiles with the new method in the interface.

---

## Task 2 — Regenerate mocks for kv

**What to do**: Run mock generation so the mock reflects the new `Append` method.

```
make clean && make mocks
```

**File updated**: `internal/storage/kv/mock/store.go`

**Done when**: Mock file includes `Append` and `make test` passes.

---

## Task 3 — Implement Append in KVMemoryStore

**File**: `internal/storage/kv/memory/memory.go`

**What to do**: Add the `Append` method to `KVMemoryStore`:

```go
func (k *KVMemoryStore) Append(_ context.Context, key string, data []byte) (int64, error) {
    existing, ok := k.store[key]
    if !ok || existing.isExpired() {
        k.store[key] = &value{data: data, enc: encodingString}
        return int64(len(data)), nil
    }
    current := existing.Bytes()
    appended := append(current, data...)
    k.store[key] = &value{data: appended, enc: encodingString, expiry: existing.expiry}
    return int64(len(appended)), nil
}
```

Key rules:
- Use `existing.Bytes()` to decode the current value (handles `encodingInteger` transparently).
- Always store the result as `encodingString`.
- Preserve `existing.expiry` when appending to a live key.
- When key is absent or expired: no TTL on the new entry.

Also add unit tests in `internal/storage/kv/memory/memory_test.go`:
- Append to absent key → returns len(value), GET returns value
- Append to existing string key → returns cumulative length, GET returns concatenation
- Append to expired key → treated as new key (no TTL preserved)
- Append to integer-encoded key → GET returns original integer string + appended bytes
- TTL preserved after append to live key

**Done when**: `make test` passes.

---

## Task 4 — Implement APPEND command

**File**: `internal/command/kv/append.go`

**What to do**: Create the command file:

```go
package kv

import (
    "avacado/internal/command"
    "avacado/internal/protocol"
    "avacado/internal/storage"
    "context"
)

type Append struct {
    Key   string
    Value []byte
}

func (a *Append) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
    length, err := storage.KV().Append(ctx, a.Key, a.Value)
    if err != nil {
        return protocol.NewErrorResponse(err)
    }
    return protocol.NewNumberResponse(length)
}

type AppendParser struct{}

func NewAppendParser() *AppendParser {
    return &AppendParser{}
}

func (p *AppendParser) Parse(msg *protocol.Message) (command.Command, error) {
    return &Append{Key: msg.Args[0], Value: []byte(msg.Args[1])}, nil
}

func (p *AppendParser) Name() string {
    return "APPEND"
}
```

Also add unit tests in `internal/command/kv/append_test.go`:
- `TestAppendParser_Parse` — produces correct `Append` struct
- `TestAppend_Execute` — mock store returns length, response is `NumberResponse`
- `TestAppend_ExecuteHandlesError` — mock store returns error, response is `ErrorResponse`

Follow the pattern from `internal/command/kv/incr_test.go`.

**Done when**: `make test` passes.

---

## Task 5 — Register APPEND in registry

**File**: `internal/command/registry/registry.go`

**What to do**: Add inside `SetupDefaultParserRegistry`, after `kv.NewExistsParser()`:

```go
registry.Register(kv.NewAppendParser())
```

**Done when**: `make test` passes.

---

## Task 6 — Write integration tests for APPEND

**File**: `integration/command/kv/append_test.go`

**What to do**: Add integration tests in the existing `kv` package (same `testClient` from `TestMain` in `incr_test.go`):

```go
package kv

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestAppend_NonExistentKey(t *testing.T) { ... }
func TestAppend_ExistingKey(t *testing.T) { ... }
func TestAppend_IntegerEncodedKey(t *testing.T) { ... }
func TestAppend_TTLPreserved(t *testing.T) { ... }
```

Test cases:
1. **APPEND to non-existent key** — returns `len(value)`, GET returns value.
2. **APPEND to existing key** — returns cumulative length, GET returns concatenated string.
3. **APPEND to integer-encoded key** — SET key "10", APPEND " items", GET returns "10 items" (length 8).
4. **TTL preserved** — SET key with EX 10, APPEND value, TTL command still returns positive.

**Done when**: `make test` passes including new integration tests.
