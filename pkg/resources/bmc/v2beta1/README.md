# BMC v2beta1 - Enhanced Authentication

This package implements version `v2beta1` of the BMC (Baseboard Management Controller) resource with enhanced authentication methods for modern Redfish deployments.

## Status

**Version:** v2beta1 (Beta)
**Stability:** Beta - API may change before v2 stable release
**Parent Version:** v1

## What's New in v2beta1

### Enhanced Authentication Methods

The v2beta1 BMC resource extends authentication beyond traditional username/password to support:

1. **Basic Authentication** (backward compatible with v1)
   - Traditional username/password
   - Same as v1 behavior

2. **Client Certificate Authentication (mTLS)**
   - Mutual TLS authentication using X.509 certificates
   - Suitable for high-security environments
   - Eliminates password management

3. **OpenID Connect (OIDC)**
   - Modern OAuth2/OIDC authentication
   - Integration with enterprise identity providers
   - Support for single sign-on (SSO)

## Usage Examples

### Basic Authentication (v1 Compatible)

```json
{
  "apiVersion": "inventory/v2",
  "kind": "BMC",
  "schemaVersion": "v2beta1",
  "metadata": {
    "name": "compute001-bmc",
    "uid": "550e8400-e29b-41d4-a716-446655440000"
  },
  "spec": {
    "address": "https://10.1.1.100",
    "type": "Redfish",
    "authentication": {
      "method": "basic",
      "basic": {
        "username": "admin",
        "password": "secure-password"
      }
    }
  }
}
```

### Client Certificate Authentication

```json
{
  "apiVersion": "inventory/v2",
  "kind": "BMC",
  "schemaVersion": "v2beta1",
  "metadata": {
    "name": "compute002-bmc",
    "uid": "660e8400-e29b-41d4-a716-446655440001"
  },
  "spec": {
    "address": "https://10.1.1.101",
    "type": "Redfish",
    "authentication": {
      "method": "client-cert",
      "clientCert": {
        "certificateRef": "secret://inventory/bmc-cert",
        "keyRef": "secret://inventory/bmc-key",
        "caBundle": "-----BEGIN CERTIFICATE-----\n..."
      }
    }
  }
}
```

### OIDC Authentication

```json
{
  "apiVersion": "inventory/v2",
  "kind": "BMC",
  "schemaVersion": "v2beta1",
  "metadata": {
    "name": "compute003-bmc",
    "uid": "770e8400-e29b-41d4-a716-446655440002"
  },
  "spec": {
    "address": "https://10.1.1.102",
    "type": "Redfish",
    "authentication": {
      "method": "oidc",
      "oidc": {
        "issuerUrl": "https://auth.example.com",
        "clientId": "bmc-inventory-client",
        "clientSecret": "your-client-secret",
        "scopes": ["openid", "profile", "bmc.access"]
      }
    }
  }
}
```

## HTTP Request Headers

Clients can request the v2beta1 version using Accept headers:

```http
GET /apis/inventory/v2/bmcs/550e8400-e29b-41d4-a716-446655440000
Accept: application/json;version=v2beta1
```

Response:

```http
HTTP/1.1 200 OK
Content-Type: application/json;version=v2beta1

{
  "apiVersion": "inventory/v2",
  "kind": "BMC",
  "schemaVersion": "v2beta1",
  ...
}
```

## Version Conversion

### v1 → v2beta1

When converting from v1 to v2beta1:
- `username` and `password` fields are mapped to `authentication.basic`
- `authentication.method` is set to `"basic"`
- All other fields are preserved

### v2beta1 → v1

When converting from v2beta1 to v1:
- Only `basic` authentication can be converted
- `client-cert` and `oidc` authentication methods **cannot** be converted to v1
- Attempting to convert non-basic auth will return an error
- `authentication.basic.username` and `password` are mapped to top-level fields

### Conversion Examples

```go
import (
    "github.com/openchami/inventory/pkg/resources/bmc"
    bmcv2beta1 "github.com/openchami/inventory/pkg/resources/bmc/v2beta1"
)

// Convert v1 to v2beta1
converter := bmcv2beta1.NewBMCConverter()
v2Beta1BMC, err := converter.Convert(v1BMC, "v1", "v2beta1")

// Convert v2beta1 back to v1 (only works with basic auth)
v1BMC, err := converter.Convert(v2Beta1BMC, "v2beta1", "v1")
```

## Testing

Run the converter tests:

```bash
go test ./pkg/resources/bmc/v2beta1/... -v
```

## API Differences from v1

| Feature | v1 | v2beta1 |
|---------|----|---------|
| Basic Auth | ✅ `username`, `password` | ✅ `authentication.basic` |
| Client Cert | ❌ | ✅ `authentication.clientCert` |
| OIDC | ❌ | ✅ `authentication.oidc` |
| Auth Method Field | Implicit | Explicit `authentication.method` |
| Status Auth Info | ❌ | ✅ `status.authenticationMethod` |

## Migration Guide

### For API Consumers

If you're using v1 BMC resources:

1. **No immediate action required** - v1 continues to work
2. **To use new auth methods:**
   - Request v2beta1 via Accept headers
   - Use new authentication structure
3. **Backward compatible** - Basic auth works the same way

### For Operators

When upgrading BMCs to v2beta1:

```bash
# View available versions
curl -H "Accept: application/json" \
  https://inventory.example.com/apis/inventory/v2/version-info

# Create BMC with OIDC auth
curl -X POST \
  -H "Accept: application/json;version=v2beta1" \
  -H "Content-Type: application/json;version=v2beta1" \
  -d @bmc-oidc.json \
  https://inventory.example.com/apis/inventory/v2/bmcs
```

## Security Considerations

### Secrets Management

- **Basic Auth Passwords:** Store in secrets management system, never in Git
- **Client Certificates:** Reference via `secret://` URIs, not inline
- **OIDC Client Secrets:** Store securely, rotate regularly

### Best Practices

1. **Use Client Certificates** for machine-to-machine authentication
2. **Use OIDC** when integrating with enterprise identity providers
3. **Rotate credentials** regularly regardless of method
4. **Use TLS** for all BMC connections (`https://` addresses)
5. **Validate CA certificates** in client certificate authentication

## Future Plans (v2 Stable)

Before promoting to v2 stable, we plan to:

- [ ] Add support for session token caching
- [ ] Implement automatic certificate renewal
- [ ] Add OIDC token refresh
- [ ] Support for hardware security modules (HSM)
- [ ] Enhanced audit logging for authentication events

## Related Documentation

- **[Version Negotiation Guide](../../../../docs/user/VERSION-NEGOTIATION.md)** - Using multi-version schemas
- **[API Reference](../../../../docs/user/API-REFERENCE.md)** - REST API documentation
- **[User Guide](../../../../docs/user/USER-GUIDE.md)** - Complete usage guide
- [BMC v1](../bmc.go) - Original BMC resource definition
- [Version Registry](../../../versioning/) - Version negotiation internals

## Feedback

This is a **beta** release. Please provide feedback on:

- Authentication method usability
- API ergonomics
- Missing authentication methods
- Documentation clarity

**File issues:** https://github.com/OpenCHAMI/inventory/issues
