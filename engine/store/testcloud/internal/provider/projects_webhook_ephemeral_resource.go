// Copyright Cloudputation, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ ephemeral.EphemeralResource = &ProjectsWebhookEphemeralResource{}

func NewProjectsWebhookEphemeralResource() ephemeral.EphemeralResource {
	return &ProjectsWebhookEphemeralResource{}
}

// ProjectsWebhookEphemeralResource defines the ephemeral resource implementation.
type ProjectsWebhookEphemeralResource struct{}

// ProjectsWebhookEphemeralResourceModel describes the ephemeral resource data model.
type ProjectsWebhookEphemeralResourceModel struct {
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

func (r *ProjectsWebhookEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_webhook"
}

func (r *ProjectsWebhookEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
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

func (r *ProjectsWebhookEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ProjectsWebhookEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call to retrieve ephemeral projects_webhook values
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to open projects_webhook, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
