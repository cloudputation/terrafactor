package terrafactor

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// BuildProvider scaffolds a new provider for the first time.
// Flow: plan → display diff (all +) → prompt → create → save state.
func BuildProvider(rootDir, providerHCLPath string) error {
	data, resources, err := parseAndBuild(rootDir, providerHCLPath)
	if err != nil {
		return err
	}

	src := filepath.Join(rootDir, ScaffoldDir)
	dst := filepath.Join(rootDir, EngineBaseDir, data.ProviderName)

	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("provider directory already exists: %s", dst)
	}

	PrintProviderDiff(data.ProviderName, map[string][]string{}, resources, os.Stdout)

	if !PromptApproval(map[string][]string{}, resources, os.Stdin, os.Stdout) {
		return nil
	}

	rendered, copied, err := scaffoldProvider(src, dst, data, resources)
	if err != nil {
		return err
	}

	log.Info("Scaffold complete",
		"rendered", rendered,
		"copied", copied,
		"total", rendered+copied,
		"resources", len(resources),
		"destination", dst,
	)

	sf := stateFilePath(rootDir, data.ProviderName)
	if err := SaveProviderState(sf, ProviderState{
		Provider:  data.ProviderName,
		Resources: resourceSpecToStateMap(resources),
	}); err != nil {
		log.Info("Warning: failed to save provider state", "error", err)
	}

	return nil
}

// PlanProvider shows what would change on re-render without writing any files.
// Flow: plan → load state → display diff → return.
func PlanProvider(rootDir, providerHCLPath string) error {
	data, resources, err := parseAndBuild(rootDir, providerHCLPath)
	if err != nil {
		return err
	}

	sf := stateFilePath(rootDir, data.ProviderName)
	prev := map[string][]string{}
	if existing, err := LoadProviderState(sf); err == nil {
		prev = existing.Resources
	}

	PrintProviderDiff(data.ProviderName, prev, resources, os.Stdout)

	return nil
}

// ApplyProvider re-renders provider files when api.yaml changes.
// Flow: plan → load state → display diff → prompt → update → save state.
func ApplyProvider(rootDir, providerHCLPath string) error {
	data, resources, err := parseAndBuild(rootDir, providerHCLPath)
	if err != nil {
		return err
	}

	sf := stateFilePath(rootDir, data.ProviderName)
	prev := map[string][]string{}
	if existing, err := LoadProviderState(sf); err == nil {
		prev = existing.Resources
	}

	PrintProviderDiff(data.ProviderName, prev, resources, os.Stdout)

	if _, err := os.Stat(sf); os.IsNotExist(err) {
		return fmt.Errorf("provider %q has not been built yet", data.ProviderName)
	}

	if !PromptApproval(prev, resources, os.Stdin, os.Stdout) {
		return nil
	}

	src := filepath.Join(rootDir, ScaffoldDir)
	dst := filepath.Join(rootDir, EngineBaseDir, data.ProviderName)

	rendered, copied, err := rerenderProvider(src, dst, data, resources)
	if err != nil {
		return err
	}

	log.Info("Apply complete",
		"rendered", rendered,
		"copied", copied,
		"total", rendered+copied,
		"resources", len(resources),
		"destination", dst,
	)

	if err := SaveProviderState(sf, ProviderState{
		Provider:  data.ProviderName,
		Resources: resourceSpecToStateMap(resources),
	}); err != nil {
		log.Info("Warning: failed to save provider state", "error", err)
	}

	return nil
}

// DestroyProvider removes the generated provider from engine/store/<name>.
// Flow: load state → display diff (all -) → prompt → delete.
func DestroyProvider(rootDir, name string) error {
	dst := filepath.Join(rootDir, EngineBaseDir, name)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return fmt.Errorf("provider %q not found in store", name)
	}

	sf := stateFilePath(rootDir, name)
	prev := map[string][]string{}
	if existing, err := LoadProviderState(sf); err == nil {
		prev = existing.Resources
	}

	PrintProviderDiff(name, prev, nil, os.Stdout)

	if !PromptApproval(prev, nil, os.Stdin, os.Stdout) {
		return nil
	}

	return removeProvider(dst)
}

// ProviderNameFromHCL parses a provider.hcl file and returns the provider name.
func ProviderNameFromHCL(providerHCLPath string) (string, error) {
	spec, err := parseProviderHCL(providerHCLPath)
	if err != nil {
		return "", err
	}
	if spec.Provider.Name == "" {
		return "", fmt.Errorf("provider name not set in %s", providerHCLPath)
	}
	return spec.Provider.Name, nil
}
