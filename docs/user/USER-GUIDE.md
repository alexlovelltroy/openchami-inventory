# OpenCHAMI Inventory User Guide

Complete guide for installing, configuring, and using the OpenCHAMI Inventory system.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Server Configuration](#server-configuration)
- [Using the CLI](#using-the-cli)
- [Using the REST API](#using-the-rest-api)
- [Authentication](#authentication)
- [Version Negotiation](#version-negotiation)
- [Common Workflows](#common-workflows)
- [Troubleshooting](#troubleshooting)

## Installation

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/openchami/inventory/releases):

```bash
# Download and extract
curl -L https://github.com/openchami/inventory/releases/latest/download/inventory-linux-amd64.tar.gz | tar xz

# Move binaries to PATH
sudo mv bin/* /usr/local/bin/
```

### From Source

Requirements:
- Go 1.24 or later
- Make

```bash
# Clone repository
git clone https://github.com/openchami/inventory.git
cd inventory

# Build everything
make build

# Binaries are in ./bin/
ls -la bin/
# bin/server
# bin/inventory-cli
# bin/populate-bmcs
```

## Quick Start

### 1. Start the Server

For testing and development, start with authentication disabled:

```bash
./bin/server --port 9999 --disable-auth
```

The server will:
- Listen on `http://localhost:9999`
- Use file-based storage in `./inventory/`
- Allow all operations without authentication
- Serve OpenAPI documentation at `/openapi.json`

### 2. Populate Sample Data

Create sample BMC resources:

```bash
./bin/populate-bmcs
```

This creates 25 sample BMCs with realistic data.

### 3. Query Resources

Using the CLI:

```bash
# List all BMCs
./bin/inventory-cli --server http://localhost:9999 bmc list

# Get specific BMC
./bin/inventory-cli --server http://localhost:9999 bmc get <uid>
```

Using curl:

```bash
# List all BMCs
curl http://localhost:9999/bmcs | jq

# Get specific BMC
curl http://localhost:9999/bmcs/<uid> | jq
```

## Server Configuration

### Command-Line Flags

```bash
./bin/server \
  --host 0.0.0.0 \
  --port 9999 \
  --disable-auth \
  --storage-path ./inventory \
  --log-level info
```

Available flags:
- `--host` - Host to bind to (default: `localhost`)
- `--port` - Port to listen on (default: `8080`)
- `--disable-auth` - Disable authentication (testing only!)
- `--storage-path` - Directory for file storage (default: `./inventory`)
- `--log-level` - Logging level: `debug`, `info`, `warn`, `error` (default: `info`)
- `--config` - Path to config file

### Environment Variables

All flags can be set via environment variables with `INVENTORY_` prefix:

```bash
export INVENTORY_HOST=0.0.0.0
export INVENTORY_PORT=9999
export INVENTORY_DISABLE_AUTH=true
export INVENTORY_STORAGE_PATH=/var/lib/inventory
export INVENTORY_LOG_LEVEL=debug

./bin/server
```

### Configuration File

Create `server.yaml`:

```yaml
host: 0.0.0.0
port: 9999
log-level: info

security:
  disable-auth: false  # Set to true only for testing!

storage:
  backend: file
  path: /var/lib/inventory

cors:
  enabled: true
  origins:
    - "https://dashboard.example.com"

openapi:
  disabled: false
  validate: false
```

Load with:
```bash
./bin/server --config server.yaml
```

## Using the CLI

The `inventory-cli` tool provides a user-friendly interface for managing inventory resources.

### Global Configuration

Set common options via environment variables:

```bash
export INVENTORY_SERVER=http://localhost:9999
export INVENTORY_OUTPUT=json
export INVENTORY_VERSION=v1
```

Or create `~/.inventory-cli.yaml`:

```yaml
server: http://localhost:9999
timeout: 30s
output: table
version: v1
```

### Resource Commands

All resources support the same operations:

```bash
# List all resources
./bin/inventory-cli <resource> list

# Get specific resource
./bin/inventory-cli <resource> get <uid>

# Create resource
./bin/inventory-cli <resource> create --spec '<json>'

# Update resource
./bin/inventory-cli <resource> update <uid> --spec '<json>'

# Delete resource
./bin/inventory-cli <resource> delete <uid>
```

Available resources:
- `bmc` - Baseboard Management Controllers
- `node` - Compute nodes
- `fru` - Field Replaceable Units
- `bootconfiguration` - Boot configurations

### Output Formats

```bash
# Table format (default)
./bin/inventory-cli bmc list

# JSON format
./bin/inventory-cli bmc list --output json

# YAML format
./bin/inventory-cli bmc list --output yaml
```

### Examples

#### List BMCs in a datacenter

```bash
./bin/inventory-cli bmc list --output json | \
  jq '.[] | select(.metadata.labels.datacenter == "dc1")'
```

#### Create a BMC

```bash
./bin/inventory-cli bmc create --spec '{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}'
```

#### Update a BMC

```bash
./bin/inventory-cli bmc update <uid> --spec '{
  "password": "new-secure-password"
}'
```

#### Delete a BMC

```bash
./bin/inventory-cli bmc delete <uid>
```

For complete CLI reference, see [CLI-REFERENCE.md](./CLI-REFERENCE.md).

## Using the REST API

### Endpoints

All resources follow RESTful conventions:

```
GET    /<resources>           List all resources
GET    /<resources>/<uid>     Get specific resource
POST   /<resources>           Create new resource
PUT    /<resources>/<uid>     Update resource
DELETE /<resources>/<uid>     Delete resource
```

### Examples

#### List BMCs

```bash
curl -X GET http://localhost:9999/bmcs | jq
```

#### Get specific BMC

```bash
curl -X GET http://localhost:9999/bmcs/<uid> | jq
```

#### Create BMC

```bash
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "node001-bmc",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "changeme",
    "type": "Redfish"
  }' | jq
```

#### Update BMC

```bash
curl -X PUT http://localhost:9999/bmcs/<uid> \
  -H "Content-Type: application/json" \
  -d '{
    "password": "new-password"
  }' | jq
```

#### Delete BMC

```bash
curl -X DELETE http://localhost:9999/bmcs/<uid>
```

For complete API reference, see [API-REFERENCE.md](./API-REFERENCE.md).

## Authentication

### Testing Mode (No Authentication)

**Warning: For development/testing only!**

Start server with `--disable-auth`:

```bash
./bin/server --port 9999 --disable-auth
```

This uses a permissive policy that allows all operations without authentication.

### Production Mode (Authentication Required)

Start server without `--disable-auth`:

```bash
./bin/server --port 9999
```

This requires JWT authentication for all operations. You'll need to:

1. Set up tokensmith middleware
2. Provide JWT tokens in requests
3. Configure authorization policies

For details, see [AUTHENTICATION.md](./AUTHENTICATION.md).

## Version Negotiation

The inventory supports multiple schema versions for each resource type.

### Requesting Specific Versions

Using CLI:

```bash
# Get as v1 (default)
./bin/inventory-cli bmc get <uid>

# Get as v2beta1
./bin/inventory-cli --version v2beta1 bmc get <uid>
```

Using curl:

```bash
# Request v1
curl -H "Accept: application/json;version=v1" \
  http://localhost:9999/bmcs/<uid>

# Request v2beta1
curl -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid>
```

### Version Conversion

The server automatically converts between versions:
- **v1 → v2beta1**: Converts `username`/`password` to `authentication.basic`
- **v2beta1 → v1**: Extracts `username`/`password` from `authentication.basic` (may lose data for other auth methods)

### Discovering Versions

Query available versions:

```bash
curl http://localhost:9999/version-info | jq
```

Response:
```json
{
  "apiVersion": "inventory/v2",
  "defaultSchemaVersion": "v1",
  "supportedResourceVersions": {
    "BMC": ["v1", "v2beta1"],
    "Node": ["v1"]
  }
}
```

For complete guide, see [VERSION-NEGOTIATION.md](./VERSION-NEGOTIATION.md).

## Common Workflows

### Workflow 1: Initial Setup

```bash
# 1. Start server
./bin/server --port 9999 --disable-auth

# 2. Populate with sample data
./bin/populate-bmcs

# 3. Verify data
./bin/inventory-cli --server http://localhost:9999 bmc list
```

### Workflow 2: Add New Hardware

```bash
# Create BMC for new node
./bin/inventory-cli bmc create --spec '{
  "name": "node042-bmc",
  "address": "https://10.1.1.142",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish",
  "metadata": {
    "labels": {
      "datacenter": "dc1",
      "rack": "r01",
      "position": "u42"
    }
  }
}'

# Create node resource
./bin/inventory-cli node create --spec '{
  "name": "node042",
  "xname": "x1000c0s42b0n0",
  "nid": 42,
  "role": "compute"
}'
```

### Workflow 3: Hardware Lifecycle

```bash
# Mark node for maintenance
./bin/inventory-cli node update <uid> --spec '{
  "status": {
    "conditions": [
      {
        "type": "Maintenance",
        "status": "True",
        "reason": "ScheduledMaintenance",
        "message": "Node offline for memory upgrade"
      }
    ]
  }
}'

# After maintenance, update status
./bin/inventory-cli node update <uid> --spec '{
  "status": {
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "reason": "MaintenanceComplete",
        "message": "Memory upgraded to 512GB"
      }
    ]
  }
}'
```

### Workflow 4: Query and Filter

```bash
# Find all BMCs in datacenter dc1
./bin/inventory-cli bmc list --output json | \
  jq '.[] | select(.metadata.labels.datacenter == "dc1")'

# Count BMCs by type
./bin/inventory-cli bmc list --output json | \
  jq 'group_by(.spec.type) | map({type: .[0].spec.type, count: length})'

# Find nodes in maintenance
./bin/inventory-cli node list --output json | \
  jq '.[] | select(.status.conditions[]? | .type == "Maintenance" and .status == "True")'
```

## Troubleshooting

### Server won't start

**Check port availability:**
```bash
lsof -i :9999
```

**Check storage directory:**
```bash
mkdir -p ./inventory
chmod 755 ./inventory
```

### 401 Unauthorized errors

**Solution:** Use `--disable-auth` for testing:
```bash
./bin/server --port 9999 --disable-auth
```

### CLI connection refused

**Check server is running:**
```bash
curl http://localhost:9999/health
```

**Verify server URL:**
```bash
./bin/inventory-cli --server http://localhost:9999 bmc list
```

### Resource not found (404)

**List all resources to find UID:**
```bash
./bin/inventory-cli bmc list --output json | jq '.[] | .metadata.uid'
```

### Version not supported error

**Check available versions:**
```bash
curl http://localhost:9999/version-info | jq '.supportedResourceVersions'
```

For more troubleshooting help, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

## Next Steps

- **[CLI Reference](./CLI-REFERENCE.md)** - Complete CLI command reference
- **[API Reference](./API-REFERENCE.md)** - REST API endpoint details
- **[Version Negotiation](./VERSION-NEGOTIATION.md)** - Multi-version schema guide
- **[Authentication](./AUTHENTICATION.md)** - Security and policies
- **[Development Guide](../developer/DEVELOPMENT.md)** - Extend the system

## Getting Help

- **Documentation**: Check `docs/` directory
- **Issues**: [GitHub Issues](https://github.com/openchami/inventory/issues)
- **Discussions**: [GitHub Discussions](https://github.com/openchami/inventory/discussions)
