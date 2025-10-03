package v2beta1

import (
	"testing"

	"github.com/alexlovelltroy/fabrica/pkg/resource"
	"github.com/openchami/inventory/pkg/resources/bmc"
)

func TestBMCConverterCanConvert(t *testing.T) {
	converter := NewBMCConverter()

	tests := []struct {
		name        string
		fromVersion string
		toVersion   string
		expected    bool
	}{
		{"v1 to v2beta1", "v1", "v2beta1", true},
		{"v2beta1 to v1", "v2beta1", "v1", true},
		{"v1 to v1", "v1", "v1", false},
		{"v2 to v1", "v2", "v1", false},
		{"invalid versions", "foo", "bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.CanConvert(tt.fromVersion, tt.toVersion)
			if result != tt.expected {
				t.Errorf("CanConvert(%s, %s) = %v, want %v",
					tt.fromVersion, tt.toVersion, result, tt.expected)
			}
		})
	}
}

func TestConvertV1ToV2Beta1(t *testing.T) {
	converter := NewBMCConverter()

	v1BMC := &bmc.BMC{
		Resource: resource.Resource{
			APIVersion:    "inventory/v1",
			Kind:          "BMC",
			SchemaVersion: "v1",
		},
		Spec: bmc.BMCSpec{
			Address:  "https://10.1.1.100",
			Username: "admin",
			Password: "secret123",
			Type:     "Redfish",
		},
		Status: bmc.BMCStatus{
			Connected: true,
			Reachable: true,
			Version:   "1.0.0",
		},
	}

	result, err := converter.Convert(v1BMC, "v1", "v2beta1")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	v2Beta1BMC, ok := result.(*BMC)
	if !ok {
		t.Fatalf("Result is not *v2beta1.BMC, got %T", result)
	}

	// Verify basic fields
	if v2Beta1BMC.SchemaVersion != "v2beta1" {
		t.Errorf("SchemaVersion = %s, want v2beta1", v2Beta1BMC.SchemaVersion)
	}

	// Verify spec conversion
	if v2Beta1BMC.Spec.Address != v1BMC.Spec.Address {
		t.Errorf("Address = %s, want %s", v2Beta1BMC.Spec.Address, v1BMC.Spec.Address)
	}

	if v2Beta1BMC.Spec.Type != v1BMC.Spec.Type {
		t.Errorf("Type = %s, want %s", v2Beta1BMC.Spec.Type, v1BMC.Spec.Type)
	}

	// Verify authentication conversion
	if v2Beta1BMC.Spec.Authentication.Method != "basic" {
		t.Errorf("Authentication.Method = %s, want basic", v2Beta1BMC.Spec.Authentication.Method)
	}

	if v2Beta1BMC.Spec.Authentication.Basic == nil {
		t.Fatal("Authentication.Basic is nil")
	}

	if v2Beta1BMC.Spec.Authentication.Basic.Username != v1BMC.Spec.Username {
		t.Errorf("Basic.Username = %s, want %s",
			v2Beta1BMC.Spec.Authentication.Basic.Username, v1BMC.Spec.Username)
	}

	if v2Beta1BMC.Spec.Authentication.Basic.Password != v1BMC.Spec.Password {
		t.Errorf("Basic.Password = %s, want %s",
			v2Beta1BMC.Spec.Authentication.Basic.Password, v1BMC.Spec.Password)
	}

	// Verify status conversion
	if v2Beta1BMC.Status.Connected != v1BMC.Status.Connected {
		t.Errorf("Status.Connected = %v, want %v",
			v2Beta1BMC.Status.Connected, v1BMC.Status.Connected)
	}

	if v2Beta1BMC.Status.AuthenticationMethod != "basic" {
		t.Errorf("Status.AuthenticationMethod = %s, want basic",
			v2Beta1BMC.Status.AuthenticationMethod)
	}
}

func TestConvertV2Beta1ToV1_BasicAuth(t *testing.T) {
	converter := NewBMCConverter()

	v2Beta1BMC := &BMC{
		Resource: resource.Resource{
			APIVersion:    "inventory/v2",
			Kind:          "BMC",
			SchemaVersion: "v2beta1",
		},
		Spec: BMCSpec{
			Address: "https://10.1.1.100",
			Type:    "Redfish",
			Authentication: AuthenticationConfig{
				Method: "basic",
				Basic: &BasicAuth{
					Username: "admin",
					Password: "secret123",
				},
			},
		},
		Status: BMCStatus{
			Connected:            true,
			Reachable:            true,
			Version:              "1.0.0",
			AuthenticationMethod: "basic",
		},
	}

	result, err := converter.Convert(v2Beta1BMC, "v2beta1", "v1")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	v1BMC, ok := result.(*bmc.BMC)
	if !ok {
		t.Fatalf("Result is not *bmc.BMC, got %T", result)
	}

	// Verify conversion
	if v1BMC.SchemaVersion != "v1" {
		t.Errorf("SchemaVersion = %s, want v1", v1BMC.SchemaVersion)
	}

	if v1BMC.Spec.Address != v2Beta1BMC.Spec.Address {
		t.Errorf("Address = %s, want %s", v1BMC.Spec.Address, v2Beta1BMC.Spec.Address)
	}

	if v1BMC.Spec.Username != v2Beta1BMC.Spec.Authentication.Basic.Username {
		t.Errorf("Username = %s, want %s",
			v1BMC.Spec.Username, v2Beta1BMC.Spec.Authentication.Basic.Username)
	}

	if v1BMC.Spec.Password != v2Beta1BMC.Spec.Authentication.Basic.Password {
		t.Errorf("Password = %s, want %s",
			v1BMC.Spec.Password, v2Beta1BMC.Spec.Authentication.Basic.Password)
	}
}

