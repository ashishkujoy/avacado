# APPEND Command Implementation Plan

## Command Specification

**Syntax:** `APPEND key value`

**Since:** Redis 2.0.0

**Complexity:** O(1) amortized

**Return value:** Integer — the total byte length of the string stored at `key` after the append.

**Behavior:**
- If the key already exists and holds a string, `value` is appended to its end.
- If the key does not exist, it is created with an empty string and then `value` is appended (equivalent to `SET key value`).
- Returns the total length of the resulting string.

**Error conditions:**
- Wrong number of arguments → standard arity error (handled by the parser).
- Existing key holds a non-string type → not applicable here (all KV store values are strings).

**Examples:**
```
APPEND mykey "Hello"   → 5
APPEND mykey " World"  → 11
GET mykey              → "Hello World"
```

---

## 1. Storage Layer Changes

### 1.1 Add `Append` to `kv.Store` interface

**File:** `internal/storage/kv/store.go`

Add the following method to the `Store` interface:

```go
Append(ctx context.Context, key string, value []byte) (int64, error)
```

- Returns the total length of the string after appending.
- No new error type is needed — the method cannot fail under normal conditions.

Re-run mock generation after updating the interface:
```
go generate ./internal/storage/kv/...
```

### 1.2 Implement `Append` in the in-memory store

**File:** `internal/storage/kv/memory/memory.go`

Implementation logic:

```go
func (k *KVMemoryStore) Append(ctx context.Context, key string, value []byte) (int64, error) {
    existing, ok := k.store[key]
    if !ok || existing.isExpired() {
        // Key absent or expired: create fresh string value
        k.store[key] = &value{data: value, enc: encodingString}
        return int64(len(value)), nil
    }
    // Decode current bytes (handles encodingInteger transparently via Bytes())
    current := existing.Bytes()
    appended := append(current, value...)
    k.store[key] = &value{data: appended, enc: encodingString, expiry: existing.expiry}
    return int64(len(appended)), nil
}
```

Key points:
- Always store the result as `encodingString` (never re-encode to integer after append).
- Preserve the existing TTL (`existing.expiry`) when appending to a live key.
- When the key is expired or absent, no TTL is set (matches Redis behaviour: APPEND on a new key sets no expiry).
- Use `existing.Bytes()` to decode the current value regardless of its encoding (handles the case where an integer-encoded value gets a string appended to it).

Note: variable name collision — the `value` parameter shadows the `value` struct type. Rename the parameter to `data` in the actual implementation:

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

### 1.3 No new storage type or encoding needed

APPEND operates on the existing KV store. No new `storage_type` directory, no new encoding, and no changes to `storage.go` or `DefaultStorage` are required.

---

## 2. Command Layer Changes

### 2.1 Create the command file

**File:** `internal/command/kv/append.go`

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

### 2.2 Unit tests for the command

**File:** `internal/command/kv/append_test.go`

Test cases to cover:
1. `AppendParser.Parse` — produces correct `Append` struct.
2. `Append.Execute` — happy path: mock store returns length, response is `NumberResponse`.
3. `Append.Execute` — error path: mock store returns error, response is `ErrorResponse`.

Follow the exact pattern used in `incr_test.go` (mock `Storage` + mock `Store`).

### 2.3 Integration test

**File:** `integration/command/kv/append_test.go`

Add tests inside the existing `kv` package (same `TestMain` / `testClient` as `incr_test.go`):

Test cases:
1. **APPEND to non-existent key** — creates the key, returns length equal to `len(value)`.
2. **APPEND to existing key** — returns cumulative length, GET returns concatenated string.
3. **APPEND to integer-encoded key** — SET key to "10", APPEND " items", result is "10 items" of length 8.
4. **TTL preserved** — SET key with EX, APPEND, verify TTL is still set (positive).

---

## 3. Command Registration

**File:** `internal/command/registry/registry.go`

Add inside `SetupDefaultParserRegistry`:

```go
registry.Register(kv.NewAppendParser())
```

Place it near the other KV command registrations (after `kv.NewExistsParser()`).

---

## 4. Testing Summary

| Layer | File | What is tested |
|---|---|---|
| Storage unit | `internal/storage/kv/memory/memory_test.go` | Append to absent key, live key, expired key, integer-encoded key, TTL preservation |
| Command unit | `internal/command/kv/append_test.go` | Parser, Execute happy path, Execute error path |
| Integration | `integration/command/kv/append_test.go` | End-to-end via go-redis client against live server |

---

## 5. Implementation Order

1. Add `Append` to `kv.Store` interface (`internal/storage/kv/store.go`)
2. Regenerate mocks (`go generate ./internal/storage/kv/...`)
3. Implement `Append` in `KVMemoryStore` (`internal/storage/kv/memory/memory.go`)
4. Add storage unit tests (`internal/storage/kv/memory/memory_test.go`)
5. Create command + unit tests (`internal/command/kv/append.go`, `append_test.go`)
6. Register the command (`internal/command/registry/registry.go`)
7. Add integration tests (`integration/command/kv/append_test.go`)
8. Run `make test` — all green
