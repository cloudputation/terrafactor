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
var _ ephemeral.EphemeralResource = &ProjectsEnvironmentEphemeralResource{}

func NewProjectsEnvironmentEphemeralResource() ephemeral.EphemeralResource {
	return &ProjectsEnvironmentEphemeralResource{}
}

// ProjectsEnvironmentEphemeralResource defines the ephemeral resource implementation.
type ProjectsEnvironmentEphemeralResource struct{}

// ProjectsEnvironmentEphemeralResourceModel describes the ephemeral resource data model.
type ProjectsEnvironmentEphemeralResourceModel struct {
	Id types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Name types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Protected types.Bool `tfsdk:"protected"`
	Slug types.String `tfsdk:"slug"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Variables types.Map `tfsdk:"variables"`
}

func (r *ProjectsEnvironmentEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_environment"
}

func (r *ProjectsEnvironmentEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
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

func (r *ProjectsEnvironmentEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ProjectsEnvironmentEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call to retrieve ephemeral projects_environment values
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to open projects_environment, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
