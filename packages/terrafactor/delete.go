package terrafactor

import (
	"fmt"
	"os"
)

// removeProvider deletes the provider directory and all its contents.
func removeProvider(dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("failed to remove provider directory: %w", err)
	}
	return nil
}
