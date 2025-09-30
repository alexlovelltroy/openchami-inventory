package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
)

// ResourceMetadata holds metadata about a resource type for code generation
type ResourceMetadata struct {
	Name         string            // e.g., "BMC"
	PluralName   string            // e.g., "bmcs"
	Package      string            // e.g., "github.com/openchami/inventory/pkg/resources/bmc"
	PackageAlias string            // e.g., "bmc"
	TypeName     string            // e.g., "*bmc.BMC"
	SpecType     string            // e.g., "bmc.BMCSpec"
	StatusType   string            // e.g., "bmc.BMCStatus"
	URLPath      string            // e.g., "/bmcs"
	StorageName  string            // e.g., "BMC" or "BootConfig" for storage function names
	Tags         map[string]string // Additional metadata
	RequiresAuth bool              // Whether this resource requires authentication
}

// Generator handles code generation for resources
type Generator struct {
	OutputDir   string
	PackageName string
	ModulePath  string
	Resources   []ResourceMetadata
	Templates   map[string]*template.Template
}

// NewGenerator creates a new code generator
func NewGenerator(outputDir, packageName, modulePath string) *Generator {
	return &Generator{
		OutputDir:   outputDir,
		PackageName: packageName,
		ModulePath:  modulePath,
		Resources:   make([]ResourceMetadata, 0),
		Templates:   make(map[string]*template.Template),
	}
}

// RegisterResource adds a resource type for code generation
func (g *Generator) RegisterResource(resourceType interface{}) error {
	t := reflect.TypeOf(resourceType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Extract resource metadata
	name := t.Name()
	pluralName := strings.ToLower(name) + "s"
	if name == "BMC" {
		pluralName = "bmcs"
	}

	// Determine spec type name (handle special cases)
	specTypeName := name + "Spec"
	if name == "BootConfiguration" {
		specTypeName = "BootConfigSpec"
	}

	// Determine storage function name (handle special cases)
	storageName := name
	if name == "BootConfiguration" {
		storageName = "BootConfig"
	}

	// Extract package path and create correct import paths
	pkgPath := t.PkgPath()
	var packageImport, typePrefix string

	// Map the new package structure
	switch {
	case strings.Contains(pkgPath, "/bmc"):
		packageImport = "github.com/openchami/inventory/pkg/resources/bmc"
		typePrefix = "bmc"
	case strings.Contains(pkgPath, "/node"):
		packageImport = "github.com/openchami/inventory/pkg/resources/node"
		typePrefix = "node"
	case strings.Contains(pkgPath, "/fru"):
		packageImport = "github.com/openchami/inventory/pkg/resources/fru"
		typePrefix = "fru"
	case strings.Contains(pkgPath, "/boot"):
		packageImport = "github.com/openchami/inventory/pkg/resources/boot"
		typePrefix = "boot"
	default:
		// Fallback for resources package
		packageImport = "github.com/openchami/inventory/pkg/resources"
		typePrefix = "resources"
	}

	metadata := ResourceMetadata{
		Name:         name,
		PluralName:   pluralName,
		Package:      packageImport,
		PackageAlias: typePrefix,
		TypeName:     fmt.Sprintf("*%s.%s", typePrefix, name),
		SpecType:     fmt.Sprintf("%s.%s", typePrefix, specTypeName),
		StatusType:   fmt.Sprintf("%s.%sStatus", typePrefix, name),
		URLPath:      fmt.Sprintf("/%s", pluralName),
		StorageName:  storageName,
		Tags:         make(map[string]string),
	}

	g.Resources = append(g.Resources, metadata)
	return nil
}

// EnableAuthForResource enables authentication for a specific resource type
func (g *Generator) EnableAuthForResource(resourceName string) error {
	for i, resource := range g.Resources {
		if resource.Name == resourceName {
			g.Resources[i].RequiresAuth = true
			return nil
		}
	}
	return fmt.Errorf("resource %s not found", resourceName)
}

// GenerateAll generates all code artifacts
func (g *Generator) GenerateAll() error {
	if err := g.LoadTemplates(); err != nil {
		return err
	}

	// Generate based on package type
	switch g.PackageName {
	case "main":
		// Server code - handlers, routes, models, and storage
		if err := g.GenerateModels(); err != nil {
			return err
		}
		if err := g.GenerateHandlers(); err != nil {
			return err
		}
		if err := g.GenerateRoutes(); err != nil {
			return err
		}
		if err := g.GenerateStorage(); err != nil {
			return err
		}
	case "client":
		// Client code - client and models only
		if err := g.GenerateClient(); err != nil {
			return err
		}
		if err := g.GenerateClientModels(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported package type: %s", g.PackageName)
	}

	return nil
}

// GenerateStorage generates storage operations for server
func (g *Generator) GenerateStorage() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["storage"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute storage template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated storage code: %w", err)
	}

	// Write storage to internal/storage directory instead of output directory
	storageDir := filepath.Join("internal", "storage")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	filename := filepath.Join(storageDir, "storage_generated.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// GenerateClientModels generates models specifically for client package
func (g *Generator) GenerateClientModels() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["clientModels"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute client models template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated client models code: %w", err)
	}

	filename := filepath.Join(g.OutputDir, "models.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write client models file: %w", err)
	}

	return nil
}

// LoadTemplates loads code generation templates
func (g *Generator) LoadTemplates() error {
	templates := map[string]string{
		"handlers":     handlersTemplate,
		"clientModels": clientModelsTemplate,
		"routes":       routesTemplate,
		"storage":      storageTemplate,
		"models":       modelsTemplate,
		"client":       clientTemplate,
		"policies":     policiesTemplate,
	}

	g.Templates = make(map[string]*template.Template)
	for name, tmplStr := range templates {
		tmpl, err := template.New(name).Funcs(templateFuncs).Parse(tmplStr)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		g.Templates[name] = tmpl
	}

	return nil
}

// GenerateHandlers generates REST API handlers for all resources
func (g *Generator) GenerateHandlers() error {
	for _, resource := range g.Resources {
		var buf bytes.Buffer
		data := struct {
			ResourceMetadata
			ModulePath string
		}{
			ResourceMetadata: resource,
			ModulePath:       g.ModulePath,
		}

		if err := g.Templates["handlers"].Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute handlers template for %s: %w", resource.Name, err)
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("failed to format generated code for %s: %w", resource.Name, err)
		}

		filename := filepath.Join(g.OutputDir, fmt.Sprintf("%s_handlers_generated.go", strings.ToLower(resource.Name)))
		if err := os.WriteFile(filename, formatted, 0644); err != nil {
			return fmt.Errorf("failed to write handlers file for %s: %w", resource.Name, err)
		}
	}

	return nil
}

