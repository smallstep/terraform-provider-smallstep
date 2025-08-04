package integration

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
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_device_integration"

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *v20250101.Client
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	integration, props, err := utils.Describe("deviceInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	jamf, jamfProps, err := utils.Describe("jamfInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Jamf Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	intune, intuneProps, err := utils.Describe("intuneInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Intune Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	// The objects can be left empty and will be populated with default values.
	// The plan modifier UseStateForUnknown prevents showing (known after apply)
	// for these.
	resp.Schema = schema.Schema{
		MarkdownDescription: integration,
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
				Optional:            true,
			},
			"jamf": schema.SingleNestedAttribute{
				MarkdownDescription: jamf,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						MarkdownDescription: jamfProps["client_id"],
						Optional:            true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: jamfProps["client_secret"],
						Optional:            true,
					},
					"tenant_url": schema.StringAttribute{
						MarkdownDescription: jamfProps["tenant_url"],
						Required:            true,
					},
				},
			},
			"intune": schema.SingleNestedAttribute{
				MarkdownDescription: intune,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: intuneProps["app_id"],
						Required:            true,
					},
					"app_secret": schema.StringAttribute{
						MarkdownDescription: intuneProps["app_secret"],
						Required:            true,
					},
					"azure_tenant_name": schema.StringAttribute{
						MarkdownDescription: intuneProps["azure_tenant_name"],
						Required:            true,
					},
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
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := state.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Device Inventory Integration Request",
			"Strategy ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetDeviceInventoryIntegration(ctx, integrationID, &v20250101.GetDeviceInventoryIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device inventory integration %q: %v", integrationID, err),
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
			fmt.Sprintf("Request %q received status %d reading device inventory integration %s: %s", reqID, httpResp.StatusCode, integrationID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	inventory := &v20250101.DeviceInventoryIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(inventory); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device inventory integration %s: %v", integrationID, err),
		)
		return
	}

	remote, d := fromAPI(ctx, inventory, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &Model{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inventory, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	reqBody := v20250101.DeviceInventoryIntegrationRequest{
		Configuration: v20250101.DeviceInventoryIntegrationRequest_Configuration(inventory.Configuration),
		Kind:          inventory.Kind,
		Name:          inventory.Name,
	}

	httpResp, err := r.client.PostDeviceInventoryIntegrations(ctx, &v20250101.PostDeviceInventoryIntegrationsParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create device inventory integration: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating device inventory integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	inventory = &v20250101.DeviceInventoryIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(inventory); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device inventory integration: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, inventory, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &Model{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	strategyID := plan.ID.ValueString()

	inventory, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, err := r.client.PatchDeviceInventoryIntegration(ctx, strategyID, &v20250101.PatchDeviceInventoryIntegrationParams{}, *inventory)
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
			fmt.Sprintf("Request %q received status %d updating device inventory integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	inventory = &v20250101.DeviceInventoryIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(inventory); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse device inventory integration update response: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, inventory, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &Model{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := state.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Device Inventory Integration Request",
			"Strategy ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteDeviceInventoryIntegration(ctx, integrationID, &v20250101.DeleteDeviceInventoryIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete device inventory integration: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting device inventory integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
