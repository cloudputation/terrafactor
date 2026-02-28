package terrafactor

import (
	"os"
	"path/filepath"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// rerenderProvider prunes files for removed resources then re-renders everything.
// Returns (rendered, copied, error).
func rerenderProvider(src, dst string, data TemplateData, resources []ResourceSpec) (int, int, error) {
	if err := pruneRemovedResources(dst, data.PrevResources, resources); err != nil {
		return 0, 0, err
	}
	return renderResources(src, dst, data, resources)
}

// pruneRemovedResources deletes all generated files for resources that existed
// in prev but are absent from next.
func pruneRemovedResources(dst string, prev map[string][]string, next []ResourceSpec) error {
	if len(prev) == 0 {
		return nil
	}

	nextSet := make(map[string]struct{}, len(next))
	for _, r := range next {
		nextSet[r.ResourceName] = struct{}{}
	}

	for name := range prev {
		if _, exists := nextSet[name]; exists {
			continue
		}

		// Flat files under internal/provider/
		flatFiles := []string{
			filepath.Join(dst, "internal", "provider", name+"_resource.go"),
			filepath.Join(dst, "internal", "provider", name+"_resource_test.go"),
			filepath.Join(dst, "internal", "provider", name+"_data_source.go"),
			filepath.Join(dst, "internal", "provider", name+"_data_source_test.go"),
			filepath.Join(dst, "internal", "provider", name+"_ephemeral_resource.go"),
			filepath.Join(dst, "internal", "provider", name+"_ephemeral_resource_test.go"),
		}
		for _, f := range flatFiles {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				return err
			}
			log.Info("Pruned resource file", "resource", name, "file", f)
		}

		// Example directories
		exampleDirs := []string{
			filepath.Join(dst, "examples", "resources", name),
			filepath.Join(dst, "examples", "data-sources", name),
			filepath.Join(dst, "examples", "ephemeral-resources", name),
		}
		for _, d := range exampleDirs {
			if err := os.RemoveAll(d); err != nil {
				return err
			}
			log.Info("Pruned resource directory", "resource", name, "dir", d)
		}
	}

	return nil
}
