package resources

// BMC represents a Baseboard Management Controller
type BMC struct {
	Resource
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
	Connected bool   `json:"connected" yaml:"connected"`
	Reachable bool   `json:"reachable" yaml:"reachable"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	LastSeen  string `json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`
}
