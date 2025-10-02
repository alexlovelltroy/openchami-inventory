# Inventory CLI

A command-line interface for managing OpenCHAMI inventory resources.

> 📖 **Complete Documentation**: [CLI Reference Guide](../../docs/user/CLI-REFERENCE.md)

## Quick Start

### Installation

```bash
# Build everything
make dev

# Or build just the CLI
make generate-client-cmd
go build -o bin/inventory-cli ./cmd/inventory-cli
```

### Basic Usage

```bash
# List all BMCs
./bin/inventory-cli --server http://localhost:9999 bmc list

# Get specific BMC
./bin/inventory-cli --server http://localhost:9999 bmc get <uid>

# Create BMC
./bin/inventory-cli --server http://localhost:9999 bmc create --spec '{...}'
```

### Quick Reference

**Available Resources:**
- `bmc` - Baseboard Management Controllers
- `node` - Compute nodes
- `fru` - Field Replaceable Units
- `bootconfiguration` - Boot configurations

**Common Commands:**
- `list` - List all resources
- `get <uid>` - Get specific resource
- `create` - Create new resource
- `update <uid>` - Update resource
- `delete <uid>` - Delete resource

**Global Flags:**
- `--server` - Server URL (default: `http://localhost:8080`)
- `--version, -v` - API version (`v1`, `v2beta1`)
- `--output, -o` - Format (`table`, `json`, `yaml`)
- `--timeout` - Request timeout (default: `30s`)

**Environment Variables:**

- `INVENTORY_SERVER` - Server URL
- `INVENTORY_VERSION` - API version
- `INVENTORY_TIMEOUT` - Timeout
- `INVENTORY_OUTPUT` - Output format

## Configuration File

Create `~/.inventory-cli.yaml`:

```yaml
server: http://localhost:9999
version: v1
timeout: 30s
output: table
```

## Examples

### Basic Operations

```bash
# List BMCs
./bin/inventory-cli bmc list

# Get specific BMC
./bin/inventory-cli bmc get <uid>

# Create BMC
./bin/inventory-cli bmc create --spec '{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}'

# Update BMC
./bin/inventory-cli bmc update <uid> --spec '{"password": "new-password"}'

# Delete BMC
./bin/inventory-cli bmc delete <uid>
```

### Version Negotiation

```bash
# Request v1 format
./bin/inventory-cli --version v1 bmc get <uid>

# Request v2beta1 format
./bin/inventory-cli --version v2beta1 bmc get <uid>

# Use environment variable
export INVENTORY_VERSION=v2beta1
./bin/inventory-cli bmc list
```

### Output Formats

```bash
# JSON output
./bin/inventory-cli bmc list --output json

# YAML output
./bin/inventory-cli bmc list --output yaml

# Table output (default)
./bin/inventory-cli bmc list
```

## Help

```bash
# General help
./bin/inventory-cli --help

# Command-specific help
./bin/inventory-cli bmc --help
./bin/inventory-cli bmc create --help
```

## Documentation

- **[Complete CLI Reference](../../docs/user/CLI-REFERENCE.md)** - Full command documentation
- **[User Guide](../../docs/user/USER-GUIDE.md)** - Usage guide and workflows
- **[Version Negotiation](../../docs/user/VERSION-NEGOTIATION.md)** - Multi-version support
- **[Troubleshooting](../../docs/user/TROUBLESHOOTING.md)** - Common issues

## Development

This CLI is automatically generated from resource definitions.

**Regenerate:**
```bash
make generate-client-cmd
```

**Template:**
`pkg/codegen/templates/client-cmd.go.tmpl`

**Learn More:**
- [Code Generation Guide](../../docs/developer/CODE-GENERATION.md)