// GenerateClient generates API client library
func (g *Generator) GenerateClient() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["client"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute client template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated client code: %w", err)
	}

	filename := filepath.Join(g.OutputDir, "client.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write client file: %w", err)
	}

	return nil
}

// GenerateModels generates request/response models
func (g *Generator) GenerateModels() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["models"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute models template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated models code: %w", err)
	}

	filename := filepath.Join(g.OutputDir, "models_generated.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write models file: %w", err)
	}

	return nil
}

// GenerateRoutes generates route registration code
func (g *Generator) GenerateRoutes() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["routes"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute routes template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated routes code: %w", err)
	}

	filename := filepath.Join(g.OutputDir, "routes_generated.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write routes file: %w", err)
	}

	return nil
}

// GeneratePolicies generates authorization policy interfaces and scaffolding
func (g *Generator) GeneratePolicies() error {
	var buf bytes.Buffer
	data := struct {
		PackageName string
		ModulePath  string
		Resources   []ResourceMetadata
	}{
		PackageName: g.PackageName,
		ModulePath:  g.ModulePath,
		Resources:   g.Resources,
	}

	if err := g.Templates["policies"].Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute policies template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated policies code: %w", err)
	}

	filename := filepath.Join(g.OutputDir, "policies_generated.go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write policies file: %w", err)
	}

	return nil
}

// Template functions
var templateFuncs = template.FuncMap{
	"toLower":    strings.ToLower,
	"toUpper":    strings.ToUpper,
	"title":      cases.Title,
	"trimPrefix": strings.TrimPrefix,
	"camelCase": func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToLower(s[:1]) + s[1:]
	},
}

