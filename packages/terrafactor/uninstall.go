package terrafactor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// uninstallBinary removes the provider binary and its dev_overrides entry from
// ~/.terraformrc. Called by lifecycle.go.
func uninstallBinary(binaryPath, registryAddress string) error {
	if err := os.Remove(binaryPath); err != nil {
		if os.IsNotExist(err) {
			log.Info("Binary not found, skipping removal", "path", binaryPath)
		} else {
			return fmt.Errorf("failed to remove binary %s: %w", binaryPath, err)
		}
	} else {
		log.Info("Removed provider binary", "path", binaryPath)
	}

	if err := removeFromTerraformRC(registryAddress); err != nil {
		return fmt.Errorf("failed to update ~/.terraformrc: %w", err)
	}

	return nil
}

// removeFromTerraformRC removes the dev_overrides entry for registryAddress from
// ~/.terraformrc. If the dev_overrides block becomes empty it is removed too.
// If provider_installation then becomes empty it is removed as well.
// Missing file or missing entry are treated as warnings, not errors.
func removeFromTerraformRC(registryAddress string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rcPath := filepath.Join(home, ".terraformrc")

	raw, err := os.ReadFile(rcPath)
	if os.IsNotExist(err) {
		log.Info("~/.terraformrc not found, nothing to clean up")
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(raw), "\n")
	addrKey := fmt.Sprintf("%q", registryAddress)

	// Step 1: drop the entry line.
	found := false
	filtered := make([]string, 0, len(lines))
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), addrKey) {
			found = true
			continue
		}
		filtered = append(filtered, l)
	}
	if !found {
		log.Info("Entry not present in ~/.terraformrc, nothing to remove", "registry_address", registryAddress)
		return nil
	}
	lines = filtered

	// Step 2: collapse empty dev_overrides blocks.
	lines = collapseEmptyBlock(lines, "dev_overrides")

	// Step 3: collapse empty provider_installation blocks.
	lines = collapseEmptyBlock(lines, "provider_installation")

	if err := os.WriteFile(rcPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return err
	}
	log.Info("Removed entry from ~/.terraformrc", "registry_address", registryAddress)
	return nil
}

// collapseEmptyBlock removes a named HCL block (e.g. "dev_overrides { ... }")
// from the lines slice if its body contains only blank lines and comments.
func collapseEmptyBlock(lines []string, blockName string) []string {
	start := -1
	depth := 0
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if start == -1 {
			if strings.HasPrefix(trimmed, blockName) && strings.Contains(trimmed, "{") {
				start = i
				depth = 1
				continue
			}
		} else {
			for _, ch := range trimmed {
				if ch == '{' {
					depth++
				} else if ch == '}' {
					depth--
				}
			}
			if depth == 0 {
				bodyEmpty := true
				for _, bodyLine := range lines[start+1 : i] {
					t := strings.TrimSpace(bodyLine)
					if t != "" && !strings.HasPrefix(t, "#") {
						bodyEmpty = false
						break
					}
				}
				if bodyEmpty {
					end := i + 1
					if end < len(lines) && strings.TrimSpace(lines[end]) == "" {
						end++
					}
					return append(lines[:start], lines[end:]...)
				}
				start = -1
				depth = 0
			}
		}
	}
	return lines
}
