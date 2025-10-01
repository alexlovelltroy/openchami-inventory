package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openchami/inventory/pkg/client"
)

func main() {
	// Get server URL from environment or use default
	serverURL := os.Getenv("INVENTORY_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:9999"
	}

	// Create client
	c, err := client.NewClient(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Populating inventory with 25 sample BMCs...\n")
	fmt.Printf("Server: %s\n\n", serverURL)

	// Create 25 BMCs
	successCount := 0
	failCount := 0

	for i := 1; i <= 25; i++ {
		// Calculate rack and position
		rack := (i-1)/5 + 1
		position := i % 5
		if position == 0 {
			position = 5
		}

		// Generate IP address (10.0.0.100-124)
		ip := fmt.Sprintf("10.0.0.%d", 99+i)

		// Generate MAC address
		mac := fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i)

		// Alternate between different BMC types
		var bmcType string
		switch i % 3 {
		case 0:
			bmcType = "iLO"
		case 1:
			bmcType = "iDRAC"
		case 2:
			bmcType = "Redfish"
		}

		// Create request - BMCSpec fields are embedded inline
		req := client.CreateBMCRequest{}
		req.Address = ip
		req.Username = "admin"
		req.Password = "changeme"
		req.Type = bmcType
		req.Name = fmt.Sprintf("bmc-%03d", i)
		req.Labels = map[string]string{
			"datacenter":  "dc1",
			"rack":        fmt.Sprintf("rack-%d", rack),
			"environment": "production",
			"mac":         mac, // Store MAC in labels since BMCSpec doesn't have it
		}
		req.Annotations = map[string]string{
			"description": fmt.Sprintf("Sample BMC #%d", i),
			"position":    fmt.Sprintf("U%d", position),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fmt.Printf("Creating bmc-%03d (%s at %s)... ", i, bmcType, ip)

		bmc, err := c.CreateBMC(ctx, req)
		if err != nil {
			fmt.Printf("✗ (failed: %v)\n", err)
			failCount++
			continue
		}

		fmt.Printf("✓ (UID: %s)\n", bmc.Metadata.UID)
		successCount++
	}

	fmt.Printf("\nDone! Created %d BMCs (%d failed)\n", successCount, failCount)
	fmt.Printf("\nView them with:\n")
	fmt.Printf("  curl %s/bmcs | jq\n", serverURL)
	fmt.Printf("  ./bin/inventory-cli --server %s bmc list\n", serverURL)
}