// Templates
const handlersTemplate = `// Code generated by codegen. DO NOT EDIT.
package main

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"{{.Package}}"
	"{{.ModulePath}}/internal/storage"
	"{{.ModulePath}}/pkg/resources"
)

// Get{{.Name}}s returns all {{.Name}} resources
func Get{{.Name}}s(c fuego.ContextNoBody) ([]{{.TypeName}}, error) {
{{if .RequiresAuth}}
	// Check authorization
	if auth, ok := GetAuthFromContext(c.Context()); ok {
		if policy, exists := policyRegistry.GetPolicy("{{.Name}}"); exists {
			if decision := policy.CanList(c.Context(), auth, c.Request()); !decision.Allowed {
				return nil, fuego.HTTPError{
					Status: http.StatusForbidden,
					Err:    fmt.Errorf("access denied: %s", decision.Reason),
				}
			}
		} else {
			return nil, fuego.HTTPError{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("no policy configured for {{.Name}}"),
			}
		}
	} else {
		return nil, fuego.HTTPError{
			Status: http.StatusUnauthorized,
			Err:    fmt.Errorf("authentication required"),
		}
	}
{{end}}
	{{camelCase .PluralName}}, err := storage.LoadAll{{.StorageName}}s()
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to load {{.PluralName}}: %w", err),
		}
	}
	return {{camelCase .PluralName}}, nil
}

// Get{{.Name}} returns a specific {{.Name}} resource by UID
func Get{{.Name}}(c fuego.ContextNoBody) ({{.TypeName}}, error) {
	uid := c.PathParam("uid")
	if uid == "" {
		return nil, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("{{.Name}} UID is required"),
		}
	}

{{if .RequiresAuth}}
	// Check authorization
	if auth, ok := GetAuthFromContext(c.Context()); ok {
		if policy, exists := policyRegistry.GetPolicy("{{.Name}}"); exists {
			if decision := policy.CanGet(c.Context(), auth, c.Request(), uid); !decision.Allowed {
				return nil, fuego.HTTPError{
					Status: http.StatusForbidden,
					Err:    fmt.Errorf("access denied: %s", decision.Reason),
				}
			}
		} else {
			return nil, fuego.HTTPError{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("no policy configured for {{.Name}}"),
			}
		}
	} else {
		return nil, fuego.HTTPError{
			Status: http.StatusUnauthorized,
			Err:    fmt.Errorf("authentication required"),
		}
	}
{{end}}

	{{camelCase .Name}}, err := storage.Load{{.StorageName}}(uid)
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("{{.Name}} not found: %w", err),
		}
	}
	return {{camelCase .Name}}, nil
}

// Create{{.Name}} creates a new {{.Name}} resource
func Create{{.Name}}(c fuego.ContextWithBody[Create{{.Name}}Request]) ({{.TypeName}}, error) {
	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("invalid request body: %w", err),
		}
	}

	uid, err := resources.GenerateUIDForResource("{{.Name}}")
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to generate UID: %w", err),
		}
	}

	{{camelCase .Name}} := &{{.PackageAlias}}.{{.Name}}{
		Resource: resources.Resource{
			APIVersion:    "v1",
			Kind:          "{{.Name}}",
			SchemaVersion: "1.0",
		},
		// Spec: TODO: Convert from req
	}

	{{camelCase .Name}}.Metadata.Initialize(req.Name, uid)

	// Set labels and annotations
	for k, v := range req.Labels {
		{{camelCase .Name}}.SetLabel(k, v)
	}
	for k, v := range req.Annotations {
		{{camelCase .Name}}.SetAnnotation(k, v)
	}

	// Set initial status
	resources.SetCondition(&{{camelCase .Name}}.Status.Conditions, "Created", "True", "{{.Name}}Created", "{{.Name}} has been created")

	if err := storage.Save{{.StorageName}}({{camelCase .Name}}); err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to save {{.Name}}: %w", err),
		}
	}

	return {{camelCase .Name}}, nil
}

// Update{{.Name}} updates an existing {{.Name}} resource
func Update{{.Name}}(c fuego.ContextWithBody[Update{{.Name}}Request]) ({{.TypeName}}, error) {
	uid := c.PathParam("uid")
	if uid == "" {
		return nil, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("{{.Name}} UID is required"),
		}
	}

	{{camelCase .Name}}, err := storage.Load{{.StorageName}}(uid)
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("{{.Name}} not found: %w", err),
		}
	}

	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("invalid request body: %w", err),
		}
	}

	// Apply updates
	if req.Name != "" {
		{{camelCase .Name}}.SetName(req.Name)
	}
	
	// Update labels and annotations
	for k, v := range req.Labels {
		{{camelCase .Name}}.SetLabel(k, v)
	}
	for k, v := range req.Annotations {
		{{camelCase .Name}}.SetAnnotation(k, v)
	}

	{{camelCase .Name}}.Touch()
	resources.SetCondition(&{{camelCase .Name}}.Status.Conditions, "Updated", "True", "{{.Name}}Updated", "{{.Name}} has been updated")

	if err := storage.Save{{.StorageName}}({{camelCase .Name}}); err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to save {{.Name}}: %w", err),
		}
	}

	return {{camelCase .Name}}, nil
}

// Delete{{.Name}} deletes a {{.Name}} resource
func Delete{{.Name}}(c fuego.ContextNoBody) (*DeleteResponse, error) {
	uid := c.PathParam("uid")
	if uid == "" {
		return nil, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("{{.Name}} UID is required"),
		}
	}

	if err := storage.Delete{{.StorageName}}(uid); err != nil {
		return nil, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to delete {{.Name}}: %w", err),
		}
	}

	return &DeleteResponse{
		Message: "{{.Name}} deleted successfully",
		UID:     uid,
	}, nil
}
`

