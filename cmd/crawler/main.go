package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish"

	"github.com/openchami/inventory/pkg/crawler"
	"github.com/openchami/inventory/pkg/resources"
	"github.com/openchami/inventory/pkg/resources/fru"
)

var (
	cfgFile  string
	hostname string
	username string
	password string
	output   string
	insecure bool
	timeout  int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crawler",
	Short: "OpenCHAMI Inventory BMC Crawler",
	Long: `A Redfish BMC crawler that discovers and collects hardware inventory data
from Baseboard Management Controllers (BMCs) using the Redfish API.

Examples:
  # Crawl a BMC and output JSON to stdout
  crawler --hostname 192.168.1.100 --username admin --password secret

  # Save results to a file
  crawler --hostname bmc.example.com --username admin --password secret --output inventory.json

  # Use environment variables
  export BMC_HOSTNAME=192.168.1.100
  export BMC_USERNAME=admin  
  export BMC_PASSWORD=secret
  crawler`,
	RunE: runCrawler,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.crawler.yaml)")

	// Local flags
	rootCmd.Flags().StringVarP(&hostname, "hostname", "H", "", "BMC hostname or IP address (required)")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "BMC username (required)")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "BMC password (required)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	rootCmd.Flags().BoolVarP(&insecure, "insecure", "k", true, "skip TLS certificate verification")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "connection timeout in seconds")

	// Mark required flags
	rootCmd.MarkFlagRequired("hostname")
	rootCmd.MarkFlagRequired("username")
	rootCmd.MarkFlagRequired("password")

	// Bind flags to viper
	viper.BindPFlag("hostname", rootCmd.Flags().Lookup("hostname"))
	viper.BindPFlag("username", rootCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	viper.BindPFlag("insecure", rootCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout"))

	// Environment variable support
	viper.SetEnvPrefix("BMC")
	viper.AutomaticEnv()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".crawler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".crawler")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func runCrawler(cmd *cobra.Command, args []string) error {
	// Get configuration values (flags override config file override env vars)
	bmcHostname := viper.GetString("hostname")
	bmcUsername := viper.GetString("username")
	bmcPassword := viper.GetString("password")
	outputFile := viper.GetString("output")
	skipTLS := viper.GetBool("insecure")

	// Validate required parameters
	if bmcHostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if bmcUsername == "" {
		return fmt.Errorf("username is required")
	}
	if bmcPassword == "" {
		return fmt.Errorf("password is required")
	}

	// Prepare endpoint URL
	endpoint := bmcHostname
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	fmt.Fprintf(os.Stderr, "Connecting to BMC at %s...\n", endpoint)

	// Create Redfish client configuration
	config := gofish.ClientConfig{
		Endpoint:  endpoint,
		Username:  bmcUsername,
		Password:  bmcPassword,
		Insecure:  skipTLS,
		BasicAuth: true,
	}

	// Connect to the Redfish service
	client, err := gofish.Connect(config)
	if err != nil {
		return fmt.Errorf("failed to connect to BMC: %w", err)
	}
	defer func() {
		client.Logout()
	}()

	fmt.Fprintf(os.Stderr, "Successfully connected. Crawling hardware inventory...\n")

	// Generate UIDs for BMC and Node (if we had them)
	bmcUID, err := resources.GenerateUIDForResource("BMC")
	if err != nil {
		return fmt.Errorf("failed to generate BMC UID: %w", err)
	}

	nodeUID, err := resources.GenerateUIDForResource("Node")
	if err != nil {
		return fmt.Errorf("failed to generate Node UID: %w", err)
	}

	// Use the crawler to collect FRU data
	startTime := time.Now()
	fruSpecs, err := crawler.CoerceAll(client, bmcUID, nodeUID)
	if err != nil {
		return fmt.Errorf("failed to crawl inventory: %w", err)
	}
	crawlDuration := time.Since(startTime)

	fmt.Fprintf(os.Stderr, "Crawl completed in %v. Found %d FRUs.\n", crawlDuration, len(fruSpecs))

	// Create inventory snapshot
	snapshot := &fru.FRUInventorySnapshot{
		Resource: resources.Resource{
			APIVersion:    "v1",
			Kind:          "FRUInventorySnapshot",
			SchemaVersion: "1.0",
		},
		Spec: fru.FRUInventorySnapshotSpec{
			SnapshotTime: time.Now().Format(time.RFC3339),
			Source:       "redfish-crawler",
			Scope:        "bmc",
			ScopeUID:     bmcUID,
			FRUIDs:       make([]string, len(fruSpecs)),
		},
		Status: fru.FRUInventorySnapshotStatus{
			Complete:       true,
			FRUCount:       len(fruSpecs),
			NewFRUs:        len(fruSpecs),
			ProcessingTime: crawlDuration.Seconds(),
		},
	}

	// Initialize snapshot metadata
	snapshotUID, err := resources.GenerateUIDForResource("FRUInventorySnapshot")
	if err != nil {
		return fmt.Errorf("failed to generate snapshot UID: %w", err)
	}
	snapshot.Metadata.Initialize(fmt.Sprintf("snapshot-%s", bmcHostname), snapshotUID)
	snapshot.SetLabel("source", "crawler")
	snapshot.SetLabel("bmc-hostname", bmcHostname)
	snapshot.SetAnnotation("crawler.command", fmt.Sprintf("%s %s", os.Args[0], strings.Join(os.Args[1:], " ")))

	// Set initial status
	resources.SetCondition(&snapshot.Status.Conditions, "Complete", "True", "CrawlComplete", "Inventory crawl completed successfully")

	// Create FRU resources from specs
	frus := make([]*fru.FRU, len(fruSpecs))
	for i, spec := range fruSpecs {
		fruUID, err := resources.GenerateUIDForResource("FRU")
		if err != nil {
			return fmt.Errorf("failed to generate FRU UID: %w", err)
		}

		fru := &fru.FRU{
			Resource: resources.Resource{
				APIVersion:    "v1",
				Kind:          "FRU",
				SchemaVersion: "1.0",
			},
			Spec: spec,
		}

		// Generate name from type and location info
		name := fmt.Sprintf("%s", spec.FRUType)
		if spec.Location.Socket != "" {
			name = fmt.Sprintf("%s-%s", spec.FRUType, spec.Location.Socket)
		} else if spec.Location.Slot != "" {
			name = fmt.Sprintf("%s-%s", spec.FRUType, spec.Location.Slot)
		} else if spec.Location.Bay != "" {
			name = fmt.Sprintf("%s-%s", spec.FRUType, spec.Location.Bay)
		} else if spec.SerialNumber != "" {
			name = fmt.Sprintf("%s-%s", spec.FRUType, spec.SerialNumber)
		} else {
			name = fmt.Sprintf("%s-%d", spec.FRUType, i)
		}

		fru.Metadata.Initialize(name, fruUID)
		fru.SetLabel("source", "redfish-crawler")
		fru.SetLabel("bmc", bmcUID)
		fru.SetLabel("node", nodeUID)
		fru.SetLabel("snapshot", snapshotUID)

		// Set discovered status
		resources.SetCondition(&fru.Status.Conditions, "Discovered", "True", "RedfishCrawl", "FRU discovered via Redfish crawl")

		frus[i] = fru
		snapshot.Spec.FRUIDs[i] = fruUID
	}

	// Create output structure
	output := struct {
		Snapshot *fru.FRUInventorySnapshot `json:"snapshot"`
		FRUs     []*fru.FRU                `json:"frus"`
	}{
		Snapshot: snapshot,
		FRUs:     frus,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Output results
	if outputFile == "" {
		// Output to stdout
		fmt.Print(string(jsonData))
	} else {
		// Output to file
		err := os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Results written to %s\n", outputFile)
	}

	return nil
}

func main() {
	Execute()
}
