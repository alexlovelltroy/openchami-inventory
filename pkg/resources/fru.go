package resources

// FRU represents a Field Replaceable Unit in the system
type FRU struct {
	Resource
	Spec   FRUSpec   `json:"spec" yaml:"spec"`
	Status FRUStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// FRUSpec defines the desired state of FRU
type FRUSpec struct {
	FRUType         string            `json:"fruType" yaml:"fruType"` // "CPU", "Memory", "Storage", "PSU", "Fan", "Network", "PCICard", "Chassis", "Other"
	SerialNumber    string            `json:"serialNumber,omitempty" yaml:"serialNumber,omitempty"`
	PartNumber      string            `json:"partNumber,omitempty" yaml:"partNumber,omitempty"`
	Manufacturer    string            `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model           string            `json:"model,omitempty" yaml:"model,omitempty"`
	Description     string            `json:"description,omitempty" yaml:"description,omitempty"`
	Version         string            `json:"version,omitempty" yaml:"version,omitempty"`
	FirmwareVersion string            `json:"firmwareVersion,omitempty" yaml:"firmwareVersion,omitempty"`
	Location        FRULocation       `json:"location" yaml:"location"`
	Parent          string            `json:"parent,omitempty" yaml:"parent,omitempty"`           // UID of parent FRU or Node
	Children        []string          `json:"children,omitempty" yaml:"children,omitempty"`       // UIDs of child FRUs
	RedfishPath     string            `json:"redfishPath,omitempty" yaml:"redfishPath,omitempty"` // Redfish endpoint path
	Properties      map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`   // Type-specific properties
}

// FRULocation describes where the FRU is physically located
type FRULocation struct {
	BMCUID   string `json:"bmcUid,omitempty" yaml:"bmcUid,omitempty"`     // BMC that reported this FRU
	NodeUID  string `json:"nodeUid,omitempty" yaml:"nodeUid,omitempty"`   // Associated node
	Rack     string `json:"rack,omitempty" yaml:"rack,omitempty"`         // Physical rack location
	Chassis  string `json:"chassis,omitempty" yaml:"chassis,omitempty"`   // Chassis identifier
	Slot     string `json:"slot,omitempty" yaml:"slot,omitempty"`         // Slot number/identifier
	Bay      string `json:"bay,omitempty" yaml:"bay,omitempty"`           // Bay identifier
	Position string `json:"position,omitempty" yaml:"position,omitempty"` // Position description
	Socket   string `json:"socket,omitempty" yaml:"socket,omitempty"`     // Socket identifier (for CPUs)
	Channel  string `json:"channel,omitempty" yaml:"channel,omitempty"`   // Channel identifier (for memory)
	Port     string `json:"port,omitempty" yaml:"port,omitempty"`         // Port identifier (for network)
}

// FRUStatus defines the observed state of FRU
type FRUStatus struct {
	Health      string      `json:"health,omitempty" yaml:"health,omitempty"`           // "OK", "Warning", "Critical", "Unknown"
	State       string      `json:"state,omitempty" yaml:"state,omitempty"`             // "Enabled", "Disabled", "Absent", "Present"
	Functional  bool        `json:"functional" yaml:"functional"`                       // Whether FRU is functioning properly
	LastSeen    string      `json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`       // Last time FRU was detected
	LastScanned string      `json:"lastScanned,omitempty" yaml:"lastScanned,omitempty"` // Last Redfish scan time
	Errors      []string    `json:"errors,omitempty" yaml:"errors,omitempty"`           // Current error conditions
	Conditions  []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// FRUAssociation represents the relationship between FRUs and other inventory items
type FRUAssociation struct {
	Resource
	Spec   FRUAssociationSpec   `json:"spec" yaml:"spec"`
	Status FRUAssociationStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// FRUAssociationSpec defines the desired state of FRUAssociation
type FRUAssociationSpec struct {
	FRUUID       string  `json:"fruUid" yaml:"fruUid"`                             // UID of the FRU
	AssociatedTo string  `json:"associatedTo" yaml:"associatedTo"`                 // UID of associated resource (Node, BMC, etc.)
	Relationship string  `json:"relationship" yaml:"relationship"`                 // "contains", "installedIn", "connectedTo", "managedBy"
	Confidence   float64 `json:"confidence,omitempty" yaml:"confidence,omitempty"` // Confidence level (0.0-1.0) for the association
	Source       string  `json:"source,omitempty" yaml:"source,omitempty"`         // Source of the association ("redfish", "manual", "discovery")
	Priority     int     `json:"priority,omitempty" yaml:"priority,omitempty"`     // Priority for conflicting associations
}

// FRUAssociationStatus defines the observed state of FRUAssociation
type FRUAssociationStatus struct {
	Active       bool        `json:"active" yaml:"active"`
	LastVerified string      `json:"lastVerified,omitempty" yaml:"lastVerified,omitempty"`
	VerifiedBy   string      `json:"verifiedBy,omitempty" yaml:"verifiedBy,omitempty"` // Method used for verification
	Conflicts    []string    `json:"conflicts,omitempty" yaml:"conflicts,omitempty"`   // UIDs of conflicting associations
	Conditions   []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// FRUInventorySnapshot represents a point-in-time snapshot of FRU inventory
type FRUInventorySnapshot struct {
	Resource
	Spec   FRUInventorySnapshotSpec   `json:"spec" yaml:"spec"`
	Status FRUInventorySnapshotStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// FRUInventorySnapshotSpec defines the desired state of FRUInventorySnapshot
type FRUInventorySnapshotSpec struct {
	SnapshotTime string   `json:"snapshotTime" yaml:"snapshotTime"`             // RFC3339 timestamp
	Source       string   `json:"source" yaml:"source"`                         // "redfish-crawler", "manual", "import"
	Scope        string   `json:"scope,omitempty" yaml:"scope,omitempty"`       // "cluster", "rack", "node", "bmc"
	ScopeUID     string   `json:"scopeUid,omitempty" yaml:"scopeUid,omitempty"` // UID of the scope resource
	FRUIDs       []string `json:"fruIds" yaml:"fruIds"`                         // UIDs of FRUs in this snapshot
}

// FRUInventorySnapshotStatus defines the observed state of FRUInventorySnapshot
type FRUInventorySnapshotStatus struct {
	Complete       bool        `json:"complete" yaml:"complete"`
	FRUCount       int         `json:"fruCount" yaml:"fruCount"`
	NewFRUs        int         `json:"newFRUs,omitempty" yaml:"newFRUs,omitempty"`
	UpdatedFRUs    int         `json:"updatedFRUs,omitempty" yaml:"updatedFRUs,omitempty"`
	RemovedFRUs    int         `json:"removedFRUs,omitempty" yaml:"removedFRUs,omitempty"`
	Errors         []string    `json:"errors,omitempty" yaml:"errors,omitempty"`
	ProcessingTime float64     `json:"processingTime,omitempty" yaml:"processingTime,omitempty"` // Seconds
	Conditions     []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