// Separate client models template
const clientModelsTemplate = `// Code generated by codegen. DO NOT EDIT.
package client

import "{{.ModulePath}}/pkg/resources"

{{range .Resources}}
// Create{{.Name}}Request represents a request to create a {{.Name}}
type Create{{.Name}}Request struct {
	Name        string            ` + "`json:\"name\" validate:\"required\"`" + `
	Labels      map[string]string ` + "`json:\"labels,omitempty\"`" + `
	Annotations map[string]string ` + "`json:\"annotations,omitempty\"`" + `
	// TODO: Add resource-specific fields based on {{.Name}}Spec
}

// Update{{.Name}}Request represents a request to update a {{.Name}}
type Update{{.Name}}Request struct {
	Name        string            ` + "`json:\"name,omitempty\"`" + `
	Labels      map[string]string ` + "`json:\"labels,omitempty\"`" + `
	Annotations map[string]string ` + "`json:\"annotations,omitempty\"`" + `
	// TODO: Add resource-specific fields based on {{.Name}}Spec
}

{{end}}

// DeleteResponse represents a successful deletion response
type DeleteResponse struct {
	Message string ` + "`json:\"message\"`" + `
	UID     string ` + "`json:\"uid\"`" + `
}
`

// Add storage template
const storageTemplate = `// Code generated by codegen. DO NOT EDIT.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"{{.ModulePath}}/pkg/resources/bmc"
	"{{.ModulePath}}/pkg/resources/node"
	"{{.ModulePath}}/pkg/resources/fru"
	"{{.ModulePath}}/pkg/resources/boot"
)

const inventoryDir = "inventory"

type Storage struct {
	baseDir string
}

var storage *Storage

func init() {
	storage = &Storage{baseDir: inventoryDir}
	
	// Ensure all directories exist
	dirs := []string{"nodes", "bmcs", "frus", "boot-configs", "boot-config-aliases", "boot-bindings", "fru-associations", "fru-snapshots"}
	for _, dir := range dirs {
		fullPath := filepath.Join(inventoryDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			panic(fmt.Sprintf("failed to create directory %s: %v", fullPath, err))
		}
	}
}

{{range .Resources}}
// {{.Name}} storage operations
func (s *Storage) LoadAll{{.StorageName}}s() ([]{{.TypeName}}, error) {
	var {{camelCase .PluralName}} []{{.TypeName}}
	
	entries, err := os.ReadDir(filepath.Join(s.baseDir, "{{.PluralName}}"))
	if err != nil {
		return nil, fmt.Errorf("failed to read {{.PluralName}} directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		
		{{camelCase .Name}}, err := s.load{{.Name}}FromFile(filepath.Join(s.baseDir, "{{.PluralName}}", entry.Name()))
		if err != nil {
			continue // Skip invalid files
		}
		{{camelCase .PluralName}} = append({{camelCase .PluralName}}, {{camelCase .Name}})
	}
	
	return {{camelCase .PluralName}}, nil
}

func (s *Storage) Load{{.StorageName}}(uid string) ({{.TypeName}}, error) {
	filename := filepath.Join(s.baseDir, "{{.PluralName}}", uid+".json")
	return s.load{{.Name}}FromFile(filename)
}

func (s *Storage) load{{.Name}}FromFile(filename string) ({{.TypeName}}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read {{.Name}} file: %w", err)
	}
	
	var {{camelCase .Name}} {{.PackageAlias}}.{{.Name}}
	if err := json.Unmarshal(data, &{{camelCase .Name}}); err != nil {
		return nil, fmt.Errorf("failed to unmarshal {{.Name}}: %w", err)
	}
	
	return &{{camelCase .Name}}, nil
}

func (s *Storage) Save{{.StorageName}}({{camelCase .Name}} {{.TypeName}}) error {
	filename := filepath.Join(s.baseDir, "{{.PluralName}}", {{camelCase .Name}}.GetUID()+".json")
	
	data, err := json.MarshalIndent({{camelCase .Name}}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal {{.Name}}: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write {{.Name}} file: %w", err)
	}
	
	return nil
}

func (s *Storage) Delete{{.StorageName}}(uid string) error {
	filename := filepath.Join(s.baseDir, "{{.PluralName}}", uid+".json")
	
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("{{.Name}} not found")
	}
	
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete {{.Name}} file: %w", err)
	}
	
	return nil
}

{{end}}

// Package-level wrapper functions for compatibility
{{range .Resources}}
func LoadAll{{.StorageName}}s() ([]{{.TypeName}}, error) {
	return storage.LoadAll{{.StorageName}}s()
}

func Load{{.StorageName}}(uid string) ({{.TypeName}}, error) {
	return storage.Load{{.StorageName}}(uid)
}

func Save{{.StorageName}}({{camelCase .Name}} {{.TypeName}}) error {
	return storage.Save{{.StorageName}}({{camelCase .Name}})
}

func Delete{{.StorageName}}(uid string) error {
	return storage.Delete{{.StorageName}}(uid)
}
{{end}}
`

