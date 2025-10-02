# CLI Reference

Complete reference for the `inventory-cli` command-line tool.

## Installation

```bash
# Build from source
make build

# Or build just the CLI
make generate-client-cmd
go build -o bin/inventory-cli ./cmd/inventory-cli
```

## Synopsis

```
inventory-cli [global flags] <resource> <command> [flags]
```

## Global Flags

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--server` | | `INVENTORY_SERVER` | `http://localhost:8080` | Inventory server URL |
| `--version` | `-v` | `INVENTORY_VERSION` | (server default) | API version to request |
| `--timeout` | | `INVENTORY_TIMEOUT` | `30s` | Request timeout |
| `--output` | `-o` | `INVENTORY_OUTPUT` | `table` | Output format: `table`, `json`, `yaml` |
| `--config` | | | `$HOME/.inventory-cli.yaml` | Config file path |
| `--help` | `-h` | | | Show help |

## Configuration

### Configuration File

Create `~/.inventory-cli.yaml`:

```yaml
server: http://localhost:9999
version: v1
timeout: 30s
output: table
```

### Environment Variables

```bash
export INVENTORY_SERVER=http://localhost:9999
export INVENTORY_VERSION=v2beta1
export INVENTORY_TIMEOUT=60s
export INVENTORY_OUTPUT=json
```

### Precedence

Configuration is loaded in this order (later sources override earlier ones):
1. Configuration file
2. Environment variables
3. Command-line flags

## Resources

The CLI supports CRUD operations for these resources:

- `bmc` - Baseboard Management Controllers
- `node` - Compute nodes
- `fru` - Field Replaceable Units
- `bootconfiguration` - Boot configurations

## Commands

### list

List all resources of a type.

**Usage:**
```bash
inventory-cli <resource> list [flags]
```

**Examples:**
```bash
# List all BMCs (table format)
inventory-cli bmc list

# List as JSON
inventory-cli bmc list --output json

# List with specific version
inventory-cli --version v2beta1 bmc list

# List nodes
inventory-cli node list
```

**Output:**
- `table` - Human-readable table
- `json` - JSON array
- `yaml` - YAML documents

### get

Get a specific resource by UID.

**Usage:**
```bash
inventory-cli <resource> get <uid> [flags]
```

**Arguments:**
- `uid` - Resource UID (required)

**Examples:**
```bash
# Get BMC by UID
inventory-cli bmc get 550e8400-e29b-41d4-a716-446655440000

# Get as JSON
inventory-cli bmc get 550e8400-e29b-41d4-a716-446655440000 --output json

# Get with specific version
inventory-cli --version v2beta1 bmc get 550e8400-e29b-41d4-a716-446655440000

# Get node
inventory-cli node get 660e8400-e29b-41d4-a716-446655440001
```

### create

Create a new resource.

**Usage:**
```bash
inventory-cli <resource> create [flags]
```

**Flags:**
- `--spec <json>` - Resource specification as JSON

**Input:**
- Reads from stdin if `--spec` not provided
- Expects JSON object with resource fields

**Examples:**
```bash
# Create from spec flag
inventory-cli bmc create --spec '{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}'

# Create from stdin
echo '{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}' | inventory-cli bmc create

# Create from file
cat bmc.json | inventory-cli bmc create

# Create with labels
inventory-cli bmc create --spec '{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish",
  "metadata": {
    "labels": {
      "datacenter": "dc1",
      "rack": "r01"
    }
  }
}'
```

### update

Update an existing resource.

**Usage:**
```bash
inventory-cli <resource> update <uid> [flags]
```

**Arguments:**
- `uid` - Resource UID (required)

**Flags:**
- `--spec <json>` - Fields to update as JSON

**Input:**
- Reads from stdin if `--spec` not provided
- Expects JSON object with fields to update
- Only specified fields are updated (partial update)

