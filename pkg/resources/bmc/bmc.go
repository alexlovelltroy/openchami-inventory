package bmc

import (
	"github.com/alexlovelltroy/fabrica/pkg/resource"
)

// BMC represents a Baseboard Management Controller
type BMC struct {
	resource.Resource
	Spec   BMCSpec   `json:"spec" yaml:"spec"`
	Status BMCStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// BMCSpec defines the desired state of BMC
type BMCSpec struct {
	Address  string `json:"address" yaml:"address"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Type     string `json:"type" yaml:"type"` // e.g., "IPMI", "Redfish", "iDRAC"
}

// BMCStatus defines the observed state of BMC
type BMCStatus struct {
	Connected  bool                  `json:"connected" yaml:"connected"`
	Reachable  bool                  `json:"reachable" yaml:"reachable"`
	Version    string                `json:"version,omitempty" yaml:"version,omitempty"`
	LastSeen   string                `json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`
	Conditions []resource.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// ToSpec converts CreateBMCRequest to BMCSpec (for code generation)
func (r *CreateBMCRequest) ToBMCSpec() BMCSpec {
	return BMCSpec{
		Address:  r.Address,
		Username: r.Username,
		Password: r.Password,
		Type:     r.Type,
	}
}

// ApplyToBMC applies UpdateBMCRequest to a BMC resource (for code generation)
func (r *UpdateBMCRequest) ApplyToBMC(bmc *BMC) {
	if r.Name != "" {
		bmc.SetName(r.Name)
	}
	if r.Address != "" {
		bmc.Spec.Address = r.Address
	}
	if r.Username != "" {
		bmc.Spec.Username = r.Username
	}
	if r.Password != "" {
		bmc.Spec.Password = r.Password
	}
	if r.Type != "" {
		bmc.Spec.Type = r.Type
	}

	// Update labels and annotations
	for k, v := range r.Labels {
		bmc.SetLabel(k, v)
	}
	for k, v := range r.Annotations {
		bmc.SetAnnotation(k, v)
	}
}

// Request types for code generation
type CreateBMCRequest struct {
	Name        string            `json:"name" validate:"required"`
	Address     string            `json:"address" validate:"required"`
	Username    string            `json:"username" validate:"required"`
	Password    string            `json:"password,omitempty"`
	Type        string            `json:"type" validate:"required"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type UpdateBMCRequest struct {
	Name        string            `json:"name,omitempty"`
	Address     string            `json:"address,omitempty"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Type        string            `json:"type,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
