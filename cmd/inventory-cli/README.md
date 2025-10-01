# Inventory CLI

A command-line interface for managing OpenCHAMI inventory resources.

## Installation

Build the CLI:

```bash
make generate-client-cmd
go build -o bin/inventory-cli ./cmd/inventory-cli
```

Or build everything:

```bash
make dev
```

## Usage

### Basic Commands

The CLI supports standard CRUD operations for all inventory resources:

- `bmc` - Manage BMC (Baseboard Management Controller) resources
- `node` - Manage Node resources
- `fru` - Manage FRU (Field Replaceable Unit) resources
- `bootconfiguration` - Manage Boot Configuration resources

Each resource supports the following subcommands:

- `list` - List all resources
- `get [uid]` - Get a specific resource by UID
- `create` - Create a new resource
- `update [uid]` - Update an existing resource
- `delete [uid]` - Delete a resource

### Global Flags

- `--server` - Inventory server URL (default: `http://localhost:8080`)
- `--timeout` - Request timeout (default: `30s`)
- `--output, -o` - Output format: `table`, `json`, `yaml` (default: `table`)
- `--config` - Config file (default: `$HOME/.inventory-cli.yaml`)

### Environment Variables

All flags can be set via environment variables with the `INVENTORY_` prefix:

- `INVENTORY_SERVER` - Server URL
- `INVENTORY_TIMEOUT` - Request timeout
- `INVENTORY_OUTPUT` - Output format

### Examples

#### List all BMCs

```bash
./bin/inventory-cli bmc list
```

#### Get a specific BMC by UID

```bash
./bin/inventory-cli bmc get <uid>
```

#### Create a new BMC from stdin

```bash
echo '{
  "name": "bmc-01",
  "ipAddress": "10.0.0.100",
  "macAddress": "aa:bb:cc:dd:ee:ff",
  "username": "admin",
  "labels": {
    "datacenter": "dc1"
  }
}' | ./bin/inventory-cli bmc create
```

#### Create a new BMC with spec flag

```bash
./bin/inventory-cli bmc create --spec '{
  "name": "bmc-01",
  "ipAddress": "10.0.0.100",
  "macAddress": "aa:bb:cc:dd:ee:ff",
  "username": "admin"
}'
```

#### Update a BMC

```bash
echo '{
  "ipAddress": "10.0.0.101"
}' | ./bin/inventory-cli bmc update <uid>
```

#### Delete a BMC

```bash
./bin/inventory-cli bmc delete <uid>
```

#### Use JSON output

```bash
./bin/inventory-cli --output json node list
```

#### Set server URL

```bash
./bin/inventory-cli --server https://inventory.example.com node list
```

Or using environment variable:

```bash
export INVENTORY_SERVER=https://inventory.example.com
./bin/inventory-cli node list
```

### Configuration File

You can create a configuration file at `~/.inventory-cli.yaml`:

```yaml
server: https://inventory.example.com
timeout: 60s
output: json
```

## Resource Examples

### Node Operations

```bash
# List all nodes
./bin/inventory-cli node list

# Get a specific node
./bin/inventory-cli node get <uid>

# Create a node
echo '{
  "name": "node-01",
  "xname": "x1000c0s0b0n0",
  "nid": 1,
  "role": "compute"
}' | ./bin/inventory-cli node create

# Update a node
echo '{
  "role": "login"
}' | ./bin/inventory-cli node update <uid>

# Delete a node
./bin/inventory-cli node delete <uid>
```

### FRU Operations

```bash
# List all FRUs
./bin/inventory-cli fru list

# Get a specific FRU
./bin/inventory-cli fru get <uid>

# Create a FRU
echo '{
  "name": "fru-01",
  "type": "chassis",
  "manufacturer": "HPE"
}' | ./bin/inventory-cli fru create
```

### Boot Configuration Operations

```bash
# List all boot configurations
./bin/inventory-cli bootconfiguration list

# Get a specific boot configuration
./bin/inventory-cli bootconfiguration get <uid>

# Create a boot configuration
echo '{
  "name": "compute-boot",
  "kernel": "vmlinuz-5.15",
  "initrd": "initrd-5.15.img"
}' | ./bin/inventory-cli bootconfiguration create
```

## Help

For detailed help on any command:

```bash
./bin/inventory-cli --help
./bin/inventory-cli bmc --help
./bin/inventory-cli bmc create --help
```

## Code Generation

This CLI is automatically generated from the resource definitions. To regenerate:

```bash
make generate-client-cmd
```

The generator uses the template at `pkg/codegen/templates/client-cmd.go.tmpl`.
