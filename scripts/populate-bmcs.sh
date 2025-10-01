#!/usr/bin/env bash
# Script to populate the inventory with 25 sample BMCs using the CLI

set -e

SERVER_URL="${INVENTORY_SERVER:-http://localhost:9999}"
CLI="${INVENTORY_CLI:-./bin/inventory-cli}"

echo "Populating inventory with 25 sample BMCs..."
echo "Server: $SERVER_URL"
echo "CLI: $CLI"
echo ""

# Check if CLI exists
if [ ! -f "$CLI" ]; then
    echo "Error: CLI not found at $CLI"
    echo "Please build it first with: make dev"
    exit 1
fi

# Create 25 BMCs with varying configurations
for i in $(seq 1 25); do
    # Calculate rack and position
    rack=$((($i - 1) / 5 + 1))
    position=$(($i % 5))
    if [ $position -eq 0 ]; then
        position=5
    fi
    
    # Generate IP address (10.0.0.100-124)
    ip="10.0.0.$((99 + i))"
    
    # Generate MAC address
    mac=$(printf "aa:bb:cc:dd:ee:%02x" $i)
    
    # Alternate between different BMC types
    case $(($i % 3)) in
        0) bmc_type="iLO" ;;
        1) bmc_type="iDRAC" ;;
        2) bmc_type="Redfish" ;;
    esac
    
    # Create JSON payload
    json=$(cat <<EOF
{
  "name": "bmc-$(printf "%03d" $i)",
  "ipAddress": "$ip",
  "macAddress": "$mac",
  "username": "admin",
  "password": "changeme",
  "bmcType": "$bmc_type",
  "labels": {
    "datacenter": "dc1",
    "rack": "rack-$rack",
    "environment": "production"
  },
  "annotations": {
    "description": "Sample BMC #$i",
    "position": "U$position"
  }
}
EOF
)
    
    echo -n "Creating bmc-$(printf "%03d" $i) ($bmc_type at $ip)... "
    
    if echo "$json" | $CLI --server "$SERVER_URL" bmc create > /dev/null 2>&1; then
        echo "✓"
    else
        echo "✗ (failed)"
    fi
done

echo ""
echo "Done! Created 25 BMCs"
echo ""
echo "View them with:"
echo "  $CLI --server $SERVER_URL bmc list"