func TestConvertV2Beta1ToV1_ClientCertAuth_Fails(t *testing.T) {
	converter := NewBMCConverter()

	v2Beta1BMC := &BMC{
		Resource: resource.Resource{
			APIVersion:    "inventory/v2",
			Kind:          "BMC",
			SchemaVersion: "v2beta1",
		},
		Spec: BMCSpec{
			Address: "https://10.1.1.100",
			Type:    "Redfish",
			Authentication: AuthenticationConfig{
				Method: "client-cert",
				ClientCert: &ClientCertAuth{
					CertificateRef: "secret://default/bmc-cert",
					KeyRef:         "secret://default/bmc-key",
				},
			},
		},
	}

	_, err := converter.Convert(v2Beta1BMC, "v2beta1", "v1")
	if err == nil {
		t.Fatal("Expected error when converting client-cert auth to v1, got nil")
	}

	expectedMsg := "cannot convert client-cert authentication to v1"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Error message = %s, want prefix %s", err.Error(), expectedMsg)
	}
}

func TestConvertV2Beta1ToV1_OIDCAuth_Fails(t *testing.T) {
	converter := NewBMCConverter()

	v2Beta1BMC := &BMC{
		Resource: resource.Resource{
			APIVersion:    "inventory/v2",
			Kind:          "BMC",
			SchemaVersion: "v2beta1",
		},
		Spec: BMCSpec{
			Address: "https://10.1.1.100",
			Type:    "Redfish",
			Authentication: AuthenticationConfig{
				Method: "oidc",
				OIDC: &OIDCAuth{
					IssuerURL:    "https://auth.example.com",
					ClientID:     "inventory-client",
					ClientSecret: "secret",
				},
			},
		},
	}

	_, err := converter.Convert(v2Beta1BMC, "v2beta1", "v1")
	if err == nil {
		t.Fatal("Expected error when converting OIDC auth to v1, got nil")
	}

	expectedMsg := "cannot convert oidc authentication to v1"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Error message = %s, want prefix %s", err.Error(), expectedMsg)
	}
}

func TestRoundTripConversion(t *testing.T) {
	converter := NewBMCConverter()

	originalV1 := &bmc.BMC{
		Resource: resource.Resource{
			APIVersion:    "inventory/v1",
			Kind:          "BMC",
			SchemaVersion: "v1",
		},
		Spec: bmc.BMCSpec{
			Address:  "https://10.1.1.100",
			Username: "admin",
			Password: "secret123",
			Type:     "Redfish",
		},
		Status: bmc.BMCStatus{
			Connected: true,
			Reachable: true,
			Version:   "1.0.0",
		},
	}

	// Convert v1 -> v2beta1
	v2Beta1Result, err := converter.Convert(originalV1, "v1", "v2beta1")
	if err != nil {
		t.Fatalf("v1->v2beta1 conversion failed: %v", err)
	}

	// Convert v2beta1 -> v1
	v1Result, err := converter.Convert(v2Beta1Result, "v2beta1", "v1")
	if err != nil {
		t.Fatalf("v2beta1->v1 conversion failed: %v", err)
	}

	v1BMC, ok := v1Result.(*bmc.BMC)
	if !ok {
		t.Fatalf("Round-trip result is not *bmc.BMC, got %T", v1Result)
	}

	// Verify round-trip preserved data
	if v1BMC.Spec.Address != originalV1.Spec.Address {
		t.Errorf("Round-trip Address = %s, want %s", v1BMC.Spec.Address, originalV1.Spec.Address)
	}
	if v1BMC.Spec.Username != originalV1.Spec.Username {
		t.Errorf("Round-trip Username = %s, want %s", v1BMC.Spec.Username, originalV1.Spec.Username)
	}
	if v1BMC.Spec.Password != originalV1.Spec.Password {
		t.Errorf("Round-trip Password = %s, want %s", v1BMC.Spec.Password, originalV1.Spec.Password)
	}
	if v1BMC.Spec.Type != originalV1.Spec.Type {
		t.Errorf("Round-trip Type = %s, want %s", v1BMC.Spec.Type, originalV1.Spec.Type)
	}
}
