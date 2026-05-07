---
name: implementation-task-planner
description: Reads a redis-command-planner output plan file and creates a concrete, ordered TaskCreate task list for implementing the command.
parameters:
  - name: plan_file
---

## Usage

```
/implementation-task-planner <PLAN_FILE_PATH>
```

Example: `/implementation-task-planner docs/plans/hashmap/hincrby.plan.md`

## Role

Read the plan file produced by `/redis-command-planner`, write a tasks markdown file that mirrors the plan's directory
and naming convention, then register each task with TaskCreate. Each task must be scoped to a single file or operation
so it can be implemented and verified with `make test` in isolation.

## Process

### Step 1: Parse the Plan File

Read the file at `<PLAN_FILE_PATH>` and extract the following into working memory before creating any tasks:

| Field                                  | Where to find it                                                    |
|----------------------------------------|---------------------------------------------------------------------|
| Command name                           | `# <COMMAND>` heading or **Syntax** line                            |
| Storage type                           | Path in **Storage Layer Changes** (e.g., `hashmaps`, `lists`, `kv`) |
| New interface methods                  | Method signatures under **Update … Interface** section              |
| New encoding needed                    | Any mention of `listpack/encoding.go` changes                       |
| Memory store implementation details    | Under **Implement … in Memory Store** section                       |
| Command struct fields + argument rules | Under **Create … Command File** section                             |
| Integration test scenarios             | Under **Testing** section                                           |

### Step 2: Write the Tasks File

Derive the output path from the plan file path by replacing the `plans/` segment with `tasks/` under `docs/` and
changing the extension:

```
Plan:  docs/plans/<storage_type>/<command_name>.plan.md
Tasks: docs/tasks/<storage_type>/<command_name>.tasks.md
```

Create the `docs/tasks/<storage_type>/` directory if it does not exist, then write the tasks file using this template:

```markdown
# <COMMAND_NAME> Implementation Tasks

**Command**: `<COMMAND_NAME>`
**Storage type**: `<storage_type>`
**Plan**: `docs/plans/<storage_type>/<command_name>.plan.md`

---

## Task 1 — <title>
...

## Task 2 — <title>
...
```

Each task section must include **File**, **What to do**, and **Done when** fields filled with specifics from Step 1.
Define all tasks (Task 1–7 below) in the file before creating any TaskCreate entries.

### Step 3: Register Tasks with TaskCreate

After the tasks file is written, call TaskCreate once per task in order, using the title and body from the tasks file.
Generic placeholders are not acceptable — every task must be concrete enough to implement without re-reading the plan.


---

#### Task 1 — Storage interface: add method signature

**Title**: `Add <CommandName> method to <StorageType> interface`

**Body must include**:

- File: `internal/storage/<storage_type>/<storage_type>.go`
- Exact method signature(s) to add, copied from the plan
- The `//go:generate` line that exists in that file (to remind the implementer mocks must be regenerated next)

**Done when**: File compiles with the new method(s) in the interface.

---

#### Task 2 — Regenerate mocks

**Title**: `Regenerate mocks for <StorageType>`

**Body must include**:

- Command: `make clean && make mocks`
- File that will be updated: `internal/storage/<storage_type>/mock/<storage_type>.go`

**Done when**: Mock file reflects the new method and `make test` passes.

---

#### Task 3 — Memory store: implement the method

**Title**: `Implement <CommandName> in <StorageType> memory store`

**Body must include**:

- File: `internal/storage/<storage_type>/memory/<storage_type>.go`
- Step-by-step implementation notes from the plan (data retrieval, initialization defaults, error cases)
- Any helper files in the same `memory/` package that need changes (e.g., `hashmap.go` alongside `hashmaps.go`)

**Done when**: `make test` passes.

---

#### Task 4 (conditional) — New listpack encoding

**Only create this task if the plan specifies changes to `internal/storage/listpack/encoding.go`.**

**Title**: `Add <type> encoding to listpack`

**Body must include**:

- File: `internal/storage/listpack/encoding.go`
- Encoding/decoding logic described in the plan

**Done when**: Encoding compiles and `make test` passes.

---

#### Task 5 — Command layer: create command file

**Title**: `Implement <CommandName> command`

**Body must include**:

- File: `internal/command/<storage_type>/<command_name_lowercase>.go`
- Package name
- Command struct with exact field names and types from the plan
- `Execute` method: which storage method to call (`storage.<Accessor>().<Method>()`) and how to map the result to a
  `*protocol.Response`
- Parser struct, `Parse` method with exact argument count validation and any type coercion (e.g., `strconv.ParseInt`)
- `Name()` method returning the uppercase command string
- Constructor `New<Command>Parser()`
- Reference: follow `internal/command/kv/get.go` for simple commands, `internal/command/hashmap/hincrby.go` for commands
  with argument validation

**Done when**: File compiles and `make test` passes.

---

#### Task 6 — Command registration

**Title**: `Register <CommandName> in registry`

**Body must include**:

- File: `internal/command/registry/registry.go`
- Exact line to add: `registry.Register(<package>.New<Command>Parser())`
- Where in the file to insert it (after the last registration in its storage-type group)

**Done when**: `make test` passes.

---

#### Task 7 — Integration tests

**Title**: `Write integration tests for <CommandName>`

**Body must include**:

- File: `integration/command/<storage_type>/<command_name_lowercase>_test.go`
- If this is the first test in the package, include the `TestMain` setup (copy from
  `integration/command/hashmap/hset_test.go`)
- Each test case from the plan's **Testing** section as a named `func Test<CommandName>_<Scenario>` entry
- For each test case: setup steps, the call under test, and the assertion

**Done when**: `make test` passes including the new integration tests.

---

### Step 4: Report to user

After all TaskCreate calls complete, print:

```
Tasks written: docs/tasks/<storage_type>/<command_name>.tasks.md

Tasks registered:
  #<id>  <title>
  #<id>  <title>
  ...

Run /implement-redis-command to execute them in order, or pick up any task manually.
```

## Output

- **Tasks file**: `docs/tasks/<storage_type>/<command_name>.tasks.md` — persists the full task list for reference and re-use
- **TaskCreate entries**: one per task, for tracking progress during execution
