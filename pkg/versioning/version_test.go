package versioning

import (
	"reflect"
	"testing"
)

func TestVersionRegistry(t *testing.T) {
	registry := NewVersionRegistry()

	// Test registering a version
	nodeV1 := SchemaVersion{
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

	typeInfo := ResourceTypeInfo{
		Type:        reflect.TypeOf(""),
		Constructor: func() interface{} { return &struct{}{} },
		Converter:   nil,
		Metadata:    nodeV1,
	}

	err := registry.RegisterVersion("Node", "v1", typeInfo)
	if err != nil {
		t.Fatalf("Failed to register version: %v", err)
	}

	// Test retrieving the version
	info, exists := registry.GetVersion("Node", "v1")
	if !exists {
		t.Fatal("Version v1 for Node should exist")
	}

	if info.Metadata.Version != "v1" {
		t.Errorf("Expected version v1, got %s", info.Metadata.Version)
	}

	// Test default version
	defaultVersion := registry.GetDefaultVersion("Node")
	if defaultVersion != "v1" {
		t.Errorf("Expected default version v1, got %s", defaultVersion)
	}

	// Test listing versions
	versions := registry.ListVersions("Node")
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	if versions[0] != "v1" {
		t.Errorf("Expected version v1, got %s", versions[0])
	}

	// Test listing kinds
	kinds := registry.ListKinds()
	if len(kinds) != 1 {
		t.Errorf("Expected 1 kind, got %d", len(kinds))
	}

	if kinds[0] != "Node" {
		t.Errorf("Expected kind Node, got %s", kinds[0])
	}
}

func TestVersionValidation(t *testing.T) {
	testCases := []struct {
		version string
		valid   bool
	}{
		{"v1", true},
		{"v2", true},
		{"v1beta1", true},
		{"v2beta3", true},
		{"v1alpha1", true},
		{"v3alpha5", true},
		{"1", false},    // missing 'v' prefix
		{"v", false},    // no number
		{"va1", false},  // invalid format
		{"v1.0", false}, // dots not allowed
	}

	for _, tc := range testCases {
		valid := isValidVersion(tc.version)
		if valid != tc.valid {
			t.Errorf("Version %s: expected valid=%v, got valid=%v", tc.version, tc.valid, valid)
		}
	}
}

func TestStabilityLevel(t *testing.T) {
	testCases := []struct {
		version   string
		stability string
	}{
		{"v1", "stable"},
		{"v2", "stable"},
		{"v1beta1", "beta"},
		{"v2beta3", "beta"},
		{"v1alpha1", "alpha"},
		{"v3alpha5", "alpha"},
	}

	for _, tc := range testCases {
		stability := GetStabilityLevel(tc.version)
		if stability != tc.stability {
			t.Errorf("Version %s: expected stability=%s, got stability=%s", tc.version, tc.stability, stability)
		}
	}
}
