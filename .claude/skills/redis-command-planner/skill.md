---
description: "Creates an implementation plan for a Redis command based on the provided command name"
---

## Role
As a Redis command implementation planner, your role is to create a detailed plan for implementing a new Redis command based on the command name provided by the user.
This plan will guide the development process, ensuring that all necessary steps are taken to successfully implement the command in the Avocado codebase.
You have to store the detailed plan in a Markdown format in docs/plans/storage_type/command_name.plan.md file.
You should not edit any other files. If the docs/plans/storage_type directory does not exist, create it.

## Fetch Command Specifications
- Using the command name fetch its specifications from the Redis documentation using url: https://redis.io/docs/latest/commands/<lowercase_command_name>
- Identify the command's behavior, expected arguments, return values, and any edge cases or error conditions.
- Discuss the architectural changes needed in the storage layer and command layer to accommodate the new command.

## Plan Structure
1. **Storage Layer Changes**:
   - Identify the storage type that need update or need to be created if not present. `internal/storage/storage.go`. Update the `DefaultStorage` if new storage type is added.
   - Define the new method signature to be added to the Store interface in `internal/storage/<storage_type>/<storage_type>.go`.
   - Identify if new storage encoding is required for the command data types and plan for its implementation. Refer `internal/storage/listpack/encoding.go` for existing encodings.
   - Outline the implementation steps for the in-memory store in `internal/storage/<storage_type>/memory/<storage_type>.go`.
   - Ensure planning mocks generation comment in the newly added storage type interface.
2. **Command Layer Changes**:
   - Plan the creation of a new command file in `internal/command/storage_type/<command_name>.go` refer existing command `internal/command/kv/get.go` for structure.
   - Define the command struct and its fields based on the command's expected arguments.
   - Outline the implementation steps for the command's logic, including how it will interact with the storage layer.
   - Add integration test in `integration/command/storage_type/command_name_test.go` to test the command's functionality in a real environment.
3. **Command Registration**:
   - Register the new command in `internal/command/registry/registry.go` to ensure it is recognized by the system.
4**Testing**:
   - Plan comprehensive tests for the new command, covering normal scenarios, edge cases, and error conditions.
