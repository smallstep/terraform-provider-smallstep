package relay

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
	relay, props, err := utils.DescribeV20260501("relay")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Relay Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: relay,
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
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: props["hostname"],
				Required:            true,
			},
			"ca_chain": schema.StringAttribute{
				MarkdownDescription: props["caChain"],
				Required:            true,
			},
			"allowed_targets": schema.ListAttribute{
				MarkdownDescription: props["allowedTargets"],
				ElementType:         types.StringType,
				Optional:            true,
			},
			"issuing_authority_id": schema.StringAttribute{
				MarkdownDescription: props["issuingAuthorityID"],
				Optional:            true,
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
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = clients.V20260501
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiRelay, _ := plan.toAPI(ctx)

	httpResp, err := r.client.PostRelays(ctx, &v20260501.PostRelaysParams{}, *apiRelay)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to create relay: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating relay: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	relay := &v20260501.Relay{}
	if err := json.NewDecoder(httpResp.Body).Decode(relay); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal relay: %v", err))
		return
	}

	model, _ := fromAPI(ctx, relay)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	relayID := state.ID.ValueString()
	if relayID == "" {
		resp.Diagnostics.AddError("Invalid Read Relay Request", "Relay ID is required")
		return
	}

	httpResp, err := r.client.GetRelay(ctx, relayID, &v20260501.GetRelayParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to read relay: %v", err))
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
			fmt.Sprintf("Request %q received status %d reading relay: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	relay := &v20260501.Relay{}
	if err := json.NewDecoder(httpResp.Body).Decode(relay); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal relay: %v", err))
		return
	}

	model, _ := fromAPI(ctx, relay)
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

	relayID := state.ID.ValueString()
	if relayID == "" {
		resp.Diagnostics.AddError("Invalid Update Relay Request", "Relay ID is required")
		return
	}

	apiRelay, _ := plan.toAPI(ctx)

	httpResp, err := r.client.PutRelay(ctx, relayID, &v20260501.PutRelayParams{}, *apiRelay)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to update relay: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating relay: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	relay := &v20260501.Relay{}
	if err := json.NewDecoder(httpResp.Body).Decode(relay); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal relay: %v", err))
		return
	}

	model, _ := fromAPI(ctx, relay)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	relayID := state.ID.ValueString()
	if relayID == "" {
		resp.Diagnostics.AddError("Invalid Delete Relay Request", "Relay ID is required")
		return
	}

	httpResp, err := r.client.DeleteRelay(ctx, relayID, &v20260501.DeleteRelayParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to delete relay: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting relay: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
