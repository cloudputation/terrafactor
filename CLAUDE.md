# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Terrafactor** is a Go CLI tool that reads JSON data and prints it in Terraform-style formatted output, with colored `+` (create) or `-` (destroy) prefixes. It is a focused utility with a minimal package architecture.

## Build & Development Commands

```bash
# Build binary
go build -o build/terrafactor .

# Run directly
go run . print <json_file> <operation>

# Run example
go run examples/main.go <json_file> <operation>

# Tidy dependencies
go mod tidy
```

## Application Initialization Order

The application follows this sequence (see `main.go`):

1. **Load Configuration** - Parse HCL config file
2. **Initialize Logger** - File + stdout output
3. **Run CLI** - Execute command-line interface

## Configuration System

### HCL Configuration
Config file default: `/etc/terrafactor/config.hcl`

Override via environment variable:
```bash
export TF_CONFIG_FILE_PATH=/path/to/config.hcl
```

### Configuration Structure
```hcl
log_dir  = "logs"
data_dir = "data"
```

## Package Architecture

```
packages/
├── bootstrap/    - Filesystem initialization (log/data directories)
├── cli/          - Cobra CLI command structure
├── config/       - HCL configuration parsing
├── logger/       - hclog-based logging (stdout + file)
└── terrafactor/  - Core output formatting logic (Print, printData, printArray)
```

### Import Path Pattern
```go
import "github.com/cloudputation/terrafactor/packages/<package>"
```

## Core Logic

The `packages/terrafactor/outputs.go` file contains the primary formatting engine:

- **`Print(data, operationTag, indentStr, w)`** - Entry point; applies colored prefix (`+` green for create, `-` red for destroy) and recurses into `printData`
- **`printData(...)`** - Recursively handles maps and arrays with proper indentation
- **`printArray(...)`** - Handles array elements including nested maps and primitives

### CLI Usage
```bash
terrafactor print plan.json create
terrafactor print plan.json destroy
```

## Logging

- Uses HashiCorp `hclog` with output to stdout + `logs/terrafactor.log`
- Log levels: debug, info, warn, error, fatal
- Logger name: `TERRAFACTOR`

## Key Design Patterns

### Configuration
- Minimal: only `log_dir` and `data_dir` fields
- Environment override via `TF_CONFIG_FILE_PATH`
- HCL format parsed with `hashicorp/hcl/v2`

### Error Handling
- Fail fast with meaningful errors
- CLI commands use `log.Fatal` for unrecoverable errors
- `Print` returns errors for invalid operation tags

## Module Structure

- **Module**: `github.com/cloudputation/terrafactor`
- **Binary**: `terrafactor`
- **Go version**: 1.23+

## Code Quality Standards

This codebase follows the software reliability rules defined in `~/.claude/CLAUDE.md`:
- Simple control flow (max 3 levels nesting, cyclomatic complexity < 10)
- Bounded operations (explicit loop termination, timeouts on external calls)
- Explicit resource management (defer cleanup immediately after acquisition)
- Small functions (20-30 lines, single responsibility)
- Defensive validation (fail fast, meaningful errors)
- Minimal scope (local > instance > global)
- Explicit error handling (never ignore errors)
- Minimal abstraction (direct code over indirection)
- Clear data flow (pure functions preferred, explicit parameters)
- Zero compiler/linter warnings
