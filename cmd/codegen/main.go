package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/openchami/inventory/pkg/codegen"
	"github.com/openchami/inventory/pkg/resources/bmc"
	"github.com/openchami/inventory/pkg/resources/boot"
	"github.com/openchami/inventory/pkg/resources/fru"
	"github.com/openchami/inventory/pkg/resources/node"
)

func main() {
	var (
		outputDir   = flag.String("output", "./generated", "Output directory for generated code")
		packageName = flag.String("package", "main", "Package name for generated code")
		modulePath  = flag.String("module", "github.com/openchami/inventory", "Go module path")
		genType     = flag.String("type", "all", "Type of code to generate: all, server, client, client-cmd, storage")
		runTidy     = flag.Bool("tidy", true, "Run go mod tidy after generation")
	)
	flag.Parse()

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create generator
	generator := codegen.NewGenerator(*outputDir, *packageName, *modulePath)

	// Register resource types for code generation
	if err := generator.RegisterResource(&bmc.BMC{}); err != nil {
		log.Fatalf("Failed to register BMC resource: %v", err)
	}

	if err := generator.RegisterResource(&node.Node{}); err != nil {
		log.Fatalf("Failed to register Node resource: %v", err)
	}

	if err := generator.RegisterResource(&fru.FRU{}); err != nil {
		log.Fatalf("Failed to register FRU resource: %v", err)
	}

	if err := generator.RegisterResource(&boot.BootConfiguration{}); err != nil {
		log.Fatalf("Failed to register BootConfiguration resource: %v", err)
	}

	// Enable authentication for sensitive resources
	if err := generator.EnableAuthForResource("BMC"); err != nil {
		log.Fatalf("Failed to enable auth for BMC: %v", err)
	}
	if err := generator.EnableAuthForResource("Node"); err != nil {
		log.Fatalf("Failed to enable auth for Node: %v", err)
	}

	// Generate code based on type
	switch *genType {
	case "server":
		if err := generator.LoadTemplates(); err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}
		if err := generator.GenerateHandlers(); err != nil {
			log.Fatalf("Failed to generate handlers: %v", err)
		}
		if err := generator.GenerateRoutes(); err != nil {
			log.Fatalf("Failed to generate routes: %v", err)
		}
		if err := generator.GenerateModels(); err != nil {
			log.Fatalf("Failed to generate models: %v", err)
		}
		if err := generator.GeneratePolicies(); err != nil {
			log.Fatalf("Failed to generate policies: %v", err)
		}
		if err := generator.GenerateOpenAPI(); err != nil {
			log.Fatalf("Failed to generate OpenAPI: %v", err)
		}
	case "client":
		if err := generator.LoadTemplates(); err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}
		if err := generator.GenerateClient(); err != nil {
			log.Fatalf("Failed to generate client: %v", err)
		}
		if err := generator.GenerateModels(); err != nil {
			log.Fatalf("Failed to generate models: %v", err)
		}
	case "storage":
		if err := generator.LoadTemplates(); err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}
		if err := generator.GenerateStorage(); err != nil {
			log.Fatalf("Failed to generate storage: %v", err)
		}
	case "client-cmd":
		if err := generator.LoadTemplates(); err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}
		if err := generator.GenerateClientCmd(); err != nil {
			log.Fatalf("Failed to generate client-cmd: %v", err)
		}
	case "all":
		if err := generator.GenerateAll(); err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}
	default:
		log.Fatalf("Unknown generation type: %s", *genType)
	}

	fmt.Printf("Successfully generated %s code in %s\n", *genType, *outputDir)

	// Run go mod tidy if requested
	if *runTidy {
		fmt.Println("Running go mod tidy...")
		if err := runGoModTidy(); err != nil {
			log.Printf("Warning: go mod tidy failed: %v", err)
		} else {
			fmt.Println("go mod tidy completed successfully")
		}
	}
}

func runGoModTidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = "." // Run in project root
	return cmd.Run()
}
