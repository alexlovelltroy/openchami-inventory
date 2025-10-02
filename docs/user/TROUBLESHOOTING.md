# Troubleshooting Guide

Common issues and solutions for OpenCHAMI Inventory.

## Table of Contents

- [Server Issues](#server-issues)
- [Authentication Errors](#authentication-errors)
- [CLI Problems](#cli-problems)
- [API Errors](#api-errors)
- [Version Negotiation Issues](#version-negotiation-issues)
- [Storage Problems](#storage-problems)
- [Code Generation Issues](#code-generation-issues)
- [Performance Issues](#performance-issues)

## Server Issues

### Server Won't Start

#### Port Already in Use

**Symptom:**
```
Error: listen tcp :9999: bind: address already in use
```

**Diagnosis:**
```bash
# Check what's using the port
lsof -i :9999

# Or on Linux
netstat -tulpn | grep 9999
```

**Solutions:**
```bash
# Option 1: Use different port
./bin/server --port 9998

# Option 2: Kill process using port
kill <pid>

# Option 3: Stop existing server
pkill -f "bin/server"
```

#### Storage Directory Permission Denied

**Symptom:**
```
Error: failed to create storage directory: permission denied
```

**Solution:**
```bash
# Create directory with correct permissions
mkdir -p ./inventory
chmod 755 ./inventory

# Or specify writable path
./bin/server --storage-path ~/inventory-data
```

#### Binary Not Found

**Symptom:**
```
zsh: command not found: server
```

**Solution:**
```bash
# Check if binary exists
ls -la bin/server

# If missing, build it
make build

# Run with explicit path
./bin/server
```

### Server Crashes

#### Panic on Startup

**Symptom:**
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Diagnosis:**
```bash
# Run with debug logging
./bin/server --log-level debug

# Check for corrupt storage files
ls -la inventory/
```

**Solution:**
```bash
# Start fresh with new storage directory
rm -rf ./inventory
mkdir -p ./inventory
./bin/server
```

## Authentication Errors

### 401 Unauthorized

**Symptom:**
```
Error: API request failed: 401 Unauthorized
{"error": "authentication required"}
```

**Cause:** Server requires authentication but none was provided.

**Solutions:**

**Option 1: Use testing mode (development only)**
```bash
# Restart server with auth disabled
./bin/server --disable-auth
```

**Option 2: Provide JWT token (production)**
```bash
# With CLI
./bin/inventory-cli --token <jwt> bmc list

# With curl
curl -H "Authorization: Bearer <jwt>" http://localhost:9999/bmcs
```

### 403 Forbidden

**Symptom:**
```
Error: API request failed: 403 Forbidden
{"error": "insufficient permissions"}
```

**Cause:** Authentication provided but user lacks required permissions.

**Diagnosis:**
```bash
# Check policy configuration in cmd/server/main.go
grep -A 10 "RegisterPolicy" cmd/server/main.go
```

**Solution:**
- Configure appropriate policies for your user roles
- See [AUTHENTICATION.md](./AUTHENTICATION.md) for policy configuration

### Policy Not Configured

**Symptom:**
```
Error: no policy configured for resource type BMC
```

**Solution:**
```bash
# Check cmd/server/main.go for policy registration
# Ensure all resources have policies registered

# Example fix in cmd/server/main.go:
policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
```

## CLI Problems

### CLI Connection Refused

**Symptom:**
```
Error: failed to connect to server: connection refused
```

**Diagnosis:**
```bash
# Check server is running
curl http://localhost:9999/health

# Check server URL
echo $INVENTORY_SERVER
```

**Solutions:**
```bash
# Start server if not running
./bin/server --disable-auth

# Verify correct URL
./bin/inventory-cli --server http://localhost:9999 bmc list

# Set environment variable
export INVENTORY_SERVER=http://localhost:9999
```

### Invalid JSON in Spec

**Symptom:**
```
Error: invalid character '}' looking for beginning of object key string
```

**Diagnosis:**
```bash
# Validate JSON
echo '{"name": "test",}' | jq .
# parse error: Expected another key-value pair at line 1, column 18
```

**Solution:**
```bash
# Fix JSON syntax (remove trailing comma)
./bin/inventory-cli bmc create --spec '{"name": "test"}'

# Use file to avoid shell escaping issues
cat > bmc.json <<EOF
{
  "name": "test-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
EOF

cat bmc.json | ./bin/inventory-cli bmc create
```

### Resource Not Found

**Symptom:**
```
Error: resource not found: 550e8400-e29b-41d4-a716-446655440000
```

**Diagnosis:**
```bash
# List all resources to verify UID
./bin/inventory-cli bmc list --output json | jq '.[] | .metadata.uid'
```

**Solution:**
```bash
# Use correct UID
./bin/inventory-cli bmc get <correct-uid>

# Or search by name
./bin/inventory-cli bmc list --output json | \
  jq '.[] | select(.metadata.name == "node001-bmc")'
```

## API Errors

### 400 Bad Request

**Symptom:**
```json
{
  "error": "invalid request body",
  "details": "missing required field: name"
}
```

**Common Causes:**
1. Missing required fields
2. Invalid JSON syntax
3. Wrong data types
4. Invalid field values

**Solution:**
```bash
# Include all required fields
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-bmc",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "changeme",
    "type": "Redfish"
  }'
```

### Cannot Unmarshal Array

**Symptom:**
```
Error: cannot unmarshal array into Go value of type client.BMCsResponse
```

**Cause:** Client/server response format mismatch (should be fixed in recent versions).

**Solution:**
```bash
# Regenerate code
make dev

# Rebuild CLI
make build
```

### 404 Not Found on List

**Symptom:**
```
GET /bmcs returns 404
```

**Diagnosis:**
```bash
# Check server logs
./bin/server --log-level debug

# Verify routes are registered
curl http://localhost:9999/openapi.json | jq '.paths | keys'
```

**Solution:**
```bash
# Regenerate server code
make generate-server
make build

# Restart server
./bin/server --disable-auth
```

## Version Negotiation Issues

### Version Not Supported

**Symptom:**
```
Error: Version 'v3' is not supported for BMC
```

**Diagnosis:**
```bash
# Check supported versions
curl http://localhost:9999/version-info | jq '.supportedResourceVersions'
```

**Solution:**
```bash
# Use supported version
./bin/inventory-cli --version v1 bmc list
./bin/inventory-cli --version v2beta1 bmc list
```

### Unexpected Version in Response

**Symptom:** Requested v2beta1 but got v1 response.

**Diagnosis:**
```bash
# Check schemaVersion in response
./bin/inventory-cli -v v2beta1 bmc get <uid> -o json | jq '.schemaVersion'
```

**Explanation:** Server may store resource in different version than requested. The system converts between versions automatically.

**Verify Conversion:**
```bash
# Get as v1
./bin/inventory-cli -v v1 bmc get <uid> -o json > v1.json

# Get as v2beta1
./bin/inventory-cli -v v2beta1 bmc get <uid> -o json > v2beta1.json

# Compare authentication structure
diff -u <(jq '.spec' v1.json) <(jq '.spec' v2beta1.json)
```

### Version Header Not Working

**Symptom:** Version header ignored, always get v1.

**Diagnosis:**
```bash
# Check server supports version negotiation
curl -v -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid>

# Look for X-Schema-Version response header
```

**Solution:**
```bash
# Ensure server is recent version with version support
git pull
make dev
./bin/server --disable-auth
```

## Storage Problems

### Storage File Corruption

**Symptom:**
```
Error: failed to load resource: invalid JSON
```

**Diagnosis:**
```bash
# Check storage files
ls -la inventory/bmcs/

# Validate JSON
jq . inventory/bmcs/<uid>.json
```

**Solution:**
```bash
# Remove corrupt file
rm inventory/bmcs/<uid>.json

# Or start fresh
rm -rf inventory/
mkdir -p inventory/
./bin/populate-bmcs
```

### Disk Space Issues

**Symptom:**
```
Error: failed to save resource: no space left on device
```

**Diagnosis:**
```bash
# Check disk space
df -h .

# Check storage directory size
du -sh inventory/
```

**Solution:**
```bash
# Clean up storage
rm -rf inventory/
mkdir -p inventory/

# Or use different location
./bin/server --storage-path /path/with/space
```

### Permission Denied on Storage

**Symptom:**
```
Error: failed to save resource: permission denied
```

**Solution:**
```bash
# Fix permissions
chmod -R 755 inventory/

# Or use writable location
./bin/server --storage-path ~/inventory-data
```

## Code Generation Issues

### Template Errors

**Symptom:**
```
Error: template: handlers.go.tmpl:42: undefined variable "ResourceType"
```

**Diagnosis:**
```bash
# Check template syntax
cat pkg/codegen/templates/handlers.go.tmpl | grep -n "ResourceType"
```

**Solution:**
```bash
# Fix template variable name
# Edit pkg/codegen/templates/handlers.go.tmpl
# Change {{.ResourceType}} to {{.Name}}

# Regenerate
make dev
```

### Generated Code Won't Compile

**Symptom:**
```
cmd/server/bmc_handlers_generated.go:25: undefined: bmc.BMC
```

**Diagnosis:**
```bash
# Check resource package imports
head -20 cmd/server/bmc_handlers_generated.go
```

**Solution:**
```bash
# Regenerate code
make clean
make dev

# Check for errors in resource definition
go build ./pkg/resources/bmc
```

### make dev Fails

**Symptom:**
```
make: *** [dev] Error 1
```

**Diagnosis:**
```bash
# Run steps individually
make clean
make generate
make build
make test
```

**Common Solutions:**
```bash
# Update dependencies
go mod tidy
go mod download

# Clean and rebuild
make clean
rm -rf bin/
make dev

# Check Go version
go version  # Requires 1.24+
```

## Performance Issues

### Slow List Operations

**Symptom:** `list` operations take several seconds.

**Cause:** File backend reads all files from disk.

**Solutions:**
```bash
# Reduce dataset size
rm -rf inventory/
./bin/populate-bmcs  # Creates 25 BMCs

# Filter on client side
./bin/inventory-cli bmc list -o json | \
  jq '.[] | select(.metadata.labels.datacenter == "dc1")'

# Future: Implement database backend
```

### High Memory Usage

**Symptom:** Server uses excessive memory.

**Diagnosis:**
```bash
# Check memory usage
ps aux | grep server

# Monitor while running
top -p $(pgrep -f bin/server)
```

**Solutions:**
```bash
# Reduce stored resources
rm -rf inventory/
mkdir -p inventory/

# Use smaller dataset
# Edit cmd/populate-bmcs/main.go to create fewer resources

# Future: Implement pagination
```

## Debug Mode

### Enable Debug Logging

```bash
# Server
./bin/server --log-level debug

# Check specific operations
./bin/server --log-level debug 2>&1 | grep -i "error\|warn"
```

### Trace API Requests

```bash
# Use curl with verbose output
curl -v http://localhost:9999/bmcs

# Check request/response headers
curl -i http://localhost:9999/bmcs
```

### Dump Request/Response

```bash
# Save request
cat > request.json <<EOF
{
  "name": "test-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
EOF

# Send and save response
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json" \
  -d @request.json \
  -w "\nStatus: %{http_code}\n" \
  -o response.json

# Examine response
cat response.json | jq .
```

## Getting More Help

### Check Logs

```bash
# Server logs (if running in foreground)
./bin/server --log-level debug

# Server logs (if running as service)
journalctl -u inventory-server -f
```

### Verify Installation

```bash
# Check binaries exist
ls -la bin/

# Check versions
./bin/server --version
./bin/inventory-cli --version

# Verify build
make discover
```

### Validate Configuration

```bash
# Check server config
cat server.yaml

# Check CLI config
cat ~/.inventory-cli.yaml

# Check environment variables
env | grep INVENTORY_
```

### Report Issues

When reporting issues, include:
1. **Error message** (full text)
2. **Steps to reproduce**
3. **Environment** (OS, Go version, commit hash)
4. **Configuration** (flags, environment variables)
5. **Logs** (with `--log-level debug`)

**Create issue:**
```bash
# Get diagnostic info
echo "OS: $(uname -a)"
echo "Go: $(go version)"
echo "Commit: $(git rev-parse HEAD)"
echo "Build: $(git log -1 --oneline)"

# Create issue at:
# https://github.com/openchami/inventory/issues/new
```

## See Also

- [User Guide](./USER-GUIDE.md) - Complete usage guide
- [CLI Reference](./CLI-REFERENCE.md) - CLI documentation
- [API Reference](./API-REFERENCE.md) - REST API documentation
- [Authentication](./AUTHENTICATION.md) - Security configuration
- [Development Guide](../developer/DEVELOPMENT.md) - Development setup