const routesTemplate = `// Code generated by codegen. DO NOT EDIT.
package main

import "github.com/go-fuego/fuego"

// RegisterGeneratedRoutes registers all generated routes
func RegisterGeneratedRoutes(s *fuego.Server) {
{{range .Resources}}
	// {{.Name}} routes
	fuego.Get(s, "{{.URLPath}}", Get{{.Name}}s)
	fuego.Get(s, "{{.URLPath}}/{uid}", Get{{.Name}})
	fuego.Post(s, "{{.URLPath}}", Create{{.Name}})
	fuego.Put(s, "{{.URLPath}}/{uid}", Update{{.Name}})
	fuego.Delete(s, "{{.URLPath}}/{uid}", Delete{{.Name}})
{{end}}
}
`

// Add models template for server-side models
const modelsTemplate = `// Code generated by codegen. DO NOT EDIT.
package {{.PackageName}}

import (
	{{range .Resources}}"{{.Package}}"
	{{end}}
)

{{range .Resources}}
// {{.Name}}Response represents the response for {{.Name}} operations
type {{.Name}}Response = {{.PackageAlias}}.{{.Name}}

// {{.Name}}sResponse represents a list of {{.Name}}s
type {{.Name}}sResponse struct {
	Items []{{.PackageAlias}}.{{.Name}} ` + "`json:\"items\"`" + `
	Total int                   ` + "`json:\"total\"`" + `
}

// Create{{.Name}}Request represents a request to create a {{.Name}}
type Create{{.Name}}Request struct {
	{{.SpecType}} ` + "`json:\",inline\"`" + `
	Name          string            ` + "`json:\"name\" validate:\"required\"`" + `
	Labels        map[string]string ` + "`json:\"labels,omitempty\"`" + `
	Annotations   map[string]string ` + "`json:\"annotations,omitempty\"`" + `
}

// Update{{.Name}}Request represents a request to update a {{.Name}}
type Update{{.Name}}Request struct {
	{{.SpecType}} ` + "`json:\",inline,omitempty\"`" + `
	Name          string            ` + "`json:\"name,omitempty\"`" + `
	Labels        map[string]string ` + "`json:\"labels,omitempty\"`" + `
	Annotations   map[string]string ` + "`json:\"annotations,omitempty\"`" + `
}

{{end}}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string ` + "`json:\"error\"`" + `
	Message string ` + "`json:\"message,omitempty\"`" + `
	Code    int    ` + "`json:\"code,omitempty\"`" + `
}

// DeleteResponse represents a successful deletion response
type DeleteResponse struct {
	Message string ` + "`json:\"message\"`" + `
	UID     string ` + "`json:\"uid\"`" + `
}
`

