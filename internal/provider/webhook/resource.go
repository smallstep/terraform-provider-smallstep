package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20230301.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = webhookTypeName
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20230301.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20230301.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	provisionerID := state.ProvisionerID.ValueString()
	authorityID := state.AuthorityID.ValueString()
	idOrName := state.ID.ValueString()
	if idOrName == "" {
		idOrName = state.Name.ValueString()
	}
	httpResp, err := r.client.GetWebhook(ctx, authorityID, provisionerID, idOrName, &v20230301.GetWebhookParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read webhook %s: %v", state.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading webhook %s: %s", httpResp.StatusCode, state.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	webhook := &v20230301.ProvisionerWebhook{}
	if err := json.NewDecoder(httpResp.Body).Decode(webhook); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal webhook %s: %v", state.ID.String(), err),
		)
		return
	}

	remote, d := fromAPI(ctx, webhook, req.State)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	remote.AuthorityID = types.StringValue(authorityID)
	remote.ProvisionerID = types.StringValue(provisionerID)

	tflog.Trace(ctx, fmt.Sprintf("read webhook %q resource", idOrName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("provisionerWebhook")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}
	basicAuth, basicAuthProps, err := utils.Describe("basicAuth")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provisioner_id": schema.StringAttribute{
				MarkdownDescription: props["provisioner_id"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authority_id"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: props["kind"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert_type": schema.StringAttribute{
				MarkdownDescription: props["certType"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_type": schema.StringAttribute{
				MarkdownDescription: props["serverType"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: props["url"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: props["secret"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Sensitive: true,
			},
			"collection_slug": schema.StringAttribute{
				MarkdownDescription: props["collectionSlug"],
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disable_tls_client_auth": schema.BoolAttribute{
				MarkdownDescription: props["disableTLSClientAuth"],
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: props["bearerToken"],
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Sensitive: true,
			},
			"basic_auth": schema.SingleNestedAttribute{
				MarkdownDescription: basicAuth,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: basicAuthProps["username"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Sensitive: true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: basicAuthProps["password"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Sensitive: true,
					},
				},
			},
		},
	}
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := toAPI(&plan)

	b, _ := json.Marshal(reqBody)
	tflog.Trace(ctx, string(b))

	authorityID := plan.AuthorityID.ValueString()
	provisionerID := plan.ProvisionerID.ValueString()

	httpResp, err := a.client.PostWebhooks(ctx, authorityID, provisionerID, &v20230301.PostWebhooksParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create webhook %q: %v", plan.Name.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	webhook := &v20230301.ProvisionerWebhook{}
	if err := json.NewDecoder(httpResp.Body).Decode(webhook); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal webhook %q: %v", plan.Name.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, webhook, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AuthorityID = types.StringValue(authorityID)
	state.ProvisionerID = types.StringValue(provisionerID)

	tflog.Trace(ctx, fmt.Sprintf("create webhook %q resource", plan.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update not supported. All changes require replacement.
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nameOrID := state.ID.ValueString()
	if nameOrID == "" {
		nameOrID = state.Name.ValueString()
	}
	httpResp, err := a.client.DeleteProvisioner(ctx, state.AuthorityID.ValueString(), nameOrID, &v20230301.DeleteProvisionerParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete provisioner %s: %v", state.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d deleting provisioner %s: %s", httpResp.StatusCode, state.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Import ID must be "<authority_id>/<provisioner_id>/<name>"`,
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("authority_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("provisioner_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
}
