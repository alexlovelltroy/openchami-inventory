package crawler

import (
	"fmt"
	"strings"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/openchami/inventory/pkg/resources/fru"
)

// CoerceAll walks a connected gofish client and returns FRUSpecs.
func CoerceAll(c *gofish.APIClient, bmcUID, nodeUID string) ([]fru.FRUSpec, error) {
	var out []fru.FRUSpec

	// Chassis-scoped FRUs
	chassis, err := c.Service.Chassis()
	if err != nil {
		return nil, fmt.Errorf("failed to get chassis: %w", err)
	}

	for _, ch := range chassis {
		chID := ch.ODataID
		out = append(out, fromChassis(ch, bmcUID, nodeUID))

		if p, err := ch.Power(); err == nil {
			for _, ps := range p.PowerSupplies {
				out = append(out, fromPSU(ps, chID, bmcUID, nodeUID))
			}
		}
		if th, err := ch.Thermal(); err == nil {
			for _, f := range th.Fans {
				out = append(out, fromFan(f, chID, bmcUID, nodeUID))
			}
		}
		if devs, err := ch.PCIeDevices(); err == nil {
			for _, d := range devs {
				out = append(out, fromPCIe(d, chID, bmcUID, nodeUID))
			}
		}
	}

	// System-scoped FRUs
	systems, err := c.Service.Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to get systems: %w", err)
	}

	for _, sys := range systems {
		sysID := sys.ODataID
		var chassisID string
		// Skip chassis lookup since Links.Chassis() doesn't exist in this version
		// We'll use the first chassis from our earlier collection if needed
		if len(chassis) > 0 {
			chassisID = chassis[0].ODataID
		}

		if mems, err := sys.Memory(); err == nil {
			for _, m := range mems {
				out = append(out, fromMemory(m, sysID, chassisID, bmcUID, nodeUID))
			}
		}
		if procs, err := sys.Processors(); err == nil {
			for _, p := range procs {
				out = append(out, fromCPU(p, sysID, chassisID, bmcUID, nodeUID))
			}
		}
		if stores, err := sys.Storage(); err == nil {
			for _, st := range stores {
				if drives, err := st.Drives(); err == nil {
					for _, d := range drives {
						out = append(out, fromDrive(d, sysID, chassisID, bmcUID, nodeUID))
					}
				}
			}
		}
		if nics, err := sys.NetworkInterfaces(); err == nil {
			for _, nic := range nics {
				out = append(out, fromNetworkInterface(nic, sysID, chassisID, bmcUID, nodeUID))
			}
		}
	}

	return out, nil
}

func fromChassis(ch *redfish.Chassis, bmcUID, nodeUID string) fru.FRUSpec {
	return fru.FRUSpec{
		FRUType:      "Chassis",
		Manufacturer: ch.Manufacturer,
		Model:        ch.Model,
		PartNumber:   ch.PartNumber,
		SerialNumber: ch.SerialNumber,
		Description:  firstNonEmpty(ch.Name, ch.Description),
		RedfishPath:  ch.ODataID,
		Location:     fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: ch.ID},
		Properties:   map[string]string{"AssetTag": ch.AssetTag},
	}
}

func fromPSU(ps redfish.PowerSupply, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if ps.PowerCapacityWatts != 0 {
		props["powerCapacityWatts"] = toStrFloat(ps.PowerCapacityWatts)
	}
	if ps.LineInputVoltage != 0 {
		props["lineInputVoltage"] = toStrFloat(ps.LineInputVoltage)
	}

	return fru.FRUSpec{
		FRUType:         "PSU",
		Manufacturer:    ps.Manufacturer,
		Model:           ps.Model,
		PartNumber:      ps.PartNumber,
		SerialNumber:    ps.SerialNumber,
		FirmwareVersion: ps.FirmwareVersion,
		Description:     ps.Name,
		RedfishPath:     chassisID + "/Power", // PSUs don't have individual ODataIDs
		Location:        fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID, Bay: ps.MemberID},
		Properties:      props,
	}
}

func fromFan(f redfish.ThermalFan, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if f.Reading != 0 {
		props["reading"] = toStr(f.Reading)
	}
	if f.ReadingUnits != "" {
		props["units"] = string(f.ReadingUnits)
	}

	return fru.FRUSpec{
		FRUType:      "Fan",
		Manufacturer: f.Manufacturer,
		Model:        f.Model,
		PartNumber:   f.PartNumber,
		SerialNumber: f.SerialNumber,
		Description:  f.Name,
		RedfishPath:  chassisID + "/Thermal", // Fans don't have individual ODataIDs
		Location:     fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID, Position: string(f.PhysicalContext)},
		Properties:   props,
	}
}

func fromPCIe(d *redfish.PCIeDevice, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if d.DeviceType != "" {
		props["deviceType"] = string(d.DeviceType)
	}
	if d.FirmwareVersion != "" {
		props["fw"] = d.FirmwareVersion
	}
	if d.ID != "" {
		props["bdf"] = d.ID
	}

	return fru.FRUSpec{
		FRUType:      "PCICard",
		Manufacturer: d.Manufacturer,
		Model:        d.Model,
		PartNumber:   d.PartNumber,
		SerialNumber: d.SerialNumber,
		RedfishPath:  d.ODataID,
		Location:     fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID},
		Properties:   props,
	}
}

