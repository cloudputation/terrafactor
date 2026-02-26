// Copyright Cloudputation, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TestcloudProvider satisfies various provider interfaces.
var _ provider.Provider = &TestcloudProvider{}
var _ provider.ProviderWithEphemeralResources = &TestcloudProvider{}

// TestcloudProvider defines the provider implementation.
type TestcloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TestcloudProviderModel describes the provider data model.
type TestcloudProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *TestcloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "testcloud"
	resp.Version = p.version
}

func (p *TestcloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint for the Testcloud provider",
				Optional:            true,
			},
		},
	}
}

func (p *TestcloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TestcloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TestcloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectsEnvironmentResource,
		NewProjectsMemberResource,
		NewProjectsWebhookResource,
	}
}

func (p *TestcloudProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewProjectEphemeralResource,
		NewProjectsEnvironmentEphemeralResource,
		NewProjectsMemberEphemeralResource,
		NewProjectsWebhookEphemeralResource,
	}
}

func (p *TestcloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectsEnvironmentDataSource,
		NewProjectsMemberDataSource,
		NewProjectsWebhookDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TestcloudProvider{
			version: version,
		}
	}
}
