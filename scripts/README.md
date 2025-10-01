# Scripts Directory

Utility scripts for the OpenCHAMI Inventory system.

## demo-version-negotiation.sh

Demonstrates the multi-version schema support in the OpenCHAMI Inventory API.

### What It Does

This interactive demo script:
1. Creates a BMC using v1 schema (basic authentication)
2. Creates a BMC using v2beta1 schema (OIDC authentication)
3. Retrieves both BMCs in v1 format (backward compatibility)
4. Retrieves both BMCs in v2beta1 format (forward compatibility)
5. Shows automatic schema conversion between versions
6. Displays API version information

### Usage

```bash
# Start the server with auth disabled for testing
./bin/server --port 9999 --disable-auth

# In another terminal, run the demo
./scripts/demo-version-negotiation.sh
```

### Environment Variables

- `INVENTORY_SERVER` - Server URL (default: `http://localhost:9999`)
- `INVENTORY_CLI` - Path to CLI binary (default: `./bin/inventory-cli`)

### Example Output

The script provides step-by-step output showing:
- Creation of v1 BMC with basic auth (username/password)
- Creation of v2beta1 BMC with OIDC auth (client credentials)
- Listing all BMCs with their schema versions
- Retrieving v1 BMC as v2beta1 (shows auth conversion)
- Retrieving v2beta1 BMC as v1 (backward compatibility)
- API version capabilities and supported versions

### Key Concepts Demonstrated

- **Version Negotiation**: Using `Accept: application/json;version=v2beta1` headers
- **Forward Compatibility**: v1 resources can be read as v2beta1
- **Backward Compatibility**: v2beta1 resources can be read as v1
- **Schema Evolution**: New authentication methods in v2beta1
- **Coexistence**: Multiple schema versions in the same database

## populate-bmcs.sh

Bash script to populate the inventory with 25 sample BMCs using the CLI.

### Usage

```bash
# With server running on default port (8080)
./scripts/populate-bmcs.sh

# With custom server URL
INVENTORY_SERVER=http://localhost:9999 ./scripts/populate-bmcs.sh

# With custom CLI path
INVENTORY_CLI=./bin/inventory-cli ./scripts/populate-bmcs.sh
```

### Environment Variables

- `INVENTORY_SERVER` - Server URL (default: `http://localhost:9999`)
- `INVENTORY_CLI` - Path to CLI binary (default: `./bin/inventory-cli`)

### What It Creates

The script creates 25 BMCs with:
- Sequential naming: `bmc-001` through `bmc-025`
- IP addresses: `10.0.0.100` through `10.0.0.124`
- MAC addresses: `aa:bb:cc:dd:ee:01` through `aa:bb:cc:dd:ee:19`
- BMC types: Alternating between `iLO`, `iDRAC`, and `Redfish`
- Rack organization: 5 BMCs per rack (rack-1 through rack-5)
- Position labels: U1 through U5 per rack
- Labels: datacenter, rack, environment
- Annotations: description, position

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

## Cleanup

To remove all created BMCs:

```bash
# List and delete all BMCs
for uid in $(./bin/inventory-cli bmc list --output json | jq -r '.[].metadata.uid'); do
    ./bin/inventory-cli bmc delete "$uid"
done
```

Or delete the storage directory:

```bash
rm -rf ./inventory/
```
