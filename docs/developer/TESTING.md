# Testing and Development Guide

Quick reference for testing the OpenCHAMI Inventory system.

## Quick Start

### 1. Build Everything

```bash
make dev
```

This builds:
- `bin/server` - API server
- `bin/inventory-cli` - Command-line client
- `bin/populate-bmcs` - Sample data generator
- All generated code (handlers, models, storage, client)

### 2. Start the Server (Testing Mode)

```bash
./bin/server --port 9999 --disable-auth
```

**Important**: The `--disable-auth` flag disables authentication checks using a permissive policy. This is for testing only!

Environment variable equivalent:
```bash
INVENTORY_DISABLE_AUTH=true ./bin/server --port 9999
```

### 3. Populate Sample Data

**Option A: Using the Go tool**
```bash
./bin/populate-bmcs
```

**Option B: Using the bash script**
```bash
./scripts/populate-bmcs.sh
```

Both create 25 sample BMCs with realistic data.

### 4. Test the API

**Using the CLI:**
```bash
./bin/inventory-cli --server http://localhost:9999 bmc list
./bin/inventory-cli --server http://localhost:9999 bmc get <uid>
```

**Using curl:**
```bash
curl http://localhost:9999/bmcs | jq
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

### Environment Variables

All flags can be set via environment variables with `INVENTORY_` prefix:

```bash
export INVENTORY_HOST=0.0.0.0
export INVENTORY_PORT=9999
export INVENTORY_DISABLE_AUTH=true
export INVENTORY_STORAGE_PATH=./inventory

./bin/server
```

### Configuration File

Create `server.yaml`:

```yaml
host: 0.0.0.0
port: 9999
log-level: info
security:
  disable-auth: true
storage:
  path: ./inventory
cors:
  enabled: false
  origins:
    - "*"
openapi:
  disabled: false
  validate: false
```

Then run:
```bash
./bin/server --config server.yaml
```

## Authentication Policies

### Permissive Policy (Testing)

When `--disable-auth` is set, the server uses a permissive policy that allows all operations:

```go
// Allows all operations without authentication
policyRegistry.RegisterPolicy("BMC", policies.NewPermissivePolicy())
```

**Warning**: Never use `--disable-auth` in production!

### Default Policies (Production)

Without `--disable-auth`, the server uses default policies that require authentication:

```go
// Requires JWT auth context for all operations
policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())
```

To enable authentication:
1. Remove `--disable-auth` flag
2. Set up tokensmith middleware (JWT)
3. Configure auth context in requests

### Custom Policies

Create resource-specific policies in `pkg/resources/<resource>/policy.go`:

```go
type MyCustomBMCPolicy struct{}

func (p *MyCustomBMCPolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    // Custom authorization logic
    if auth == nil {
        return policies.Deny("authentication required")
    }
    if policies.HasRole(auth, "bmc-viewer") {
        return policies.Allow()
    }
    return policies.Deny("bmc-viewer role required")
}
```

## Testing Scenarios

### Scenario 1: Basic CRUD Operations

```bash
# Create
echo '{"name":"test-bmc","address":"10.0.0.1","username":"admin","password":"secret","type":"Redfish"}' | \
  ./bin/inventory-cli --server http://localhost:9999 bmc create

# List
./bin/inventory-cli --server http://localhost:9999 bmc list

# Get
./bin/inventory-cli --server http://localhost:9999 bmc get <uid>

# Delete
./bin/inventory-cli --server http://localhost:9999 bmc delete <uid>
```

### Scenario 2: Version Negotiation

Test version negotiation with the CLI:

```bash
# Get BMC as default version (v1)
./bin/inventory-cli --server http://localhost:9999 bmc get <uid>

# Get BMC as v2beta1
./bin/inventory-cli --server http://localhost:9999 --version v2beta1 bmc get <uid>

