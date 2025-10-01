#!/bin/bash
#
# demo-version-negotiation.sh
# Demonstrates multi-version schema support in the OpenCHAMI Inventory API
#
# This script:
#   1. Creates a BMC using v1 schema (basic auth)
#   2. Creates a BMC using v2beta1 schema (OIDC auth)
#   3. Retrieves both BMCs as v1 (shows basic auth only)
#   4. Retrieves both BMCs as v2beta1 (shows enhanced auth)
#   5. Demonstrates backward/forward compatibility
#
# Usage:
#   ./scripts/demo-version-negotiation.sh
#
# Environment Variables:
#   INVENTORY_SERVER - Server URL (default: http://localhost:9999)
#   INVENTORY_CLI    - Path to CLI binary (default: ./bin/inventory-cli)

set -e

# Configuration
INVENTORY_SERVER=${INVENTORY_SERVER:-http://localhost:9999}
INVENTORY_CLI=${INVENTORY_CLI:-./bin/inventory-cli}

echo "================================================"
echo "OpenCHAMI Inventory - Version Negotiation Demo"
echo "================================================"
echo ""
echo "Server: $INVENTORY_SERVER"
echo "CLI: $INVENTORY_CLI"
echo ""

# Check if CLI exists
if [ ! -f "$INVENTORY_CLI" ]; then
    echo "Error: CLI not found at $INVENTORY_CLI"
    echo "Run 'make build' to build the CLI"
    exit 1
fi

# Check if server is running
if ! curl -s -f "$INVENTORY_SERVER/health" > /dev/null 2>&1; then
    echo "Error: Server not responding at $INVENTORY_SERVER"
    echo "Start the server with: ./bin/server --port 9999 --disable-auth"
    exit 1
fi

echo "✓ Server is running"
echo ""

#
# Step 1: Create a v1 BMC (basic authentication)
#
echo "Step 1: Creating BMC with v1 schema (basic auth)"
echo "================================================"

V1_BMC_JSON=$(cat <<'EOF'
{
  "name": "bmc-v1-demo",
  "address": "https://10.99.1.100",
  "username": "admin",
  "password": "secret123",
  "type": "Redfish",
  "labels": {
    "datacenter": "dc1",
    "version": "v1",
    "demo": "true"
  },
  "annotations": {
    "description": "BMC created with v1 schema using basic authentication",
    "created-by": "demo-script"
  }
}
EOF
)

echo "$V1_BMC_JSON" | $INVENTORY_CLI --server "$INVENTORY_SERVER" bmc create > /tmp/v1_bmc_response.json
V1_BMC_UID=$(jq -r '.metadata.uid' /tmp/v1_bmc_response.json)
echo "✓ Created v1 BMC with UID: $V1_BMC_UID"
echo ""

#
# Step 2: Create a v2beta1 BMC (OIDC authentication)
#
echo "Step 2: Creating BMC with v2beta1 schema (OIDC auth)"
echo "====================================================="

V2BETA1_BMC_JSON=$(cat <<'EOF'
{
  "name": "bmc-v2beta1-demo",
  "address": "https://10.99.1.101",
  "type": "Redfish",
  "authentication": {
    "method": "oidc",
    "oidc": {
      "issuerUrl": "https://auth.example.com",
      "clientId": "bmc-inventory-client",
      "clientSecret": "super-secret-client-secret",
      "scopes": ["openid", "profile", "bmc.access"]
    }
  },
  "labels": {
    "datacenter": "dc1",
    "version": "v2beta1",
    "demo": "true"
  },
  "annotations": {
    "description": "BMC created with v2beta1 schema using OIDC authentication",
    "created-by": "demo-script"
  }
}
EOF
)

# Note: v2beta1 creation requires Accept header with version
# Using curl for now as it's easier to specify version in both headers
echo "Creating v2beta1 BMC via curl with version headers..."
V2BETA1_RESPONSE=$(curl -s -X POST "$INVENTORY_SERVER/bmcs" \
  -H "Content-Type: application/json;version=v2beta1" \
  -H "Accept: application/json;version=v2beta1" \
  -d "$V2BETA1_BMC_JSON")

V2BETA1_BMC_UID=$(echo "$V2BETA1_RESPONSE" | jq -r '.metadata.uid')

if [ "$V2BETA1_BMC_UID" == "null" ] || [ -z "$V2BETA1_BMC_UID" ]; then
    echo "Warning: v2beta1 creation may have failed or server doesn't support it yet"
    echo "Response: $V2BETA1_RESPONSE"
    echo ""
    echo "Creating a second v1 BMC for demonstration instead..."
    
    V1_BMC2_JSON=$(cat <<'EOF'
{
  "name": "bmc-v1-demo-2",
  "address": "https://10.99.1.101",
  "username": "admin",
  "password": "secret456",
  "type": "iDRAC",
  "labels": {
    "datacenter": "dc1",
    "version": "v1",
    "demo": "true"
  },
  "annotations": {
    "description": "Second BMC created with v1 schema",
    "created-by": "demo-script"
  }
}
EOF
)
    
    echo "$V1_BMC2_JSON" | $INVENTORY_CLI --server "$INVENTORY_SERVER" bmc create > /tmp/v1_bmc2_response.json
    V2BETA1_BMC_UID=$(jq -r '.metadata.uid' /tmp/v1_bmc2_response.json)
    echo "✓ Created second v1 BMC with UID: $V2BETA1_BMC_UID"
else
    echo "✓ Created v2beta1 BMC with UID: $V2BETA1_BMC_UID"
fi
echo ""

#
# Step 3: Retrieve both BMCs (default version - v1)
#
echo "Step 3: Listing all BMCs (default v1 format)"
echo "============================================="
echo ""

$INVENTORY_CLI --server "$INVENTORY_SERVER" bmc list --output json | jq -c '.[] | {name: .metadata.name, uid: .metadata.uid, address: .spec.address, schemaVersion: .schemaVersion}'
echo ""

#
# Step 4: Retrieve specific BMC as v1
#
echo "Step 4: Retrieving v1 BMC as v1 (using CLI with --version flag)"
echo "================================================================="
echo ""

$INVENTORY_CLI --server "$INVENTORY_SERVER" --version v1 bmc get "$V1_BMC_UID" --output json | jq '{
    name: .metadata.name,
    schemaVersion: .schemaVersion,
    address: .spec.address,
    username: .spec.username,
    type: .spec.type,
    hasPassword: (.spec.password != null)
}'
echo ""

