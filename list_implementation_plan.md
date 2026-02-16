# Plan: Redis List Data Type Implementation

## Context

The project is a Redis-compatible in-memory store in Go. The listpack encoding layer
(encoding, traversal, push, pop) is already implemented. The goal is to build on top of
it by implementing a **quicklist** (Redis's actual list encoding — a doubly-linked list
of listpack nodes) and exposing it through the standard Redis list commands: LPUSH,
RPUSH, LPOP, RPOP, LLEN, LRANGE.

## Internal Representation: Quicklist

Redis stores lists as a **quicklist**: a doubly-linked list where each node is a
**listpack**. Small lists live in a single listpack node. When a node fills up
(`isFull()` → true), a new node is created. This gives O(1) push/pop at both ends and
bounded memory per node.

```
head ──► [listpack: a b c] ◄──► [listpack: d e f] ◄── tail
```

## Changes Required

### 1. Extend `listPack` with head-side operations

**File:** `internal/storage/lists/memory/listPack.go`

The current `push()` / `pop()` operate on the **tail** (right side). Add:

- `rpush(values ...[]byte) (int, error)` — alias/rename of existing `push()`
- `rpop(count int) [][]byte` — alias/rename of existing `pop()`
- `lpush(values ...[]byte) (int, error)` — prepend at head (offset 6):
  1. Encode each value to a temp buffer
  2. `copy` existing entries right by encoded size (memmove)
  3. Write new entry at offset 6
  4. Update header count + size
- `lpop(count int) [][]byte` — remove from head using forward `traverse()`:
  1. Traverse first `count` entries, collect values and final byte offset
  2. `copy` remaining entries left to offset 6
  3. Write `0xFF` at new end, update header
- `getRange(start, stop int) [][]byte` — forward `traverse()`, skip `start` entries,
  collect until `stop` (inclusive), return as `[][]byte`

### 2. Add `quickList` struct

**New file:** `internal/storage/lists/memory/quickList.go`

```go
type quickListNode struct {
    lp   *listPack
    prev *quickListNode
    next *quickListNode
}

type quickList struct {
    mu          sync.RWMutex
    head        *quickListNode
    tail        *quickListNode
    count       int64  // total element count across all nodes
    nodeCount   int
    maxNodeSize int    // default 8192 (8 KB, matching Redis -2 config)
}
```

**Operations:**

- `newQuickList() *quickList` — initialize with one empty listpack node
- `rpush(values ...[]byte) int64` — append to tail node; if `tail.lp.isFull()`, create
  new tail node first
- `lpush(values ...[]byte) int64` — prepend to head node; if `head.lp.isFull()`, create
  new head node first
- `rpop(count int) [][]byte` — pop from tail node; if node becomes empty, unlink it
- `lpop(count int) [][]byte` — pop from head node; if node becomes empty, unlink it
- `llen() int64` — return `count`
- `lrange(start, stop int64) [][]byte` — handle negative indices, clamp stop to
  `count-1`, traverse nodes collecting elements in the requested range

### 3. Expand the `Lists` interface

**File:** `internal/storage/lists/lists.go`

```go
type Lists interface {
    LPush(key string, values ...[]byte) (int64, error)
    RPush(key string, values ...[]byte) (int64, error)
    LPop(key string, count int) ([][]byte, error)
    RPop(key string, count int) ([][]byte, error)
    LLen(key string) (int64, error)
    LRange(key string, start, stop int64) ([][]byte, error)
}
```

### 4. Implement `ListsMemoryStore`

**File:** `internal/storage/lists/memory/listsMemoryStore.go`

```go
type ListsMemoryStore struct {
    mu    sync.RWMutex
    lists map[string]*quickList
}
```

Implements all 6 `Lists` interface methods. Each method acquires the store-level lock,
looks up or creates a `quickList` for the key, delegates to the quickList method, and
returns results.

- `LPush` / `RPush` return the new list length
- `LPop` / `RPop` return nil + error when key does not exist (Redis returns nil)
- `LLen` returns 0 for missing keys (not an error)
- `LRange` returns empty slice for missing keys or out-of-range indices

