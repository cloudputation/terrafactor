package main

import (
	"fmt"
	"os"

	"github.com/cloudputation/terrafactor/packages/cli"
	"github.com/cloudputation/terrafactor/packages/config"
	log "github.com/cloudputation/terrafactor/packages/logger"
)

func main() {
	fmt.Printf("INFO: Starting terrafactor..\n\n")

	// Load main configuration file
	err := config.LoadConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure log directory exists before initializing logger
	if err := os.MkdirAll(config.AppConfig.LogDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging system
	err = log.InitLogger(config.AppConfig.LogDir, "info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logs: %v\n", err)
		os.Exit(1)
	}
	defer log.CloseLogger()

	// Run CLI
	rootCmd := cli.SetupRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Error executing command: %v", err)
	}
}
