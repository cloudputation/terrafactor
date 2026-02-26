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
var _ datasource.DataSource = &ProjectsMemberDataSource{}

func NewProjectsMemberDataSource() datasource.DataSource {
	return &ProjectsMemberDataSource{}
}

// ProjectsMemberDataSource defines the data source implementation.
type ProjectsMemberDataSource struct {
	client *http.Client
}

// ProjectsMemberDataSourceModel describes the data source data model.
type ProjectsMemberDataSourceModel struct {
	Id types.String `tfsdk:"id"`
	Accepted types.Bool `tfsdk:"accepted"`
	CreatedAt types.String `tfsdk:"created_at"`
	Email types.String `tfsdk:"email"`
	InvitedBy types.String `tfsdk:"invited_by"`
	ProjectId types.String `tfsdk:"project_id"`
	Role types.String `tfsdk:"role"`
}

func (d *ProjectsMemberDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects_member"
}

func (d *ProjectsMemberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

func (d *ProjectsMemberDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsMemberDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement API call
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read projects_member, got error: %s", err))
	//     return
	// }

	tflog.Trace(ctx, "read projects_member data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