func fromMemory(m *redfish.Memory, sysID, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if m.CapacityMiB != 0 {
		props["capacityMiB"] = toStr(m.CapacityMiB)
	}
	if m.OperatingSpeedMhz != 0 {
		props["speedMHz"] = toStr(m.OperatingSpeedMhz)
	}
	if m.BaseModuleType != "" {
		props["moduleType"] = string(m.BaseModuleType)
	}
	if m.MemoryDeviceType != "" {
		props["memoryType"] = string(m.MemoryDeviceType)
	}

	return fru.FRUSpec{
		FRUType:      "Memory",
		Manufacturer: m.Manufacturer,
		Model:        m.Model,
		PartNumber:   m.PartNumber,
		SerialNumber: m.SerialNumber,
		Description:  m.Name,
		Version:      m.FirmwareRevision,
		RedfishPath:  m.ODataID,
		Location:     fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID, Slot: m.DeviceLocator, Channel: m.DeviceLocator},
		Parent:       sysID,
		Properties:   props,
	}
}

func fromCPU(p *redfish.Processor, sysID, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if p.TotalCores != 0 {
		props["totalCores"] = toStr(p.TotalCores)
	}
	if p.TotalThreads != 0 {
		props["totalThreads"] = toStr(p.TotalThreads)
	}
	if p.MaxSpeedMHz != 0 {
		props["maxMHz"] = toStrFloat(p.MaxSpeedMHz)
	}

	// Use processor ID step if available - ProcessorID is a struct, not a pointer
	var firmwareVersion string
	if p.ProcessorID.Step != "" {
		firmwareVersion = p.ProcessorID.Step
	}

	return fru.FRUSpec{
		FRUType:         "CPU",
		Manufacturer:    p.Manufacturer,
		Model:           p.Model,
		PartNumber:      p.PartNumber,
		SerialNumber:    p.SerialNumber,
		FirmwareVersion: firmwareVersion,
		Description:     p.Name,
		RedfishPath:     p.ODataID,
		Location:        fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID, Socket: p.Socket},
		Parent:          sysID,
		Properties:      props,
	}
}

func fromDrive(d *redfish.Drive, sysID, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}
	if d.Protocol != "" {
		props["protocol"] = string(d.Protocol)
	}
	if d.MediaType != "" {
		props["mediaType"] = string(d.MediaType)
	}
	if d.CapacityBytes != 0 {
		props["capacityBytes"] = toStr(d.CapacityBytes)
	}

	// Handle identifiers properly
	if len(d.Identifiers) > 0 {
		for _, id := range d.Identifiers {
			switch id.DurableName {
			case "EUI":
				props["eui64"] = string(id.DurableNameFormat)
			case "NGUID":
				props["nguid"] = string(id.DurableNameFormat)
			case "WWN":
				props["wwn"] = string(id.DurableNameFormat)
			}
		}
	}

	location := fru.FRULocation{
		BMCUID:  bmcUID,
		NodeUID: nodeUID,
		Chassis: chassisID,
	}

	// Try to get location information
	if d.PhysicalLocation.PartLocation.LocationOrdinalValue != 0 {
		location.Bay = toStr(d.PhysicalLocation.PartLocation.LocationOrdinalValue)
	} else if d.Name != "" {
		location.Bay = d.Name
	}

	return fru.FRUSpec{
		FRUType:         "Storage",
		Manufacturer:    d.Manufacturer,
		Model:           d.Model,
		PartNumber:      d.PartNumber,
		SerialNumber:    d.SerialNumber,
		Description:     d.Name,
		Version:         d.Revision,
		FirmwareVersion: d.FirmwareVersion,
		RedfishPath:     d.ODataID,
		Location:        location,
		Parent:          sysID,
		Properties:      props,
	}
}

func fromNetworkInterface(nic *redfish.NetworkInterface, sysID, chassisID, bmcUID, nodeUID string) fru.FRUSpec {
	props := map[string]string{}

	// Get MACs from NetworkPorts
	if ports, err := nic.NetworkPorts(); err == nil && len(ports) > 0 {
		var macs []string
		for _, p := range ports {
			if p.AssociatedNetworkAddresses != nil {
				for _, addr := range p.AssociatedNetworkAddresses {
					if addr != "" {
						macs = append(macs, addr)
					}
				}
			}
		}
		if len(macs) > 0 {
			props["mac"] = strings.Join(macs, ",")
		}
	}

	return fru.FRUSpec{
		FRUType:     "Network",
		Description: nic.Name,
		RedfishPath: nic.ODataID,
		Location:    fru.FRULocation{BMCUID: bmcUID, NodeUID: nodeUID, Chassis: chassisID},
		Parent:      sysID,
		Properties:  props,
	}
}

// ---- tiny helpers
func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func toStr[T ~int | ~int64 | ~uint64 | ~float64](v T) string {
	return fmt.Sprintf("%v", v)
}

func toStrFloat(v float32) string {
	return fmt.Sprintf("%v", v)
}
