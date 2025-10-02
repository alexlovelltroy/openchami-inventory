# Version Negotiation Guide

Complete guide to multi-version schema support in OpenCHAMI Inventory.

## Table of Contents

- [Overview](#overview)
- [Supported Versions](#supported-versions)
- [How It Works](#how-it-works)
- [Using the CLI](#using-the-cli)
- [Using the API](#using-the-api)
- [Version Conversion](#version-conversion)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

OpenCHAMI Inventory supports multiple schema versions for each resource type, allowing:

- **Backward compatibility**: Old clients work with new servers
- **Forward compatibility**: New features without breaking old clients
- **Gradual migration**: Update clients at your own pace
- **Schema evolution**: Add features to new versions

### Key Concepts

- **API Version**: Overall API version (e.g., `inventory/v2`)
- **Schema Version**: Resource-specific version (e.g., `v1`, `v2beta1`)
- **Default Version**: Version used when none specified (typically `v1`)
- **Version Negotiation**: Client requests specific version via headers
- **Version Conversion**: Server automatically converts between versions

## Supported Versions

Query the server for supported versions:

```bash
curl http://localhost:9999/version-info | jq
```

**Response:**
```json
{
  "apiVersion": "inventory/v2",
  "defaultSchemaVersion": "v1",
  "supportedResourceVersions": {
    "BMC": ["v1", "v2beta1"],
    "Node": ["v1"],
    "FRU": ["v1"],
    "BootConfiguration": ["v1"]
  }
}
```

### BMC Versions

#### v1 (Stable)
- Simple username/password authentication
- Flat structure
- Fully stable and supported

#### v2beta1 (Beta)
- Structured authentication object
- Multiple auth methods: basic, client-cert, OIDC
- Enhanced features for modern Redfish
- Beta stability - may change before v2 stable

### Other Resources

Currently, Node, FRU, and BootConfiguration only have v1 versions.

## How It Works

### Request Flow

```
Client Request
  ├─ Accept: application/json;version=v2beta1
  └─ Content-Type: application/json;version=v2beta1
         ↓
    Version Registry
  ├─ Check if version supported
  ├─ Get resource in storage version
  └─ Convert to requested version
         ↓
    Response
  ├─ X-Schema-Version: v2beta1
  └─ Body: Resource in v2beta1 format
```

### Storage

Resources are stored in their **native** version:
- Created as v1 → Stored as v1
- Created as v2beta1 → Stored as v2beta1
- Converted on read as needed

### Conversion

The server performs automatic conversion:
- **v1 → v2beta1**: Converts flat auth to structured format
- **v2beta1 → v1**: Extracts basic auth (lossy for other auth methods)

## Using the CLI

### Global Version Flag

```bash
# Use default version (v1)
./bin/inventory-cli bmc list

# Request v2beta1
./bin/inventory-cli --version v2beta1 bmc list

# Short form
./bin/inventory-cli -v v2beta1 bmc list
```

### Environment Variable

```bash
export INVENTORY_VERSION=v2beta1

# All commands use v2beta1
./bin/inventory-cli bmc list
./bin/inventory-cli bmc get <uid>
```

### Configuration File

Create `~/.inventory-cli.yaml`:

```yaml
server: http://localhost:9999
version: v2beta1
output: json
```

### Get Resource in Different Versions

```bash
# Get as v1
./bin/inventory-cli -v v1 bmc get <uid> --output json

# Get as v2beta1
./bin/inventory-cli -v v2beta1 bmc get <uid> --output json
```

### Comparison Example

```bash
# Save both versions
./bin/inventory-cli -v v1 bmc get <uid> -o json > v1.json
./bin/inventory-cli -v v2beta1 bmc get <uid> -o json > v2beta1.json

# Compare authentication structures
echo "v1 auth:"
jq '.spec | {username, password, type}' v1.json

echo "v2beta1 auth:"
jq '.spec.authentication' v2beta1.json
```

## Using the API

### Request Headers

Specify version in `Accept` and `Content-Type` headers:

```bash
# Request v1
curl -H "Accept: application/json;version=v1" \
  http://localhost:9999/bmcs/<uid>

# Request v2beta1
curl -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid>
```

### Create with Version

```bash
# Create v1 BMC
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json;version=v1" \
  -H "Accept: application/json;version=v1" \
  -d '{
    "name": "node001-bmc",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "changeme",
    "type": "Redfish"
  }'

# Create v2beta1 BMC
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json;version=v2beta1" \
  -H "Accept: application/json;version=v2beta1" \
  -d '{
    "name": "node002-bmc",
    "address": "https://10.1.1.101",
    "type": "Redfish",
    "authentication": {
      "method": "oidc",
      "oidc": {
        "issuerUrl": "https://auth.example.com",
        "clientId": "bmc-client",
        "clientSecret": "secret123"
      }
    }
  }'
```

### Response Headers

Check which version was returned:

```bash
curl -i -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid>

# Look for:
# X-Schema-Version: v2beta1
```

## Version Conversion

### v1 → v2beta1 (Lossless)

**v1 Input:**
```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
```

**v2beta1 Output:**
```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "type": "Redfish",
  "authentication": {
    "method": "basic",
    "basic": {
      "username": "admin",
      "password": "changeme"
    }
  }
}
```

**Conversion:** `username` and `password` become `authentication.basic`

### v2beta1 → v1 (Potentially Lossy)

#### Basic Auth (Lossless)

**v2beta1 Input:**
```json
{
  "authentication": {
    "method": "basic",
    "basic": {
      "username": "admin",
      "password": "changeme"
    }
  }
}
```

**v1 Output:**
```json
{
  "username": "admin",
  "password": "changeme"
}
```

**Conversion:** `authentication.basic` becomes flat `username`/`password`

#### OIDC Auth (Lossy)

**v2beta1 Input:**
```json
{
  "authentication": {
    "method": "oidc",
    "oidc": {
      "issuerUrl": "https://auth.example.com",
      "clientId": "bmc-client"
    }
  }
}
```

**v1 Output:**
```json
{
  "username": "",
  "password": ""
}
```

**Conversion:** OIDC auth cannot be represented in v1 (data loss)

## Best Practices

### 1. Use Stable Versions in Production

```bash
# ✅ Good: Use v1 (stable)
./bin/inventory-cli bmc list

# ⚠️ Caution: v2beta1 may change
./bin/inventory-cli -v v2beta1 bmc list
```

### 2. Always Specify Version for New Features

```bash
# When using v2beta1-specific features
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json;version=v2beta1" \
  -d '{"authentication": {"method": "oidc", ...}}'
```

### 3. Check Schema Version in Responses

```bash
# Verify you got the version you requested
curl -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid> | jq '.schemaVersion'
```

### 4. Test Conversion Paths

```bash
# Create as v1
echo '{"username": "admin", "password": "secret"}' | \
  ./bin/inventory-cli bmc create

BMC_UID=$(./bin/inventory-cli bmc list -o json | jq -r '.[0].metadata.uid')

# Read as v2beta1 to verify conversion
./bin/inventory-cli -v v2beta1 bmc get "$BMC_UID" -o json | \
  jq '.spec.authentication.method'
# Should output: "basic"
```

### 5. Document Version Requirements

When distributing code, specify required version:

```go
// Requires BMC v2beta1 for OIDC support
client := inventory.NewClient("http://localhost:9999").WithVersion("v2beta1")
```

## Examples

### Example 1: Create and Retrieve in Multiple Versions

```bash
#!/bin/bash
# demo-version-conversion.sh

SERVER="http://localhost:9999"

# Create v1 BMC
echo "Creating v1 BMC..."
RESPONSE=$(curl -s -X POST "$SERVER/bmcs" \
  -H "Content-Type: application/json;version=v1" \
  -d '{
    "name": "test-v1-bmc",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "secret",
    "type": "Redfish"
  }')

BMC_UID=$(echo "$RESPONSE" | jq -r '.metadata.uid')
echo "Created BMC: $BMC_UID"

# Read as v1
echo -e "\n=== As v1 ==="
curl -s -H "Accept: application/json;version=v1" \
  "$SERVER/bmcs/$BMC_UID" | jq '.spec | {username, password}'

# Read as v2beta1
echo -e "\n=== As v2beta1 ==="
curl -s -H "Accept: application/json;version=v2beta1" \
  "$SERVER/bmcs/$BMC_UID" | jq '.spec.authentication'
```

### Example 2: Bulk Version Comparison

```bash
#!/bin/bash
# compare-all-bmcs.sh

SERVER="http://localhost:9999"

# Get all BMC UIDs
UIDS=$(curl -s "$SERVER/bmcs" | jq -r '.[] | .metadata.uid')

for uid in $UIDS; do
  echo "=== BMC: $uid ==="
  
  # Get v1 auth
  V1_AUTH=$(curl -s -H "Accept: application/json;version=v1" \
    "$SERVER/bmcs/$uid" | jq -c '.spec | {username, password}')
  echo "v1: $V1_AUTH"
  
  # Get v2beta1 auth
  V2_AUTH=$(curl -s -H "Accept: application/json;version=v2beta1" \
    "$SERVER/bmcs/$uid" | jq -c '.spec.authentication')
  echo "v2beta1: $V2_AUTH"
  
  echo
done
```

### Example 3: Using CLI for Version Testing

```bash
#!/bin/bash
# test-version-negotiation.sh

# Create v1 BMC
echo "Creating v1 BMC with basic auth..."
./bin/inventory-cli bmc create --spec '{
  "name": "test-basic",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "secret",
  "type": "Redfish"
}'

# Get UID
BMC_UID=$(./bin/inventory-cli bmc list -o json | \
  jq -r '.[] | select(.metadata.name == "test-basic") | .metadata.uid')

# Test version negotiation
echo -e "\n=== Reading as v1 ==="
./bin/inventory-cli -v v1 bmc get "$BMC_UID" -o json | \
  jq '{name: .metadata.name, username: .spec.username}'

echo -e "\n=== Reading as v2beta1 ==="
./bin/inventory-cli -v v2beta1 bmc get "$BMC_UID" -o json | \
  jq '{name: .metadata.name, authMethod: .spec.authentication.method}'
```

### Example 4: Version Discovery

```bash
#!/bin/bash
# discover-versions.sh

SERVER="http://localhost:9999"

echo "=== API Version Info ==="
curl -s "$SERVER/version-info" | jq '.'

echo -e "\n=== Supported BMC Versions ==="
curl -s "$SERVER/version-info" | jq -r '.supportedResourceVersions.BMC[]'

echo -e "\n=== Default Schema Version ==="
curl -s "$SERVER/version-info" | jq -r '.defaultSchemaVersion'
```

## Troubleshooting

### Version Not Supported Error

**Symptom:**
```
Error: Version 'v3' is not supported for BMC
```

**Solution:**
```bash
# Check supported versions
curl http://localhost:9999/version-info | jq '.supportedResourceVersions'

# Use supported version
./bin/inventory-cli --version v1 bmc list
```

### Unexpected Conversion Results

**Symptom:** Created as v2beta1 with OIDC, but v1 shows empty username.

**Explanation:** This is expected - OIDC cannot be represented in v1 format.

**Solution:** Always use v2beta1 or later for OIDC authentication.

### CLI Version Flag Not Working

**Symptom:** `--version` flag not recognized.

**Solution:**
```bash
# Rebuild CLI
make build

# Verify flag exists
./bin/inventory-cli --help | grep version
```

## Demo Script

Run the comprehensive version negotiation demo:

```bash
# Start server
./bin/server --port 9999 --disable-auth

# Run demo
./scripts/demo-version-negotiation.sh
```

This demonstrates:
- Creating BMCs in different versions
- Retrieving resources with version negotiation
- Automatic schema conversion
- Version discovery

## See Also

- [User Guide](./USER-GUIDE.md) - Complete usage guide
- [CLI Reference](./CLI-REFERENCE.md) - CLI documentation
- [API Reference](./API-REFERENCE.md) - REST API documentation
- [BMC v2beta1 Documentation](../../pkg/resources/bmc/v2beta1/README.md) - v2beta1 features
