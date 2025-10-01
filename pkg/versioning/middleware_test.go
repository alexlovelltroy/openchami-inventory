package versioning

import (
	"testing"
)

func TestExtractGroupVersionFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/apis/inventory/v2/nodes", "v2"},
		{"/apis/inventory/v1beta1/bmcs", "v1beta1"},
		{"/apis/inventory/v3alpha1/frus", "v3alpha1"},
		{"/v2/nodes", "v2"},
		{"/nodes", "v1"},   // fallback
		{"/invalid", "v1"}, // fallback
	}

	for _, tc := range testCases {
		result := extractGroupVersionFromPath(tc.path)
		if result != tc.expected {
			t.Errorf("Path %s: expected %s, got %s", tc.path, tc.expected, result)
		}
	}
}

func TestExtractResourceKindFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/apis/inventory/v2/nodes", "Node"},
		{"/apis/inventory/v1/bmcs", "BMC"},
		{"/apis/inventory/v1/frus", "FRU"},
		{"/v2/nodes", "Node"},
		{"/nodes", "Node"},
		{"/bmcs", "BMC"},
	}

	for _, tc := range testCases {
		result := extractResourceKindFromPath(tc.path)
		if result != tc.expected {
			t.Errorf("Path %s: expected %s, got %s", tc.path, tc.expected, result)
		}
	}
}

func TestParseVersionFromAcceptHeader(t *testing.T) {
	testCases := []struct {
		header   string
		expected string
	}{
		{"application/json;version=v2beta1", "v2beta1"},
		{"application/json;v=v1alpha1", "v1alpha1"},
		{"application/json", ""},
		{"application/vnd.inventory.node+json;version=v3", "v3"},
		{"text/plain;version=v2", "v2"},
	}

	for _, tc := range testCases {
		result := parseVersionFromAcceptHeader(tc.header)
		if result != tc.expected {
			t.Errorf("Header %s: expected %s, got %s", tc.header, tc.expected, result)
		}
	}
}

func TestSingularizeResourceName(t *testing.T) {
	testCases := []struct {
		plural   string
		expected string
	}{
		{"nodes", "Node"},
		{"bmcs", "BMC"},
		{"frus", "FRU"},
		{"bootconfigurations", "BootConfiguration"},
		{"bootconfigs", "BootConfiguration"},
		{"items", "Item"},
		{"data", "Data"}, // no 's' suffix
	}

	for _, tc := range testCases {
		result := singularizeResourceName(tc.plural)
		if result != tc.expected {
			t.Errorf("Plural %s: expected %s, got %s", tc.plural, tc.expected, result)
		}
	}
}

func TestValidateVersion(t *testing.T) {
	testCases := []struct {
		version   string
		shouldErr bool
	}{
		{"v1", false},
		{"v2", false},
		{"v1beta1", false},
		{"v2beta3", false},
		{"v1alpha1", false},
		{"v3alpha5", false},
		{"1", true},       // missing 'v' prefix
		{"v", true},       // no number
		{"va1", true},     // invalid format
		{"v1.0", true},    // dots not allowed
		{"v1beta", true},  // missing number after beta
		{"v1alpha", true}, // missing number after alpha
	}

	for _, tc := range testCases {
		err := ValidateVersion(tc.version)
		hasErr := err != nil
		if hasErr != tc.shouldErr {
			t.Errorf("Version %s: expected error=%v, got error=%v (err: %v)",
				tc.version, tc.shouldErr, hasErr, err)
		}
	}
}
