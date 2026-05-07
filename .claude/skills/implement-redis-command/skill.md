---
name: implement-redis-command
description: Implements a new Redis command in the avacado project following the established architecture pattern.
parameters:
  - name: command_name
---

## Usage

```
/implement-redis-command <COMMAND_NAME>
```

Example: `/implement-redis-command APPEND`

## What This Skill Does

End-to-end automation for implementing a new Redis command:

1. **Plan** — invoke `/redis-command-planner` to produce a detailed plan file
2. **Create tasks** — invoke `/implementation-task-planner` on that plan to register an ordered task list
3. **Execute** — work through each task in order, running `make test` after every task

## Process

### Phase 1: Plan

Invoke `/redis-command-planner <COMMAND_NAME>`.

Wait for it to finish. It will write a plan file to:
```
docs/plans/<storage_type>/<command_name_lowercase>.plan.md
```

Do not proceed until the plan file exists on disk.

### Phase 2: Create Tasks

Invoke `/implementation-task-planner docs/plans/<storage_type>/<command_name_lowercase>.plan.md`.

This writes a tasks file to:
```
docs/tasks/<storage_type>/<command_name_lowercase>.tasks.md
```
and registers all implementation tasks via TaskCreate. Note the task IDs it returns — you will update each one as you work through Phase 3.

### Phase 3: Execute Tasks

For each task (in the order returned by `/implementation-task-planner`):

1. Mark the task `in_progress` with TaskUpdate.
2. Implement exactly what the task specifies — no more, no less.
3. Run `make test`.
   - **Tests pass** → mark the task `completed` with TaskUpdate, move to the next task.
   - **Tests fail** →
     a. Read the failure output in full.
     b. Fix only the code introduced in this task (do not touch other tasks' files).
     c. Re-run `make test`.
     d. Repeat until green, then mark `completed`.
     e. If still failing after 3 fix attempts, stop and report to the user: show the exact failing test name, the error output, and what you tried.

### Completion Criteria

The command is complete when **all** of the following are true:

- Every task is marked `completed`
- `make test` is green with no failures
- The command appears in `internal/command/registry/registry.go`
- At least one integration test in `integration/command/` covers the happy path

Do not report success if any task is pending or any test is failing.
