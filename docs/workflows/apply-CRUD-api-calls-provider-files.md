# Plan: Implement Real HTTP CRUD in Terrafactor Generated Providers

## Context

Terrafactor generates Terraform provider Go code from OpenAPI specs. The generated CRUD methods are stubs (`// TODO: implement API call`) — they save Terraform state locally but never call the actual API. `terraform apply` succeeds but nothing happens on the real service.

The fix: make Terrafactor extract CRUD operation metadata (paths, HTTP methods, path params) from the OpenAPI spec and pass it to templates, which then generate real HTTP client code.

## Files to Modify

| File | Changes |
|------|---------|
| `packages/terrafactor/plan.go` | Add `Patch` to path item; add `CRUDOp` struct; extend `ResourceSpec` + `TemplateData` with CRUD fields; modify `ParseOpenAPI()` to extract ops; add `findDetailPath()` |
| `packages/terrafactor/create.go` | Populate new CRUD fields in per-resource rendering loop (line 53-58) |
| `engine/.../provider.go.tmpl` | Add `APIClient` struct; wire endpoint URL; pass `*APIClient` to resources |
| `engine/.../resource.go.tmpl` | Replace `*http.Client` with `*APIClient`; add `APIModel` + conversions; generate real Create/Read/Update/Delete HTTP calls |
| `engine/.../data_source.go.tmpl` | Same `*APIClient` pattern; generate real Read |
| `engine/.../ephemeral_resource.go.tmpl` | Same `*APIClient` pattern; generate real Open |

## Changes

### 1. `plan.go` — Parser Enhancement

**1a. Add `Patch` to `openAPIPathItem` (line 105-109):**
```go
type openAPIPathItem struct {
    Get    *openAPIOperation `yaml:"get"`
    Post   *openAPIOperation `yaml:"post"`
    Put    *openAPIOperation `yaml:"put"`
    Patch  *openAPIOperation `yaml:"patch"`
    Delete *openAPIOperation `yaml:"delete"`
}
```

**1b. Add CRUD operation struct:**
```go
type CRUDOp struct {
    Method      string // "POST", "GET", "PUT", "PATCH", "DELETE"
    Path        string // e.g. "/sentinel/surveillance/policies/{name}"
    OperationID string // e.g. "createPolicy"
}
```

**1c. Extend `ResourceSpec`:**
```go
CollectionPath string   // e.g. "/sentinel/surveillance/policies"
DetailPath     string   // e.g. "/sentinel/surveillance/policies/{name}"
PathParam      string   // e.g. "name" — from {name} segment
PathParamField string   // snake_case model field that maps to path param
Create         *CRUDOp
Read           *CRUDOp
Update         *CRUDOp
Delete         *CRUDOp
```

**1d. Extend `TemplateData`:** Same fields + `PathParamGo string` (PascalCase for Go struct access).

**1e. Add `findDetailPath()` helper:**
Scans all paths for one matching `collectionPath + "/{param}"`. Returns the detail path and param name.

**1f. Modify `ParseOpenAPI()`:**
After extracting schema/fields for a collection path:
- Find detail path via `findDetailPath()`
- Map: POST on collection → Create, GET on detail → Read, PUT/PATCH on detail → Update, DELETE on detail → Delete
- Prefer PUT for update; fall back to PATCH if no PUT

### 2. `create.go` — Wire CRUD fields (line 53-58)

Add after `resData.Fields = res.Fields`:
```go
resData.CollectionPath = res.CollectionPath
resData.DetailPath = res.DetailPath
resData.PathParam = res.PathParam
resData.PathParamField = res.PathParamField
resData.PathParamGo = snakeToPascal(res.PathParamField)
resData.Create = res.Create
resData.Read = res.Read
resData.Update = res.Update
resData.Delete = res.Delete
```

### 3. `provider.go.tmpl` — APIClient

Replace bare `http.DefaultClient` with a wrapper that carries the endpoint URL:

```go
type APIClient struct {
    HTTPClient *http.Client
    Endpoint   string
}
```

In `Configure()`: read `data.Endpoint.ValueString()`, trim trailing `/`, create `&APIClient{HTTPClient: http.DefaultClient, Endpoint: endpoint}`, pass to `resp.ResourceData` and `resp.DataSourceData`.

### 4. `resource.go.tmpl` — Real CRUD Methods

