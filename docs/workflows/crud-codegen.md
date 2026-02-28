# CRUD HTTP Code Generation in Terrafactor

## Overview

Terrafactor generates Terraform provider Go code from OpenAPI specs. This document describes how real HTTP CRUD operations are extracted from the spec and wired into generated provider code.

## Architecture

### APIClient

All generated providers use a shared `APIClient` wrapper defined in `provider.go`:

```go
type APIClient struct {
    HTTPClient *http.Client
    Endpoint   string
}
```

The provider's `Configure()` reads the `endpoint` attribute, trims trailing slashes, and passes `*APIClient` to all resources and data sources via `resp.ResourceData` / `resp.DataSourceData`.

### APIModel

Each resource generates a parallel `APIModel` struct with native Go types for JSON serialization:

- `types.String` → `string`
- `types.Int64` → `*int64`
- `types.Bool` → `*bool`
- `types.List` → `[]string`
- `types.Map` → `map[string]string`
- `types.Object` → `map[string]interface{}`

`toAPIModel()` and `fromAPIModel()` conversion methods are generated per resource.

### CRUDOp

Each resource spec carries four optional CRUD operation descriptors:

```go
type CRUDOp struct {
    Method      string // "POST", "GET", "PUT", "PATCH", "DELETE"
    Path        string // e.g. "/sentinel/surveillance/policies/{name}"
    OperationID string // e.g. "createPolicy"
}
```

Mapped from OpenAPI paths:
- POST on collection path → Create
- GET on detail path → Read
- PUT on detail path → Update (PATCH as fallback)
- DELETE on detail path → Delete

### Path Parameters

Detail paths follow the pattern `{collectionPath}/{param}`. The param name is extracted and stored as `PathParam`. Path substitution in generated code:

```go
url = strings.Replace(url, "{name}", data.Name.ValueString(), 1)
```

If `PathParam != "id"`, the generated code sets `data.Id = data.Name` so `ImportStatePassthroughID` works.

## Response Envelope

Generated code expects API responses in the form:

```json
{
  "success": true,
  "data": { ... },
  "error": { "code": "...", "message": "..." }
}
```

On `!success`, an error diagnostic is added with `error.code` and `error.message`. On 404 during Read, `resp.State.RemoveResource()` is called.

## Template Variables

These fields are available in all per-resource templates:

| Field | Type | Description |
|-------|------|-------------|
| `CollectionPath` | `string` | e.g. `/sentinel/surveillance/policies` |
| `DetailPath` | `string` | e.g. `/sentinel/surveillance/policies/{name}` |
| `PathParam` | `string` | e.g. `name` |
| `PathParamField` | `string` | snake_case model field mapping to path param |
| `PathParamGo` | `string` | PascalCase version for Go struct access |
| `Create` | `*CRUDOp` | nil if no POST found |
| `Read` | `*CRUDOp` | nil if no GET found |
| `Update` | `*CRUDOp` | nil if no PUT/PATCH found |
| `Delete` | `*CRUDOp` | nil if no DELETE found |

Guard clauses `{{if .Create}}...{{else}}` ensure missing operations emit diagnostics instead of panics.

## Files Involved

| File | Role |
|------|------|
| `packages/terrafactor/plan.go` | Extracts CRUD ops from OpenAPI spec |
| `packages/terrafactor/create.go` | Wires CRUD fields into template data |
| `engine/templates/provider.go.tmpl` | Defines APIClient, wires endpoint |
| `engine/templates/resource.go.tmpl` | Generates APIModel + real CRUD methods |
| `engine/templates/data_source.go.tmpl` | Generates real Read via GET |
| `engine/templates/ephemeral_resource.go.tmpl` | Generates real Open via GET |
