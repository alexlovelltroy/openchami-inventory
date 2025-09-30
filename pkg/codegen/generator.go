// Package codegen provides code generation for REST API resources.
//
// This package generates consistent CRUD operations, storage, and client code
// for all resource types. The goal is to eliminate boilerplate while maintaining
// type safety and consistency across the API.
//
// Architecture:
//   - Templates define the code patterns
//   - ResourceMetadata describes each resource type
//   - Generator applies templates to metadata
//   - Output is formatted Go code
//
// Usage:
//
//	generator := NewGenerator(outputDir, packageName, modulePath)
//	generator.RegisterResource(&myresource.MyResource{})
//	generator.GenerateAll()
//
// Generated artifacts:
//   - REST API handlers (CRUD operations)
//   - Storage operations (file-based persistence)
//   - HTTP client library
//   - Request/response models
//   - Route registration
//   - Authorization integration
//
// Customization:
//   - Edit templates to change generated code patterns
//   - Implement resource-specific policies
//   - Override storage methods for custom behavior
//
// See docs/DEVELOPMENT.md for detailed developer guide.
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

// LoadTemplates loads code generation templates from files
func (g *Generator) LoadTemplates() error {
	templateDir := filepath.Join("pkg", "codegen", "templates")

	templateFiles := map[string]string{
		"handlers":     "handlers.go.tmpl",
		"clientModels": "client-models.go.tmpl",
		"routes":       "routes.go.tmpl",
		"storage":      "storage.go.tmpl",
		"models":       "models.go.tmpl",
		"client":       "client.go.tmpl",
		"policies":     "policies.go.tmpl",
	}

	g.Templates = make(map[string]*template.Template)
	for name, filename := range templateFiles {
		templatePath := filepath.Join(templateDir, filename)

		// Read template content from file
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
		}

		// Parse template with functions
		tmpl, err := template.New(name).Funcs(templateFuncs).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
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
