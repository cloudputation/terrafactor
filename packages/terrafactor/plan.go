package terrafactor

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"go.yaml.in/yaml/v3"

	log "github.com/cloudputation/terrafactor/packages/logger"
)

// ---------------------------------------------------------------------------
// HCL types
// ---------------------------------------------------------------------------

// ProviderSpec holds the parsed contents of a provider.hcl manifest.
type ProviderSpec struct {
	Provider ProviderBlock `hcl:"provider,block"`
}

type ProviderBlock struct {
	Name           string       `hcl:"name"`
	Org            string       `hcl:"org"`
	Registry       string       `hcl:"registry"`
	GoModulePrefix string       `hcl:"go_module_prefix"`
	Schema         SchemaBlock  `hcl:"schema,block"`
	Release        ReleaseBlock `hcl:"release,block"`
	Author         AuthorBlock  `hcl:"author,block"`
}

type SchemaBlock struct {
	Spec string `hcl:"spec"`
}

type ReleaseBlock struct {
	Version string `hcl:"version"`
	License string `hcl:"license"`
}

type AuthorBlock struct {
	Name   string `hcl:"name"`
	Email  string `hcl:"email"`
	Github string `hcl:"github"`
}

// ---------------------------------------------------------------------------
// Resource types
// ---------------------------------------------------------------------------

// ResourceSpec describes one API resource extracted from the OpenAPI spec.
type ResourceSpec struct {
	ResourceName        string
	ResourceNamePascal  string
	ResourceDescription string
	Fields              []ResourceField
}

// ResourceField describes a single attribute of a resource.
type ResourceField struct {
	Name        string // snake_case tfsdk name
	GoName      string // PascalCase Go struct field name
	GoType      string // e.g. "types.String"
	TFType      string // e.g. "schema.StringAttribute"
	Required    bool
	Optional    bool
	Computed    bool
	Sensitive   bool
	Description string
}

// TemplateData is the dot-context passed to every .tmpl file.
type TemplateData struct {
	ProviderName        string
	ProviderNamePascal  string
	ResourceName        string          // set per-resource during generation
	ResourceNamePascal  string          // set per-resource during generation
	ResourceDescription string          // set per-resource during generation
	Fields              []ResourceField // set per-resource during generation
	Resources           []ResourceSpec  // all resources — used by provider.go.tmpl
	GoModule            string
	RegistryAddress     string
	OrgName             string
	GithubOwner         string
	CopyrightHolder     string
	License             string
	APISpec             string
}

// ---------------------------------------------------------------------------
// OpenAPI internal types
// ---------------------------------------------------------------------------

type openAPIDoc struct {
	Paths      map[string]openAPIPathItem `yaml:"paths"`
	Components openAPIComponents          `yaml:"components"`
}

type openAPIPathItem struct {
	Get    *openAPIOperation `yaml:"get"`
	Post   *openAPIOperation `yaml:"post"`
	Put    *openAPIOperation `yaml:"put"`
	Delete *openAPIOperation `yaml:"delete"`
}

