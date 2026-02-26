// Copyright Cloudputation, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectsEnvironmentDataSource{}

func NewProjectsEnvironmentDataSource() datasource.DataSource {
	return &ProjectsEnvironmentDataSource{}
}

// ProjectsEnvironmentDataSource defines the data source implementation.
type ProjectsEnvironmentDataSource struct {
	client *http.Client
}

// ProjectsEnvironmentDataSourceModel describes the data source data model.
type ProjectsEnvironmentDataSourceModel struct {
	Id types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Name types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Protected types.Bool `tfsdk:"protected"`
	Slug types.String `tfsdk:"slug"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Variables types.Map `tfsdk:"variables"`
}

func (d *ProjectsEnvironmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_environment"
}

func (d *ProjectsEnvironmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ProjectsEnvironment resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Environment name (e.g. production, staging).",
				Required:  true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project this environment belongs to.",
				Required:  true,
			},
			"protected": schema.BoolAttribute{
				MarkdownDescription: "Whether the environment requires approval to deploy.",
				Optional:  true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe identifier derived from name.",
				Computed:  true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"variables": schema.MapAttribute{
				MarkdownDescription: "Environment variable key-value pairs.",
				Optional:  true,
			},
		},
	}
}

func (d *ProjectsEnvironmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ProjectsEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsEnvironmentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_environment, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "read projects_environment data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
