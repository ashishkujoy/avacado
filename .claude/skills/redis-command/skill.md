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

This skill automates the implementation of a new Redis command by:
1. Create a detailed implementation using /redis-command-planner skill to create detailed implementation plan at `plans/storage_type/command_name.plan.md`
2. Based on the implementation plan create a detailed list of task to implement the command.
3. After each task runs `make test` to ensure all tests are passing and no regressions are introduced.
