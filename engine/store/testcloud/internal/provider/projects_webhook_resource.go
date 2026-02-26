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
var _ resource.Resource = &ProjectsWebhookResource{}
var _ resource.ResourceWithImportState = &ProjectsWebhookResource{}

func NewProjectsWebhookResource() resource.Resource {
	return &ProjectsWebhookResource{}
}

// ProjectsWebhookResource defines the resource implementation.
type ProjectsWebhookResource struct {
	client *http.Client
}

// ProjectsWebhookResourceModel describes the resource data model.
type ProjectsWebhookResourceModel struct {
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

func (r *ProjectsWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_webhook"
}

func (r *ProjectsWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ProjectsWebhook resource.",

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

func (r *ProjectsWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectsWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectsWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create projects_webhook, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "created projects_webhook resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectsWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_webhook, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectsWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update projects_webhook, got error: %s", err))
	//     return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectsWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectsWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete projects_webhook, got error: %s", err))
	//     return
	// }
}

func (r *ProjectsWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
