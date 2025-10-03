// Package versioning demonstrates multi-version support capability.
//
// This example shows how to register multiple schema versions for the same
// resource type and test version negotiation.
package versioning

import (
	"fmt"
	"reflect"

	"github.com/alexlovelltroy/fabrica/pkg/versioning"
	bmcv2beta1 "github.com/openchami/inventory/pkg/resources/bmc/v2beta1"
)

// DemoMultiVersionSupport demonstrates the multi-version capability
func DemoMultiVersionSupport() error {
	fmt.Println("=== Multi-Version Support Demo ===")

	// Create a version registry
	registry := versioning.NewVersionRegistry()

	// Register Node v1 (stable)
	nodeV1 := versioning.SchemaVersion{
		Version:    "v1",
		IsDefault:  true,
		Stability:  "stable",
		Deprecated: false,
		SpecType:   "node.NodeSpec",
		StatusType: "node.NodeStatus",
		TypeName:   "*node.Node",
		Package:    "github.com/openchami/inventory/pkg/resources/node",
		Transforms: []string{},
	}

	v1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(""),
		Constructor: func() interface{} { return &struct{}{} },
		Converter:   nil,
		Metadata:    nodeV1,
	}

	err := registry.RegisterVersion("Node", "v1", v1TypeInfo)
	if err != nil {
		return fmt.Errorf("failed to register Node v1: %w", err)
	}
	fmt.Println("✓ Registered Node v1 (stable)")

	// Register Node v2beta1 (preview of v2 features)
	nodeV2Beta1 := versioning.SchemaVersion{
		Version:    "v2beta1",
		IsDefault:  false,
		Stability:  "beta",
		Deprecated: false,
		SpecType:   "nodev2beta1.NodeSpecV2Beta1",
		StatusType: "nodev2beta1.NodeStatusV2Beta1",
		TypeName:   "*nodev2beta1.NodeV2Beta1",
		Package:    "github.com/openchami/inventory/pkg/resources/node/v2beta1",
		Transforms: []string{"ConvertV1ToV2Beta1", "ConvertV2Beta1ToV1"},
	}

	v2Beta1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(""),
		Constructor: func() interface{} { return &struct{}{} },
		Converter:   nil, // Would implement actual converter in Phase 3
		Metadata:    nodeV2Beta1,
	}

	err = registry.RegisterVersion("Node", "v2beta1", v2Beta1TypeInfo)
	if err != nil {
		return fmt.Errorf("failed to register Node v2beta1: %w", err)
	}
	fmt.Println("✓ Registered Node v2beta1 (beta)")

	// Register BMC v1
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
		Type:        reflect.TypeOf(""),
		Constructor: func() interface{} { return &struct{}{} },
		Converter:   nil,
		Metadata:    bmcV1,
	}

	err = registry.RegisterVersion("BMC", "v1", bmcV1TypeInfo)
	if err != nil {
		return fmt.Errorf("failed to register BMC v1: %w", err)
	}
	fmt.Println("✓ Registered BMC v1 (stable)")

	// Register BMC v2beta1 (with enhanced authentication)
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

	err = registry.RegisterVersion("BMC", "v2beta1", bmcV2Beta1TypeInfo)
	if err != nil {
		return fmt.Errorf("failed to register BMC v2beta1: %w", err)
	}
	fmt.Println("✓ Registered BMC v2beta1 (beta - enhanced authentication)")

	// Demonstrate version queries
	fmt.Println("\n=== Version Registry Queries ===")

	// List all resource kinds
	kinds := registry.ListKinds()
	fmt.Printf("Registered resource kinds: %v\n", kinds)

	// List versions for each kind
	for _, kind := range kinds {
		versions := registry.ListVersions(kind)
		defaultVersion := registry.GetDefaultVersion(kind)
		fmt.Printf("%s versions: %v (default: %s)\n", kind, versions, defaultVersion)

		// Show version details
		versionInfo := registry.GetVersionInfo(kind)
		for version, info := range versionInfo {
			fmt.Printf("  %s: stability=%s, deprecated=%v, package=%s\n",
				version, info.Stability, info.Deprecated, info.Package)
		}
	}

	// Test version negotiation scenarios
	fmt.Println("\n=== Version Negotiation Scenarios ===")

	scenarios := []struct {
		name             string
		resourceKind     string
		requestedVersion string
		expectedVersion  string
	}{
		{
			name:             "Client requests default Node version",
			resourceKind:     "Node",
			requestedVersion: "",
			expectedVersion:  "v1",
		},
		{
			name:             "Client requests Node v2beta1",
			resourceKind:     "Node",
			requestedVersion: "v2beta1",
			expectedVersion:  "v2beta1",
		},
		{
			name:             "Client requests non-existent Node v3",
			resourceKind:     "Node",
			requestedVersion: "v3",
			expectedVersion:  "v1", // fallback to default
		},
		{
			name:             "Client requests default BMC version",
			resourceKind:     "BMC",
			requestedVersion: "",
			expectedVersion:  "v1",
		},
		{
			name:             "Client requests BMC v2beta1 (enhanced auth)",
			resourceKind:     "BMC",
			requestedVersion: "v2beta1",
			expectedVersion:  "v2beta1",
		},
	}

	for _, scenario := range scenarios {
		ctx := &versioning.VersionContext{
			ResourceKind:     scenario.resourceKind,
			RequestedVersion: scenario.requestedVersion,
			DefaultVersion:   registry.GetDefaultVersion(scenario.resourceKind),
			GroupVersion:     "v1",
		}

		// Simulate version negotiation (would normally be done by middleware)
		availableVersions := registry.ListVersions(scenario.resourceKind)
		servedVersion := scenario.expectedVersion
		if scenario.requestedVersion != "" {
			found := false
			for _, v := range availableVersions {
				if v == scenario.requestedVersion {
					found = true
					break
				}
			}
			if !found {
				servedVersion = ctx.DefaultVersion
			}
		} else {
			servedVersion = ctx.DefaultVersion
		}
		status := "✓"
		if servedVersion != scenario.expectedVersion {
			status = "✗"
		}

		fmt.Printf("%s %s: requested=%s, served=%s\n",
			status, scenario.name,
			scenario.requestedVersion, servedVersion)
	}

	fmt.Println("\n=== Demo Complete ===")
	return nil
}