**4a. APIModel struct for JSON serialization:**

Terraform framework types (`types.String`, `types.Int64`, etc.) don't marshal/unmarshal with `encoding/json`. Generate a parallel struct with native Go types:

```go
type {{.ResourceNamePascal}}APIModel struct {
    // types.String  → string
    // types.Int64   → *int64  (pointer for omitempty on zero)
    // types.Bool    → *bool
    // types.List    → []string (for string element arrays)
    // types.Map     → map[string]string
    // types.Object  → map[string]interface{}
}
```

Generate `toAPIModel()` (ResourceModel → APIModel) and `fromAPIModel()` (APIModel → ResourceModel) conversion methods. Each field converts based on its `GoType`:
- `types.String`: `ValueString()` / `types.StringValue()`
- `types.Int64`: `ValueInt64()` / `types.Int64Value()`
- `types.Bool`: `ValueBool()` / `types.BoolValue()`
- `types.Object` (nested): `map[string]interface{}` with recursive handling via template-generated helpers

**4b. CRUD methods pattern (Create example):**

```go
func (r *{{.ResourceNamePascal}}Resource) Create(...) {
    // 1. Read plan into ResourceModel
    // 2. Convert to APIModel via toAPIModel()
    // 3. json.Marshal the APIModel
    // 4. Build URL: r.client.Endpoint + "{{.Create.Path}}"
    // 5. http.NewRequestWithContext(ctx, "{{.Create.Method}}", url, body)
    // 6. Set Content-Type header
    // 7. r.client.HTTPClient.Do(httpReq)
    // 8. Decode response envelope {success, data, error}
    // 9. If !success, AddError with error.code + error.message
    // 10. fromAPIModel(envelope.Data) back into ResourceModel
    // 11. Set ID (from name field or response id field)
    // 12. Save to state
}
```

Read: same but GET to detail path with path param substitution. 404 → `resp.State.RemoveResource()`.
Update: same as Create but to detail path with PUT/PATCH.
Delete: same but no request body, no response data parsing.

**4c. Path param substitution:**
```go
url = strings.Replace(url, "{name}", data.Name.ValueString(), 1)
```

**4d. ID mapping:**
- If `PathParam == "id"`: ID comes from the response naturally
- If `PathParam != "id"` (e.g., `"name"`): set `data.Id = data.Name` so Terraform's `ImportStatePassthroughID` works

**4e. Guard clauses:**
All CRUD ops use `{{if .Create}}...{{else}}` so missing operations generate a diagnostic error or no-op.

### 5. `data_source.go.tmpl` + `ephemeral_resource.go.tmpl`

Same `*APIClient` pattern. Data source Read and ephemeral Open generate GET requests using the same envelope parsing.

## Execution Strategy

**Phase 1** (2 parallel agents):
- Agent A: `plan.go` — all parser changes (CRUDOp, ResourceSpec, ParseOpenAPI, findDetailPath)
- Agent B: `provider.go.tmpl` — APIClient struct and endpoint wiring

**Phase 2** (2 parallel agents, after Phase 1):
- Agent A: `resource.go.tmpl` — APIModel, conversions, all 4 CRUD methods
- Agent B: `data_source.go.tmpl` + `ephemeral_resource.go.tmpl`

**Phase 3** (sequential):
- `create.go` — wire CRUD fields (small, depends on Phase 1 struct names)
- Build terrafactor: `go build ./...`
- Regenerate: `provider apply`
- Install: `provider install`
- Test: `terraform plan` + `terraform apply` against running Sentinel

## Pre-Implementation Step

Write this plan as project documentation at `docs/workflows/crud-codegen.md` in the terrafactor project before beginning implementation.

## Verification

1. `go build ./...` in terrafactor root — compiles clean
2. `go vet ./...` — no warnings
3. Regenerate sentinel provider, verify generated resource file has:
   - `APIClient` instead of `*http.Client`
   - `APIModel` struct with native Go types
   - Real HTTP calls in Create/Read/Update/Delete
   - Correct URL construction with path param substitution
4. Generated provider compiles: `go build ./...` in `engine/store/sentinel/`
5. `terraform plan` shows resources with match sub-fields
6. `terraform apply` — Sentinel docker logs show incoming POST requests
7. `terraform destroy` — Sentinel logs show DELETE requests
