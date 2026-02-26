// Copyright Cloudputation, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectsMemberResource{}
var _ resource.ResourceWithImportState = &ProjectsMemberResource{}

func NewProjectsMemberResource() resource.Resource {
	return &ProjectsMemberResource{}
}

// ProjectsMemberResource defines the resource implementation.
type ProjectsMemberResource struct {
	client *http.Client
}

// ProjectsMemberResourceModel describes the resource data model.
type ProjectsMemberResourceModel struct {
	Id types.String `tfsdk:"id"`
	Accepted types.Bool `tfsdk:"accepted"`
	CreatedAt types.String `tfsdk:"created_at"`
	Email types.String `tfsdk:"email"`
	InvitedBy types.String `tfsdk:"invited_by"`
	ProjectId types.String `tfsdk:"project_id"`
	Role types.String `tfsdk:"role"`
}

func (r *ProjectsMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_member"
}

func (r *ProjectsMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ProjectsMember resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *ProjectsMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ProjectsMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectsMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create projects_member, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "created projects_member resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectsMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_member, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectsMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update projects_member, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectsMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete projects_member, got error: %s", err))
	//     return
	// }
}

func (r *ProjectsMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
