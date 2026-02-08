# Redis Command Implementation Skill

Implements a new Redis command in the avacado project following the established architecture pattern.

## Usage

```
/redis-command <COMMAND_NAME>
```

Example: `/redis-command APPEND`

## What This Skill Does

This skill automates the implementation of a new Redis command by:
1. Gathering command specifications
2. Adding storage layer method
3. Implementing memory store logic
4. Creating command and parser
5. Writing comprehensive tests
6. Registering the command
7. Running all tests

## Process

### Step 1: Gather Requirements

Ask the user for:
- Command name (e.g., APPEND, STRLEN, INCRBY)
- Command description and Redis specification
- Arguments (names and types)
- Return type (integer, string, bulk string, array, etc.)
- Edge cases to handle

### Step 2: Storage Layer

1. **Add method to Store interface** (`internal/storage/kv/store.go`)
   - Add method signature with appropriate parameters
   - Include context as first parameter
   - Use Go conventions for return types

2. **Implement in memory store** (`internal/storage/kv/memory/memory.go`)
   - Add method with proper mutex locking (`k.mu.Lock()` / `defer k.mu.Unlock()`)
   - Handle non-existent keys
   - Handle expired keys (treat as non-existent)
   - Implement the core logic
   - Return appropriate values

3. **Add memory store test** (`internal/storage/kv/memory/memory_test.go`)
   - Test with non-existent key
   - Test with existing key
   - Test with expired key
   - Test error cases
   - Test edge cases specific to the command

### Step 3: Command Layer

1. **Create command file** (`internal/command/kv/<command_name>.go`)
   ```go
   package kv

   import (
       "avacado/internal/command"
       "avacado/internal/protocol"
       "avacado/internal/storage"
       "context"
   )

   // CommandName represents the command with its arguments
   type CommandName struct {
       // Fields for arguments
   }

   func (c *CommandName) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
       result, err := storage.KV().MethodName(ctx, ...)
       if err != nil {
           return protocol.NewErrorResponse(err)
       }
       // Return appropriate response type
       return protocol.NewNumberResponse(result)
   }

   type CommandNameParser struct {}

   func NewCommandNameParser() *CommandNameParser {
       return &CommandNameParser{}
   }

   func (p *CommandNameParser) Parse(msg *protocol.Message) (command.Command, error) {
       // Parse arguments from msg.Args
       // Handle parsing errors
       return &CommandName{...}, nil
   }

   func (p *CommandNameParser) Name() string {
       return "COMMANDNAME"
   }
   ```

2. **Create unit tests** (`internal/command/kv/<command_name>_test.go`)
   - TestParser_Parse: Test parsing with valid input
   - TestParser_ParseMultiple: If command takes multiple args
   - TestCommand_Execute: Test successful execution with mocks
   - TestCommand_ExecuteHandlesError: Test error handling

### Step 4: Integration Tests

Create integration test file (`integration/command/kv/<command_name>_test.go`)
- Use `t.Parallel()` for all tests
- Use unique key names to avoid conflicts
- Test typical usage scenarios
- Test edge cases
- Test with expired keys
- Test error conditions
- Use `testClient` (go-redis client) to call the command
- Use `testify/assert` for assertions

### Step 5: Register and Test

1. **Register parser** (`internal/command/registry/registry.go`)
   - Add `registry.Register(kv.NewCommandNameParser())` in `SetupDefaultParserRegistry`

2. **Regenerate mocks**
   ```bash
   go generate ./...
   ```

3. **Run tests**
   ```bash
   # Memory store tests
   go test ./internal/storage/kv/memory -run TestKVMemoryStore_CommandName -v

   # Command unit tests
   go test ./internal/command/kv -run TestCommandName -v

   # Integration tests
   go test ./integration/command/kv -run TestCommandName -v

   # All KV tests
   go test ./internal/command/kv -v
   go test ./integration/command/kv -v

   # Build project
   go build ./...
   ```

4. **Create task list** to track progress through the implementation

## Response Types

Common protocol response types:
- `protocol.NewNumberResponse(int64)` - For integer returns
- `protocol.NewBulkStringResponse([]byte)` - For string values
- `protocol.NewSimpleStringResponse(string)` - For simple strings like "OK"
- `protocol.NewNullResponse()` - For null/nil values
- `protocol.NewErrorResponse(error)` - For errors
- `protocol.NewArrayResponse([]*protocol.Response)` - For arrays

## Key Patterns

**Thread Safety:**
- Always use `k.mu.Lock()` and `defer k.mu.Unlock()` in storage methods

**Expired Keys:**
- Check `v.isExpired()` after retrieving values
- Treat expired keys as non-existent

**Testing:**
- Use `gomock` for mocking in unit tests
- Use `go-redis/v9` client in integration tests
- Use unique key prefixes in integration tests (e.g., `commandname_key1`)

**Parsing:**
- Use `msg.Args[i].AsString()` to parse string arguments
- Use `strconv.ParseInt()` for integer arguments
- Return parsing errors immediately

## Example Commands

Reference these existing implementations:
- **INCR/DECR** - Simple key operations, integer arithmetic
- **DECRBY** - Commands with numeric arguments
- **DEL** - Commands with multiple keys (variadic)
- **SET** - Commands with options
- **GET** - Simple key retrieval

## Notes

- Follow Redis specifications exactly for command behavior
- Maintain consistency with existing code style
- Write descriptive test names
- Handle all edge cases
- Verify no regressions by running all tests