**Examples:**
```bash
# Update from spec flag
inventory-cli bmc update 550e8400-e29b-41d4-a716-446655440000 --spec '{
  "password": "new-secure-password"
}'

# Update from stdin
echo '{
  "address": "https://10.1.1.101"
}' | inventory-cli bmc update 550e8400-e29b-41d4-a716-446655440000

# Update labels
inventory-cli bmc update 550e8400-e29b-41d4-a716-446655440000 --spec '{
  "metadata": {
    "labels": {
      "status": "maintenance"
    }
  }
}'

# Update node role
inventory-cli node update <uid> --spec '{
  "role": "login"
}'
```

### delete

Delete a resource.

**Usage:**
```bash
inventory-cli <resource> delete <uid>
```

**Arguments:**
- `uid` - Resource UID (required)

**Examples:**
```bash
# Delete BMC
inventory-cli bmc delete 550e8400-e29b-41d4-a716-446655440000

# Delete node
inventory-cli node delete 660e8400-e29b-41d4-a716-446655440001

# Delete with confirmation prompt (not yet implemented)
# inventory-cli bmc delete <uid> --confirm
```

## Version Negotiation

The `--version` flag controls which API version to request.

### Available Versions

- `v1` - Stable version with basic features
- `v2beta1` - Beta version with enhanced features

### Version Examples

#### List with Different Versions

```bash
# Default version (v1)
inventory-cli bmc list

# Request v2beta1
inventory-cli --version v2beta1 bmc list

# Short form
inventory-cli -v v2beta1 bmc list
```

#### Get with Version Comparison

```bash
# Get as v1
inventory-cli -v v1 bmc get <uid> --output json > v1.json

# Get as v2beta1
inventory-cli -v v2beta1 bmc get <uid> --output json > v2beta1.json

# Compare
diff -u v1.json v2beta1.json
```

#### Authentication Format Differences

**v1 BMC (flat auth structure):**
```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
```

**v2beta1 BMC (structured auth):**
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

### Version Discovery

Check supported versions:

```bash
curl $(inventory-cli config get server)/version-info | jq
```

Or query server directly:
```bash
curl http://localhost:9999/version-info | jq '.supportedResourceVersions'
```

## Resource Specifications

### BMC (Baseboard Management Controller)

**v1 Fields:**
```json
{
  "name": "string",
  "address": "string (URL)",
  "username": "string",
  "password": "string",
  "type": "string (Redfish, iLO, iDRAC)",
  "metadata": {
    "labels": {"key": "value"},
    "annotations": {"key": "value"}
  }
}
```

**v2beta1 Fields:**
```json
{
  "name": "string",
  "address": "string (URL)",
  "type": "string (Redfish, iLO, iDRAC)",
  "authentication": {
    "method": "basic|client-cert|oidc",
    "basic": {
      "username": "string",
      "password": "string"
    },
    "clientCert": {
      "certificateRef": "string",
      "keyRef": "string",
      "caBundle": "string"
    },
    "oidc": {
      "issuerUrl": "string",
      "clientId": "string",
      "clientSecret": "string"
    }
  },
  "metadata": {
    "labels": {"key": "value"},
    "annotations": {"key": "value"}
  }
}
```

### Node

```json
{
  "name": "string",
  "xname": "string (HSN identifier)",
  "nid": "integer (Node ID)",
  "role": "string (compute, login, service)",
  "metadata": {
    "labels": {"key": "value"},
    "annotations": {"key": "value"}
  },
  "status": {
    "conditions": [
      {
        "type": "string",
        "status": "True|False|Unknown",
        "reason": "string",
        "message": "string"
      }
    ]
  }
}
```

### FRU (Field Replaceable Unit)

```json
{
  "name": "string",
  "type": "string (chassis, blade, psu, fan)",
  "manufacturer": "string",
  "model": "string",
  "serialNumber": "string",
  "partNumber": "string",
  "metadata": {
    "labels": {"key": "value"},
    "annotations": {"key": "value"}
  }
}
```

