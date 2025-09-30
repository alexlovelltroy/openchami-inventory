package node

import (
	"github.com/openchami/inventory/pkg/resources"
)

// Node represents a physical or virtual machine
type Node struct {
	resources.Resource
	Spec   NodeSpec   `json:"spec" yaml:"spec"`
	Status NodeStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// NodeEthernetSpec defines the desired state of Ethernet interfaces on the Node
type NodeEthernetSpec struct {
	Interfaces []EthernetInterface `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
}

// EthernetInterface represents a single Ethernet interface on the Node
type EthernetInterface struct {
	Name       string            `json:"name" yaml:"name"`
	MACAddress string            `json:"macAddress" yaml:"macAddress"`
	Addresses  map[string]string `json:"ipAddresses,omitempty" yaml:"ipAddresses,omitempty"` // hostname -> ipAddress
	Notes      string            `json:"notes,omitempty" yaml:"notes,omitempty"`
	Priority   int               `json:"priority,omitempty" yaml:"priority,omitempty"` // used for ordering
}

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	BMCUID          string            `json:"bmcUid,omitempty" yaml:"bmcUid,omitempty"`
	Hostname        string            `json:"hostname" yaml:"hostname"`
	Interfaces      NodeEthernetSpec  `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	Rack            string            `json:"rack,omitempty" yaml:"rack,omitempty"`
	Position        string            `json:"position,omitempty" yaml:"position,omitempty"`
	AssetTag        string            `json:"assetTag,omitempty" yaml:"assetTag,omitempty"`
	PowerState      string            `json:"powerState,omitempty" yaml:"powerState,omitempty"` // "on", "off", "reboot"
	Attributes      map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	MaintenanceMode bool              `json:"maintenanceMode,omitempty" yaml:"maintenanceMode,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	PowerState  string                `json:"powerState,omitempty" yaml:"powerState,omitempty"`
	Online      bool                  `json:"online" yaml:"online"`
	LastBooted  string                `json:"lastBooted,omitempty" yaml:"lastBooted,omitempty"`
	HealthState string                `json:"healthState,omitempty" yaml:"healthState,omitempty"` // "healthy", "degraded", "failed"
	Conditions  []resources.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// Request types for code generation
type CreateNodeRequest struct {
	Name        string            `json:"name" validate:"required"`
	Hostname    string            `json:"hostname" validate:"required"`
	BMCUID      string            `json:"bmcUid,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type UpdateNodeRequest struct {
	Name        string            `json:"name,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	BMCUID      string            `json:"bmcUid,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
