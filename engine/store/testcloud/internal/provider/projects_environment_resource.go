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
var _ resource.Resource = &ProjectsEnvironmentResource{}
var _ resource.ResourceWithImportState = &ProjectsEnvironmentResource{}

func NewProjectsEnvironmentResource() resource.Resource {
	return &ProjectsEnvironmentResource{}
}

// ProjectsEnvironmentResource defines the resource implementation.
type ProjectsEnvironmentResource struct {
	client *http.Client
}

// ProjectsEnvironmentResourceModel describes the resource data model.
type ProjectsEnvironmentResourceModel struct {
	Id types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Name types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Protected types.Bool `tfsdk:"protected"`
	Slug types.String `tfsdk:"slug"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Variables types.Map `tfsdk:"variables"`
}

func (r *ProjectsEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_environment"
}

func (r *ProjectsEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ProjectsEnvironment resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *ProjectsEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectsEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectsEnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create projects_environment, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "created projects_environment resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectsEnvironmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_environment, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectsEnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update projects_environment, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectsEnvironmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete projects_environment, got error: %s", err))
	//     return
	// }
}

func (r *ProjectsEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
