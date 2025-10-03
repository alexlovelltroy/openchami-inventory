# Chi Migration Proposal: Replacing go-fuego with Chi + kin-openapi

## Executive Summary

This proposal outlines the migration from **go-fuego** to **Chi** as the web framework, while maintaining automatic OpenAPI documentation generation through code-generated **kin-openapi** specifications. This approach provides the flexibility and maturity of Chi while preserving the developer experience of automatic API documentation.

**Key Benefits:**
- ✅ More mature, battle-tested web framework (Chi: 8+ years vs go-fuego: <2 years)
- ✅ Flexible routing with middleware composition
- ✅ Automatic OpenAPI generation (no docstring annotations required)
- ✅ Seamless integration with existing code generation architecture
- ✅ Lower risk for production deployments
- ✅ Idiomatic Go patterns

**Migration Effort:** 2-3 days
**Risk Level:** Low
**Complexity:** Low-Medium

---

## Table of Contents

1. [Motivation](#motivation)
2. [Current State Analysis](#current-state-analysis)
3. [Proposed Architecture](#proposed-architecture)
4. [Implementation Plan](#implementation-plan)
5. [Code Examples](#code-examples)
6. [Migration Strategy](#migration-strategy)
7. [Testing Strategy](#testing-strategy)
8. [Risk Assessment](#risk-assessment)
9. [Timeline](#timeline)

---

## Motivation

### Why Switch from go-fuego?

**Current Framework: go-fuego (v0.18.8)**
- **Age:** < 2 years old (first release 2023)
- **Stars:** ~1.1k
- **Production Use:** Limited
- **Community:** Small, growing
- **API Stability:** Breaking changes possible

**Concerns:**
1. ⚠️ **Maturity Risk** - Very young project, not battle-tested
2. ⚠️ **Community Size** - Small community means slower issue resolution
3. ⚠️ **Production Track Record** - Limited real-world validation
4. ⚠️ **Breaking Changes** - Pre-1.0 status means API instability
5. ⚠️ **Dependency Risk** - Project could be abandoned or change direction

### Why Chi?

**Proposed Framework: Chi (v5)**
- **Age:** 8+ years (first release 2016)
- **Stars:** 18k+
- **Production Use:** Extensive (used by major companies)
- **Community:** Large, active
- **API Stability:** Stable v5 API

**Benefits:**
1. ✅ **Battle-Tested** - 8+ years in production environments
2. ✅ **Mature** - Stable API, predictable behavior
3. ✅ **Idiomatic Go** - Built on net/http, follows Go conventions
4. ✅ **Flexible Routing** - Powerful URL patterns and sub-routers
5. ✅ **Middleware Composition** - Rich middleware ecosystem
6. ✅ **Performance** - Excellent performance benchmarks
7. ✅ **Documentation** - Comprehensive docs and examples

### Why Keep Automatic OpenAPI?

**Requirement:** Maintain automatic OpenAPI documentation without manual annotations

**Solution:** Generate OpenAPI specs via **kin-openapi** (already in dependencies)

**Why This Works:**
- ✅ **Already Dependency** - `kin-openapi` in go.mod (v0.133.0)
- ✅ **Type-Driven** - Generate specs from Go types automatically
- ✅ **Code Generation** - Fits perfectly with existing template system
- ✅ **No Annotations** - No docstring burden
- ✅ **Framework Agnostic** - Works with any HTTP framework

---

## Current State Analysis

### Current go-fuego Usage

**File:** `cmd/server/main.go`
```go
import (
    "github.com/go-fuego/fuego"
)

func runServer(cmd *cobra.Command, args []string) {
    s := fuego.NewServer(
        fuego.WithAddr(fmt.Sprintf("%s:%d", host, port)),
    )

    // Middleware
    fuego.Use(s, versioning.VersionNegotiationMiddleware(versionRegistry))

    // Register routes (generated)
    RegisterGeneratedRoutes(s)

    // Health check
    fuego.Get(s, "/health", func(c fuego.ContextNoBody) (map[string]string, error) {
        return map[string]string{
            "status":  "ok",
            "version": "1.0.0",
        }, nil
    })

    s.Run()
}
```

**Generated Handler Pattern (go-fuego):**
```go
// From templates/handlers.go.tmpl
func ListBMCs(c fuego.ContextNoBody) ([]bmc.BMC, error) {
    bmcs, err := storage.LoadAllBMCs(c.Context())
    if err != nil {
        return nil, fuego.HTTPError{
            Status: http.StatusInternalServerError,
            Err:    fmt.Errorf("failed to load BMCs: %w", err),
        }
    }
    return bmcs, nil
}

func GetBMC(c fuego.ContextNoBody) (*bmc.BMC, error) {
    uid := c.PathParam("uid")
    bmc, err := storage.LoadBMC(c.Context(), uid)
    if err != nil {
        return nil, fuego.HTTPError{
            Status: http.StatusNotFound,
            Err:    fmt.Errorf("BMC not found: %s", uid),
        }
    }
    return bmc, nil
}

func CreateBMC(c fuego.ContextWithBody[bmc.CreateBMCRequest]) (*bmc.BMC, error) {
    req := c.Body()

    // Create BMC
    newBMC := &bmc.BMC{
        Spec: req.ToBMCSpec(),
    }

    if err := storage.SaveBMC(c.Context(), newBMC); err != nil {
        return nil, fuego.HTTPError{
            Status: http.StatusInternalServerError,
            Err:    err,
        }
    }

    return newBMC, nil
}
```

**OpenAPI Generation (go-fuego):**
- Automatic via fuego's built-in reflection
- Available at `/swagger/openapi.json`
- Generated from handler signatures

---

## Proposed Architecture

### New Stack: Chi + kin-openapi

```
┌─────────────────────────────────────────────────────────────┐
│                     Code Generation                          │
│                                                               │
│  Resource Definitions (pkg/resources/)                       │
│             ↓                                                 │
│  Templates (pkg/codegen/templates/)                          │
│             ↓                                                 │
│  ┌──────────────────┬──────────────────┬──────────────────┐ │
│  │  Chi Handlers    │  OpenAPI Spec    │  Route Registry  │ │
│  │  (generated)     │  (generated)     │  (generated)     │ │
│  └──────────────────┴──────────────────┴──────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│                      Runtime                                 │
│                                                               │
│  Chi Router ──→ Middleware ──→ Handlers ──→ Storage         │
│      ↓                                                        │
│  OpenAPI Endpoints:                                          │
│    - GET /openapi.json  (generated spec)                     │
│    - GET /docs          (Swagger UI)                         │
└─────────────────────────────────────────────────────────────┘
```

### Component Breakdown

**1. Chi Router**
- URL routing and matching
- Sub-router composition
- Middleware pipeline

**2. kin-openapi Generator**
- Generate specs from Go types
- Type-safe schema generation
- No manual annotations

**3. Code Generation Templates**
- `handlers.go.tmpl` - Chi handlers
- `openapi-generator.go.tmpl` - OpenAPI spec generation
- `routes.go.tmpl` - Route registration

---

## Implementation Plan

### Phase 1: Add Dependencies (30 minutes)

**Update go.mod:**
```bash
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/go-chi/httprate
# kin-openapi already present
```

**Dependencies to add:**
- `github.com/go-chi/chi/v5` - Router
- `github.com/go-chi/cors` - CORS middleware
- `github.com/go-chi/httprate` - Rate limiting (optional)

**Dependencies to remove (after migration):**
- `github.com/go-fuego/fuego`

---

### Phase 2: Update Templates (4-6 hours)

#### Template 1: `pkg/codegen/templates/handlers.go.tmpl`

**Current (go-fuego):**
```go
func List{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s(c.Context())
    return {{camelCase .PluralName}}, err
}
```

**New (Chi):**
```go
func List{{.Name}}s(w http.ResponseWriter, r *http.Request) {
    {{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s(r.Context())
    if err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, {{camelCase .PluralName}})
}

func Get{{.Name}}(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")
    resource, err := storage.Load{{.StorageName}}(r.Context(), uid)
    if err != nil {
        respondError(w, http.StatusNotFound, err)
        return
    }
    respondJSON(w, http.StatusOK, resource)
}

func Create{{.Name}}(w http.ResponseWriter, r *http.Request) {
    var req {{.PackageAlias}}.Create{{.Name}}Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    resource := &{{.TypeName}}{
        Spec: req.To{{.Name}}Spec(),
    }

    if err := storage.Save{{.StorageName}}(r.Context(), resource); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusCreated, resource)
}

func Update{{.Name}}(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")

    var req {{.PackageAlias}}.Update{{.Name}}Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    resource, err := storage.Load{{.StorageName}}(r.Context(), uid)
    if err != nil {
        respondError(w, http.StatusNotFound, err)
        return
    }

    req.ApplyTo{{.Name}}(resource)

    if err := storage.Save{{.StorageName}}(r.Context(), resource); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusOK, resource)
}

func Delete{{.Name}}(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")

    if err := storage.Delete{{.StorageName}}(r.Context(), uid); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// Helper functions (generated once in common file)
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{
        "error": err.Error(),
    })
}
```

#### Template 2: `pkg/codegen/templates/routes.go.tmpl`

**New (Chi):**
```go
// Code generated by codegen. DO NOT EDIT.
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

func RegisterGeneratedRoutes(r chi.Router) {
    // Standard middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // CORS (if enabled)
    if corsEnabled {
        r.Use(cors.Handler(cors.Options{
            AllowedOrigins:   corsOrigins,
            AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
            ExposedHeaders:   []string{"Link"},
            AllowCredentials: true,
            MaxAge:           300,
        }))
    }

    // Version negotiation middleware
    r.Use(versioning.VersionNegotiationMiddleware(versionRegistry))

    {{range .Resources}}
    // {{.Name}} routes
    r.Route("{{.URLPath}}", func(r chi.Router) {
        r.Get("/", List{{.Name}}s)
        r.Post("/", Create{{.Name}})

        r.Route("/{uid}", func(r chi.Router) {
            r.Get("/", Get{{.Name}})
            r.Put("/", Update{{.Name}})
            r.Delete("/", Delete{{.Name}})
        })
    })
    {{end}}

    // OpenAPI endpoints
    r.Get("/openapi.json", ServeOpenAPISpec)
    r.Get("/docs", ServeSwaggerUI)

    // Health check
    r.Get("/health", HealthCheck)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, 200, map[string]string{
        "status":  "ok",
        "version": "1.0.0",
    })
}
```

#### Template 3: `pkg/codegen/templates/openapi-generator.go.tmpl` (NEW)

```go
// Code generated by codegen. DO NOT EDIT.
package main

import (
    "encoding/json"
    "net/http"

    "github.com/getkin/kin-openapi/openapi3"
    "github.com/getkin/kin-openapi/openapi3gen"

    {{range .Resources}}
    {{.PackageAlias}} "{{.Package}}"
    {{end}}
)

// GenerateOpenAPISpec creates the OpenAPI 3.0 specification
func GenerateOpenAPISpec() *openapi3.T {
    spec := &openapi3.T{
        OpenAPI: "3.0.0",
        Info: &openapi3.Info{
            Title:       "OpenCHAMI Inventory API",
            Description: "HPC hardware inventory management system",
            Version:     "1.0.0",
            Contact: &openapi3.Contact{
                Name:  "OpenCHAMI Project",
                URL:   "https://openchami.org",
                Email: "support@openchami.org",
            },
            License: &openapi3.License{
                Name: "MIT",
                URL:  "https://opensource.org/licenses/MIT",
            },
        },
        Servers: openapi3.Servers{
            {
                URL:         "http://localhost:8080",
                Description: "Development server",
            },
            {
                URL:         "https://api.example.com",
                Description: "Production server",
            },
        },
        Paths: openapi3.Paths{},
        Components: &openapi3.Components{
            Schemas: make(openapi3.Schemas),
        },
    }

    {{range .Resources}}
    register{{.Name}}Paths(spec)
    {{end}}

    return spec
}

{{range .Resources}}
func register{{.Name}}Paths(spec *openapi3.T) {
    // Generate schemas from types
    resourceSchema, _ := openapi3gen.NewSchemaRefForValue(&{{.PackageAlias}}.{{.Name}}{}, spec.Components.Schemas)
    createSchema, _ := openapi3gen.NewSchemaRefForValue(&{{.PackageAlias}}.Create{{.Name}}Request{}, spec.Components.Schemas)
    updateSchema, _ := openapi3gen.NewSchemaRefForValue(&{{.PackageAlias}}.Update{{.Name}}Request{}, spec.Components.Schemas)

    // Store in components for reuse
    spec.Components.Schemas["{{.Name}}"] = resourceSchema
    spec.Components.Schemas["Create{{.Name}}Request"] = createSchema
    spec.Components.Schemas["Update{{.Name}}Request"] = updateSchema

    // Collection operations
    spec.Paths["{{.URLPath}}"] = &openapi3.PathItem{
        Summary:     "{{.Name}} collection operations",
        Description: "Operations on the collection of {{.Name}} resources",
        Get: &openapi3.Operation{
            OperationID: "list{{.Name}}s",
            Summary:     "List all {{.Name}} resources",
            Description: "Retrieve a list of all {{.Name}} resources in the inventory",
            Tags:        []string{"{{.Name}}"},
            Responses: openapi3.Responses{
                "200": &openapi3.ResponseRef{
                    Value: &openapi3.Response{
                        Description: strPtr("Successful response with list of {{.Name}} resources"),
                        Content: openapi3.Content{
                            "application/json": &openapi3.MediaType{
                                Schema: &openapi3.SchemaRef{
                                    Value: &openapi3.Schema{
                                        Type: "array",
                                        Items: &openapi3.SchemaRef{
                                            Ref: "#/components/schemas/{{.Name}}",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "500": errorResponse("Internal server error"),
            },
        },
        Post: &openapi3.Operation{
            OperationID: "create{{.Name}}",
            Summary:     "Create a new {{.Name}}",
            Description: "Create a new {{.Name}} resource in the inventory",
            Tags:        []string{"{{.Name}}"},
            RequestBody: &openapi3.RequestBodyRef{
                Value: &openapi3.RequestBody{
                    Description: "{{.Name}} creation request",
                    Required:    true,
                    Content: openapi3.Content{
                        "application/json": &openapi3.MediaType{
                            Schema: &openapi3.SchemaRef{
                                Ref: "#/components/schemas/Create{{.Name}}Request",
                            },
                        },
                    },
                },
            },
            Responses: openapi3.Responses{
                "201": &openapi3.ResponseRef{
                    Value: &openapi3.Response{
                        Description: strPtr("{{.Name}} created successfully"),
                        Content: openapi3.Content{
                            "application/json": &openapi3.MediaType{
                                Schema: &openapi3.SchemaRef{
                                    Ref: "#/components/schemas/{{.Name}}",
                                },
                            },
                        },
                    },
                },
                "400": errorResponse("Invalid request body"),
                "500": errorResponse("Internal server error"),
            },
        },
    }

    // Individual resource operations
    spec.Paths["{{.URLPath}}/{uid}"] = &openapi3.PathItem{
        Summary:     "{{.Name}} resource operations",
        Description: "Operations on individual {{.Name}} resources",
        Parameters: openapi3.Parameters{
            {
                Value: &openapi3.Parameter{
                    Name:        "uid",
                    In:          "path",
                    Description: "Unique identifier of the {{.Name}} resource",
                    Required:    true,
                    Schema: &openapi3.SchemaRef{
                        Value: &openapi3.Schema{
                            Type:    "string",
                            Format:  "uuid",
                            Example: "{{.PluralName}}-abc123",
                        },
                    },
                },
            },
        },
        Get: &openapi3.Operation{
            OperationID: "get{{.Name}}",
            Summary:     "Get a {{.Name}} by UID",
            Description: "Retrieve a single {{.Name}} resource by its unique identifier",
            Tags:        []string{"{{.Name}}"},
            Responses: openapi3.Responses{
                "200": &openapi3.ResponseRef{
                    Value: &openapi3.Response{
                        Description: strPtr("{{.Name}} resource"),
                        Content: openapi3.Content{
                            "application/json": &openapi3.MediaType{
                                Schema: &openapi3.SchemaRef{
                                    Ref: "#/components/schemas/{{.Name}}",
                                },
                            },
                        },
                    },
                },
                "404": errorResponse("{{.Name}} not found"),
                "500": errorResponse("Internal server error"),
            },
        },
        Put: &openapi3.Operation{
            OperationID: "update{{.Name}}",
            Summary:     "Update a {{.Name}}",
            Description: "Update an existing {{.Name}} resource",
            Tags:        []string{"{{.Name}}"},
            RequestBody: &openapi3.RequestBodyRef{
                Value: &openapi3.RequestBody{
                    Description: "{{.Name}} update request",
                    Required:    true,
                    Content: openapi3.Content{
                        "application/json": &openapi3.MediaType{
                            Schema: &openapi3.SchemaRef{
                                Ref: "#/components/schemas/Update{{.Name}}Request",
                            },
                        },
                    },
                },
            },
            Responses: openapi3.Responses{
                "200": &openapi3.ResponseRef{
                    Value: &openapi3.Response{
                        Description: strPtr("{{.Name}} updated successfully"),
                        Content: openapi3.Content{
                            "application/json": &openapi3.MediaType{
                                Schema: &openapi3.SchemaRef{
                                    Ref: "#/components/schemas/{{.Name}}",
                                },
                            },
                        },
                    },
                },
                "400": errorResponse("Invalid request body"),
                "404": errorResponse("{{.Name}} not found"),
                "500": errorResponse("Internal server error"),
            },
        },
        Delete: &openapi3.Operation{
            OperationID: "delete{{.Name}}",
            Summary:     "Delete a {{.Name}}",
            Description: "Delete a {{.Name}} resource from the inventory",
            Tags:        []string{"{{.Name}}"},
            Responses: openapi3.Responses{
                "204": &openapi3.ResponseRef{
                    Value: &openapi3.Response{
                        Description: strPtr("{{.Name}} deleted successfully"),
                    },
                },
                "404": errorResponse("{{.Name}} not found"),
                "500": errorResponse("Internal server error"),
            },
        },
    }
}
{{end}}

// Helper functions
func strPtr(s string) *string {
    return &s
}

func errorResponse(description string) *openapi3.ResponseRef {
    return &openapi3.ResponseRef{
        Value: &openapi3.Response{
            Description: strPtr(description),
            Content: openapi3.Content{
                "application/json": &openapi3.MediaType{
                    Schema: &openapi3.SchemaRef{
                        Value: &openapi3.Schema{
                            Type: "object",
                            Properties: openapi3.Schemas{
                                "error": &openapi3.SchemaRef{
                                    Value: &openapi3.Schema{
                                        Type: "string",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

// ServeOpenAPISpec serves the generated OpenAPI specification
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
    spec := GenerateOpenAPISpec()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(spec)
}

// ServeSwaggerUI serves the Swagger UI HTML page
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
    html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenCHAMI Inventory API - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
            });
        };
    </script>
</body>
</html>`
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

---

### Phase 3: Update Server Setup (1-2 hours)

**File:** `cmd/server/main.go`

**Current (go-fuego):**
```go
func runServer(cmd *cobra.Command, args []string) {
    s := fuego.NewServer(
        fuego.WithAddr(fmt.Sprintf("%s:%d", host, port)),
    )

    fuego.Use(s, versioning.VersionNegotiationMiddleware(versionRegistry))
    RegisterGeneratedRoutes(s)

    s.Run()
}
```

**New (Chi):**
```go
func runServer(cmd *cobra.Command, args []string) {
    r := chi.NewRouter()

    // Register all generated routes (includes middleware)
    RegisterGeneratedRoutes(r)

    // Start server
    addr := fmt.Sprintf("%s:%d", host, port)
    log.Printf("Starting OpenCHAMI Inventory Server on %s", addr)
    log.Printf("OpenAPI spec available at: http://%s/openapi.json", addr)
    log.Printf("Swagger UI available at: http://%s/docs", addr)

    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
```

---

### Phase 4: Update Code Generator (1 hour)

**File:** `cmd/codegen/main.go`

Add OpenAPI spec generation to the generator workflow:

```go
func main() {
    generator := codegen.NewGenerator()

    // Register resources
    if err := generator.RegisterResource(&bmc.BMC{}); err != nil {
        log.Fatalf("Failed to register BMC: %v", err)
    }
    // ... other resources

    // Generate handlers
    if err := generator.GenerateHandlers("cmd/server"); err != nil {
        log.Fatalf("Failed to generate handlers: %v", err)
    }

    // Generate routes
    if err := generator.GenerateRoutes("cmd/server"); err != nil {
        log.Fatalf("Failed to generate routes: %v", err)
    }

    // Generate OpenAPI spec generator
    if err := generator.GenerateOpenAPISpec("cmd/server"); err != nil {
        log.Fatalf("Failed to generate OpenAPI spec: %v", err)
    }

    // ... other generators
}
```

---

### Phase 5: Update Middleware (1 hour)

**Middleware needs to be updated for Chi patterns:**

**Before (go-fuego):**
```go
func VersionNegotiationMiddleware(registry *VersionRegistry) func(fuego.ContextNoBody) error {
    return func(c fuego.ContextNoBody) error {
        // fuego-specific middleware
    }
}
```

**After (Chi):**
```go
func VersionNegotiationMiddleware(registry *VersionRegistry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract version from Accept header or query param
            version := extractVersion(r)

            // Store in context
            ctx := context.WithValue(r.Context(), versionKey, version)

            // Call next handler
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## Code Examples

### Complete Handler Example (Chi)

```go
// Generated: cmd/server/bmc_handlers_generated.go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/openchami/inventory/pkg/resources/bmc"
    "github.com/openchami/inventory/internal/storage"
)

// ListBMCs returns all BMC resources
func ListBMCs(w http.ResponseWriter, r *http.Request) {
    bmcs, err := storage.LoadAllBMCs(r.Context())
    if err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, bmcs)
}

// GetBMC returns a single BMC by UID
func GetBMC(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")
    bmc, err := storage.LoadBMC(r.Context(), uid)
    if err != nil {
        respondError(w, http.StatusNotFound, err)
        return
    }
    respondJSON(w, http.StatusOK, bmc)
}

// CreateBMC creates a new BMC resource
func CreateBMC(w http.ResponseWriter, r *http.Request) {
    var req bmc.CreateBMCRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    newBMC := &bmc.BMC{
        Spec: req.ToBMCSpec(),
    }
    newBMC.SetName(req.Name)
    newBMC.GenerateUID("bmc")

    for k, v := range req.Labels {
        newBMC.SetLabel(k, v)
    }
    for k, v := range req.Annotations {
        newBMC.SetAnnotation(k, v)
    }

    if err := storage.SaveBMC(r.Context(), newBMC); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusCreated, newBMC)
}

// UpdateBMC updates an existing BMC resource
func UpdateBMC(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")

    var req bmc.UpdateBMCRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    existingBMC, err := storage.LoadBMC(r.Context(), uid)
    if err != nil {
        respondError(w, http.StatusNotFound, err)
        return
    }

    req.ApplyToBMC(existingBMC)

    if err := storage.SaveBMC(r.Context(), existingBMC); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusOK, existingBMC)
}

// DeleteBMC deletes a BMC resource
func DeleteBMC(w http.ResponseWriter, r *http.Request) {
    uid := chi.URLParam(r, "uid")

    if err := storage.DeleteBMC(r.Context(), uid); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

### Complete Routes Example (Chi)

```go
// Generated: cmd/server/routes_generated.go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func RegisterGeneratedRoutes(r chi.Router) {
    // Standard Chi middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // Custom middleware
    r.Use(versioning.VersionNegotiationMiddleware(versionRegistry))

    // BMC routes
    r.Route("/bmcs", func(r chi.Router) {
        r.Get("/", ListBMCs)
        r.Post("/", CreateBMC)

        r.Route("/{uid}", func(r chi.Router) {
            r.Get("/", GetBMC)
            r.Put("/", UpdateBMC)
            r.Delete("/", DeleteBMC)
        })
    })

    // Node routes
    r.Route("/nodes", func(r chi.Router) {
        r.Get("/", ListNodes)
        r.Post("/", CreateNode)

        r.Route("/{uid}", func(r chi.Router) {
            r.Get("/", GetNode)
            r.Put("/", UpdateNode)
            r.Delete("/", DeleteNode)
        })
    })

    // ... other resources

    // OpenAPI endpoints
    r.Get("/openapi.json", ServeOpenAPISpec)
    r.Get("/docs", ServeSwaggerUI)

    // Health check
    r.Get("/health", HealthCheck)
}
```

---

## Migration Strategy

### Step-by-Step Migration

**Day 1: Preparation**

1. ✅ Create feature branch: `git checkout -b chi-migration`
2. ✅ Add Chi dependencies: `go get github.com/go-chi/chi/v5`
3. ✅ Create new template files (don't delete old ones yet)
4. ✅ Add OpenAPI generator template

**Day 2: Implementation**

5. ✅ Update handlers template for Chi
6. ✅ Update routes template for Chi
7. ✅ Create OpenAPI generator template
8. ✅ Update code generator to use new templates
9. ✅ Regenerate code: `make dev`
10. ✅ Update `cmd/server/main.go` to use Chi
11. ✅ Fix compilation errors

**Day 3: Testing & Validation**

12. ✅ Run existing tests: `make test`
13. ✅ Manual API testing
14. ✅ Verify OpenAPI spec: `curl http://localhost:8080/openapi.json`
15. ✅ Test Swagger UI: `open http://localhost:8080/docs`
16. ✅ Test version negotiation
17. ✅ Load testing / performance validation
18. ✅ Remove go-fuego dependency
19. ✅ Clean up old templates
20. ✅ Update documentation

### Parallel Implementation Strategy

**Option:** Run Chi and go-fuego in parallel temporarily

```go
// cmd/server/main.go
func runServer(cmd *cobra.Command, args []string) {
    if useChiRouter {
        runChiServer()
    } else {
        runFuegoServer() // legacy
    }
}
```

**Benefits:**
- Easy rollback if issues found
- A/B testing between implementations
- Gradual migration

**Recommendation:** Direct migration is simple enough, parallel not needed

---

## Testing Strategy

### Unit Tests

**Update handler tests:**

**Before (go-fuego):**
```go
func TestListBMCs(t *testing.T) {
    // Test with fuego context
}
```

**After (Chi):**
```go
func TestListBMCs(t *testing.T) {
    req := httptest.NewRequest("GET", "/bmcs", nil)
    w := httptest.NewRecorder()

    ListBMCs(w, req)

    assert.Equal(t, 200, w.Code)

    var bmcs []bmc.BMC
    json.Unmarshal(w.Body.Bytes(), &bmcs)
    assert.Greater(t, len(bmcs), 0)
}
```

### Integration Tests

**Test complete request flow:**

```go
func TestBMCCRUD(t *testing.T) {
    // Setup
    r := chi.NewRouter()
    RegisterGeneratedRoutes(r)
    ts := httptest.NewServer(r)
    defer ts.Close()

    // Create BMC
    createReq := bmc.CreateBMCRequest{
        Name:     "test-bmc",
        Address:  "10.0.0.1",
        Username: "admin",
        Type:     "Redfish",
    }
    body, _ := json.Marshal(createReq)
    resp, _ := http.Post(ts.URL+"/bmcs", "application/json", bytes.NewBuffer(body))
    assert.Equal(t, 201, resp.StatusCode)

    var created bmc.BMC
    json.NewDecoder(resp.Body).Decode(&created)

    // Get BMC
    resp, _ = http.Get(ts.URL + "/bmcs/" + created.GetUID())
    assert.Equal(t, 200, resp.StatusCode)

    // Update BMC
    updateReq := bmc.UpdateBMCRequest{Address: "10.0.0.2"}
    body, _ = json.Marshal(updateReq)
    req, _ := http.NewRequest("PUT", ts.URL+"/bmcs/"+created.GetUID(), bytes.NewBuffer(body))
    resp, _ = http.DefaultClient.Do(req)
    assert.Equal(t, 200, resp.StatusCode)

    // Delete BMC
    req, _ = http.NewRequest("DELETE", ts.URL+"/bmcs/"+created.GetUID(), nil)
    resp, _ = http.DefaultClient.Do(req)
    assert.Equal(t, 204, resp.StatusCode)
}
```

### OpenAPI Validation

**Validate generated spec:**

```go
func TestOpenAPISpec(t *testing.T) {
    spec := GenerateOpenAPISpec()

    // Validate spec is valid OpenAPI 3.0
    loader := openapi3.NewLoader()
    err := loader.ResolveRefsIn(spec, nil)
    assert.NoError(t, err)

    // Validate all paths exist
    assert.NotNil(t, spec.Paths["/bmcs"])
    assert.NotNil(t, spec.Paths["/bmcs/{uid}"])

    // Validate schemas
    assert.NotNil(t, spec.Components.Schemas["BMC"])
    assert.NotNil(t, spec.Components.Schemas["CreateBMCRequest"])
}
```

### Performance Testing

**Compare Chi vs go-fuego performance:**

```bash
# Benchmark before (go-fuego)
hey -n 10000 -c 100 http://localhost:8080/bmcs

# Benchmark after (Chi)
hey -n 10000 -c 100 http://localhost:8080/bmcs

# Expected: Similar or better performance with Chi
```

---

## Risk Assessment

### Risk Matrix

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| **API incompatibility** | Low | Medium | Comprehensive testing, feature parity validation |
| **Performance regression** | Very Low | Medium | Benchmark testing, Chi is proven fast |
| **OpenAPI gaps** | Low | Medium | Validate generated specs, manual review |
| **Middleware issues** | Low | Low | Chi has robust middleware ecosystem |
| **Breaking client apps** | Very Low | High | REST API is unchanged, only internal implementation |
| **Template bugs** | Medium | Medium | Careful template review, thorough testing |

### Mitigation Strategies

**1. API Compatibility**
- ✅ REST endpoints remain identical
- ✅ Request/response formats unchanged
- ✅ HTTP status codes consistent
- ✅ Existing clients unaffected

**2. Rollback Plan**
- ✅ Keep go-fuego templates until migration validated
- ✅ Feature branch allows easy revert
- ✅ Tag current version before migration

**3. Testing Coverage**
- ✅ Run full test suite before and after
- ✅ Add Chi-specific integration tests
- ✅ Manual API testing checklist
- ✅ Performance benchmarks

---

## Timeline

### Detailed Schedule

**Day 1: Preparation (4 hours)**
- ☐ Create feature branch
- ☐ Add Chi dependencies
- ☐ Review current fuego usage patterns
- ☐ Create new template files

**Day 2: Implementation (6-8 hours)**
- ☐ Update handlers template
- ☐ Update routes template
- ☐ Create OpenAPI generator template
- ☐ Update code generator
- ☐ Regenerate all code
- ☐ Update main.go
- ☐ Fix compilation errors

**Day 3: Testing & Validation (4-6 hours)**
- ☐ Run unit tests
- ☐ Run integration tests
- ☐ Manual API testing
- ☐ OpenAPI validation
- ☐ Performance benchmarks
- ☐ Documentation updates
- ☐ Remove go-fuego dependency
- ☐ Clean up old templates

**Total Effort:** 2-3 days (14-18 hours)

---

## Benefits Summary

### Immediate Benefits

1. **Stability** - Chi is battle-tested (8+ years)
2. **Flexibility** - Rich routing and middleware options
3. **Performance** - Excellent performance characteristics
4. **Community** - Large ecosystem (18k+ stars)
5. **Idiomatic** - Standard Go patterns

### Long-Term Benefits

1. **Production Readiness** - Proven in large-scale deployments
2. **Maintainability** - Stable API, predictable behavior
3. **Extensibility** - Easy to add custom middleware
4. **Documentation** - Better OpenAPI through kin-openapi
5. **Risk Reduction** - Lower dependency risk

### OpenAPI Benefits

1. **Type-Safe** - Specs generated from actual Go types
2. **No Annotations** - No docstring burden
3. **Always Accurate** - Generated code = generated docs
4. **Extensible** - Easy to add custom spec details
5. **Standard** - OpenAPI 3.0 compliance

---

## Success Criteria

### Must-Have

- ✅ All existing tests pass
- ✅ API endpoints function identically
- ✅ OpenAPI spec validates
- ✅ Swagger UI works correctly
- ✅ Performance maintains or improves
- ✅ Zero breaking changes for clients

### Nice-to-Have

- ✅ Improved middleware composition
- ✅ Better error messages
- ✅ Enhanced OpenAPI documentation
- ✅ Additional Chi middleware integration

---

## Post-Migration Tasks

### Documentation Updates

1. Update developer guide
2. Update code generation guide
3. Add Chi best practices
4. Document OpenAPI generation
5. Update examples

### Code Cleanup

1. Remove go-fuego dependency
2. Delete old fuego templates
3. Remove unused imports
4. Clean up intermediate files

### Communication

1. Announce migration to team
2. Update README with new endpoints
3. Document any breaking changes (if any)
4. Update deployment guides

---

## Conclusion

Migrating from **go-fuego to Chi** with **kin-openapi** for automatic documentation generation is a low-risk, high-value improvement that:

1. ✅ Increases production stability
2. ✅ Maintains automatic OpenAPI generation
3. ✅ Provides better routing flexibility
4. ✅ Reduces long-term dependency risk
5. ✅ Fits perfectly with existing code generation

**Recommendation:** **Proceed with migration**

The migration is straightforward, well-defined, and provides significant benefits with minimal risk. The code generation approach makes this migration cleaner than typical framework switches, as most of the work is updating templates rather than hand-editing code.

---

## Appendix A: Template Mapping

| Component | go-fuego Template | Chi Template |
|-----------|-------------------|--------------|
| Handlers | Uses `fuego.Context*` | Uses `http.ResponseWriter`, `*http.Request` |
| Routes | `fuego.Get/Post/...` | `r.Get/Post/...` |
| Middleware | `fuego.Use` | `r.Use` (standard pattern) |
| Params | `c.PathParam()` | `chi.URLParam()` |
| Body | `c.Body()` | `json.NewDecoder(r.Body)` |
| Response | `return data, err` | `respondJSON(w, status, data)` |

## Appendix B: Dependency Changes

**Remove:**
```go
github.com/go-fuego/fuego v0.18.8
```

**Add:**
```go
github.com/go-chi/chi/v5 v5.0.12
github.com/go-chi/cors v1.2.1
```

**Already Present (leverage):**
```go
github.com/getkin/kin-openapi v0.133.0
```

## Appendix C: Example curl Commands

**Test OpenAPI spec:**
```bash
curl http://localhost:8080/openapi.json | jq .
```

**View Swagger UI:**
```bash
open http://localhost:8080/docs
```

**Test endpoints (unchanged):**
```bash
# List BMCs
curl http://localhost:8080/bmcs

# Get BMC
curl http://localhost:8080/bmcs/bmc-abc123

# Create BMC
curl -X POST http://localhost:8080/bmcs \
  -H "Content-Type: application/json" \
  -d '{"name":"test","address":"10.0.0.1","username":"admin","type":"Redfish"}'
```

---

**Document Version:** 1.0
**Last Updated:** 2025-10-02
**Status:** Proposal - Ready for Review
**Estimated Effort:** 2-3 days
**Risk Level:** Low
**Recommendation:** Approve and Implement
