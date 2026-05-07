# HMGET Implementation Plan

## Command Specification

**Syntax:** `HMGET key field [field ...]`

**Behaviour:** Returns the values associated with the specified fields in the hash stored at `key`. For each field:
- If the field exists, its value is returned as a bulk string.
- If the field does not exist, `nil` (null bulk string) is returned.
- If `key` does not exist, it is treated as an empty hash and all results are `nil`.

**Return value:** Array of values in the same order as the requested fields, with `nil` for missing fields/keys.

**Time complexity:** O(N) where N is the number of requested fields.

**Example:**
```
> HSET myhash field1 "Hello" field2 "World"
(integer) 2
> HMGET myhash field1 field2 nofield
1) "Hello"
2) "World"
3) (nil)
```

---

## 1. Storage Layer Changes

### 1.1 Add `HMGet` to the `HashMaps` interface

**File:** `internal/storage/hashmaps/hashmaps.go`

Add the following method to the `HashMaps` interface:

```go
HMGet(ctx context.Context, key string, fields []string) []any
```

The return type is `[]any` where:
- Each element is `[]byte` when the field exists.
- Each element is `nil` when the field or key does not exist.

The slice preserves the order of the requested `fields`.

The `//go:generate` directive at the top of the file will re-generate the mock automatically.

### 1.2 Implement `HMGet` in the in-memory store

**File:** `internal/storage/hashmaps/memory/hashmaps.go`

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

### 1.3 Regenerate mock

Run `go generate ./internal/storage/hashmaps/...` to regenerate `internal/storage/hashmaps/mock/hashmaps.go` with the new `HMGet` method.

No new storage encoding is required — HMGET reuses the existing `HashMap` byte-slice storage.

---

## 2. Command Layer Changes

### 2.1 Command implementation

**File:** `internal/command/hashmap/hmget.go`

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

Note: `len(msg.Args) < 2` enforces the minimum of one key and one field. Any number of additional fields is valid.

### 2.2 Unit tests

**File:** `internal/command/hashmap/hmget_test.go`

Cover the following scenarios using mocks (`mockhashmaps`, `mocksstorage`):

| Test | Description |
|------|-------------|
| `TestHMGetCommand_Execute_AllFound` | All requested fields exist; array of bulk strings returned |
| `TestHMGetCommand_Execute_SomeNotFound` | Some fields missing; nil values at correct positions |
| `TestHMGetCommand_Execute_KeyNotFound` | Key does not exist; all results are nil |
| `TestHMGetParser_Parse_Valid` | Two or more args parse correctly |
| `TestHMGetParser_Parse_TooFewArgs` | Zero or one arg returns error |
| `TestHMGetParser_Name` | Name returns `"HMGET"` |

### 2.3 Integration tests

**File:** `integration/command/hashmap/hmget_test.go`

Use the existing `testClient` from the `hashmap` package (defined in `hset_test.go`). No new `TestMain` needed — all hashmap integration tests share the same package.

Cover the following scenarios:

| Test | Description |
|------|-------------|
| `TestHMGet_AllFieldsExist` | HMGET returns values in order for all existing fields |
| `TestHMGet_SomeMissing` | Mixed existing/missing fields; nil at correct positions |
| `TestHMGet_KeyNotFound` | Key does not exist; all nil values |
| `TestHMGet_SingleField` | HMGET with exactly one field |
| `TestHMGet_DuplicateFields` | Requesting the same field twice; each occurrence returns its value |

---

## 3. Command Registration

**File:** `internal/command/registry/registry.go`

Add the following line inside `SetupDefaultParserRegistry`, alongside the other hashmap registrations:

```go
registry.Register(hashmap.NewHMGetParser())
```

---

## 4. Implementation Order

1. Add `HMGet` to `internal/storage/hashmaps/hashmaps.go`
2. Implement `HMGet` in `internal/storage/hashmaps/memory/hashmaps.go`
3. Regenerate mock: `go generate ./internal/storage/hashmaps/...`
4. Create `internal/command/hashmap/hmget.go`
5. Create `internal/command/hashmap/hmget_test.go`
6. Register command in `internal/command/registry/registry.go`
7. Create `integration/command/hashmap/hmget_test.go`
