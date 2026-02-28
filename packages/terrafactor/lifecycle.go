package terrafactor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// InstallProvider orchestrates building the provider binary and registering it
// in ~/.terraformrc so Terraform uses the local dev build.
func InstallProvider(rootDir, providerHCLPath string) error {
	data, _, err := parseAndBuild(rootDir, providerHCLPath)
	if err != nil {
		return err
	}

	providerDir := filepath.Join(rootDir, EngineBaseDir, data.ProviderName)
	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
		return fmt.Errorf("provider %q has not been built yet — run `terrafactor provider build` first", data.ProviderName)
	}

	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go binary not found on PATH: %w", err)
	}

	if err := installBinary(providerDir, data.RegistryAddress); err != nil {
		return err
	}

	gobin, _ := resolveGOBIN()
	binaryName, _ := resolveBinaryName(providerDir)
	log.Info("Installation complete",
		"binary", filepath.Join(gobin, binaryName),
		"registry_address", data.RegistryAddress,
		"terraformrc", "~/.terraformrc",
	)
	return nil
}

// UninstallProvider orchestrates removing the provider binary and its
// dev_overrides entry from ~/.terraformrc.
func UninstallProvider(rootDir, providerHCLPath string) error {
	data, _, err := parseAndBuild(rootDir, providerHCLPath)
	if err != nil {
		return err
	}

	providerDir := filepath.Join(rootDir, EngineBaseDir, data.ProviderName)

	gobin, err := resolveGOBIN()
	if err != nil {
		return err
	}

	binaryName, err := resolveBinaryName(providerDir)
	if err != nil {
		return err
	}

	binaryPath := filepath.Join(gobin, binaryName)

	if err := uninstallBinary(binaryPath, data.RegistryAddress); err != nil {
		return err
	}

	log.Info("Uninstall complete",
		"binary", binaryPath,
		"registry_address", data.RegistryAddress,
	)
	return nil
}
