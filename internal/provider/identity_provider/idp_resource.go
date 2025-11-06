package identity_provider

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

var _ resource.ResourceWithImportState = (*IdentityProviderResource)(nil)

func NewIdentityProviderResource() resource.Resource {
	return &IdentityProviderResource{}
}

type IdentityProviderResource struct {
	client *v20250101.Client
}

func (r *IdentityProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	idp, props, err := utils.Describe("identityProvider")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Identity Provider Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: idp,

		Attributes: map[string]schema.Attribute{
			"trust_roots": schema.StringAttribute{
				MarkdownDescription: props["trustRoots"],
				Required:            true,
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: props["issuer"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"authorize_endpoint": schema.StringAttribute{
				MarkdownDescription: props["authorizeEndpoint"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token_endpoint": schema.StringAttribute{
				MarkdownDescription: props["tokenEndpoint"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"jwks_endpoint": schema.StringAttribute{
				MarkdownDescription: props["jwksEndpoint"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IdentityProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = idp_name
}

// Configure adds the Smallstep API client to the resource.
func (r *IdentityProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	httpResp, err := r.client.GetIdentityProvider(ctx, &v20250101.GetIdentityProviderParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read identity provider: %v", err),
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
			fmt.Sprintf("Request %q received status %d reading identity provider: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	idp := &v20250101.IdentityProvider{}
	if err := json.NewDecoder(httpResp.Body).Decode(idp); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal identity provider: %v", err),
		)
		return
	}

	remote := idpFromAPI(idp)

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (a *IdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &IdentityProviderModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := idpToAPI(plan)

	// Check for conflicts since each team has a single identity provider and we
	// don't want to transparently overwrite trust roots. Force the user to
	// import their identity provider before modifying it.
	getResp, err := a.client.GetIdentityProvider(ctx, &v20250101.GetIdentityProviderParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to check for identity provider: %v", err),
		)
		return
	}
	defer getResp.Body.Close()
	switch {
	case getResp.StatusCode == 200:
		resp.Diagnostics.AddError(
			"Identity Provider Conflict",
			"Team already has an identity provider. Import it to manage trust roots.",
		)
		return
	case getResp.StatusCode == 404:
		// ok
	default:
		resp.Diagnostics.AddError(
			"Smallstep API Error",
			fmt.Sprintf("Failed to check for identity provider: %d: %s", getResp.StatusCode, utils.APIErrorMsg(getResp.Body)),
		)
		return
	}

	putResp, err := a.client.PutIdentityProvider(ctx, &v20250101.PutIdentityProviderParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create identity provider: %v", err),
		)
		return
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK {
		reqID := putResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating identity provider: %s", reqID, putResp.StatusCode, utils.APIErrorMsg(putResp.Body)),
		)
		return
	}

	idp := &v20250101.IdentityProvider{}
	if err := json.NewDecoder(putResp.Body).Decode(idp); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal idp: %v", err),
		)
		return
	}

	model := idpFromAPI(idp)
	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *IdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &IdentityProviderModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	reqBody := idpToAPI(plan)

	httpResp, err := r.client.PutIdentityProvider(ctx, &v20250101.PutIdentityProviderParams{}, reqBody)
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
			fmt.Sprintf("Request %q received status %d updating identity provider: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	idp := &v20250101.IdentityProvider{}
	if err := json.NewDecoder(httpResp.Body).Decode(idp); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse identity provider update response: %v", err),
		)
		return
	}

	model := idpFromAPI(idp)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *IdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResp, err := r.client.DeleteIdentityProvider(ctx, &v20250101.DeleteIdentityProviderParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete identity provider: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting identity provider: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *IdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	diags := resp.State.SetAttribute(ctx, path.Root("issuer"), "")
	resp.Diagnostics.Append(diags...)
}
