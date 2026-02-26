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
var _ datasource.DataSource = &ProjectsWebhookDataSource{}

func NewProjectsWebhookDataSource() datasource.DataSource {
	return &ProjectsWebhookDataSource{}
}

// ProjectsWebhookDataSource defines the data source implementation.
type ProjectsWebhookDataSource struct {
	client *http.Client
}

// ProjectsWebhookDataSourceModel describes the data source data model.
type ProjectsWebhookDataSourceModel struct {
	Id types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Enabled types.Bool `tfsdk:"enabled"`
	Events types.List `tfsdk:"events"`
	Name types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	RetryPolicy types.Object `tfsdk:"retry_policy"`
	Secret types.String `tfsdk:"secret"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Url types.String `tfsdk:"url"`
}

func (d *ProjectsWebhookDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_webhook"
}

func (d *ProjectsWebhookDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ProjectsWebhook resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:  true,
			},
			"events": schema.ListAttribute{
				MarkdownDescription: "Event types this webhook subscribes to.",
				Required:  true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Required:  true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"retry_policy": schema.SingleNestedAttribute{
				MarkdownDescription: "",
				Optional:  true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "HMAC signing secret. Write-only — not returned after creation.
",
				Optional:  true,
				Sensitive: true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "HTTPS endpoint that receives events.",
				Required:  true,
			},
		},
	}
}

func (d *ProjectsWebhookDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsWebhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsWebhookDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_webhook, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "read projects_webhook data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
