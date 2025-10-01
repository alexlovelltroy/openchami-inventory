package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/go-fuego/fuego"
	"github.com/openchami/inventory/pkg/policies"
	"github.com/openchami/inventory/pkg/resources/bmc"
	bmcv2beta1 "github.com/openchami/inventory/pkg/resources/bmc/v2beta1"
	"github.com/openchami/inventory/pkg/resources/node"
	"github.com/openchami/inventory/pkg/versioning"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global version registry
var versionRegistry *versioning.VersionRegistry

// Server configuration
var (
	cfgFile         string
	host            string
	port            int
	corsEnabled     bool
	corsOrigins     []string
	logLevel        string
	storagePath     string
	disableOpenAPI  bool
	openAPIValidate bool
	disableAuth     bool // Disable authentication checks (testing only)
)

// initializeVersionRegistry sets up all resource versions
func initializeVersionRegistry() {
	versionRegistry = versioning.NewVersionRegistry()

	// Register BMC v1 (stable)
	bmcV1 := versioning.SchemaVersion{
		Version:    "v1",
		IsDefault:  true,
		Stability:  "stable",
		Deprecated: false,
		SpecType:   "bmc.BMCSpec",
		StatusType: "bmc.BMCStatus",
		TypeName:   "*bmc.BMC",
		Package:    "github.com/openchami/inventory/pkg/resources/bmc",
		Transforms: []string{},
	}

	bmcV1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(&bmc.BMC{}),
		Constructor: func() interface{} { return &bmc.BMC{} },
		Converter:   nil, // v1 doesn't need converter to itself
		Metadata:    bmcV1,
	}

	if err := versionRegistry.RegisterVersion("BMC", "v1", bmcV1TypeInfo); err != nil {
		log.Fatalf("Failed to register BMC v1: %v", err)
	}

	// Register BMC v2beta1 (beta - enhanced authentication)
	bmcV2Beta1 := versioning.SchemaVersion{
		Version:    "v2beta1",
		IsDefault:  false,
		Stability:  "beta",
		Deprecated: false,
		SpecType:   "bmcv2beta1.BMCSpec",
		StatusType: "bmcv2beta1.BMCStatus",
		TypeName:   "*bmcv2beta1.BMC",
		Package:    "github.com/openchami/inventory/pkg/resources/bmc/v2beta1",
		Transforms: []string{"ConvertV1ToV2Beta1", "ConvertV2Beta1ToV1"},
	}

	bmcV2Beta1TypeInfo := versioning.ResourceTypeInfo{
		Type:        reflect.TypeOf(&bmcv2beta1.BMC{}),
		Constructor: func() interface{} { return &bmcv2beta1.BMC{} },
		Converter:   bmcv2beta1.NewBMCConverter(),
		Metadata:    bmcV2Beta1,
	}

	if err := versionRegistry.RegisterVersion("BMC", "v2beta1", bmcV2Beta1TypeInfo); err != nil {
		log.Fatalf("Failed to register BMC v2beta1: %v", err)
	}

	log.Println("Version registry initialized:")
	log.Printf("  BMC: %v (default: %s)", versionRegistry.ListVersions("BMC"), versionRegistry.GetDefaultVersion("BMC"))
}

