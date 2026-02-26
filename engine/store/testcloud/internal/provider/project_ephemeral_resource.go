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
var _ ephemeral.EphemeralResource = &ProjectEphemeralResource{}

func NewProjectEphemeralResource() ephemeral.EphemeralResource {
	return &ProjectEphemeralResource{}
}

// ProjectEphemeralResource defines the ephemeral resource implementation.
type ProjectEphemeralResource struct{}

// ProjectEphemeralResourceModel describes the ephemeral resource data model.
type ProjectEphemeralResourceModel struct {
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

func (r *ProjectEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
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

func (r *ProjectEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ProjectEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call to retrieve ephemeral project values
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to open project, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
