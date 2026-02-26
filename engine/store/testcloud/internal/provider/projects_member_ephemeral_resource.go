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
var _ ephemeral.EphemeralResource = &ProjectsMemberEphemeralResource{}

func NewProjectsMemberEphemeralResource() ephemeral.EphemeralResource {
	return &ProjectsMemberEphemeralResource{}
}

// ProjectsMemberEphemeralResource defines the ephemeral resource implementation.
type ProjectsMemberEphemeralResource struct{}

// ProjectsMemberEphemeralResourceModel describes the ephemeral resource data model.
type ProjectsMemberEphemeralResourceModel struct {
	Id types.String `tfsdk:"id"`
	Accepted types.Bool `tfsdk:"accepted"`
	CreatedAt types.String `tfsdk:"created_at"`
	Email types.String `tfsdk:"email"`
	InvitedBy types.String `tfsdk:"invited_by"`
	ProjectId types.String `tfsdk:"project_id"`
	Role types.String `tfsdk:"role"`
}

func (r *ProjectsMemberEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_member"
}

func (r *ProjectsMemberEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Manages a ProjectsMember resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
			},
			"accepted": schema.BoolAttribute{
				MarkdownDescription: "Whether the invitation has been accepted.",
				Computed:  true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:  true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Member email address.",
				Required:  true,
			},
			"invited_by": schema.StringAttribute{
				MarkdownDescription: "ID of the member who sent the invitation.",
				Computed:  true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project this member belongs to.",
				Computed:  true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Access role within the project.",
				Required:  true,
			},
		},
	}
}

func (r *ProjectsMemberEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ProjectsMemberEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call to retrieve ephemeral projects_member values
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to open projects_member, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
