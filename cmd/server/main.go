package main

import (
	"log"

	"github.com/go-fuego/fuego"
	"github.com/openchami/inventory/pkg/policies"
	"github.com/openchami/inventory/pkg/resources/bmc"
	"github.com/openchami/inventory/pkg/resources/node"
)

func main() {
	// Initialize policy registry with default policies
	policyRegistry = policies.NewPolicyRegistry()
	policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
	policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())

	// Create fuego server
	s := fuego.NewServer()

	// Register generated routes
	RegisterGeneratedRoutes(s)

	// Add health check endpoint
	fuego.Get(s, "/health", func(c fuego.ContextNoBody) (map[string]string, error) {
		return map[string]string{"status": "ok"}, nil
	})

	log.Println("Starting server on :8080")
	s.Run()
}
