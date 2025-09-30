# Code Generation Templates

This directory contains Go templates used to generate consistent code across all resource types.

## Templates Overview

| Template | Purpose | Output File |
|----------|---------|-------------|
| `handlers.go.tmpl` | REST API endpoints (CRUD) | `cmd/server/*_handlers_generated.go` |
| `storage.go.tmpl` | Data persistence operations | `internal/storage/storage_generated.go` |
| `client.go.tmpl` | HTTP client library | `pkg/client/client.go` |
| `client-models.go.tmpl` | Client request/response types | `pkg/client/models.go` |
| `models.go.tmpl` | Server request/response types | `cmd/server/models_generated.go` |
| `routes.go.tmpl` | URL routing registration | `cmd/server/routes_generated.go` |
| `policies.go.tmpl` | Authentication integration | `cmd/server/policies_generated.go` |

## Template Variables

Each template receives the following data:

### Resource Metadata
- `{{.Name}}` - Resource type name (e.g., "BMC", "Node")
- `{{.PluralName}}` - Plural form (e.g., "bmcs", "nodes") 
- `{{.Package}}` - Import path (e.g., "github.com/openchami/inventory/pkg/resources/bmc")
- `{{.PackageAlias}}` - Package alias (e.g., "bmc", "node")
- `{{.TypeName}}` - Full type reference (e.g., "*bmc.BMC")
- `{{.SpecType}}` - Spec type name (e.g., "bmc.BMCSpec")
- `{{.StatusType}}` - Status type name (e.g., "bmc.BMCStatus")
- `{{.URLPath}}` - REST endpoint path (e.g., "/bmcs")
- `{{.StorageName}}` - Storage function suffix (e.g., "BMC", "BootConfig")
- `{{.RequiresAuth}}` - Boolean indicating if auth is required

### Global Data
- `{{.ModulePath}}` - Go module path
- `{{.PackageName}}` - Target package name
- `{{.Resources}}` - Array of all resources

## Template Functions

Available functions for data transformation:
- `{{camelCase .Name}}` - Convert to camelCase
- `{{toLower .Name}}` - Convert to lowercase
- `{{toUpper .Name}}` - Convert to uppercase

## Modifying Templates

1. **Edit the template file** directly (e.g., `handlers.go.tmpl`)
2. **Run code generation**: `make dev`
3. **Test the changes**: Generated files will reflect your modifications

### Example: Adding a New Handler Function

```go
// Add to handlers.go.tmpl
// Get{{.Name}}Count returns the count of {{.Name}} resources
func Get{{.Name}}Count(c fuego.ContextNoBody) (int, error) {
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s()
    if err != nil {
        return 0, fuego.HTTPError{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to load {{.PluralName}}: %w", err),
        }
    }
    return len({{camelCase .PluralName}}), nil
}
```

## Best Practices

- **Keep templates focused**: Each template handles one type of generated artifact
- **Use descriptive comments**: Help developers understand generated code purpose
- **Include error handling**: Generated code should be robust
- **Follow Go conventions**: Generated code should be idiomatic
- **Add TODO comments**: Mark areas where manual customization is needed

## Testing Templates

After modifying templates:

1. Run `make dev` to regenerate all code
2. Run `make test` to verify generated code compiles
3. Test API endpoints with generated handlers
4. Verify client library works with generated types

## Debugging

If template generation fails:

1. Check template syntax with `go run cmd/codegen/main.go`
2. Verify template variables are correctly referenced
3. Use `make templates` to view template content
4. Check generated code for compilation errors

## File Structure

```
pkg/codegen/templates/
├── README.md                    # This file
├── handlers.go.tmpl            # REST API handlers
├── storage.go.tmpl             # Data persistence  
├── client.go.tmpl              # HTTP client
├── client-models.go.tmpl       # Client types
├── models.go.tmpl              # Server types
├── routes.go.tmpl              # URL routing
└── policies.go.tmpl            # Authentication
```