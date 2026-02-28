# Terrafactor

Terraform provider generation engine — builds fully functional Terraform providers from OpenAPI specs and the official terraform-provider-scaffold template.

## Overview

Terrafactor reads an OpenAPI 3.0 spec and a `provider.hcl` manifest, then scaffolds a complete, compilable Terraform provider written in Go. The generated provider includes real HTTP CRUD operations, an `APIClient` wired to the configured endpoint, typed resource models, and proper Terraform state management — not stubs.

Provider generation is the core of terrafactor. A secondary feature (`terrafactor print`) provides Terraform-style colored output formatting that can also be embedded directly into Go applications.

---

## Provider Generation Engine

### How it works

1. Parse `provider.hcl` — declares provider name, org, registry, OpenAPI spec path, and release metadata
2. Parse the OpenAPI spec — extracts collection paths, detail paths, path parameters, and CRUD operations (POST/GET/PUT/PATCH/DELETE)
3. Scaffold or update a full Terraform provider in `engine/store/<name>/`

Each resource in the spec becomes:
- A `resource.go` with `APIModel`, `toAPIModel()`/`fromAPIModel()` conversions, and real Create/Read/Update/Delete HTTP methods
- A `data_source.go` with a real GET
- An `ephemeral_resource.go` with a real GET open
- Schema, examples, and test stubs

The generated `provider.go` defines an `APIClient` struct that carries the HTTP client and configured endpoint, wired through provider `Configure()` to all resources.

### provider.hcl

```hcl
provider {
  name             = "myprovider"
  org              = "myorg"
  registry         = "registry.terraform.io"
  go_module_prefix = "github.com/myorg"

  schema {
    spec = "path/to/openapi.yaml"
  }

  release {
    version = "0.1.0"
    license = "MPL-2.0"
  }

  author {
    name   = "My Name"
    email  = "me@example.com"
    github = "myorg"
  }
}
```

### CLI commands

```sh
# Scaffold a new provider for the first time
terrafactor provider build provider.hcl

# Preview changes without writing files
terrafactor provider plan provider.hcl

# Re-render after OpenAPI spec changes
terrafactor provider apply provider.hcl

# Remove generated provider (can be rebuilt)
terrafactor provider destroy provider.hcl

# Build binary and wire dev_overrides in ~/.terraformrc
terrafactor provider install provider.hcl

# Remove binary and dev_overrides entry
terrafactor provider uninstall provider.hcl
```

All commands show a Terraform-style diff before making changes and require `yes` confirmation.

### Generated provider structure

```
engine/store/<name>/
├── internal/provider/
│   ├── provider.go               # APIClient, Configure, resource/datasource registration
│   ├── <resource>_resource.go    # APIModel, CRUD HTTP calls, state management
│   ├── <resource>_data_source.go
│   └── <resource>_ephemeral_resource.go
├── examples/
└── ...
```

---

## Terraform-Style Output (embeddable)

A secondary utility for printing JSON data in Terraform-style colored output. Useful for CLIs or tools that want to show resource changes in a familiar format.

### Installation

```sh
go get github.com/cloudputation/terrafactor
```

### Import

```go
import terrafactor "github.com/cloudputation/terrafactor/packages/terrafactor"
```

### Usage

```go
err := terrafactor.Print(data, "create", "    ", os.Stdout)
// operationTag: "create" (green +) or "destroy" (red -)
```

`Print` accepts any `interface{}` (typically an unmarshaled JSON object), an operation tag, an indent string, and an `io.Writer`.

### Error handling

Passing an unsupported `operationTag` returns:
```
invalid operation: <operation>. Supported operations are 'create' or 'destroy'
```

### CLI

```sh
terrafactor print <json_file> <create|destroy>
```
