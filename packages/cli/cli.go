package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudputation/terrafactor/packages/bootstrap"
	"github.com/cloudputation/terrafactor/packages/config"
	log "github.com/cloudputation/terrafactor/packages/logger"
	terrafactor "github.com/cloudputation/terrafactor/packages/terrafactor"
)

const defaultIndentLevel = 4

func SetupRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "terrafactor",
		Short: "Terrafactor - Terraform provider generator and output formatter",
	}
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(cmdPrint())
	rootCmd.AddCommand(cmdProvider())

	return rootCmd
}

func cmdPrint() *cobra.Command {
	return &cobra.Command{
		Use:   "print <json_file> <operation>",
		Short: "Print JSON file in Terraform-style output (operation: create|destroy)",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := bootstrap.BootstrapFileSystem()
			if err != nil {
				log.Fatal("Failed to bootstrap the filesystem: %v", err)
			}

			jsonFile := args[0]
			operationTag := args[1]
			indentStr := strings.Repeat(" ", defaultIndentLevel)

			dataBytes, err := os.ReadFile(jsonFile)
			if err != nil {
				log.Fatal("Error reading JSON file: %v", err)
			}

			var data interface{}
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				log.Fatal("Error parsing JSON: %v", err)
			}

			if err := terrafactor.Print(data, operationTag, indentStr, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error printing data: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func cmdProvider() *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Terraform provider operations",
	}

	buildCmd := &cobra.Command{
		Use:   "build <provider.hcl>",
		Short: "Scaffold a new Terraform provider from the template",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			providerHCL := args[0]
			log.Info("Building provider scaffold", "config", providerHCL)
			if err := terrafactor.BuildProvider(config.RootDir, providerHCL); err != nil {
				log.Info(err.Error())
			}
		},
	}

	planCmd := &cobra.Command{
		Use:   "plan <provider.hcl>",
		Short: "Show what would change on re-render without writing any files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			providerHCL := args[0]
			log.Info("Planning provider update", "config", providerHCL)
			if err := terrafactor.PlanProvider(config.RootDir, providerHCL); err != nil {
				log.Info(err.Error())
			}
		},
	}

	applyCmd := &cobra.Command{
		Use:   "apply <provider.hcl>",
		Short: "Re-render provider files and update state when api.yaml changes",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			providerHCL := args[0]
			log.Info("Applying provider update", "config", providerHCL)
			if err := terrafactor.ApplyProvider(config.RootDir, providerHCL); err != nil {
				log.Info(err.Error())
			}
		},
	}

	destroyCmd := &cobra.Command{
		Use:   "destroy <provider.hcl>",
		Short: "Remove a generated provider from engine/store so it can be rebuilt from scratch",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			providerHCL := args[0]
			name, err := terrafactor.ProviderNameFromHCL(providerHCL)
			if err != nil {
				log.Info(err.Error())
				return
			}
			if err := terrafactor.DestroyProvider(config.RootDir, name); err != nil {
				log.Info(err.Error())
				return
			}
			log.Info("Provider destroyed", "name", name)
		},
	}

	providerCmd.AddCommand(buildCmd)
	providerCmd.AddCommand(planCmd)
	providerCmd.AddCommand(applyCmd)
	providerCmd.AddCommand(destroyCmd)

	return providerCmd
}