// Add client template for generating HTTP client
const clientTemplate = `// Code generated by codegen. DO NOT EDIT.
package {{.PackageName}}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	{{range .Resources}}"{{.Package}}"
	{{end}}
)

// Client provides access to the inventory API
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		baseURL:    u,
		httpClient: httpClient,
	}, nil
}

// doRequest performs an HTTP request and handles the response
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	u := *c.baseURL
	u.Path = path.Join(u.Path, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Error)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

{{range .Resources}}
// Get{{.Name}}s retrieves all {{.PluralName}}
func (c *Client) Get{{.Name}}s(ctx context.Context) ([]{{.PackageAlias}}.{{.Name}}, error) {
	var response {{.Name}}sResponse
	if err := c.doRequest(ctx, "GET", "{{.URLPath}}", nil, &response); err != nil {
		return nil, err
	}
	return response.Items, nil
}

// Get{{.Name}} retrieves a specific {{.Name}} by UID
func (c *Client) Get{{.Name}}(ctx context.Context, uid string) ({{.TypeName}}, error) {
	var result {{.PackageAlias}}.{{.Name}}
	endpoint := fmt.Sprintf("{{.URLPath}}/%s", uid)
	if err := c.doRequest(ctx, "GET", endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create{{.Name}} creates a new {{.Name}}
func (c *Client) Create{{.Name}}(ctx context.Context, req Create{{.Name}}Request) ({{.TypeName}}, error) {
	var result {{.PackageAlias}}.{{.Name}}
	if err := c.doRequest(ctx, "POST", "{{.URLPath}}", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update{{.Name}} updates an existing {{.Name}}
func (c *Client) Update{{.Name}}(ctx context.Context, uid string, req Update{{.Name}}Request) ({{.TypeName}}, error) {
	var result {{.PackageAlias}}.{{.Name}}
	endpoint := fmt.Sprintf("{{.URLPath}}/%s", uid)
	if err := c.doRequest(ctx, "PUT", endpoint, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete{{.Name}} deletes a {{.Name}} by UID
func (c *Client) Delete{{.Name}}(ctx context.Context, uid string) error {
	endpoint := fmt.Sprintf("{{.URLPath}}/%s", uid)
	var response DeleteResponse
	if err := c.doRequest(ctx, "DELETE", endpoint, nil, &response); err != nil {
		return err
	}
	return nil
}

{{end}}
`

// Policies template for authorization interfaces and scaffolding
const policiesTemplate = `// Code generated by codegen. DO NOT EDIT.
// This file provides auth integration for generated handlers.

package main

import (
	"context"
	
	"{{.ModulePath}}/pkg/policies"
)

// Global policy registry - should be initialized in main.go
var policyRegistry *policies.PolicyRegistry

// GetAuthFromContext extracts auth context from the request context
// This should match your tokensmith middleware's context key
func GetAuthFromContext(ctx context.Context) (*policies.AuthContext, bool) {
	// Adjust this to match your tokensmith middleware's context key
	auth, ok := ctx.Value("auth").(*policies.AuthContext)
	return auth, ok
}
`
