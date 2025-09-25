package resources

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
	Resource
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
	Ready           bool        `json:"ready" yaml:"ready"`
	Validated       bool        `json:"validated" yaml:"validated"`
	Error           string      `json:"error,omitempty" yaml:"error,omitempty"`
	UsedBy          []string    `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
	IsLatest        bool        `json:"isLatest" yaml:"isLatest"`                                   // True if this is the latest version
	IsDefault       bool        `json:"isDefault" yaml:"isDefault"`                                 // True if this is the default version
	DeprecatedAfter string      `json:"deprecatedAfter,omitempty" yaml:"deprecatedAfter,omitempty"` // ISO 8601 timestamp when version becomes deprecated
	Conditions      []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// BootConfigAlias represents version aliases for BootConfigurations.
//
// Aliases allow users to reference configurations by semantic names rather than
// specific versions. This provides flexibility for deployments and rollbacks.
//
// Common Aliases:
//   - "latest": Points to the newest version
//   - "default": Points to the designated default version
//   - "stable": Points to a tested, stable version
//   - "previous": Points to the previous version (for rollbacks)
type BootConfigAlias struct {
	Resource
	Spec   BootConfigAliasSpec   `json:"spec" yaml:"spec"`
	Status BootConfigAliasStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// BootConfigAliasSpec defines the desired state of BootConfigAlias.
type BootConfigAliasSpec struct {
	BaseConfigUID string `json:"baseConfigUid" yaml:"baseConfigUid"`                 // UID of the base configuration
	AliasName     string `json:"aliasName" yaml:"aliasName"`                         // Name of the alias (e.g., "latest", "default")
	TargetVersion string `json:"targetVersion" yaml:"targetVersion"`                 // Version this alias points to
	AutoUpdate    bool   `json:"autoUpdate,omitempty" yaml:"autoUpdate,omitempty"`   // Whether to auto-update (for "latest" alias)
	Description   string `json:"description,omitempty" yaml:"description,omitempty"` // Purpose of this alias
}

// BootConfigAliasStatus defines the observed state of BootConfigAlias.
type BootConfigAliasStatus struct {
	ResolvedVersion string      `json:"resolvedVersion" yaml:"resolvedVersion"`                     // Current resolved version
	LastUpdated     string      `json:"lastUpdated,omitempty" yaml:"lastUpdated,omitempty"`         // When alias was last updated
	ReferencedBy    []string    `json:"referencedBy,omitempty" yaml:"referencedBy,omitempty"`       // UIDs of resources using this alias
	ValidationError string      `json:"validationError,omitempty" yaml:"validationError,omitempty"` // Error if target version doesn't exist
	Conditions      []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// BootBinding binds a Node to a BootConfiguration with version support.
//
// BootBindings can reference specific versions or version aliases. Version
// resolution happens at binding time to allow for flexible deployment strategies.
type BootBinding struct {
	Resource
	Spec   BootBindingSpec   `json:"spec" yaml:"spec"`
	Status BootBindingStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// BootBindingSpec defines the desired state of BootBinding with version support.
type BootBindingSpec struct {
	NodeUID       string `json:"nodeUid" yaml:"nodeUid"`
	BaseConfigUID string `json:"baseConfigUid" yaml:"baseConfigUid"`                     // UID of the base boot configuration
	ConfigVersion string `json:"configVersion,omitempty" yaml:"configVersion,omitempty"` // Specific version or alias ("latest", "default", "1.0.0")
	Priority      int    `json:"priority,omitempty" yaml:"priority,omitempty"`
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	OnNextBoot    bool   `json:"onNextBoot,omitempty" yaml:"onNextBoot,omitempty"`
	BootCount     int    `json:"bootCount,omitempty" yaml:"bootCount,omitempty"` // Number of times to use this config
	ExpiresAt     string `json:"expiresAt,omitempty" yaml:"expiresAt,omitempty"`
	VersionLock   bool   `json:"versionLock,omitempty" yaml:"versionLock,omitempty"` // If true, don't auto-resolve aliases
}

// BootBindingStatus defines the observed state of BootBinding with version tracking.
type BootBindingStatus struct {
	Active                 bool        `json:"active" yaml:"active"`
	ResolvedConfigUID      string      `json:"resolvedConfigUid,omitempty" yaml:"resolvedConfigUid,omitempty"` // UID of actual config version being used
	ResolvedVersion        string      `json:"resolvedVersion,omitempty" yaml:"resolvedVersion,omitempty"`     // Actual version resolved from alias
	LastVersionUpdate      string      `json:"lastVersionUpdate,omitempty" yaml:"lastVersionUpdate,omitempty"` // When version was last resolved
	LastBooted             string      `json:"lastBooted,omitempty" yaml:"lastBooted,omitempty"`
	BootCount              int         `json:"bootCount" yaml:"bootCount"`
	LastError              string      `json:"lastError,omitempty" yaml:"lastError,omitempty"`
	ReadyForBoot           bool        `json:"readyForBoot" yaml:"readyForBoot"`
	VersionResolutionError string      `json:"versionResolutionError,omitempty" yaml:"versionResolutionError,omitempty"` // Error resolving version/alias
	Conditions             []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
