package credential

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
	credential, props, err := utils.Describe("credential")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Schema",
			err.Error(),
		)
		return
	}

	cert, certProps, err := utils.Describe("credentialCertificate")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Certificate Schema",
			err.Error(),
		)
		return
	}

	policy, policyProps, err := utils.Describe("policyMatchCriteria")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Policy Schema",
			err.Error(),
		)
		return
	}

	files, filesProps, err := utils.Describe("credentialFiles")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Files Schema",
			err.Error(),
		)
		return
	}

	x509, _, err := utils.Describe("x509Fields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	_, certFieldProps, err := utils.Describe("certificateField")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	_, certFieldListProps, err := utils.Describe("certificateFieldList")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	name := schema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.StringAttribute{
				MarkdownDescription: certFieldProps["static"],
				Optional:            true,
			},
			"device_metadata": schema.StringAttribute{
				MarkdownDescription: certFieldProps["deviceMetadata"],
				Optional:            true,
			},
		},
	}

	nameList := schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.ListAttribute{
				MarkdownDescription: certFieldListProps["static"],
				ElementType:         types.StringType,
				Optional:            true,
			},
			"device_metadata": schema.ListAttribute{
				MarkdownDescription: certFieldListProps["deviceMetadata"],
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}

	key, keyProps, err := utils.Describe("credentialKey")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Key Info Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: credential,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
			},
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: cert,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"common_name":         name,
							"sans":                nameList,
							"organization":        nameList,
							"organizational_unit": nameList,
							"locality":            nameList,
							"country":             nameList,
							"province":            nameList,
							"street_address":      nameList,
							"postal_code":         nameList,
						},
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certProps["duration"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							// If unset the duration will default to 24h.
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"authority_id": schema.StringAttribute{
						MarkdownDescription: certProps["authorityID"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"key": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: key,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: keyProps["type"],
					},
					"pub_file": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: keyProps["pubFile"],
					},
					"protection": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: keyProps["protection"],
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIf(
								func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
									if isAttested(req.StateValue) != isAttested(req.PlanValue) {
										resp.RequiresReplace = true
									}
								},
								"If the key protection changes to/from attested, the credential must be replaced.",
								"If the key protection changes to/from attested, the credential must be replaced.",
							),
						},
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: policy,
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"assurance": schema.ListAttribute{
						MarkdownDescription: policyProps["assurance"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"ownership": schema.ListAttribute{
						MarkdownDescription: policyProps["ownership"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"os": schema.ListAttribute{
						MarkdownDescription: policyProps["operatingSystem"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"source": schema.ListAttribute{
						MarkdownDescription: policyProps["source"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: policyProps["tags"],
						ElementType:         types.StringType,
						Optional:            true,
					},
				},
			},
			"files": schema.SingleNestedAttribute{
				MarkdownDescription: files,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"root_file": schema.StringAttribute{
						MarkdownDescription: filesProps["rootFile"],
						Optional:            true,
					},
					"crt_file": schema.StringAttribute{
						MarkdownDescription: filesProps["crtFile"],
						Optional:            true,
					},
					"key_file": schema.StringAttribute{
						MarkdownDescription: filesProps["keyFile"],
						Optional:            true,
					},
					"key_format": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: filesProps["keyFormat"],
					},
					"uid": schema.Int64Attribute{
						MarkdownDescription: filesProps["uid"],
						Optional:            true,
					},
					"gid": schema.Int64Attribute{
						MarkdownDescription: filesProps["gid"],
						Optional:            true,
					},
					"mode": schema.Int64Attribute{
						MarkdownDescription: filesProps["mode"],
						Optional:            true,
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
	state := &CredentialModel{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialID := state.ID.ValueString()
	if credentialID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Credential Request",
			"Credential ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetCredential(ctx, credentialID, &v20250101.GetCredentialParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read credential %q: %v", credentialID, err),
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
			fmt.Sprintf("Request %q received status %d reading credential %s: %s", reqID, httpResp.StatusCode, credentialID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	credential := &v20250101.Credential{}
	if err := json.NewDecoder(httpResp.Body).Decode(credential); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal credential %s: %v", credentialID, err),
		)
		return
	}

	remote := fromAPI(ctx, &resp.Diagnostics, credential, req.State)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &CredentialModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := toAPI(ctx, &resp.Diagnostics, plan)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PostCredentials(ctx, &v20250101.PostCredentialsParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create credential: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating credential: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	credential := &v20250101.Credential{}
	if err := json.NewDecoder(httpResp.Body).Decode(credential); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal credential: %v", err),
		)
		return
	}

	model := fromAPI(ctx, &resp.Diagnostics, credential, req.Plan)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &CredentialModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	credentialID := plan.ID.ValueString()

	reqBody := toAPI(ctx, &resp.Diagnostics, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, err := r.client.PutCredential(ctx, credentialID, &v20250101.PutCredentialParams{}, reqBody)
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
			fmt.Sprintf("Request %q received status %d updating credential: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	credential := &v20250101.Credential{}
	if err := json.NewDecoder(httpResp.Body).Decode(credential); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse credential update response: %v", err),
		)
		return
	}

	model := fromAPI(ctx, &resp.Diagnostics, credential, req.Plan)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &CredentialModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialID := state.ID.ValueString()
	if credentialID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Credential Request",
			"Credential ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteCredential(ctx, credentialID, &v20250101.DeleteCredentialParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete credential: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting credential: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
