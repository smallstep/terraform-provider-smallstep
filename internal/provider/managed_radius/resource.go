package managed_radius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *v20250101.Client
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	radius, props, err := utils.Describe("managedRadius")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Managed RADIUS Schema",
			err.Error(),
		)
		return
	}

	replyAttrs, replyAttrsProps, err := utils.Describe("replyAttribute")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Reply Attributes Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: radius,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				// TODO remove once this field is documented in the API spec
				MarkdownDescription: utils.Default(props["id"], "The UUID of this managed RADIUS server."),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				// TODO remove once this field is documented in the API spec
				MarkdownDescription: utils.Default(props["name"], "A descriptive name for this resource. Must be unique across the team."),

				Required: true,
			},
			"nas_ips": schema.ListAttribute{
				// TODO remove once this field is documented in the API spec
				MarkdownDescription: utils.Default(props["nasIPs"], "The ip addresses the Network Access Server (NAS) may connect to the RADIUS server from."),
				Required:            true,
				ElementType:         types.StringType,
			},
			"client_ca": schema.StringAttribute{
				MarkdownDescription: props["clientCA"],
				Required:            true,
			},
			"reply_attributes": schema.ListNestedAttribute{
				MarkdownDescription: replyAttrs,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: replyAttrsProps["name"],
						},
						"value": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: replyAttrsProps["value"],
						},
						"value_from_certificate": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: replyAttrsProps["valueFromCertificate"],
						},
					},
				},
			},

			"server_ca": schema.StringAttribute{
				MarkdownDescription: props["serverCA"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: props["serverIP"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_port": schema.StringAttribute{
				MarkdownDescription: props["serverPort"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_hostname": schema.StringAttribute{
				MarkdownDescription: props["serverHostname"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = name
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id string
	diags := req.State.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Managed RADIUS Request",
			"ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetManagedRadius(ctx, id, &v20250101.GetManagedRadiusParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read managed radius %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading managed radius %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	radius := &v20250101.ManagedRadius{}
	if err := json.NewDecoder(httpResp.Body).Decode(radius); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal managed radius %s: %v", id, err),
		)
		return
	}

	remote := fromAPI(ctx, &resp.Diagnostics, radius, req.State)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &ManagedRadiusModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PostManagedRadius(ctx, &v20250101.PostManagedRadiusParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create managed radius: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating managed radius: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	radius := &v20250101.ManagedRadius{}
	if err := json.NewDecoder(httpResp.Body).Decode(radius); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal managed radius: %v", err),
		)
		return
	}

	model := fromAPI(ctx, &resp.Diagnostics, radius, req.Plan)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &ManagedRadiusModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	id := plan.ID.ValueString()

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, err := r.client.PutManagedRadius(ctx, id, &v20250101.PutManagedRadiusParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating managed radius: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	radius := &v20250101.ManagedRadius{}
	if err := json.NewDecoder(httpResp.Body).Decode(radius); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse managed radius update response: %v", err),
		)
		return
	}

	model := fromAPI(ctx, &resp.Diagnostics, radius, req.Plan)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id string
	diags := req.State.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Managed RADIUS Request",
			"ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteManagedRadius(ctx, id, &v20250101.DeleteManagedRadiusParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete managed radius: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting managed radius: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
