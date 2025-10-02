# Code Generation Guide

Complete guide to the template-based code generation system in OpenCHAMI Inventory.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Templates](#templates)
- [Template Variables](#template-variables)
- [Workflow](#workflow)
- [Modifying Templates](#modifying-templates)
- [Adding Features](#adding-features)
- [Debugging](#debugging)
- [Best Practices](#best-practices)

## Overview

OpenCHAMI Inventory uses **code generation** to maintain consistency across resource types while minimizing boilerplate. When you define a resource struct, the generator creates:

- REST API handlers (CRUD operations)
- Storage layer operations
- Type-safe HTTP client library
- CLI commands
- Request/response models
- Route registration
- Policy integration

### Benefits

✅ **Consistency** - All resources have identical patterns
✅ **Maintainability** - Fix once in template, applies everywhere  
✅ **Type Safety** - Compile-time checking across entire stack
✅ **Productivity** - Add new resource types in minutes
✅ **Documentation** - Generated code includes comprehensive comments

## Architecture

```
Resource Definition (Go struct)
    ↓
Code Generator (cmd/codegen)
    ↓
Templates (pkg/codegen/templates/*.tmpl)
    ↓
Generated Code
    ├─ REST API (cmd/server/*_generated.go)
    ├─ Storage (internal/storage/storage_generated.go)
    ├─ Client Library (pkg/client/*.go)
    └─ CLI (cmd/inventory-cli/generated.go)
```

### Generated vs Manual Code

**✅ Edit These (Manual Code):**
- `pkg/resources/*/` - Resource type definitions
- `pkg/codegen/templates/` - Code generation templates
- `cmd/server/main.go` - Server configuration
- `pkg/policies/` - Authorization policies

**⚠️ Don't Edit These (Generated Code):**
- `cmd/server/*_generated.go` - Auto-generated handlers
- `internal/storage/storage_generated.go` - Auto-generated storage
- `pkg/client/client.go` - Auto-generated client
- `cmd/inventory-cli/*_generated.go` - Auto-generated CLI

## Templates

### Template Files

| Template | Generates | Output Location |
|----------|-----------|-----------------|
| `handlers.go.tmpl` | REST API endpoints | `cmd/server/*_handlers_generated.go` |
| `storage.go.tmpl` | Data persistence | `internal/storage/storage_generated.go` |
| `client.go.tmpl` | HTTP client library | `pkg/client/client.go` |
| `client-models.go.tmpl` | Client types | `pkg/client/models.go` |
| `client-cmd.go.tmpl` | CLI application | `cmd/inventory-cli/*_generated.go` |
| `models.go.tmpl` | Server types | `cmd/server/models_generated.go` |
| `routes.go.tmpl` | URL routing | `cmd/server/routes_generated.go` |
| `policies.go.tmpl` | Auth integration | `cmd/server/policies_generated.go` |

### Template Structure

Each template is a Go text template with:

1. **Header Comment** - Documentation about the template
2. **Package Declaration** - Go package for generated code
3. **Imports** - Required packages
4. **Generated Code** - Template logic with {{variables}}

**Example:**
```go
// handlers.go.tmpl
package main

import (
    "net/http"
    {{.PackageAlias}} "{{.Package}}"
)

// List{{.Name}}s returns all {{.Name}} resources
func List{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s()
    if err != nil {
        return nil, err
    }
    return {{camelCase .PluralName}}, nil
}
```

## Template Variables

### Resource Metadata

Templates receive metadata about each resource:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.Name}}` | Resource type name | `BMC`, `Node` |
| `{{.PluralName}}` | Plural form | `bmcs`, `nodes` |
| `{{.Package}}` | Full import path | `github.com/openchami/inventory/pkg/resources/bmc` |
| `{{.PackageAlias}}` | Package alias | `bmc`, `node` |
| `{{.TypeName}}` | Full type reference | `*bmc.BMC` |
| `{{.SpecType}}` | Spec type name | `bmc.BMCSpec` |
| `{{.StatusType}}` | Status type name | `bmc.BMCStatus` |
| `{{.URLPath}}` | REST endpoint path | `/bmcs` |
| `{{.StorageName}}` | Storage function suffix | `BMC`, `BootConfig` |
| `{{.RequiresAuth}}` | Requires authentication | `true`, `false` |

### Global Data

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.ModulePath}}` | Go module path | `github.com/openchami/inventory` |
| `{{.PackageName}}` | Target package | `main`, `storage`, `client` |
| `{{.Resources}}` | Array of all resources | `[BMC, Node, FRU, BootConfiguration]` |

### Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `{{camelCase .Name}}` | Convert to camelCase | `BMC` → `bmc` |
| `{{toLower .Name}}` | Convert to lowercase | `BMC` → `bmc` |
| `{{toUpper .Name}}` | Convert to uppercase | `bmc` → `BMC` |

## Workflow

### Development Cycle

```bash
# 1. Make changes to resource or template
vim pkg/resources/bmc/bmc.go
# or
vim pkg/codegen/templates/handlers.go.tmpl

# 2. Regenerate all code
make dev

# 3. Test changes
make test

# 4. Run server with new code
./bin/server --disable-auth
```

### What `make dev` Does

```bash
make dev
  ├─ make clean          # Remove generated files
  ├─ make generate       # Run code generation
  │   ├─ generate-storage
  │   ├─ generate-server
  │   ├─ generate-client
  │   └─ generate-client-cmd
  ├─ go mod tidy         # Update dependencies
  ├─ go fmt              # Format code
  ├─ make build          # Build binaries
  └─ make test           # Run tests
```

## Modifying Templates

### Example 1: Add New Handler Function

**Goal:** Add count endpoint for each resource.

**Edit:** `pkg/codegen/templates/handlers.go.tmpl`

```go
// Add after existing handlers

// Get{{.Name}}Count returns the count of {{.Name}} resources
func Get{{.Name}}Count(c fuego.ContextNoBody) (map[string]int, error) {
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s(c.Context())
    if err != nil {
        return nil, fuego.HTTPError{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to load {{.PluralName}}: %w", err),
        }
    }
    
    return map[string]int{
        "count": len({{camelCase .PluralName}}),
    }, nil
}
```

**Register route:** `pkg/codegen/templates/routes.go.tmpl`

```go
// Add in route registration section
fuego.Get(server, "{{.URLPath}}/count", Get{{.Name}}Count)
```

**Regenerate:**
```bash
make dev
```

**Test:**
```bash
curl http://localhost:9999/bmcs/count
# {"count": 25}
```

### Example 2: Add Validation

**Goal:** Validate resources before creation.

**Edit:** `pkg/codegen/templates/handlers.go.tmpl`

```go
// Modify Create handler
func Create{{.Name}}(c fuego.ContextWithBody[{{.TypeName}}]) ({{.TypeName}}, error) {
    resource := c.Body()
    
    // Add validation
    if err := validate{{.Name}}(resource); err != nil {
        return nil, fuego.BadRequestError{
            Err: err,
        }
    }
    
    // ... rest of create logic
}

// Add validation function
func validate{{.Name}}(resource {{.TypeName}}) error {
    if resource.Metadata.Name == "" {
        return fmt.Errorf("name is required")
    }
    // Add more validation rules
    return nil
}
```

### Example 3: Add Query Parameters

**Goal:** Add filtering by label.

**Edit:** `pkg/codegen/templates/handlers.go.tmpl`

```go
func List{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
    // Load all resources
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s(c.Context())
    if err != nil {
        return nil, err
    }
    
    // Filter by label if provided
    labelFilter := c.QueryParam("label")
    if labelFilter != "" {
        filtered := make([]{{.TypeName}}, 0)
        for _, resource := range {{camelCase .PluralName}} {
            for key, value := range resource.Metadata.Labels {
                if fmt.Sprintf("%s=%s", key, value) == labelFilter {
                    filtered = append(filtered, resource)
                    break
                }
            }
        }
        return filtered, nil
    }
    
    return {{camelCase .PluralName}}, nil
}
```

**Usage:**
```bash
curl "http://localhost:9999/bmcs?label=datacenter=dc1"
```

## Adding Features

### Add New Endpoint

1. **Edit handler template** - Add handler function
2. **Edit routes template** - Register new route
3. **Edit client template** (optional) - Add client method
4. **Regenerate** - `make dev`

### Add Middleware

1. **Edit routes template** - Add middleware
2. **Regenerate** - `make dev`

**Example:**
```go
// In routes.go.tmpl
server.Use(myCustomMiddleware)
fuego.Get(server, "{{.URLPath}}", List{{.Name}}s)
```

### Add Response Headers

1. **Edit handlers template** - Set headers in handlers
2. **Regenerate** - `make dev`

**Example:**
```go
func List{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
    c.Response().Header().Set("X-Total-Count", strconv.Itoa(len(resources)))
    return resources, nil
}
```

## Debugging

### Template Syntax Errors

**Error:**
```
Error: template: handlers.go.tmpl:42: undefined variable "ResourceType"
```

**Solution:**
```bash
# Check template variable names
grep -n "ResourceType" pkg/codegen/templates/handlers.go.tmpl

# Fix variable name (should be .Name or .TypeName)
# Edit template and regenerate
make dev
```

### Generated Code Won't Compile

**Error:**
```
cmd/server/bmc_handlers_generated.go:25: undefined: storage.LoadAllBMCs
```

**Diagnosis:**
```bash
# Check what got generated
cat cmd/server/bmc_handlers_generated.go | grep LoadAll

# Check storage template
cat pkg/codegen/templates/storage.go.tmpl | grep LoadAll
```

**Solution:**
```bash
# Verify template has correct function name
# Fix template and regenerate
make dev
```

### View Generated Code

```bash
# View specific generated file
cat cmd/server/bmc_handlers_generated.go

# View all generated files
make discover

# View template before generation
cat pkg/codegen/templates/handlers.go.tmpl
```

### Test Template Changes

```bash
# Clean and regenerate
make clean
make generate

# Check for compilation errors
go build ./cmd/server

# Run tests
make test
```

## Best Practices

### 1. Keep Templates Focused

Each template should generate one type of code:
- `handlers.go.tmpl` → API handlers only
- `storage.go.tmpl` → Storage operations only
- Don't mix concerns

### 2. Use Descriptive Variable Names

```go
// ❌ Bad
{{.N}} {{.T}} {{.P}}

// ✅ Good
{{.Name}} {{.TypeName}} {{.PluralName}}
```

### 3. Add Comments to Generated Code

```go
// ✅ Generated code should be documented
// List{{.Name}}s returns all {{.Name}} resources from storage
func List{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
    // ...
}
```

### 4. Handle Errors Consistently

```go
// ✅ Consistent error handling
if err != nil {
    return nil, fuego.HTTPError{
        Status: http.StatusInternalServerError,
        Err:    fmt.Errorf("failed to load {{.PluralName}}: %w", err),
    }
}
```

### 5. Use Template Functions

```go
// ✅ Use provided functions
{{camelCase .PluralName}}  // bmc.BMCs → bmcs

// ❌ Don't hardcode transformations
{{.PluralName}}  // Might not match expected case
```

### 6. Test After Every Change

```bash
# Always test after template changes
make dev && make test
```

### 7. Document Template Changes

Add comments explaining complex template logic:

```go
{{/* 
  This section generates version-aware handlers.
  It checks if the resource has multiple versions registered
  and generates conversion logic if needed.
*/}}
{{if .HasMultipleVersions}}
    // Version conversion code here
{{end}}
```

## Template Documentation

Each generated file includes comprehensive header comments:

```go
// Code generated by codegen. DO NOT EDIT.
//
// This file contains REST API handlers for BMC resources.
// Generated from: pkg/codegen/templates/handlers.go.tmpl
// 
// To modify this code:
//   1. Edit the template file: pkg/codegen/templates/handlers.go.tmpl
//   2. Run 'make dev' to regenerate
//   3. Do NOT edit this file directly - changes will be lost
//
// Features:
//   - CRUD operations (List, Get, Create, Update, Delete)
//   - Policy-based authorization
//   - Version negotiation support
//   - Error handling with appropriate HTTP status codes
//
// Extension Points:
//   - Add new handlers: Edit handlers.go.tmpl
//   - Modify authorization: Edit policy files in pkg/resources/
//   - Change response format: Edit models.go.tmpl
```

## See Also

- [Development Guide](./DEVELOPMENT.md) - Complete development guide
- [Testing Guide](./TESTING.md) - Testing and development mode
- [Template README](../../pkg/codegen/templates/README.md) - Template details
- [User Guide](../user/USER-GUIDE.md) - Using the generated code
