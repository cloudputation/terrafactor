package terrafactor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// installBinary runs go mod tidy + go install for the provider at providerDir,
// verifies the binary landed in GOBIN, then merges the dev_overrides entry into
// ~/.terraformrc. Called by lifecycle.go.
func installBinary(providerDir, registryAddress string) error {
	log.Info("Running go mod tidy", "dir", providerDir)
	if err := runGoCommand(providerDir, "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	log.Info("Running go install", "dir", providerDir)
	if err := runGoCommand(providerDir, "install"); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	gobin, err := resolveGOBIN()
	if err != nil {
		return err
	}

	binaryName, err := resolveBinaryName(providerDir)
	if err != nil {
		return err
	}

	binaryPath := filepath.Join(gobin, binaryName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("expected binary not found at %s", binaryPath)
	}

	log.Info("Provider binary installed", "path", binaryPath)

	if err := writeTerraformRC(registryAddress, gobin); err != nil {
		return fmt.Errorf("failed to update ~/.terraformrc: %w", err)
	}

	return nil
}

// runGoCommand runs `go <args>` with the given working directory, streaming output to stdout/stderr.
func runGoCommand(dir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// resolveGOBIN returns $GOBIN if set, otherwise $GOPATH/bin.
func resolveGOBIN() (string, error) {
	if v := os.Getenv("GOBIN"); v != "" {
		return v, nil
	}

	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", fmt.Errorf("could not determine GOPATH: %w", err)
	}
	gopath := strings.TrimSpace(string(out))
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	return filepath.Join(gopath, "bin"), nil
}

// resolveBinaryName reads go.mod in the provider directory and returns the
// basename of the module path (e.g. "terraform-provider-testcloud").
func resolveBinaryName(providerDir string) (string, error) {
	gomodPath := filepath.Join(providerDir, "go.mod")
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", gomodPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimPrefix(line, "module ")
			module = strings.TrimSpace(module)
			return filepath.Base(module), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module directive not found in %s", gomodPath)
}

// writeTerraformRC merges the dev_override entry for registryAddress into the
// existing ~/.terraformrc without touching any other content.
//
// Strategy (text-based, avoids pulling in an HCL library):
//  1. If the file does not exist → write a minimal one from scratch.
//  2. If a line already matching `"<registryAddress>"` exists inside
//     dev_overrides → replace that line in-place.
//  3. If a dev_overrides block exists but the address is absent → insert a
//     new line just before the closing `}` of that block.
//  4. If a provider_installation block exists but no dev_overrides → insert
//     a dev_overrides block at the top of provider_installation.
//  5. If none of the above → append a provider_installation block.
func writeTerraformRC(registryAddress, gobin string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rcPath := filepath.Join(home, ".terraformrc")

	entry := fmt.Sprintf("    %q = %q", registryAddress, gobin)

	// File does not exist — write from scratch.
	raw, err := os.ReadFile(rcPath)
	if os.IsNotExist(err) {
		content := fmt.Sprintf("provider_installation {\n  dev_overrides {\n%s\n  }\n\n  direct {}\n}\n", entry)
		if werr := os.WriteFile(rcPath, []byte(content), 0644); werr != nil {
			return werr
		}
		log.Info("Created ~/.terraformrc", "registry_address", registryAddress)
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(raw), "\n")

	// Pass 1: replace an existing entry for this registry address.
	addrKey := fmt.Sprintf("%q", registryAddress)
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, addrKey) {
			lines[i] = entry
			if werr := os.WriteFile(rcPath, []byte(strings.Join(lines, "\n")), 0644); werr != nil {
				return werr
			}
			log.Info("Updated ~/.terraformrc entry", "registry_address", registryAddress)
			return nil
		}
	}

	// Pass 2: insert inside an existing dev_overrides block.
	inDevOverrides := false
	depth := 0
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !inDevOverrides {
			if strings.HasPrefix(trimmed, "dev_overrides") && strings.Contains(trimmed, "{") {
				inDevOverrides = true
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
				lines = append(lines[:i], append([]string{entry}, lines[i:]...)...)
				if werr := os.WriteFile(rcPath, []byte(strings.Join(lines, "\n")), 0644); werr != nil {
					return werr
				}
				log.Info("Added entry to existing dev_overrides in ~/.terraformrc", "registry_address", registryAddress)
				return nil
			}
		}
	}

	// Pass 3: insert a dev_overrides block inside an existing provider_installation block.
	devOverridesBlock := fmt.Sprintf("  dev_overrides {\n%s\n  }\n", entry)
	inProviderInstall := false
	depth = 0
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !inProviderInstall {
			if strings.HasPrefix(trimmed, "provider_installation") && strings.Contains(trimmed, "{") {
				inProviderInstall = true
				depth = 1
				lines = append(lines[:i+1], append([]string{devOverridesBlock}, lines[i+1:]...)...)
				if werr := os.WriteFile(rcPath, []byte(strings.Join(lines, "\n")), 0644); werr != nil {
					return werr
				}
				log.Info("Added dev_overrides block to ~/.terraformrc", "registry_address", registryAddress)
				return nil
			}
		}
		_ = inProviderInstall
		_ = depth
	}

	// Pass 4: no provider_installation at all — append.
	appendBlock := fmt.Sprintf("\nprovider_installation {\n  dev_overrides {\n%s\n  }\n\n  direct {}\n}\n", entry)
	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, werr := f.WriteString(appendBlock); werr != nil {
		return werr
	}
	log.Info("Appended provider_installation block to ~/.terraformrc", "registry_address", registryAddress)
	return nil
}
