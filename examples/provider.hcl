# provider.hcl — provider ownership, release metadata, and API definition reference

provider {
  name             = "testcloud"
  org              = "cloudputation"
  registry         = "registry.terraform.io"
  go_module_prefix = "github.com"

  # path to OpenAPI 3.0 definition (relative to terrafactor root)
  schema {
    spec = "examples/api.yaml"
  }

  release {
    version = "0.1.0"
    license = "MPL-2.0"
  }

  author {
    name   = "Cloudputation, Inc."
    email  = "engineering@cloudputation.io"
    github = "cloudputation"
  }
}
