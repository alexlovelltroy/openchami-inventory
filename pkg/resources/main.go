// Package resources provides the core types and helper functions for infrastructure inventory management.
//
// This package implements a Kubernetes-inspired resource model for managing physical and virtual
// infrastructure components including nodes, BMCs, FRUs (Field Replaceable Units), and boot
// configurations. It provides a unified, extensible framework for inventory tracking, lifecycle
// management, and operational state monitoring.
//
// # Architecture Overview
//
// The package is organized into several key components:
//
//   - Core Types (resource.go, metadata.go) - Base Resource struct and metadata handling
//   - Conditions (conditions.go) - Status tracking and health monitoring
//   - Infrastructure Resources (node.go, bmc.go) - Physical/virtual machines and management controllers
//   - Hardware Inventory (fru.go) - Field Replaceable Unit tracking with Redfish integration
//   - Boot Management (boot.go) - Versioned boot configurations with immutable specs
//
// All resource types follow a consistent pattern:
//   - Embed the base Resource struct for common functionality
//   - Define a Spec struct for desired state (immutable in some cases like boot configs)
//   - Define a Status struct for observed state with conditions
//   - Use structured UIDs for better log readability
//   - Support JSON and YAML serialization
//
// # Quick Start Examples
//
// Creating and managing a Node resource:
//
//	// Create a node with structured UID
//	node := &Node{
//	    Resource: Resource{
//	        APIVersion: "v1",
//	        Kind: "Node",
//	        SchemaVersion: "1.0",
//	    },
//	    Spec: NodeSpec{
//	        Hostname: "worker-001",
//	        Rack: "rack-01",
//	        Interfaces: NodeEthernetSpec{
//	            Interfaces: []EthernetInterface{
//	                {Name: "eth0", MACAddress: "aa:bb:cc:dd:ee:ff"},
//	            },
//	        },
//	    },
//	}
//
//	// Initialize with generated UID
//	uid, _ := GenerateUIDForResource("Node")  // Generates "nd-1a2b3c4d"
//	node.Metadata.Initialize("worker-001", uid)
//
//	// Add operational metadata
//	node.SetLabel("environment", "production")
//	node.SetLabel("rack", "rack-01")
//	node.SetAnnotation("deployed.by", "automation-system")
//
//	// Track operational status
//	SetCondition(&node.Status.Conditions, "Ready", "True", "NodeHealthy", "All systems operational")
//
// Working with FRU inventory:
//
//	// Create a CPU FRU discovered via Redfish
//	cpu := &FRU{
//	    Resource: Resource{APIVersion: "v1", Kind: "FRU"},
//	    Spec: FRUSpec{
//	        FRUType: "CPU",
//	        SerialNumber: "CPU123456",
//	        Manufacturer: "Intel",
//	        Model: "Xeon E5-2680v4",
//	        Location: FRULocation{
//	            BMCUID: "bmc-abc12345",
//	            Socket: "CPU1",
//	        },
//	        RedfishPath: "/redfish/v1/Systems/1/Processors/CPU1",
//	    },
//	}
//
//	fruUID, _ := GenerateUIDForResource("FRU")
//	cpu.Metadata.Initialize("cpu-socket1", fruUID)
//
//	// Associate with a node
//	association := &FRUAssociation{
//	    Spec: FRUAssociationSpec{
//	        FRUUID: cpu.GetUID(),
//	        AssociatedTo: node.GetUID(),
//	        Relationship: "installedIn",
//	        Source: "redfish",
//	        Confidence: 1.0,
//	    },
//	}
//
// Managing versioned boot configurations:
//
//	// Create base boot configuration (version 1.0.0)
//	bootConfig := &BootConfiguration{
//	    Resource: Resource{APIVersion: "v1", Kind: "BootConfiguration"},
//	    Spec: BootConfigSpec{
//	        Version: "1.0.0",
//	        BaseConfigUID: "bc-base001",  // Same across all versions
//	        KernelURI: "http://repo.example.com/kernels/vmlinuz-5.15.0",
//	        InitrdURI: "http://repo.example.com/kernels/initrd-5.15.0",
//	        KernelParams: "console=ttyS0,115200 crashkernel=auto",
//	        BootMode: "pxe",
//	    },
//	}
//
//	// Create an alias pointing to latest version
//	alias := &BootConfigAlias{
//	    Spec: BootConfigAliasSpec{
//	        BaseConfigUID: "bc-base001",
//	        AliasName: "latest",
//	        TargetVersion: "1.0.0",
//	        AutoUpdate: true,
//	    },
//	}
//
//	// Bind node to boot config using alias
//	binding := &BootBinding{
//	    Spec: BootBindingSpec{
//	        NodeUID: node.GetUID(),
//	        BaseConfigUID: "bc-base001",
//	        ConfigVersion: "latest",  // Will resolve to current "latest" alias target
//	        Enabled: true,
//	    },
//	}
//
// # Structured UIDs
//
// The package uses human-readable structured UIDs instead of UUIDs for better log analysis
// and debugging. UIDs follow the format: <prefix>-<hex-digits>
//
// Built-in prefixes:
//   - nd-xxxxxxxx    - Node resources
//   - bmc-xxxxxxxx   - BMC resources
//   - fru-xxxxxxxx   - FRU resources
//   - bc-xxxxxxxx    - BootConfiguration resources
//   - bca-xxxxxxxx   - BootConfigAlias resources
//   - bb-xxxxxxxx    - BootBinding resources
//   - fa-xxxxxxxx    - FRUAssociation resources
//   - fis-xxxxxxxx   - FRUInventorySnapshot resources
//
// Register custom resource types:
//
//	func init() {
//	    resources.RegisterResourcePrefix("CustomResource", "cr")
//	}
//
// # Labels vs Annotations
//
// Labels and annotations serve different purposes:
//
// Labels (for selection, grouping, automation):
//   - Short, structured key-value pairs (max 63 chars)
//   - Used by selectors and queries
//   - Examples: environment=production, rack=rack-01, role=compute
//
// Annotations (for metadata, documentation):
//
//   - Arbitrary key-value pairs, can be longer
//
//   - Not used for selection, purely informational
//
//   - Examples: deployment.notes, contact.email, last-maintenance-date
//
//     resource.SetLabel("environment", "production")        // Queryable
//     resource.SetAnnotation("maintenance.notes", "...")    // Informational
//
// # Condition Management
//
// Conditions track operational status using the Kubernetes pattern. Each condition has:
//   - Type: What aspect is being tracked ("Ready", "Healthy", "Reachable")
//   - Status: Current state ("True", "False", "Unknown")
//   - Reason: Machine-readable reason code
//   - Message: Human-readable description
//   - LastTransitionTime: When status last changed
//
// Common condition patterns:
//
//	// Set a condition (creates or updates)
//	SetCondition(&resource.Status.Conditions, "Ready", "True", "AllChecksPass", "Resource is operational")
//
//	// Check condition status
//	if IsConditionTrue(resource.Status.Conditions, "Ready") {
//	    // Resource is ready
//	}
//
//	// Get condition details
//	if condition := FindCondition(resource.Status.Conditions, "Healthy"); condition != nil {
//	    fmt.Printf("Health status: %s (for %v)\n", condition.Status, condition.Age())
//	}
//
// # File Organization
//
// The package is organized into focused files:
//   - main.go: Package documentation and initialization
//   - resource.go: Core Resource struct and UID management
//   - metadata.go: Metadata struct and helper methods
//   - conditions.go: Condition struct and management functions
//   - node.go: Node resource definitions
//   - bmc.go: BMC resource definitions
//   - fru.go: FRU and inventory management
//   - boot.go: Boot configuration with versioning
//
// # Integration Patterns
//
// Common integration patterns for using this package:
//
//	// Discovery agent pattern
//	func discoverNodes() ([]*Node, error) {
//	    var nodes []*Node
//	    // ... discovery logic ...
//	    for _, discovered := range discoveredNodes {
//	        node := &Node{...}
//	        uid, _ := GenerateUIDForResource("Node")
//	        node.Metadata.Initialize(discovered.Hostname, uid)
//	        SetCondition(&node.Status.Conditions, "Discovered", "True", "AutoDiscovery", "Found via network scan")
//	        nodes = append(nodes, node)
//	    }
//	    return nodes, nil
//	}
//
//	// Health monitoring pattern
//	func checkResourceHealth(resource *Resource) {
//	    if isHealthy(resource) {
//	        SetCondition(&status.Conditions, "Healthy", "True", "HealthCheck", "All metrics within normal range")
//	    } else {
//	        SetCondition(&status.Conditions, "Healthy", "False", "HealthCheck", "Performance degraded")
//	    }
//	    resource.Touch() // Update timestamp
//	}
package resources

// init registers the built-in resource types with their standard prefixes.
//
// This ensures that the core inventory resource types are available by default.
// Other packages can register additional resource types by calling
// RegisterResourcePrefix in their own init functions.
func init() {
	// Register built-in resource types
	RegisterResourcePrefix("Node", "nd")
	RegisterResourcePrefix("BMC", "bmc")
	RegisterResourcePrefix("FRU", "fru")
	RegisterResourcePrefix("BootConfiguration", "bc")
	RegisterResourcePrefix("BootConfigAlias", "bca")
	RegisterResourcePrefix("BootBinding", "bb")
	RegisterResourcePrefix("FRUAssociation", "fa")
	RegisterResourcePrefix("FRUInventorySnapshot", "fis")
}
