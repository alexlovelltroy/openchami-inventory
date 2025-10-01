// Package main provides a demonstration of the multi-version support capability.
//
// This command demonstrates how to register multiple schema versions for the same
// resource type and test version negotiation.
//
// Usage:
//
//	go run cmd/version-demo/main.go
//
// Or build and run:
//
//	go build -o bin/version-demo ./cmd/version-demo
//	./bin/version-demo
package main

import (
	"fmt"
	"os"

	"github.com/openchami/inventory/pkg/versioning"
)

func main() {
	fmt.Println("OpenCHAMI Inventory - Multi-Version Support Demonstration")
	fmt.Println("=" + "=========================================================")
	fmt.Println()

	if err := versioning.DemoMultiVersionSupport(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running demo: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Demo completed successfully!")
}
