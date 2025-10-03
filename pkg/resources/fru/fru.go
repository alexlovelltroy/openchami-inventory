package fru

import (
	"github.com/alexlovelltroy/fabrica/pkg/resource"
)

// FRU represents a Field Replaceable Unit in the system
type FRU struct {
	resource.Resource
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
	Health      string                `json:"health,omitempty" yaml:"health,omitempty"`           // "OK", "Warning", "Critical", "Unknown"
	State       string                `json:"state,omitempty" yaml:"state,omitempty"`             // "Enabled", "Disabled", "Absent", "Present"
	Functional  bool                  `json:"functional" yaml:"functional"`                       // Whether FRU is functioning properly
	LastSeen    string                `json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`       // Last time FRU was detected
	LastScanned string                `json:"lastScanned,omitempty" yaml:"lastScanned,omitempty"` // Last Redfish scan time
	Errors      []string              `json:"errors,omitempty" yaml:"errors,omitempty"`           // Current error conditions
	Conditions  []resource.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// Request types for code generation
type CreateFRURequest struct {
	Name         string            `json:"name" validate:"required"`
	FRUType      string            `json:"fruType" validate:"required"`
	SerialNumber string            `json:"serialNumber,omitempty"`
	PartNumber   string            `json:"partNumber,omitempty"`
	Manufacturer string            `json:"manufacturer,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

type UpdateFRURequest struct {
	Name         string            `json:"name,omitempty"`
	FRUType      string            `json:"fruType,omitempty"`
	SerialNumber string            `json:"serialNumber,omitempty"`
	PartNumber   string            `json:"partNumber,omitempty"`
	Manufacturer string            `json:"manufacturer,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// FRUInventorySnapshot represents a snapshot of FRU inventory at a point in time
type FRUInventorySnapshot struct {
	resource.Resource
	Spec   FRUInventorySnapshotSpec   `json:"spec" yaml:"spec"`
	Status FRUInventorySnapshotStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// FRUInventorySnapshotSpec defines the desired state of FRUInventorySnapshot
type FRUInventorySnapshotSpec struct {
	SnapshotTime string   `json:"snapshotTime" yaml:"snapshotTime"` // RFC3339 timestamp
	Source       string   `json:"source" yaml:"source"`             // Source of the snapshot (e.g., "redfish-crawler")
	Scope        string   `json:"scope" yaml:"scope"`               // Scope of the snapshot (e.g., "bmc", "node", "cluster")
	ScopeUID     string   `json:"scopeUid" yaml:"scopeUid"`         // UID of the scoped resource
	FRUIDs       []string `json:"fruIds" yaml:"fruIds"`             // List of FRU UIDs in this snapshot
}

// FRUInventorySnapshotStatus defines the observed state of FRUInventorySnapshot
type FRUInventorySnapshotStatus struct {
	Complete       bool                  `json:"complete" yaml:"complete"`                         // Whether snapshot is complete
	FRUCount       int                   `json:"fruCount" yaml:"fruCount"`                         // Total number of FRUs found
	NewFRUs        int                   `json:"newFrus" yaml:"newFrus"`                           // Number of new FRUs discovered
	ProcessingTime float64               `json:"processingTime" yaml:"processingTime"`             // Time taken to process in seconds
	Conditions     []resource.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"` // Status conditions
}