type openAPIOperation struct {
	OperationID string                     `yaml:"operationId"`
	RequestBody *openAPIRequestBody        `yaml:"requestBody"`
	Responses   map[string]openAPIResponse `yaml:"responses"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `yaml:"content"`
}

type openAPIResponse struct {
	Content map[string]openAPIMediaType `yaml:"content"`
}

type openAPIMediaType struct {
	Schema openAPISchemaRef `yaml:"schema"`
}

type openAPIComponents struct {
	Schemas map[string]openAPISchema `yaml:"schemas"`
}

type openAPISchema struct {
	Ref                  string                   `yaml:"$ref"`
	Type                 string                   `yaml:"type"`
	Format               string                   `yaml:"format"`
	Description          string                   `yaml:"description"`
	ReadOnly             bool                     `yaml:"readOnly"`
	WriteOnly            bool                     `yaml:"writeOnly"`
	Nullable             bool                     `yaml:"nullable"`
	Required             []string                 `yaml:"required"`
	Properties           map[string]openAPISchema `yaml:"properties"`
	Items                *openAPISchema            `yaml:"items"`
	AdditionalProperties *openAPISchema            `yaml:"additionalProperties"`
	Enum                 []interface{}            `yaml:"enum"`
}

type openAPISchemaRef struct {
	Ref    string `yaml:"$ref"`
	Type   string `yaml:"type"`
	Format string `yaml:"format"`
}

var paramSegment = regexp.MustCompile(`^\{.+\}$`)

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

// parseAndBuild parses the provider.hcl and OpenAPI spec, returning the
// TemplateData and ResourceSpec list used by all engine operations.
func parseAndBuild(rootDir, providerHCLPath string) (TemplateData, []ResourceSpec, error) {
	spec, err := parseProviderHCL(providerHCLPath)
	if err != nil {
		return TemplateData{}, nil, fmt.Errorf("failed to parse provider.hcl: %w", err)
	}

	if spec.Provider.Name == "" {
		return TemplateData{}, nil, fmt.Errorf("provider name is required in provider.hcl")
	}

	p := spec.Provider
	log.Info("Loaded provider manifest", "path", providerHCLPath)
	log.Info("Provider config",
		"name", p.Name,
		"org", p.Org,
		"registry", p.Registry,
		"go_module_prefix", p.GoModulePrefix,
	)
	log.Info("Schema", "spec", p.Schema.Spec)
	log.Info("Release", "version", p.Release.Version, "license", p.Release.License)
	log.Info("Author", "name", p.Author.Name, "email", p.Author.Email, "github", p.Author.Github)

	resources, err := ParseOpenAPI(p.Schema.Spec)
	if err != nil {
		return TemplateData{}, nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}
	log.Info("OpenAPI parsed", "spec", p.Schema.Spec, "resources", len(resources))
	for _, r := range resources {
		log.Info("Resource found", "name", r.ResourceName, "fields", len(r.Fields))
	}

	data := buildTemplateData(spec, resources)
	log.Info("Derived values",
		"go_module", data.GoModule,
		"registry_address", data.RegistryAddress,
	)
	log.Info("Rendering as",
		"provider", data.ProviderName,
		"provider_pascal", data.ProviderNamePascal,
		"copyright_holder", data.CopyrightHolder,
		"license", data.License,
	)

	return data, resources, nil
}

// ---------------------------------------------------------------------------
// HCL parsing
// ---------------------------------------------------------------------------

// parseProviderHCL reads and decodes a provider.hcl file.
func parseProviderHCL(path string) (*ProviderSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL parse error: %s", diags)
	}

	var spec ProviderSpec
	diags = gohcl.DecodeBody(file.Body, nil, &spec)
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL decode error: %s", diags)
	}

	return &spec, nil
}

// buildTemplateData derives all template variables from the parsed spec and resource list.
func buildTemplateData(spec *ProviderSpec, resources []ResourceSpec) TemplateData {
	p := spec.Provider
	goModule := fmt.Sprintf("%s/%s/terraform-provider-%s", p.GoModulePrefix, p.Org, p.Name)
	registryAddr := fmt.Sprintf("%s/%s/%s", p.Registry, p.Org, p.Name)

	githubOwner := p.Author.Github
	if githubOwner == "" {
		githubOwner = p.Org
	}

	copyrightHolder := p.Author.Name
	if copyrightHolder == "" {
		copyrightHolder = p.Org
	}

	return TemplateData{
		ProviderName:       p.Name,
		ProviderNamePascal: pascal(p.Name),
		Resources:          resources,
		GoModule:           goModule,
		RegistryAddress:    registryAddr,
		OrgName:            p.Org,
		GithubOwner:        githubOwner,
		CopyrightHolder:    copyrightHolder,
		License:            p.Release.License,
		APISpec:            p.Schema.Spec,
	}
}

// pascal converts snake_case or hyphen-case to PascalCase.
func pascal(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// OpenAPI parsing
// ---------------------------------------------------------------------------

// ParseOpenAPI reads an OpenAPI 3.0 YAML spec and returns one ResourceSpec per collection path.
func ParseOpenAPI(specPath string) ([]ResourceSpec, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read OpenAPI spec %s: %w", specPath, err)
	}

	var doc openAPIDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("cannot parse OpenAPI spec %s: %w", specPath, err)
	}

	var collectionPaths []string
	for path := range doc.Paths {
		segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
		last := segs[len(segs)-1]
		if !paramSegment.MatchString(last) {
			collectionPaths = append(collectionPaths, path)
		}
	}

	sortStrings(collectionPaths)

	var resources []ResourceSpec
	for _, path := range collectionPaths {
		item := doc.Paths[path]

		name := pathToResourceName(path)
		schemaName := resolveSchemaName(path, item, &doc)
		if schemaName == "" {
			continue
		}

		schema, ok := doc.Components.Schemas[schemaName]
		if !ok {
			continue
		}

		fields := extractFields(&schema, &doc)

		resources = append(resources, ResourceSpec{
			ResourceName:        name,
			ResourceNamePascal:  pascal(name),
			ResourceDescription: fmt.Sprintf("Manages a %s resource.", pascal(name)),
			Fields:              fields,
		})
	}

	return resources, nil
}

func pathToResourceName(path string) string {
	segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var parts []string
	for _, seg := range segs {
		if paramSegment.MatchString(seg) {
			continue
		}
		parts = append(parts, strings.ReplaceAll(seg, "-", "_"))
	}
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		parts[len(parts)-1] = depluralize(last)
	}
	return strings.Join(parts, "_")
}

func depluralize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

func resolveSchemaName(path string, item openAPIPathItem, doc *openAPIDoc) string {
	if item.Post != nil && item.Post.RequestBody != nil {
		for _, mt := range item.Post.RequestBody.Content {
			if name := refToName(mt.Schema.Ref); name != "" {
				return name
			}
		}
	}

	if item.Get != nil {
		if resp, ok := item.Get.Responses["200"]; ok {
			for _, mt := range resp.Content {
				if name := refToName(mt.Schema.Ref); name != "" {
					if schema, exists := doc.Components.Schemas[name]; exists {
						for _, prop := range schema.Properties {
							if prop.Type == "array" && prop.Items != nil {
								if itemName := refToName(prop.Items.Ref); itemName != "" {
									return itemName
								}
							}
						}
					}
					return name
				}
			}
		}
	}

	return ""
}

func refToName(ref string) string {
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		return ref[idx+1:]
	}
	return ""
}

func extractFields(schema *openAPISchema, doc *openAPIDoc) []ResourceField {
	required := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		required[r] = true
	}

	var fields []ResourceField

	fields = append(fields, ResourceField{
		Name:        "id",
		GoName:      "Id",
		GoType:      "types.String",
		TFType:      "schema.StringAttribute",
		Computed:    true,
		Description: "Unique identifier.",
	})

	for name, prop := range schema.Properties {
		if name == "id" {
			continue
		}

		resolved := prop
		if prop.Ref != "" {
			schemaName := refToName(prop.Ref)
			if s, ok := doc.Components.Schemas[schemaName]; ok {
				resolved = s
			}
		}

		field := mapField(name, resolved, required[name])
		fields = append(fields, field)
	}

	sortFields(fields)
	return fields
}

func mapField(name string, prop openAPISchema, isRequired bool) ResourceField {
	goName := snakeToPascal(name)
	goType, tfType := mapType(prop)

	computed := prop.ReadOnly
	sensitive := prop.WriteOnly
	required := isRequired && !prop.ReadOnly
	optional := !required && !computed

	return ResourceField{
		Name:        name,
		GoName:      goName,
		GoType:      goType,
		TFType:      tfType,
		Required:    required,
		Optional:    optional,
		Computed:    computed,
		Sensitive:   sensitive,
		Description: prop.Description,
	}
}

func mapType(prop openAPISchema) (string, string) {
	switch prop.Type {
	case "boolean":
		return "types.Bool", "schema.BoolAttribute"
	case "integer":
		return "types.Int64", "schema.Int64Attribute"
	case "number":
		return "types.Float64", "schema.Float64Attribute"
	case "array":
		return "types.List", "schema.ListAttribute"
	case "object":
		if prop.AdditionalProperties != nil {
			return "types.Map", "schema.MapAttribute"
		}
		return "types.Object", "schema.SingleNestedAttribute"
	default:
		return "types.String", "schema.StringAttribute"
	}
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && ss[j] < ss[j-1]; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}

func sortFields(fields []ResourceField) {
	for i, f := range fields {
		if f.Name == "id" && i != 0 {
			fields[0], fields[i] = fields[i], fields[0]
			break
		}
	}
	rest := fields[1:]
	for i := 1; i < len(rest); i++ {
		for j := i; j > 0 && rest[j].Name < rest[j-1].Name; j-- {
			rest[j], rest[j-1] = rest[j-1], rest[j]
		}
	}
}
