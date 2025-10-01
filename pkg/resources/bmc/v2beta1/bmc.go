// Package v2beta1 provides the v2beta1 version of the BMC resource with enhanced authentication.
//
// This version introduces multiple authentication methods for Redfish including:
//   - Traditional username/password (backward compatible with v1)
//   - Client certificate authentication (mTLS)
//   - OpenID Connect (OIDC) authentication
//
// Version: v2beta1 (beta)
// Stability: Beta - API may change before v2 stable release
package v2beta1

import (
	"github.com/openchami/inventory/pkg/resources"
)

// BMC represents a Baseboard Management Controller with enhanced authentication in v2beta1
type BMC struct {
	resources.Resource
	Spec   BMCSpec   `json:"spec" yaml:"spec"`
	Status BMCStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// BMCSpec defines the desired state of BMC with multiple authentication methods
type BMCSpec struct {
	// Address is the BMC endpoint (e.g., "https://10.1.1.100", "ipmi://10.1.1.100")
	Address string `json:"address" yaml:"address"`

	// Type specifies the BMC protocol type (e.g., "IPMI", "Redfish", "iDRAC")
	Type string `json:"type" yaml:"type"`

	// Authentication configuration - one method must be specified
	Authentication AuthenticationConfig `json:"authentication" yaml:"authentication"`
}

// AuthenticationConfig defines multiple authentication methods for BMC access
type AuthenticationConfig struct {
	// Method specifies which authentication method to use
	// Valid values: "basic", "client-cert", "oidc"
	Method string `json:"method" yaml:"method"`

	// Basic authentication (username/password) - backward compatible with v1
	Basic *BasicAuth `json:"basic,omitempty" yaml:"basic,omitempty"`

	// ClientCert authentication (mTLS)
	ClientCert *ClientCertAuth `json:"clientCert,omitempty" yaml:"clientCert,omitempty"`

	// OIDC authentication
	OIDC *OIDCAuth `json:"oidc,omitempty" yaml:"oidc,omitempty"`
}

// BasicAuth provides traditional username/password authentication
type BasicAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"` // Omit in responses
}

// ClientCertAuth provides mutual TLS authentication
type ClientCertAuth struct {
	// CertificateRef references a certificate stored in the system
	// Format: "secret://namespace/name" or "file:///path/to/cert.pem"
	CertificateRef string `json:"certificateRef" yaml:"certificateRef"`

	// KeyRef references the private key for the certificate
	// Format: "secret://namespace/name" or "file:///path/to/key.pem"
	KeyRef string `json:"keyRef" yaml:"keyRef"`

	// CABundle is the CA certificate bundle for server verification (optional)
	CABundle string `json:"caBundle,omitempty" yaml:"caBundle,omitempty"`
}

// OIDCAuth provides OpenID Connect authentication
type OIDCAuth struct {
	// IssuerURL is the OIDC provider's issuer URL
	IssuerURL string `json:"issuerUrl" yaml:"issuerUrl"`

	// ClientID is the OAuth2 client identifier
	ClientID string `json:"clientId" yaml:"clientId"`

	// ClientSecret is the OAuth2 client secret (omit in responses)
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`

	// Scopes are the OAuth2 scopes to request (default: ["openid"])
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`

	// TokenEndpoint is the token endpoint URL (optional, auto-discovered if not specified)
	TokenEndpoint string `json:"tokenEndpoint,omitempty" yaml:"tokenEndpoint,omitempty"`
}

// BMCStatus defines the observed state of BMC (unchanged from v1)
type BMCStatus struct {
	Connected  bool                  `json:"connected" yaml:"connected"`
	Reachable  bool                  `json:"reachable" yaml:"reachable"`
	Version    string                `json:"version,omitempty" yaml:"version,omitempty"`
	LastSeen   string                `json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`
	Conditions []resources.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`

	// AuthenticationMethod reports which authentication method is currently active
	AuthenticationMethod string `json:"authenticationMethod,omitempty" yaml:"authenticationMethod,omitempty"`
}

// Request types for code generation (v2beta1)
type CreateBMCRequest struct {
	Name           string               `json:"name" validate:"required"`
	Address        string               `json:"address" validate:"required"`
	Type           string               `json:"type" validate:"required"`
	Authentication AuthenticationConfig `json:"authentication" validate:"required"`
	Labels         map[string]string    `json:"labels,omitempty"`
	Annotations    map[string]string    `json:"annotations,omitempty"`
}

type UpdateBMCRequest struct {
	Name           string                `json:"name,omitempty"`
	Address        string                `json:"address,omitempty"`
	Type           string                `json:"type,omitempty"`
	Authentication *AuthenticationConfig `json:"authentication,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	Annotations    map[string]string     `json:"annotations,omitempty"`
}

// ToBMCSpec converts CreateBMCRequest to BMCSpec
func (r *CreateBMCRequest) ToBMCSpec() BMCSpec {
	return BMCSpec{
		Address:        r.Address,
		Type:           r.Type,
		Authentication: r.Authentication,
	}
}

// ApplyToBMC applies UpdateBMCRequest to a BMC resource
func (r *UpdateBMCRequest) ApplyToBMC(bmc *BMC) {
	if r.Name != "" {
		bmc.SetName(r.Name)
	}
	if r.Address != "" {
		bmc.Spec.Address = r.Address
	}
	if r.Type != "" {
		bmc.Spec.Type = r.Type
	}
	if r.Authentication != nil {
		bmc.Spec.Authentication = *r.Authentication
	}

	// Update labels and annotations
	for k, v := range r.Labels {
		bmc.SetLabel(k, v)
	}
	for k, v := range r.Annotations {
		bmc.SetAnnotation(k, v)
	}
}
