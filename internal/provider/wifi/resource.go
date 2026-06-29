package wifi

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
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
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
	wifi, props, err := utils.Describe("wifi")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Wifi Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: wifi,

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
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssid": schema.StringAttribute{
				MarkdownDescription: props["ssid"],
				Required:            true,
			},
			"radius_server_ca": schema.StringAttribute{
				MarkdownDescription: props["radiusServerCA"],
				Required:            true,
			},
			"radius_server_domain": schema.StringAttribute{
				MarkdownDescription: props["radiusServerDomain"],
				Optional:            true,
			},
			"autojoin": schema.BoolAttribute{
				MarkdownDescription: props["autojoin"],
				Optional:            true,
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: props["hidden"],
				Optional:            true,
			},
			"credentials": schema.SetAttribute{
				MarkdownDescription: props["credentials"],
				ElementType:         types.StringType,
				Optional:            true,
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

	clients, ok := req.ProviderData.(*clientset.Clients)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = clients.V20250101
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &WifiModel{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifiID := state.ID.ValueString()
	if wifiID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Wifi Request",
			"Wifi ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetWifi(ctx, wifiID, &v20250101.GetWifiParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read wifi %q: %v", wifiID, err),
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
			fmt.Sprintf("Request %q received status %d reading wifi %s: %s", reqID, httpResp.StatusCode, wifiID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	wifi := &v20250101.Wifi{}
	if err := json.NewDecoder(httpResp.Body).Decode(wifi); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal wifi %s: %v", wifiID, err),
		)
		return
	}

	remote := FromAPI(ctx, wifi, &resp.Diagnostics, req.State)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &WifiModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostWifi(ctx, &v20250101.PostWifiParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create wifi: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating wifi: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	wifi := &v20250101.Wifi{}
	if err := json.NewDecoder(httpResp.Body).Decode(wifi); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal wifi: %v", err),
		)
		return
	}

	model := FromAPI(ctx, wifi, &resp.Diagnostics, req.Plan)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &WifiModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	wifiID := plan.ID.ValueString()

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PutWifi(ctx, wifiID, &v20250101.PutWifiParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			err.Error(),
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
			fmt.Sprintf("Request %q received status %d updating wifi: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	wifi := &v20250101.Wifi{}
	if err := json.NewDecoder(httpResp.Body).Decode(wifi); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse wifi update response: %v", err),
		)
		return
	}

	model := FromAPI(ctx, wifi, &resp.Diagnostics, req.Plan)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &WifiModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifiID := state.ID.ValueString()
	if wifiID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Wifi Request",
			"Wifi ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteWifi(ctx, wifiID, &v20250101.DeleteWifiParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete wifi: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting wifi: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