var rootCmd = &cobra.Command{
	Use:   "inventory-server",
	Short: "OpenCHAMI Inventory API Server",
	Long: `A REST API server for managing OpenCHAMI inventory resources.

This server provides endpoints for managing BMCs, Nodes, FRUs, and Boot Configurations
with support for authentication, authorization, and multi-version schema support.`,
	Run: runServer,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/inventory/server.yaml or $HOME/.inventory-server.yaml)")

	// Server configuration
	rootCmd.Flags().StringVar(&host, "host", "0.0.0.0", "server host address")
	rootCmd.Flags().IntVar(&port, "port", 8080, "server port")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")

	// CORS configuration
	rootCmd.Flags().BoolVar(&corsEnabled, "cors-enabled", false, "enable CORS support")
	rootCmd.Flags().StringSliceVar(&corsOrigins, "cors-origins", []string{"*"}, "allowed CORS origins")

	// Storage configuration
	rootCmd.Flags().StringVar(&storagePath, "storage-path", "./inventory", "path to storage directory")

	// OpenAPI configuration
	rootCmd.Flags().BoolVar(&disableOpenAPI, "disable-openapi", false, "disable OpenAPI documentation generation")
	rootCmd.Flags().BoolVar(&openAPIValidate, "openapi-validate", false, "enable strict OpenAPI schema validation (shows warnings)")

	// Security configuration
	rootCmd.Flags().BoolVar(&disableAuth, "disable-auth", false, "disable authentication checks (WARNING: for testing only)")

	// Bind flags to viper
	viper.BindPFlag("host", rootCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("cors.enabled", rootCmd.Flags().Lookup("cors-enabled"))
	viper.BindPFlag("cors.origins", rootCmd.Flags().Lookup("cors-origins"))
	viper.BindPFlag("storage.path", rootCmd.Flags().Lookup("storage-path"))
	viper.BindPFlag("openapi.disabled", rootCmd.Flags().Lookup("disable-openapi"))
	viper.BindPFlag("openapi.validate", rootCmd.Flags().Lookup("openapi-validate"))
	viper.BindPFlag("security.disable-auth", rootCmd.Flags().Lookup("disable-auth"))

	// Environment variable support
	viper.SetEnvPrefix("INVENTORY")
	viper.AutomaticEnv()
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		viper.AddConfigPath("/etc/inventory/")
		viper.AddConfigPath("$HOME/.inventory/")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("server")
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Get configuration from viper
	host := viper.GetString("host")
	port := viper.GetInt("port")
	logLevel := viper.GetString("log-level")
	corsEnabled := viper.GetBool("cors.enabled")
	corsOrigins := viper.GetStringSlice("cors.origins")
	storagePath := viper.GetString("storage.path")

	disableAuth := viper.GetBool("security.disable-auth")

	log.Printf("Server configuration:")
	log.Printf("  Host: %s", host)
	log.Printf("  Port: %d", port)
	log.Printf("  Log Level: %s", logLevel)
	log.Printf("  CORS Enabled: %v", corsEnabled)
	if corsEnabled {
		log.Printf("  CORS Origins: %v", corsOrigins)
	}
	log.Printf("  Storage Path: %s", storagePath)
	log.Printf("  Authentication: %v", !disableAuth)
	if disableAuth {
		log.Printf("  WARNING: Authentication is DISABLED - this is for testing only!")
	}

	// Initialize policy registry
	policyRegistry = policies.NewPolicyRegistry()

	if disableAuth {
		// Use permissive policy for all resources (testing only)
		permissivePolicy := policies.NewPermissivePolicy()
		policyRegistry.RegisterPolicy("BMC", permissivePolicy)
		policyRegistry.RegisterPolicy("Node", permissivePolicy)
		policyRegistry.RegisterPolicy("FRU", permissivePolicy)
		policyRegistry.RegisterPolicy("BootConfiguration", permissivePolicy)
		log.Printf("  Using permissive policy for all resources (no authentication)")
	} else {
		// Use default policies with authentication
		policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
		policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())
		log.Printf("  Using default policies with authentication required")
	}

	// Initialize version registry
	initializeVersionRegistry()

	// Create fuego server with custom options
	serverOptions := []func(*fuego.Server){
		fuego.WithAddr(fmt.Sprintf("%s:%d", host, port)),
	}

	// Add CORS if enabled
	if corsEnabled {
		log.Printf("Note: CORS configuration requested but needs custom middleware implementation")
		// TODO: Implement CORS middleware based on fuego's actual CORS support
		// For now, CORS can be handled by a reverse proxy (nginx, traefik, etc.)
	}

	s := fuego.NewServer(serverOptions...)

	// Add version negotiation middleware
	fuego.Use(s, versioning.VersionNegotiationMiddleware(versionRegistry))

	// Register generated routes
	RegisterGeneratedRoutes(s)

	// Add health check endpoint
	fuego.Get(s, "/health", func(c fuego.ContextNoBody) (map[string]string, error) {
		return map[string]string{
			"status":  "ok",
			"version": "1.0.0",
		}, nil
	})

	// Add version discovery endpoint
	fuego.Get(s, "/version-info", GetVersionInfo)

	log.Printf("Starting OpenCHAMI Inventory Server on %s:%d", host, port)
	log.Printf("Health check available at: http://%s:%d/health", host, port)
	log.Printf("Version info available at: http://%s:%d/version-info", host, port)

	if err := s.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
