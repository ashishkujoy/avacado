# HINCRBY Command Implementation Plan

## Command Overview
**HINCRBY** increments the integer value of a hash field by the specified amount.

### Redis Specification
- **Syntax**: `HINCRBY key field increment`
- **Arguments**:
  - `key`: The hash key
  - `field`: The field name within the hash
  - `increment`: The integer amount to increment by (can be negative for decrement)
- **Return Value**: Integer reply — the value of the field after the increment operation
- **Time Complexity**: O(1)
- **ACL Categories**: `@write`, `@hash`, `@fast`
- **Available since**: Redis 2.0.0

### Key Behaviors
- If the key doesn't exist, a new hash is created
- If the field doesn't exist, it's initialized to 0 before incrementing
- The increment argument is signed, allowing both increment and decrement
- Value range limited to 64-bit signed integers
- Error if the field value is not an integer

## Implementation Architecture

### Storage Layer Changes

#### 1. Update HashMaps Interface
**File**: `internal/storage/hashmaps/hashmaps.go`

Add a new method to the `HashMaps` interface:
```go
HIncrBy(ctx context.Context, key string, field string, increment int64) (int64, error)
```

This method will:
- Increment the integer value of a field in the hash by the specified amount
- Create the hash if it doesn't exist
- Initialize the field to 0 if it doesn't exist
- Return the value after the increment
- Return an error if the field contains a non-integer value

#### 2. Implement HIncrBy in Memory Store
**File**: `internal/storage/hashmaps/memory/hashmaps.go`

Implement the `HIncrBy` method in the `HashMaps` struct:
- Retrieve or create the HashMap for the given key
- Call the HashMap's increment method to perform the operation
- Handle conversion of string values to integers
- Return the new value and any errors

#### 3. Implement HIncrBy in HashMap
**File**: `internal/storage/hashmaps/memory/hashmap.go`

Add a new method `IncrBy(field string, increment int64) (int64, error)` to the `HashMap` struct:
- Get the current value of the field (or 0 if not exists)
- Convert the string value to int64
- Add the increment value
- Check for integer overflow/underflow
- Update the field with the new value
- Return the new value
- Handle both listpack and hash encodings for internal storage

### Command Layer Changes

#### 1. Create HIncrBy Command File
**File**: `internal/command/hashmap/hincrby.go`

Create the command and parser following the established pattern:

```go
type HIncrBy struct {
    key       string
    field     string
    increment int64
}

func (h *HIncrBy) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
    // Call storage.Maps().HIncrBy() and return response
}

type HIncrByParser struct {}

func (p *HIncrByParser) Parse(msg *protocol.Message) (command.Command, error) {
    // Validate arguments (must have exactly 4: command, key, field, increment)
    // Parse increment as int64
    // Return HIncrBy command or error
}

func (p *HIncrByParser) Name() string {
    return "HINCRBY"
}
```

Argument validation:
- Must have exactly 3 arguments (key, field, increment)
- Increment must be parseable as a signed 64-bit integer
- Return appropriate error messages for invalid arguments

#### 2. Command Registration
**File**: `internal/command/registry/registry.go`

Add to `SetupDefaultParserRegistry()`:
```go
registry.Register(hashmap.NewHIncrByParser())
```

### Testing Strategy

#### Unit Tests
**File**: `internal/command/hashmap/hincrby_test.go`

Test cases for the parser:
- Valid arguments with positive increment
- Valid arguments with negative increment
- Missing arguments
- Non-integer increment value
- Increment that overflows 64-bit integer
- Parser name returns "HINCRBY"

#### Integration Tests
**File**: `integration/command/hashmap/hincrby_test.go`

Test cases for the command execution:
- Increment a non-existent field (initialize to 0, then increment)
- Increment an existing integer field
- Decrement using negative increment
- Increment on non-existent key (creates hash)
- Error when field contains non-integer value
- Multiple increments on the same field
- Return value verification after increment
- Edge cases near integer boundaries

### Implementation Steps

1. **Storage Layer**:
   - Add `HIncrBy` method signature to `HashMaps` interface
   - Implement `HIncrBy` in `HashMaps` struct (hashmaps.go)
   - Add `IncrBy` method to `HashMap` struct (hashmap.go)
   - Generate mocks with `go generate`

2. **Command Layer**:
   - Create `hincrby.go` command file with HIncrBy struct and HIncrByParser
   - Create `hincrby_test.go` unit tests
   - Register command in registry

3. **Integration Tests**:
   - Create integration test file
   - Implement comprehensive test cases

4. **Testing & Validation**:
   - Run unit tests
   - Run integration tests
   - Run full test suite with `make test`
   - Verify no regressions in existing commands

## Dependencies & Error Handling

### Error Conditions to Handle
1. **ERR hash value is not an integer or out of range** - Field contains non-integer value or would overflow
2. **ERR Protocol error** - Invalid number of arguments
3. **ERR increment is not an integer** - Increment parameter is not a valid integer

### Integer Validation
- Parse increment as int64
- When retrieving field value, verify it can be parsed as int64
- Check for overflow/underflow when adding: `if increment > 0 && current > math.MaxInt64-increment`
- Return appropriate Redis error format

## Compatibility Notes
- HINCRBY requires Redis 2.0.0+ (widely supported)
- Compatible with both listpack and hash encoding in the internal HashMap implementation
- Should handle migration from listpack to hash encoding if needed
