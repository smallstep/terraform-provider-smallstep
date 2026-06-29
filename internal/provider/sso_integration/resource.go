package sso_integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *v20260501.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = typeName
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	integration, props, err := utils.DescribeV20260501("ssoIntegration")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI SSO Integration Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: integration,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"redirect_uri": schema.StringAttribute{
				MarkdownDescription: props["redirectURI"],
				Required:            true,
			},
			"lifecycle_failure_uri": schema.StringAttribute{
				MarkdownDescription: props["lifecycleFailureURI"],
				Optional:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: props["secret"],
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"store_secret": schema.BoolAttribute{
				MarkdownDescription: "Whether to store the secret in terraform state when it is created. The secret cannot be recovered later.",
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace(), boolplanmodifier.UseStateForUnknown()},
			},
			"write_secret_file": schema.StringAttribute{
				MarkdownDescription: "If non-empty the secret will be written to this filepath when it is created. The secret cannot be recovered later.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*clientset.Clients)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	r.client = clients.V20260501
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Extract all config including store_secret and write_secret_file
	type planWithSecretConfig struct {
		ID                   types.String `tfsdk:"id"`
		RedirectURI          types.String `tfsdk:"redirect_uri"`
		LifecycleFailureURI  types.String `tfsdk:"lifecycle_failure_uri"`
		Secret               types.String `tfsdk:"secret"`
		StoreSecret          types.Bool   `tfsdk:"store_secret"`
		WriteSecretFile      types.String `tfsdk:"write_secret_file"`
	}

	var fullPlan planWithSecretConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &fullPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract store_secret and write_secret_file
	storeSecret := fullPlan.StoreSecret.ValueBool()
	writeSecretFile := fullPlan.WriteSecretFile.ValueString()

	// Convert to the actual model (without secret config fields)
	plan := &Model{
		ID:                  fullPlan.ID,
		RedirectURI:         fullPlan.RedirectURI,
		LifecycleFailureURI: fullPlan.LifecycleFailureURI,
		Secret:              fullPlan.Secret,
	}

	apiIntegration, _ := plan.toAPI(ctx)
	httpResp, err := r.client.PostSsoIntegrations(ctx, &v20260501.PostSsoIntegrationsParams{}, *apiIntegration)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to create sso integration: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating sso integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	integration := &v20260501.SsoIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(integration); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal sso integration: %v", err))
		return
	}

	model, _ := fromAPI(ctx, integration)
	// Only set store_secret if it was explicitly provided
	if !fullPlan.StoreSecret.IsNull() {
		model.StoreSecret = types.BoolValue(storeSecret)
	} else {
		model.StoreSecret = types.BoolNull()
	}
	// Only set write_secret_file if it was explicitly provided
	if !fullPlan.WriteSecretFile.IsNull() {
		model.WriteSecretFile = types.StringValue(writeSecretFile)
	} else {
		model.WriteSecretFile = types.StringNull()
	}
	// Handle secret: preserve if store_secret is true, otherwise null
	if storeSecret && integration.Secret != nil {
		model.Secret = types.StringValue(*integration.Secret)
	} else {
		model.Secret = types.StringNull()
	}
	if writeSecretFile != "" && integration.Secret != nil {
		if err := os.WriteFile(writeSecretFile, []byte(*integration.Secret), 0600); err != nil {
			resp.Diagnostics.AddError("Write secret to file", err.Error())
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := state.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError("Invalid Read SSO Integration Request", "SSO Integration ID is required")
		return
	}

	httpResp, err := r.client.GetSsoIntegration(ctx, integrationID, &v20260501.GetSsoIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to read sso integration: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading sso integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	integration := &v20260501.SsoIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(integration); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal sso integration: %v", err))
		return
	}

	model, _ := fromAPI(ctx, integration)
	model.Secret = state.Secret
	model.StoreSecret = state.StoreSecret
	model.WriteSecretFile = state.WriteSecretFile

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := state.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError("Invalid Update SSO Integration Request", "SSO Integration ID is required")
		return
	}

	apiIntegration, _ := plan.toAPI(ctx)
	httpResp, err := r.client.PutSsoIntegration(ctx, integrationID, &v20260501.PutSsoIntegrationParams{}, *apiIntegration)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to update sso integration: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating sso integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	integration := &v20260501.SsoIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(integration); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal sso integration: %v", err))
		return
	}

	model, _ := fromAPI(ctx, integration)
	model.Secret = state.Secret

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := state.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError("Invalid Delete SSO Integration Request", "SSO Integration ID is required")
		return
	}

	httpResp, err := r.client.DeleteSsoIntegration(ctx, integrationID, &v20260501.DeleteSsoIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to delete sso integration: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting sso integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
