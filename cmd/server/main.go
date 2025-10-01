package main

import (
	"log"
	"reflect"

	"github.com/go-fuego/fuego"
	"github.com/openchami/inventory/pkg/policies"
	"github.com/openchami/inventory/pkg/resources/bmc"
	bmcv2beta1 "github.com/openchami/inventory/pkg/resources/bmc/v2beta1"
	"github.com/openchami/inventory/pkg/resources/node"
	"github.com/openchami/inventory/pkg/versioning"
)

// Global version registry
var versionRegistry *versioning.VersionRegistry

// initializeVersionRegistry sets up all resource versions
func initializeVersionRegistry() {
	versionRegistry = versioning.NewVersionRegistry()

	// Register BMC v1 (stable)
	bmcV1 := versioning.SchemaVersion{
		Version:    "v1",
		IsDefault:  true,
		Stability:  "stable",
		Deprecated: false,
		SpecType:   "bmc.BMCSpec",
		StatusType: "bmc.BMCStatus",
		TypeName:   "*bmc.BMC",
		Package:    "github.com/openchami/inventory/pkg/resources/bmc",
		Transforms: []string{},
	}

	bmcV1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(&bmc.BMC{}),
		Constructor: func() interface{} { return &bmc.BMC{} },
		Converter:   nil, // v1 doesn't need converter to itself
		Metadata:    bmcV1,
	}

	if err := versionRegistry.RegisterVersion("BMC", "v1", bmcV1TypeInfo); err != nil {
		log.Fatalf("Failed to register BMC v1: %v", err)
	}

	// Register BMC v2beta1 (beta - enhanced authentication)
	bmcV2Beta1 := versioning.SchemaVersion{
		Version:    "v2beta1",
		IsDefault:  false,
		Stability:  "beta",
		Deprecated: false,
		SpecType:   "bmcv2beta1.BMCSpec",
		StatusType: "bmcv2beta1.BMCStatus",
		TypeName:   "*bmcv2beta1.BMC",
		Package:    "github.com/openchami/inventory/pkg/resources/bmc/v2beta1",
		Transforms: []string{"ConvertV1ToV2Beta1", "ConvertV2Beta1ToV1"},
	}

	bmcV2Beta1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(&bmcv2beta1.BMC{}),
		Constructor: func() interface{} { return &bmcv2beta1.BMC{} },
		Converter:   bmcv2beta1.NewBMCConverter(),
		Metadata:    bmcV2Beta1,
	}

	if err := versionRegistry.RegisterVersion("BMC", "v2beta1", bmcV2Beta1TypeInfo); err != nil {
		log.Fatalf("Failed to register BMC v2beta1: %v", err)
	}

	log.Println("Version registry initialized:")
	log.Printf("  BMC: %v (default: %s)", versionRegistry.ListVersions("BMC"), versionRegistry.GetDefaultVersion("BMC"))
}

func main() {
	// Initialize policy registry with default policies
	policyRegistry = policies.NewPolicyRegistry()
	policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
	policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())

	// Initialize version registry
	initializeVersionRegistry()

	// Create fuego server
	s := fuego.NewServer()

	// Add version negotiation middleware
	fuego.Use(s, versioning.VersionNegotiationMiddleware(versionRegistry))

	// Register generated routes
	RegisterGeneratedRoutes(s)

	// Add health check endpoint
	fuego.Get(s, "/health", func(c fuego.ContextNoBody) (map[string]string, error) {
		return map[string]string{"status": "ok"}, nil
	})

	// Add version discovery endpoint
	fuego.Get(s, "/version-info", GetVersionInfo)

	log.Println("Starting server on :8080")
	s.Run()
}
