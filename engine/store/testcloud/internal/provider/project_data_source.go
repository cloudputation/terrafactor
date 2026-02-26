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
var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	client *http.Client
}

// ProjectDataSourceModel describes the data source data model.
type ProjectDataSourceModel struct {
	Id types.String `tfsdk:"id"`
	Budget types.Float64 `tfsdk:"budget"`
	CreatedAt types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	Enabled types.Bool `tfsdk:"enabled"`
	Name types.String `tfsdk:"name"`
	OrgId types.String `tfsdk:"org_id"`
	Region types.String `tfsdk:"region"`
	Tags types.Map `tfsdk:"tags"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Project resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
			},
			"budget": schema.Float64Attribute{
				MarkdownDescription: "Monthly spend budget in USD. Zero means unlimited.",
				Optional:  true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional project description.",
				Optional:  true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the project is active.",
				Optional:  true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable project name.",
				Required:  true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Organization this project belongs to.",
				Required:  true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Deployment region.",
				Optional:  true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Arbitrary key-value tags.",
				Optional:  true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "read project data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
