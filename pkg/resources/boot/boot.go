package boot

import (
	"github.com/alexlovelltroy/fabrica/pkg/resource"
)

// BootConfiguration represents a boot configuration with versioning support.
//
// BootConfigurations use an immutable versioning system where each change
// creates a new version. This provides audit trails and safe rollback capabilities.
//
// Versioning Behavior:
//   - Each BootConfigSpec is immutable once created
//   - Changes create new versions with incremented version numbers
//   - Users can reference specific versions or use aliases like "latest" or "default"
//   - Version aliases are resolved at binding time for flexibility
type BootConfiguration struct {
	resource.Resource
	Spec   BootConfigSpec   `json:"spec" yaml:"spec"`
	Status BootConfigStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// BootConfigSpec defines the desired state of BootConfiguration.
//
// This spec is immutable once created. Any changes should result in a new
// BootConfiguration resource with an incremented version number.
//
// Version Management:
//   - Version: Semantic version (e.g., "1.0.0", "1.1.0")
//   - BaseConfigUID: UID of the base configuration (same across all versions)
//   - ParentVersion: Version this was derived from (for audit trail)
type BootConfigSpec struct {
	// Version information
	Version       string `json:"version" yaml:"version"`                                 // Semantic version (e.g., "1.0.0")
	BaseConfigUID string `json:"baseConfigUid" yaml:"baseConfigUid"`                     // UID of the base config (same across versions)
	ParentVersion string `json:"parentVersion,omitempty" yaml:"parentVersion,omitempty"` // Version this was derived from
	ChangeNote    string `json:"changeNote,omitempty" yaml:"changeNote,omitempty"`       // Human-readable change description

	// Boot configuration fields (immutable within a version)
	KernelURI    string            `json:"kernelUri,omitempty" yaml:"kernelUri,omitempty"`
	InitrdURI    string            `json:"initrdUri,omitempty" yaml:"initrdUri,omitempty"`
	KernelParams string            `json:"kernelParams,omitempty" yaml:"kernelParams,omitempty"`
	UEFI         bool              `json:"uefi,omitempty" yaml:"uefi,omitempty"`
	PXEConfig    string            `json:"pxeConfig,omitempty" yaml:"pxeConfig,omitempty"`
	ImageURI     string            `json:"imageUri,omitempty" yaml:"imageUri,omitempty"`
	Variables    map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	BootMode     string            `json:"bootMode,omitempty" yaml:"bootMode,omitempty"` // "pxe", "local", "iso"
}

// BootConfigStatus defines the observed state of BootConfiguration.
//
// Status includes validation state, usage tracking, and version lifecycle information.
type BootConfigStatus struct {
	Ready           bool                  `json:"ready" yaml:"ready"`
	Validated       bool                  `json:"validated" yaml:"validated"`
	Error           string                `json:"error,omitempty" yaml:"error,omitempty"`
	UsedBy          []string              `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
	IsLatest        bool                  `json:"isLatest" yaml:"isLatest"`                                   // True if this is the latest version
	IsDefault       bool                  `json:"isDefault" yaml:"isDefault"`                                 // True if this is the default version
	DeprecatedAfter string                `json:"deprecatedAfter,omitempty" yaml:"deprecatedAfter,omitempty"` // ISO 8601 timestamp when version becomes deprecated
	Conditions      []resource.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// Request types for code generation
type CreateBootConfigurationRequest struct {
	Name         string            `json:"name" validate:"required"`
	Version      string            `json:"version" validate:"required"`
	KernelURI    string            `json:"kernelUri,omitempty"`
	InitrdURI    string            `json:"initrdUri,omitempty"`
	KernelParams string            `json:"kernelParams,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

type UpdateBootConfigurationRequest struct {
	Name         string            `json:"name,omitempty"`
	Version      string            `json:"version,omitempty"`
	KernelURI    string            `json:"kernelUri,omitempty"`
	InitrdURI    string            `json:"initrdUri,omitempty"`
	KernelParams string            `json:"kernelParams,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}