### Boot Configuration

```json
{
  "name": "string",
  "kernel": "string (kernel image path)",
  "initrd": "string (initrd image path)",
  "cmdline": "string (kernel command line)",
  "metadata": {
    "labels": {"key": "value"},
    "annotations": {"key": "value"}
  }
}
```

## Scripting Examples

### Bulk Operations

```bash
#!/bin/bash
# Create multiple BMCs

for i in {1..10}; do
  inventory-cli bmc create --spec "{
    \"name\": \"node$(printf %03d $i)-bmc\",
    \"address\": \"https://10.1.1.$((100+i))\",
    \"username\": \"admin\",
    \"password\": \"changeme\",
    \"type\": \"Redfish\"
  }"
done
```

### Query and Filter

```bash
#!/bin/bash
# Find BMCs in specific datacenter

inventory-cli bmc list --output json | \
  jq '.[] | select(.metadata.labels.datacenter == "dc1") | .metadata.name'
```

### Export/Import

```bash
#!/bin/bash
# Export all BMCs

inventory-cli bmc list --output json > bmcs-backup.json

# Import BMCs (create each one)
jq -c '.[]' bmcs-backup.json | while read bmc; do
  echo "$bmc" | inventory-cli bmc create
done
```

### Status Monitoring

```bash
#!/bin/bash
# Monitor node conditions

watch -n 5 'inventory-cli node list --output json | \
  jq -r ".[] | select(.status.conditions[]?.type == \"Maintenance\") | 
  \"\(.metadata.name): \(.status.conditions[0].message)\""'
```

## Tips and Tricks

### Use jq for JSON Processing

```bash
# Extract specific fields
inventory-cli bmc list -o json | jq '.[] | {name: .metadata.name, address: .spec.address}'

# Count resources
inventory-cli bmc list -o json | jq 'length'

# Group by label
inventory-cli bmc list -o json | jq 'group_by(.metadata.labels.datacenter)'
```

### Use Environment Variables

```bash
# Set once, use everywhere
export INVENTORY_SERVER=http://localhost:9999
export INVENTORY_OUTPUT=json

inventory-cli bmc list
inventory-cli node list
```

### Create Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias inv='inventory-cli --server http://localhost:9999'
alias invj='inventory-cli --server http://localhost:9999 --output json'

# Use aliases
inv bmc list
invj bmc get <uid> | jq .spec
```

### Debug API Calls

```bash
# Use curl with same headers
curl -v \
  -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs
```

## Troubleshooting

### CLI not found

**Check installation:**
```bash
which inventory-cli
ls -la bin/inventory-cli
```

**Build if missing:**
```bash
make build
```

### Connection refused

**Check server URL:**
```bash
curl http://localhost:9999/health
```

**Verify configuration:**
```bash
inventory-cli config get server
```

### 401 Unauthorized

**Server requires authentication.** Either:
1. Restart server with `--disable-auth` (testing only)
2. Configure authentication (production)

### Version not supported

**Check available versions:**
```bash
curl http://localhost:9999/version-info | jq
```

**Use supported version:**
```bash
inventory-cli --version v1 bmc list
```

### Invalid JSON spec

**Validate JSON:**
```bash
echo '{"name": "test"}' | jq .
```

**Use proper quoting:**
```bash
# Good
inventory-cli bmc create --spec '{"name": "test"}'

# Bad (shell interprets as multiple args)
inventory-cli bmc create --spec {"name": "test"}
```

## See Also

- [User Guide](./USER-GUIDE.md) - Complete usage guide
- [API Reference](./API-REFERENCE.md) - REST API documentation
- [Version Negotiation](./VERSION-NEGOTIATION.md) - Multi-version guide
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues

## Getting Help

```bash
# General help
inventory-cli --help

# Resource help
inventory-cli bmc --help

# Command help
inventory-cli bmc create --help
```
