package v2beta1

import (
	"fmt"

	"github.com/openchami/inventory/pkg/resources/bmc"
)

// BMCConverter implements bidirectional conversion between BMC v1 and v2beta1
type BMCConverter struct{}

// NewBMCConverter creates a new BMC version converter
func NewBMCConverter() *BMCConverter {
	return &BMCConverter{}
}

// CanConvert checks if conversion is supported between versions
func (c *BMCConverter) CanConvert(fromVersion, toVersion string) bool {
	supportedConversions := map[string]map[string]bool{
		"v1": {
			"v2beta1": true,
		},
		"v2beta1": {
			"v1": true,
		},
	}

	if toVersions, exists := supportedConversions[fromVersion]; exists {
		return toVersions[toVersion]
	}
	return false
}

// Convert transforms a resource from one version to another
func (c *BMCConverter) Convert(resource interface{}, fromVersion, toVersion string) (interface{}, error) {
	switch {
	case fromVersion == "v1" && toVersion == "v2beta1":
		v1BMC, ok := resource.(*bmc.BMC)
		if !ok {
			return nil, fmt.Errorf("expected *bmc.BMC, got %T", resource)
		}
		return c.convertV1ToV2Beta1(v1BMC)

	case fromVersion == "v2beta1" && toVersion == "v1":
		v2Beta1BMC, ok := resource.(*BMC)
		if !ok {
			return nil, fmt.Errorf("expected *v2beta1.BMC, got %T", resource)
		}
		return c.convertV2Beta1ToV1(v2Beta1BMC)

	default:
		return nil, fmt.Errorf("unsupported conversion: %s -> %s", fromVersion, toVersion)
	}
}

// ConvertSpec transforms just the spec portion
func (c *BMCConverter) ConvertSpec(spec interface{}, fromVersion, toVersion string) (interface{}, error) {
	switch {
	case fromVersion == "v1" && toVersion == "v2beta1":
		v1Spec, ok := spec.(bmc.BMCSpec)
		if !ok {
			return nil, fmt.Errorf("expected bmc.BMCSpec, got %T", spec)
		}
		return c.convertV1SpecToV2Beta1(v1Spec), nil

	case fromVersion == "v2beta1" && toVersion == "v1":
		v2Beta1Spec, ok := spec.(BMCSpec)
		if !ok {
			return nil, fmt.Errorf("expected v2beta1.BMCSpec, got %T", spec)
		}
		return c.convertV2Beta1SpecToV1(v2Beta1Spec)

	default:
		return nil, fmt.Errorf("unsupported spec conversion: %s -> %s", fromVersion, toVersion)
	}
}

// ConvertStatus transforms just the status portion
func (c *BMCConverter) ConvertStatus(status interface{}, fromVersion, toVersion string) (interface{}, error) {
	switch {
	case fromVersion == "v1" && toVersion == "v2beta1":
		v1Status, ok := status.(bmc.BMCStatus)
		if !ok {
			return nil, fmt.Errorf("expected bmc.BMCStatus, got %T", status)
		}
		return c.convertV1StatusToV2Beta1(v1Status), nil

	case fromVersion == "v2beta1" && toVersion == "v1":
		v2Beta1Status, ok := status.(BMCStatus)
		if !ok {
			return nil, fmt.Errorf("expected v2beta1.BMCStatus, got %T", status)
		}
		return c.convertV2Beta1StatusToV1(v2Beta1Status), nil

	default:
		return nil, fmt.Errorf("unsupported status conversion: %s -> %s", fromVersion, toVersion)
	}
}

// convertV1ToV2Beta1 converts a v1 BMC to v2beta1
func (c *BMCConverter) convertV1ToV2Beta1(v1BMC *bmc.BMC) (*BMC, error) {
	v2Beta1BMC := &BMC{
		Resource: v1BMC.Resource,
		Spec:     c.convertV1SpecToV2Beta1(v1BMC.Spec),
		Status:   c.convertV1StatusToV2Beta1(v1BMC.Status),
	}

	// Update schema version
	v2Beta1BMC.SchemaVersion = "v2beta1"

	return v2Beta1BMC, nil
}

// convertV1SpecToV2Beta1 converts v1 spec to v2beta1 spec
func (c *BMCConverter) convertV1SpecToV2Beta1(v1Spec bmc.BMCSpec) BMCSpec {
	return BMCSpec{
		Address: v1Spec.Address,
		Type:    v1Spec.Type,
		Authentication: AuthenticationConfig{
			Method: "basic", // v1 always uses basic auth
			Basic: &BasicAuth{
				Username: v1Spec.Username,
				Password: v1Spec.Password,
			},
		},
	}
}

// convertV1StatusToV2Beta1 converts v1 status to v2beta1 status
func (c *BMCConverter) convertV1StatusToV2Beta1(v1Status bmc.BMCStatus) BMCStatus {
	return BMCStatus{
		Connected:            v1Status.Connected,
		Reachable:            v1Status.Reachable,
		Version:              v1Status.Version,
		LastSeen:             v1Status.LastSeen,
		Conditions:           v1Status.Conditions,
		AuthenticationMethod: "basic", // v1 always uses basic auth
	}
}

// convertV2Beta1ToV1 converts a v2beta1 BMC to v1
func (c *BMCConverter) convertV2Beta1ToV1(v2Beta1BMC *BMC) (*bmc.BMC, error) {
	v1Spec, err := c.convertV2Beta1SpecToV1(v2Beta1BMC.Spec)
	if err != nil {
		return nil, err
	}

	v1BMC := &bmc.BMC{
		Resource: v2Beta1BMC.Resource,
		Spec:     v1Spec,
		Status:   c.convertV2Beta1StatusToV1(v2Beta1BMC.Status),
	}

	// Update schema version
	v1BMC.SchemaVersion = "v1"

	return v1BMC, nil
}

// convertV2Beta1SpecToV1 converts v2beta1 spec to v1 spec
func (c *BMCConverter) convertV2Beta1SpecToV1(v2Beta1Spec BMCSpec) (bmc.BMCSpec, error) {
	v1Spec := bmc.BMCSpec{
		Address: v2Beta1Spec.Address,
		Type:    v2Beta1Spec.Type,
	}

	// Only basic auth is supported in v1
	switch v2Beta1Spec.Authentication.Method {
	case "basic":
		if v2Beta1Spec.Authentication.Basic != nil {
			v1Spec.Username = v2Beta1Spec.Authentication.Basic.Username
			v1Spec.Password = v2Beta1Spec.Authentication.Basic.Password
		}
	case "client-cert", "oidc":
		return v1Spec, fmt.Errorf("cannot convert %s authentication to v1 (only basic auth supported in v1)", v2Beta1Spec.Authentication.Method)
	default:
		return v1Spec, fmt.Errorf("unknown authentication method: %s", v2Beta1Spec.Authentication.Method)
	}

	return v1Spec, nil
}

// convertV2Beta1StatusToV1 converts v2beta1 status to v1 status
func (c *BMCConverter) convertV2Beta1StatusToV1(v2Beta1Status BMCStatus) bmc.BMCStatus {
	return bmc.BMCStatus{
		Connected:  v2Beta1Status.Connected,
		Reachable:  v2Beta1Status.Reachable,
		Version:    v2Beta1Status.Version,
		LastSeen:   v2Beta1Status.LastSeen,
		Conditions: v2Beta1Status.Conditions,
	}
}