#
# Step 5: Retrieve v1 BMC as v2beta1 (demonstrates forward conversion)
#
echo "Step 5: Retrieving v1 BMC as v2beta1 (forward conversion)"
echo "=========================================================="
echo "This shows how v1 basic auth is converted to v2beta1 format"
echo ""

$INVENTORY_CLI --server "$INVENTORY_SERVER" --version v2beta1 bmc get "$V1_BMC_UID" --output json | jq '{
    name: .metadata.name,
    schemaVersion: .schemaVersion,
    address: .spec.address,
    type: .spec.type,
    authentication: .spec.authentication
}'
echo ""

#
# Step 6: Retrieve second BMC as v1
#
echo "Step 6: Retrieving second BMC as v1 (backward compatibility)"
echo "============================================================="
echo ""

$INVENTORY_CLI --server "$INVENTORY_SERVER" --version v1 bmc get "$V2BETA1_BMC_UID" --output json | jq '{
    name: .metadata.name,
    schemaVersion: .schemaVersion,
    address: .spec.address,
    username: .spec.username,
    type: .spec.type,
    hasPassword: (.spec.password != null)
}'
echo ""

#
# Step 7: Retrieve second BMC as v2beta1
#
echo "Step 7: Retrieving second BMC as v2beta1"
echo "========================================="
echo ""

$INVENTORY_CLI --server "$INVENTORY_SERVER" --version v2beta1 bmc get "$V2BETA1_BMC_UID" --output json | jq '{
    name: .metadata.name,
    schemaVersion: .schemaVersion,
    address: .spec.address,
    type: .spec.type,
    authentication: .spec.authentication
}'
echo ""

#
# Step 8: Query version info
#
echo "Step 8: Querying API version information"
echo "========================================="
echo ""

curl -s "$INVENTORY_SERVER/version-info" | jq '{
    apiVersion: .apiVersion,
    supportedResourceVersions: .supportedResourceVersions,
    defaultVersions: .defaultVersions
}'
echo ""

#
# Summary
#
echo "================================================"
echo "Demo Complete!"
echo "================================================"
echo ""
echo "Key Takeaways:"
echo "1. Both v1 and v2beta1 BMCs can coexist in the system"
echo "2. Clients can request specific versions via Accept headers"
echo "3. The API automatically converts between versions"
echo "4. v1 basic auth is converted to v2beta1 authentication structure"
echo "5. Default version (v1) is used when no version is specified"
echo ""
echo "Cleanup:"
echo "  To delete demo BMCs:"
echo "    $INVENTORY_CLI --server $INVENTORY_SERVER bmc delete $V1_BMC_UID"
echo "    $INVENTORY_CLI --server $INVENTORY_SERVER bmc delete $V2BETA1_BMC_UID"
echo ""