# List all BMCs as v2beta1
./bin/inventory-cli --server http://localhost:9999 --version v2beta1 bmc list
```

Or run the comprehensive demo script:
```bash
./scripts/demo-version-negotiation.sh
```

This demonstrates:
- Creating v1 and v2beta1 BMCs
- Retrieving resources in different versions
- Automatic schema conversion
- Version discovery

### Scenario 3: Integration Testing

```bash
# Run the integration test script
./cmd/server/integration_test.sh
```

This tests:
- Server health check
- Version info endpoint
- Create/retrieve/update/delete operations
- Version conversion (v1 ↔ v2beta1)

## Development Workflow

### 1. Make Changes to Templates

Edit files in `pkg/codegen/templates/`:
- `handlers.go.tmpl` - API handlers
- `client.go.tmpl` - Client library
- `storage.go.tmpl` - Storage layer
- `models.go.tmpl` - Request/response models
- `client-cmd.go.tmpl` - CLI commands

### 2. Regenerate Code

```bash
make dev
```

This runs:
1. `make clean` - Removes generated files
2. Code generation for storage, server, client, CLI
3. `go mod tidy` - Updates dependencies
4. `go fmt` - Formats code
5. `make build` - Builds all binaries
6. `go test` - Runs tests

### 3. Test Changes

```bash
# Start server
./bin/server --port 9999 --disable-auth

# Test with CLI
./bin/inventory-cli --server http://localhost:9999 bmc list

# Run integration tests
./cmd/server/integration_test.sh
```

## Troubleshooting

### "401 Unauthorized" Errors

**Cause**: Authentication is enabled but no auth context provided.

**Solution**: Add `--disable-auth` flag when starting the server:
```bash
./bin/server --port 9999 --disable-auth
```

### "Cannot unmarshal array into Go value"

**Cause**: Client expects wrapped response but server returns plain array.

**Solution**: This was fixed by updating the client template to expect plain arrays. Regenerate code:
```bash
make dev
```

### "Policy not configured" Errors

**Cause**: No policy registered for the resource type.

**Solution**: Register a policy in `cmd/server/main.go`:
```go
policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
```

### Storage Errors

**Problem**: "Failed to save resource"

**Check**:
1. Storage directory exists: `mkdir -p ./inventory`
2. Directory is writable: `chmod 755 ./inventory`
3. Path is correct: `--storage-path ./inventory`

## Useful Commands

### View all BMCs
```bash
./bin/inventory-cli --server http://localhost:9999 bmc list --output json | jq
```

### View BMCs with specific version
```bash
./bin/inventory-cli --server http://localhost:9999 --version v2beta1 bmc list --output json | jq
```

### Count resources
```bash
./bin/inventory-cli --server http://localhost:9999 bmc list --output json | jq 'length'
```

### Compare versions
```bash
# Get as v1
./bin/inventory-cli --server http://localhost:9999 --version v1 bmc get <uid> --output json > v1.json

# Get as v2beta1
./bin/inventory-cli --server http://localhost:9999 --version v2beta1 bmc get <uid> --output json > v2beta1.json

# Compare
diff -u v1.json v2beta1.json
```

### Filter by label
```bash
curl http://localhost:9999/bmcs | jq '.[] | select(.metadata.labels.datacenter == "dc1")'
```

### Check server health
```bash
curl http://localhost:9999/health
```

### Query version info
```bash
curl http://localhost:9999/version-info | jq
```

### Clean storage
```bash
rm -rf ./inventory/
```

### Watch server logs
```bash
./bin/server --port 9999 --disable-auth --log-level debug
```

## Next Steps

1. **Explore the API**: Try different endpoints and operations
2. **Run demos**: Execute `./scripts/demo-version-negotiation.sh`
3. **Read documentation**: Check `docs/` directory for detailed guides
4. **Customize policies**: Implement your own authorization logic
5. **Add resources**: Create new resource types (FRU, BootConfiguration)

## Additional Resources

- [Version Negotiation Guide](./VERSION-NEGOTIATION.md)
- [Development Guide](./DEVELOPMENT.md)
- [BMC v2beta1 Documentation](../pkg/resources/bmc/v2beta1/README.md)
- [Multi-Schema Versioning Proposal](../PROPOSAL-Multi-Schema-Versioning.md)
