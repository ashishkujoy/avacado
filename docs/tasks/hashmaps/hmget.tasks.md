# HMGET Implementation Tasks

**Command**: `HMGET`
**Storage type**: `hashmaps`
**Plan**: `docs/plans/hashmaps/hmget.plan.md`

---

## Task 1 — Add HMGet method to HashMaps interface

**File**: `internal/storage/hashmaps/hashmaps.go`

**What to do**: Add the following method to the `HashMaps` interface:

```go
HMGet(ctx context.Context, key string, fields []string) []any
```

Return type is `[]any` where each element is `[]byte` (field value) or `nil` (field/key missing). Order matches the requested `fields` slice.

The file already contains:
```go
//go:generate sh -c "rm -f mock/hashmaps.go && mockgen -source=hashmaps.go -destination=mock/hashmaps.go -package=mockhashmaps"
```
Mocks must be regenerated in the next task.

**Done when**: File compiles with `HMGet` in the interface.

---

## Task 2 — Regenerate mocks for HashMaps

**What to do**: Run:
```
make clean && make mocks
```

**File updated**: `internal/storage/hashmaps/mock/hashmaps.go`

**Done when**: Mock file reflects the new `HMGet` method and `make test` passes.

---

## Task 3 — Implement HMGet in HashMaps memory store

**File**: `internal/storage/hashmaps/memory/hashmaps.go`

**What to do**: Add the following method to `*HashMaps`:

```go
func (h *HashMaps) HMGet(_ context.Context, key string, fields []string) []any {
    result := make([]any, len(fields))
    hMap, found := h.maps[key]
    if !found {
        return result // all nil
    }
    for i, field := range fields {
        if value, ok := hMap.Get(field); ok {
            result[i] = value
        }
        // else result[i] remains nil
    }
    return result
}
```

Logic:
- Allocate `result` slice of length `len(fields)`, zero-value is `nil`.
- If the key does not exist in `h.maps`, return the all-nil slice immediately.
- Otherwise iterate over `fields`; for each field call `hMap.Get(field)` — if found, assign `[]byte` value to `result[i]`.

No helper files need changes; `hMap.Get` is already defined in `memory/hashmap.go`.

**Done when**: `make test` passes.

---

## Task 4 — Implement HMGET command

**File**: `internal/command/hashmap/hmget.go`

**What to do**: Create the file with the following content:

```go
package hashmap

import (
    "avacado/internal/command"
    "avacado/internal/protocol"
    "avacado/internal/storage"
    "context"
)

type hMGet struct {
    key    string
    fields []string
}

func (h *hMGet) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
    values := storage.Maps().HMGet(ctx, h.key, h.fields)
    return protocol.NewArrayResponse(values)
}

type HMGetParser struct{}

func (p *HMGetParser) Parse(msg *protocol.Message) (command.Command, error) {
    if len(msg.Args) < 2 {
        return nil, command.NewInvalidArgumentsCount(p.Name(), 2, len(msg.Args))
    }
    return &hMGet{key: msg.Args[0], fields: msg.Args[1:]}, nil
}

func (p *HMGetParser) Name() string {
    return "HMGET"
}

func NewHMGetParser() *HMGetParser {
    return &HMGetParser{}
}
```

Also create `internal/command/hashmap/hmget_test.go` with unit tests using `mockhashmaps` and `mocksstorage`:

| Test | Description |
|------|-------------|
| `TestHMGetCommand_Execute_AllFound` | All requested fields exist; response array contains bulk string values in order |
| `TestHMGetCommand_Execute_SomeNotFound` | Some fields missing; nil slots at correct positions in the array |
| `TestHMGetCommand_Execute_KeyNotFound` | Key does not exist; all array elements are null bulk strings |
| `TestHMGetParser_Parse_Valid` | Two or more args parse into correct `key` and `fields` |
| `TestHMGetParser_Parse_TooFewArgs` | Zero or one arg returns error |
| `TestHMGetParser_Name` | Returns `"HMGET"` |

**Done when**: File compiles and `make test` passes.

---

## Task 5 — Register HMGET in registry

**File**: `internal/command/registry/registry.go`

**What to do**: Add the following line inside `SetupDefaultParserRegistry`, after the existing hashmap registrations (after `registry.Register(hashmap.NewHIncrByParser())`):

```go
registry.Register(hashmap.NewHMGetParser())
```

**Done when**: `make test` passes.

---

## Task 6 — Write integration tests for HMGET

**File**: `integration/command/hashmap/hmget_test.go`

**What to do**: Add integration tests to the existing `hashmap` package (do NOT add a new `TestMain` — it already exists in `hset_test.go`).

Test scenarios:

| Test | Setup | Call | Assert |
|------|-------|------|--------|
| `TestHMGet_AllFieldsExist` | HSET `hmget_hash1` with `field1=v1`, `field2=v2` | HMGet `field1`, `field2` | `["v1", "v2"]` |
| `TestHMGet_SomeMissing` | HSET `hmget_hash2` with `field1=v1` | HMGet `field1`, `nofield` | `["v1", nil]` |
| `TestHMGet_KeyNotFound` | (no setup) | HMGet `hmget_nonexistent` with `field1` | `[nil]` |
| `TestHMGet_SingleField` | HSET `hmget_hash3` with `only=val` | HMGet `only` | `["val"]` |
| `TestHMGet_DuplicateFields` | HSET `hmget_hash4` with `f1=v1` | HMGet `f1`, `f1` | `["v1", "v1"]` |

Use `testClient.HMGet(ctx, key, fields...).Result()` which returns `[]interface{}`.

**Done when**: `make test` passes including the new integration tests.
