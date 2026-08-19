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
3. **Execute** — work through each task in order, running `make test` and `make lint` after every task, then
   `make vulncheck` once all tasks are complete

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
   - **Tests pass** → proceed to step 4.
   - **Tests fail** →
     a. Read the failure output in full.
     b. Fix only the code introduced in this task (do not touch other tasks' files).
     c. Re-run `make test`.
     d. Repeat until green, then proceed to step 4.
     e. If still failing after 3 fix attempts, stop and report to the user: show the exact failing test name, the error output, and what you tried.
4. Run `make lint`.
   - **No issues** → mark the task `completed` with TaskUpdate, move to the next task.
   - **Issues reported** →
     a. Fix only the code introduced in this task (do not touch other tasks' files).
     b. Re-run `make lint`.
     c. Repeat until clean, then mark `completed`.
     d. If still failing after 3 fix attempts, stop and report to the user: show the exact lint output and what you tried.

After the last task is marked `completed`, run `make vulncheck` once for the whole change:
   - **No vulnerabilities** → proceed to reporting completion.
   - **Vulnerabilities reported** →
     a. If a flagged dependency is one this task pulled in, upgrade or replace it and re-run `make vulncheck`.
     b. If the finding is pre-existing (unrelated to this change), stop and report it to the user rather than
        silently proceeding — do not attempt to fix vulnerabilities outside the scope of this command.

### Completion Criteria

The command is complete when **all** of the following are true:

- Every task is marked `completed`
- `make test` is green with no failures
- `make lint` reports no issues
- `make vulncheck` reports no new vulnerabilities introduced by this change
- The command appears in `internal/command/registry/registry.go`
- At least one integration test in `integration/command/` covers the happy path

Do not report success if any task is pending or any test is failing.