### 5. Add `Lists()` to `Storage` interface

**File:** `internal/storage/storage.go`

- Add `Lists() lists.Lists` to the `Storage` interface
- Add `lists *memory.ListsMemoryStore` field to `DefaultStorage`
- Add `Lists()` method returning that field
- Initialize in `NewDefaultStorage()`

### 6. Implement list commands

**New directory:** `internal/command/list/`

Each file follows the same pattern as `kv/set.go`: a `Command` struct + `Parser` struct
+ `Execute()` + `Parse()` + `Name()`.

| File | Command | Parse rules |
|------|---------|-------------|
| `lpush.go` | `LPUSH key value [value ...]` | ≥2 args |
| `rpush.go` | `RPUSH key value [value ...]` | ≥2 args |
| `lpop.go`  | `LPOP key [count]`           | 1–2 args; count defaults to 1; returns bulk string if no count arg, array if count given |
| `rpop.go`  | `RPOP key [count]`           | same as LPOP |
| `llen.go`  | `LLEN key`                   | 1 arg; returns integer |
| `lrange.go`| `LRANGE key start stop`      | 3 args; returns array |

LPUSH / RPUSH return an integer (new length).
LPOP / RPOP with no count return a bulk string (or null bulk string if empty).
LPOP / RPOP with count return an array.
LRANGE returns an array (empty array for out-of-range, not error).

All commands call `storage.Lists().<Method>()`.

### 7. Register commands

**File:** `internal/command/registry/registry.go`

Add to `SetupDefaultParserRegistry()`:
```go
registry.Register(list.NewLPushParser())
registry.Register(list.NewRPushParser())
registry.Register(list.NewLPopParser())
registry.Register(list.NewRPopParser())
registry.Register(list.NewLLenParser())
registry.Register(list.NewLRangeParser())
```

### 8. Regenerate mocks

Run `go generate` in:
- `internal/storage/lists/` (regenerates `mock/lists.go`)
- `internal/storage/` (regenerates `mock/storage.go`)

## Critical File Paths

| File | Action |
|------|--------|
| `internal/storage/lists/memory/listPack.go` | Add `lpush`, `lpop`, `getRange`; rename `push`→`rpush`, `pop`→`rpop` |
| `internal/storage/lists/memory/quickList.go` | **Create** — quickList struct + all operations |
| `internal/storage/lists/memory/listsMemoryStore.go` | Implement fully |
| `internal/storage/lists/lists.go` | Expand interface |
| `internal/storage/storage.go` | Add `Lists()` |
| `internal/command/list/lpush.go` | **Create** |
| `internal/command/list/rpush.go` | **Create** |
| `internal/command/list/lpop.go` | **Create** |
| `internal/command/list/rpop.go` | **Create** |
| `internal/command/list/llen.go` | **Create** |
| `internal/command/list/lrange.go` | **Create** |
| `internal/command/registry/registry.go` | Register 6 new parsers |

## Reused Functions

- `encode()` — `internal/storage/lists/memory/encoding.go:103`
- `decode()` — `internal/storage/lists/memory/encoding.go:130`
- `traverse()` — `internal/storage/lists/memory/encoding.go:535`
- `traverseReverse()` — `internal/storage/lists/memory/encoding.go:559`
- `newEmptyListPack()` — `internal/storage/lists/memory/listPack.go:16`
- `listPack.isFull()` — `internal/storage/lists/memory/listPack.go:37`

## Verification

1. **Unit tests** — add `quickList_test.go` covering push/pop/range at both ends,
   node promotion when `isFull`, empty-list edge cases
2. **Existing tests** — `make test` must stay green (listPack tests must still pass
   after rename of `push`→`rpush` / `pop`→`rpop`)
3. **Integration** — run `redis-cli` against the server and verify:
   ```
   RPUSH mylist a b c      → 3
   LPUSH mylist x y        → 5
   LRANGE mylist 0 -1      → y x a b c
   LLEN mylist             → 5
   LPOP mylist             → y
   RPOP mylist 2           → c b
   LRANGE mylist 0 -1      → x a
   ```
