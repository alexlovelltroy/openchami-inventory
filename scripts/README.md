# Scripts Directory

> 📖 **Documentation**: [Testing Guide](../docs/developer/TESTING.md) | [Version Negotiation Guide](../docs/user/VERSION-NEGOTIATION.md)

Utility and demo scripts for the OpenCHAMI Inventory system.

## Quick Reference

| Script | Purpose | Usage |
|--------|---------|-------|
| `demo-version-negotiation.sh` | Version negotiation demo | Run after starting server |
| `populate-bmcs.sh` | Create sample data | Run to populate test data |
| `discover.sh` | Codebase explorer | Run with `make discover` |
| `templates.sh` | View templates | Run with `make templates` |

## demo-version-negotiation.sh

**Purpose:** Demonstrates multi-version schema support in the API.

**Quick Start:**
```bash
# Terminal 1: Start server
./bin/server --port 9999 --disable-auth

# Terminal 2: Run demo
./scripts/demo-version-negotiation.sh
```

**What It Demonstrates:**
1. Creating BMCs in v1 and v2beta1 formats
2. Version negotiation with CLI `--version` flag
3. Automatic schema conversion (v1 ↔ v2beta1)
4. Forward and backward compatibility
5. Version discovery endpoint

**Environment Variables:**
- `INVENTORY_SERVER` - Server URL (default: `http://localhost:9999`)
- `INVENTORY_CLI` - CLI path (default: `./bin/inventory-cli`)

**Learn More:**
- [Version Negotiation Guide](../docs/user/VERSION-NEGOTIATION.md)
- [BMC v2beta1 Documentation](../pkg/resources/bmc/v2beta1/README.md)

## populate-bmcs.sh

**Purpose:** Create 25 sample BMCs for testing.

**Quick Start:**
```bash
# With default settings
./scripts/populate-bmcs.sh

# With custom server
INVENTORY_SERVER=http://localhost:9999 ./scripts/populate-bmcs.sh
```

**Creates:**
- 25 BMCs: `bmc-001` through `bmc-025`
- IP addresses: `10.0.0.100` through `10.0.0.124`
- Types: iLO, iDRAC, Redfish (alternating)
- Organized in 5 racks with labels and annotations

**Environment Variables:**
- `INVENTORY_SERVER` - Server URL (default: `http://localhost:9999`)
- `INVENTORY_CLI` - CLI path (default: `./bin/inventory-cli`)

## populate-bmcs (Go Tool)

Standalone Go tool to populate the inventory using the client library directly.

### Building

```bash
make build
# or
go build -o bin/populate-bmcs ./cmd/populate-bmcs
```

### Usage

```bash
# With server running on default port (8080)
./bin/populate-bmcs

# With custom server URL
INVENTORY_SERVER=http://localhost:9999 ./bin/populate-bmcs
```

### Environment Variables

- `INVENTORY_SERVER` - Server URL (default: `http://localhost:9999`)

### Output

```
Populating inventory with 25 sample BMCs...
Server: http://localhost:9999

Creating bmc-001 (Redfish at 10.0.0.100)... ✓ (UID: abc123...)
Creating bmc-002 (iLO at 10.0.0.101)... ✓ (UID: def456...)
Creating bmc-003 (iDRAC at 10.0.0.102)... ✓ (UID: ghi789...)
...

Done! Created 25 BMCs (0 failed)

View them with:
  curl http://localhost:9999/bmcs | jq
  ./bin/inventory-cli --server http://localhost:9999 bmc list
```

## Verifying the Data

After running either script, you can verify the BMCs were created:

### Using curl

```bash
# List all BMCs
curl http://localhost:9999/bmcs | jq

# Get a specific BMC
curl http://localhost:9999/bmcs/<uid> | jq
```

### Using the CLI

```bash
# List all BMCs
./bin/inventory-cli bmc list

# Get JSON output
./bin/inventory-cli bmc list --output json

# Get a specific BMC
./bin/inventory-cli bmc get <uid>
```

### Using the GUI (if available)

Navigate to `http://localhost:9999` in your browser.

## Development

### Adding New Sample Data Scripts

To create similar scripts for other resources (Nodes, FRUs, etc.):

1. **Bash Version**:
   ```bash
   cp scripts/populate-bmcs.sh scripts/populate-nodes.sh
   # Edit to use appropriate resource type and fields
   ```

2. **Go Version**:
   ```bash
   cp -r cmd/populate-bmcs cmd/populate-nodes
   # Edit main.go to use Node client methods
   # Update Makefile build target
   ```

### Script Template

Both scripts follow the same pattern:
1. Parse environment variables for configuration
2. Connect to the server
3. Loop to create sample resources
4. Report success/failure for each
5. Print summary with verification commands

## discover.sh

**Purpose:** Interactive codebase explorer showing structure and status.

**Usage:**
```bash
make discover
# or
./scripts/discover.sh
```

**Shows:**
- Project structure
- Generated vs manual files
- Resource definitions
- Template locations
- Build status

## templates.sh

**Purpose:** View code generation template content.

**Usage:**
```bash
make templates
# or
./scripts/templates.sh
```

**Shows:**
- All template files
- Template purposes
- Quick reference

## Cleanup

**Remove created BMCs:**
```bash
# Using CLI
for uid in $(./bin/inventory-cli bmc list --output json | jq -r '.[].metadata.uid'); do
    ./bin/inventory-cli bmc delete "$uid"
done

# Or delete storage
rm -rf ./inventory/
```

## Documentation

- **[Testing Guide](../docs/developer/TESTING.md)** - Development and testing workflows
- **[Version Negotiation](../docs/user/VERSION-NEGOTIATION.md)** - Multi-version schema guide
- **[User Guide](../docs/user/USER-GUIDE.md)** - Using the system
