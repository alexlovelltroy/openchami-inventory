# CLI Version Support Examples

The inventory-cli now supports API version negotiation via the `--version` flag.

## Basic Usage

### List Resources with Different Versions

```bash
# Default version (v1)
./bin/inventory-cli --server http://localhost:9999 bmc list

# Request v2beta1
./bin/inventory-cli --server http://localhost:9999 --version v2beta1 bmc list

# Short form
./bin/inventory-cli --server http://localhost:9999 -v v2beta1 bmc list
```

### Get Specific Resource with Version

```bash
# Get as v1 (shows username/password)
./bin/inventory-cli --server http://localhost:9999 \
  --version v1 \
  bmc get 550e8400-e29b-41d4-a716-446655440000

# Get as v2beta1 (shows authentication structure)
./bin/inventory-cli --server http://localhost:9999 \
  --version v2beta1 \
  bmc get 550e8400-e29b-41d4-a716-446655440000 \
  --output json
```

## Environment Variables

Set the version once for multiple commands:

```bash
export INVENTORY_SERVER=http://localhost:9999
export INVENTORY_VERSION=v2beta1

# All commands will use v2beta1
./bin/inventory-cli bmc list
./bin/inventory-cli bmc get <uid>
./bin/inventory-cli node list
```

## Configuration File

Create `~/.inventory-cli.yaml`:

```yaml
server: http://localhost:9999
version: v2beta1
timeout: 30s
output: json
```

Then commands automatically use v2beta1:

```bash
./bin/inventory-cli bmc list
```

## Comparing Versions

### Side-by-side Comparison

```bash
# Get same resource as different versions
./bin/inventory-cli -v v1 bmc get <uid> --output json > v1-response.json
./bin/inventory-cli -v v2beta1 bmc get <uid> --output json > v2beta1-response.json

# Compare with jq
echo "v1 authentication:"
jq '.spec | {username, password, type}' v1-response.json

echo ""
echo "v2beta1 authentication:"
jq '.spec.authentication' v2beta1-response.json
```

### See Schema Version in Response

```bash
./bin/inventory-cli -v v2beta1 bmc get <uid> --output json | jq '.schemaVersion'
# Output: "v2beta1"
```

## Create/Update with Versions

### Create BMC with Specific Version

Currently, create operations should use curl for precise version control:

```bash
# Create v2beta1 BMC
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json;version=v2beta1" \
  -H "Accept: application/json;version=v2beta1" \
  -d '{
    "name": "bmc-v2",
    "address": "https://10.0.0.1",
    "type": "Redfish",
    "authentication": {
      "method": "oidc",
      "oidc": {
        "issuerUrl": "https://auth.example.com",
        "clientId": "bmc-client"
      }
    }
  }'
```

The CLI will use the specified version for reading the created resource.

## Advanced Examples

### List All BMCs and Show Their Schema Versions

```bash
./bin/inventory-cli bmc list --output json | \
  jq '.[] | {name: .metadata.name, version: .schemaVersion}'
```

### Get BMC with v1 and Extract Auth Fields

```bash
./bin/inventory-cli -v v1 bmc get <uid> --output json | \
  jq '{
    name: .metadata.name,
    username: .spec.username,
    address: .spec.address,
    type: .spec.type
  }'
```

### Get BMC with v2beta1 and Extract Auth Method

```bash
./bin/inventory-cli -v v2beta1 bmc get <uid> --output json | \
  jq '{
    name: .metadata.name,
    authMethod: .spec.authentication.method,
    address: .spec.address,
    type: .spec.type
  }'
```

### Test Version Conversion

```bash
# Create a v1 BMC (basic auth)
echo '{
  "name": "test-bmc",
  "address": "https://10.0.0.100",
  "username": "admin",
  "password": "secret",
  "type": "Redfish"
}' | ./bin/inventory-cli bmc create --output json > created.json

# Extract UID
BMC_UID=$(jq -r '.metadata.uid' created.json)

# Read it as v2beta1 to see conversion
./bin/inventory-cli -v v2beta1 bmc get "$BMC_UID" --output json | \
  jq '.spec.authentication'

# Output shows basic auth converted to v2beta1 format:
# {
#   "method": "basic",
#   "basic": {
#     "username": "admin",
#     "password": "secret"
#   }
# }
```

## Troubleshooting

### "Version not supported" Error

```bash
./bin/inventory-cli -v v3 bmc list
# Error: API error (400): Version 'v3' is not supported for BMC
```

Check supported versions:
```bash
curl http://localhost:9999/version-info | jq '.supportedResourceVersions'
```

### Version Flag Not Working

Verify you're using the latest CLI:
```bash
make build
./bin/inventory-cli --help | grep version
```

Should show:
```
-v, --version string     API version to request (e.g., v1, v2beta1)
```

### Different Results with Same Version

The server may have converted the resource internally. Check the actual `schemaVersion` in the response:

```bash
./bin/inventory-cli -v v1 bmc get <uid> --output json | jq '.schemaVersion'
```

If this returns `"v1"` but you created it as v2beta1, the server stored it as v1 (lossy conversion).

## Tips

1. **Use JSON output** for scripting: `--output json`
2. **Check schema version** in responses to understand conversions
3. **Use environment variables** for consistent testing
4. **Compare versions** side-by-side to understand differences
5. **Default to stable versions** (v1) in production

## See Also

- [Version Negotiation Guide](../docs/VERSION-NEGOTIATION.md)
- [Testing Guide](../docs/TESTING.md)
- [Demo Script](./demo-version-negotiation.sh)
