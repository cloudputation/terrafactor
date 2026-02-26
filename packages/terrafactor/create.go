package terrafactor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

const (
	ScaffoldDir   = "engine/terraform-provider-scaffold"
	EngineBaseDir = "engine/store"
)

// perResourceTemplates lists scaffold-relative paths of templates that must be
// rendered once per resource (not once globally).
var perResourceTemplates = []string{
	"internal/provider/resource.go.tmpl",
	"internal/provider/resource_test.go.tmpl",
	"internal/provider/data_source.go.tmpl",
	"internal/provider/data_source_test.go.tmpl",
	"internal/provider/ephemeral_resource.go.tmpl",
	"internal/provider/ephemeral_resource_test.go.tmpl",
	"examples/resources/resource.tf.tmpl",
	"examples/resources/import.sh.tmpl",
	"examples/data-sources/data-source.tf.tmpl",
	"examples/ephemeral-resources/ephemeral-resource.tf.tmpl",
}

// scaffoldProvider renders the full scaffold into dst for the first time.
// Returns (rendered, copied, error).
func scaffoldProvider(src, dst string, data TemplateData, resources []ResourceSpec) (int, int, error) {
	return renderResources(src, dst, data, resources)
}

// renderResources runs the global renderDir pass and the per-resource template loop.
func renderResources(src, dst string, data TemplateData, resources []ResourceSpec) (int, int, error) {
	skipSet := make(map[string]bool, len(perResourceTemplates))
	for _, t := range perResourceTemplates {
		skipSet[filepath.ToSlash(t)] = true
	}

	rendered, copied, err := renderDir(src, dst, data, skipSet)
	if err != nil {
		return rendered, copied, err
	}

	for _, res := range resources {
		resData := data
		resData.ResourceName = res.ResourceName
		resData.ResourceNamePascal = res.ResourceNamePascal
		resData.ResourceDescription = res.ResourceDescription
		resData.Fields = res.Fields

		for _, tmplRel := range perResourceTemplates {
			tmplSrc := filepath.Join(src, filepath.FromSlash(tmplRel))
			outRel := resourceOutputPath(tmplRel, res.ResourceName)
			outDst := filepath.Join(dst, filepath.FromSlash(outRel))

			info, err := os.Stat(tmplSrc)
			if err != nil {
				return rendered, copied, fmt.Errorf("template not found: %s: %w", tmplSrc, err)
			}
			if err := renderTemplate(tmplSrc, outDst, resData, info.Mode()); err != nil {
				return rendered, copied, err
			}
			log.Info("Rendered resource file", "resource", res.ResourceName, "file", outRel)
			rendered++
		}
	}

	return rendered, copied, nil
}

// resourceOutputPath maps a scaffold-relative template path to the per-resource output path.
func resourceOutputPath(tmplRel, resourceName string) string {
	base := filepath.Base(tmplRel)
	dir := filepath.Dir(tmplRel)

	base = strings.TrimSuffix(base, ".tmpl")

	switch {
	case base == "resource.go":
		base = resourceName + "_resource.go"
	case base == "resource_test.go":
		base = resourceName + "_resource_test.go"
	case base == "data_source.go":
		base = resourceName + "_data_source.go"
	case base == "data_source_test.go":
		base = resourceName + "_data_source_test.go"
	case base == "ephemeral_resource.go":
		base = resourceName + "_ephemeral_resource.go"
	case base == "ephemeral_resource_test.go":
		base = resourceName + "_ephemeral_resource_test.go"
	case base == "resource.tf":
		dir = filepath.Join("examples", "resources", resourceName)
		base = "resource.tf"
	case base == "import.sh":
		dir = filepath.Join("examples", "resources", resourceName)
		base = "import.sh"
	case base == "data-source.tf":
		dir = filepath.Join("examples", "data-sources", resourceName)
		base = "data-source.tf"
	case base == "ephemeral-resource.tf":
		dir = filepath.Join("examples", "ephemeral-resources", resourceName)
		base = "ephemeral-resource.tf"
	}

	return filepath.Join(dir, base)
}

// renderDir walks src, skipping per-resource templates, renders/copies everything else.
func renderDir(src, dst string, data TemplateData, skipSet map[string]bool) (int, int, error) {
	var rendered, copied int

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		relSlash := filepath.ToSlash(rel)

		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), info.Mode())
		}

		if skipSet[relSlash] {
			return nil
		}

		if strings.HasSuffix(path, ".tmpl") {
			outPath := stripTmpl(filepath.Join(dst, rel))
			if err := renderTemplate(path, outPath, data, info.Mode()); err != nil {
				return err
			}
			log.Info("Rendered template", "file", stripTmpl(relSlash))
			rendered++
			return nil
		}

		if err := copyFile(path, filepath.Join(dst, rel), info.Mode()); err != nil {
			return err
		}
		log.Info("Copied file", "file", relSlash)
		copied++
		return nil
	})

	return rendered, copied, err
}

func stripTmpl(p string) string {
	return strings.TrimSuffix(p, ".tmpl")
}

// renderTemplate executes a Go text/template file and writes the result.
func renderTemplate(src, dst string, data TemplateData, mode os.FileMode) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(src)).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return fmt.Errorf("template parse error in %s: %w", src, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template execute error in %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, buf.Bytes(), mode)
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
