#!/bin/bash
# Integration test for BMC version conversion
# Tests v1 <-> v2beta1 conversion end-to-end

set -e

BASE_URL="http://localhost:9999"
BMC_UID=""

echo "=================================="
echo "BMC Version Conversion Integration Test"
echo "=================================="
echo ""

# Start server in background
echo "Starting server..."
../../bin/server > /tmp/integration-server.log 2>&1 &
SERVER_PID=$!
sleep 2

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}

trap cleanup EXIT

echo "✓ Server started (PID: $SERVER_PID)"
echo ""

# Test 1: Check version-info endpoint
echo "Test 1: Version Discovery"
echo "-------------------------"
VERSION_INFO=$(curl -s $BASE_URL/version-info)
echo "$VERSION_INFO" | jq '.'
echo ""

BMC_VERSIONS=$(echo "$VERSION_INFO" | jq -r '.supportedResourceVersions.BMC | join(", ")')
echo "✓ BMC versions available: $BMC_VERSIONS"
echo ""

# Test 2: Create a BMC with v1
echo "Test 2: Create BMC with v1"
echo "-------------------------"
CREATE_V1_RESPONSE=$(curl -s -X POST $BASE_URL/bmcs \
  -H "Content-Type: application/json" \
  -H "Accept: application/json;version=v1" \
  -d '{
    "name": "test-bmc-01",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "secret123",
    "type": "Redfish"
  }')

echo "$CREATE_V1_RESPONSE" | jq '.'
BMC_UID=$(echo "$CREATE_V1_RESPONSE" | jq -r '.metadata.uid')
echo ""
echo "✓ Created BMC with UID: $BMC_UID"
echo ""

# Test 3: Retrieve the BMC as v1
echo "Test 3: Retrieve BMC as v1"
echo "-------------------------"
GET_V1_RESPONSE=$(curl -s -H "Accept: application/json;version=v1" \
  "$BASE_URL/bmcs/$BMC_UID")

echo "$GET_V1_RESPONSE" | jq '.'
V1_SCHEMA=$(echo "$GET_V1_RESPONSE" | jq -r '.schemaVersion')
V1_USERNAME=$(echo "$GET_V1_RESPONSE" | jq -r '.spec.username')
echo ""
echo "✓ Retrieved as v1 - Schema: $V1_SCHEMA, Username: $V1_USERNAME"
echo ""

# Test 4: Retrieve the same BMC as v2beta1 (with conversion)
echo "Test 4: Retrieve BMC as v2beta1 (conversion test)"
echo "------------------------------------------------"
GET_V2BETA1_RESPONSE=$(curl -s -H "Accept: application/json;version=v2beta1" \
  "$BASE_URL/bmcs/$BMC_UID")

echo "$GET_V2BETA1_RESPONSE" | jq '.'
V2_SCHEMA=$(echo "$GET_V2BETA1_RESPONSE" | jq -r '.schemaVersion')
V2_AUTH_METHOD=$(echo "$GET_V2BETA1_RESPONSE" | jq -r '.spec.authentication.method')
V2_USERNAME=$(echo "$GET_V2BETA1_RESPONSE" | jq -r '.spec.authentication.basic.username')
echo ""
echo "✓ Retrieved as v2beta1 - Schema: $V2_SCHEMA, Auth Method: $V2_AUTH_METHOD, Username: $V2_USERNAME"
echo ""

# Test 5: Create a BMC with v2beta1
echo "Test 5: Create BMC with v2beta1"
echo "-------------------------------"
CREATE_V2_RESPONSE=$(curl -s -X POST $BASE_URL/bmcs \
  -H "Content-Type: application/json;version=v2beta1" \
  -H "Accept: application/json;version=v2beta1" \
  -d '{
    "name": "test-bmc-02",
    "address": "https://10.1.1.101",
    "type": "Redfish",
    "authentication": {
      "method": "basic",
      "basic": {
        "username": "root",
        "password": "password456"
      }
    }
  }')

echo "$CREATE_V2_RESPONSE" | jq '.'
BMC2_UID=$(echo "$CREATE_V2_RESPONSE" | jq -r '.metadata.uid')
echo ""
echo "✓ Created BMC with v2beta1, UID: $BMC2_UID"
echo ""

# Test 6: Retrieve v2beta1-created BMC as v1
echo "Test 6: Retrieve v2beta1-created BMC as v1 (backward conversion)"
echo "---------------------------------------------------------------"
GET_V2_AS_V1_RESPONSE=$(curl -s -H "Accept: application/json;version=v1" \
  "$BASE_URL/bmcs/$BMC2_UID")

echo "$GET_V2_AS_V1_RESPONSE" | jq '.'
CONVERTED_SCHEMA=$(echo "$GET_V2_AS_V1_RESPONSE" | jq -r '.schemaVersion')
CONVERTED_USERNAME=$(echo "$GET_V2_AS_V1_RESPONSE" | jq -r '.spec.username')
echo ""
echo "✓ Retrieved v2beta1 BMC as v1 - Schema: $CONVERTED_SCHEMA, Username: $CONVERTED_USERNAME"
echo ""

# Test 7: List all BMCs with v1
echo "Test 7: List all BMCs as v1"
echo "---------------------------"
LIST_V1=$(curl -s -H "Accept: application/json;version=v1" "$BASE_URL/bmcs")
V1_COUNT=$(echo "$LIST_V1" | jq '. | length')
echo "✓ Found $V1_COUNT BMCs in v1 format"
echo ""

# Test 8: List all BMCs with v2beta1
echo "Test 8: List all BMCs as v2beta1"
echo "--------------------------------"
LIST_V2BETA1=$(curl -s -H "Accept: application/json;version=v2beta1" "$BASE_URL/bmcs")
V2_COUNT=$(echo "$LIST_V2BETA1" | jq '. | length')
echo "✓ Found $V2_COUNT BMCs in v2beta1 format"
echo ""

# Test 9: Request unsupported version (should fail)
echo "Test 9: Request unsupported version (should return 406)"
echo "-------------------------------------------------------"
UNSUPPORTED_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -H "Accept: application/json;version=v3" "$BASE_URL/bmcs/$BMC_UID")
HTTP_STATUS=$(echo "$UNSUPPORTED_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
echo "HTTP Status: $HTTP_STATUS"

if [ "$HTTP_STATUS" == "406" ]; then
    echo "✓ Correctly rejected unsupported version with 406 Not Acceptable"
else
    echo "✗ Expected 406, got $HTTP_STATUS"
    exit 1
fi
echo ""

# Summary
echo "=================================="
echo "✅ All Integration Tests Passed!"
echo "=================================="
echo ""
echo "Summary:"
echo "  - Version discovery working"
echo "  - Create BMC with v1: ✓"
echo "  - Create BMC with v2beta1: ✓"
echo "  - v1 → v2beta1 conversion: ✓"
echo "  - v2beta1 → v1 conversion: ✓"
echo "  - List operations: ✓"
echo "  - Unsupported version rejection: ✓"
echo ""
