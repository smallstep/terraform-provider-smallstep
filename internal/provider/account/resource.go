package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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

// This value is stored in the private state to indicate an optional, computed
// value in state is computed. An optional, computed object is one that will be
// set to a default by the server if not set by the client, e.g. if the x509
// object is not set in a request it will be set in the reply with a default
// common name and sans. When the terraform client sets the object and then
// later removes it in a subsequent request, the object will be unknown. We
// can use the value from state, but only if it's the computed default value
// set by the server, which is what this flag tracks. This has to be valid json
// even though it's never parsed.
var computed = []byte(`{"computed": true}`)

const x509Private = "x509"

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	account, props, err := utils.Describe("account")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Account Schema",
			err.Error(),
		)
		return
	}

	wifi, wifiProps, err := utils.Describe("wifiAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WiFi Account Schema",
			err.Error(),
		)
		return
	}

	vpn, vpnProps, err := utils.Describe("vpnAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN Account Schema",
			err.Error(),
		)
		return
	}

	ike, ikeProps, err := utils.Describe("ikeV2Config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN IKEv2 Account Schema",
			err.Error(),
		)
		return
	}

	browser, _, err := utils.Describe("browserAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Browser Account Schema",
			err.Error(),
		)
		return
	}

	ethernet, ethernetProps, err := utils.Describe("ethernetAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Ethernet Account Schema",
			err.Error(),
		)
		return
	}

	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Certificate Schema",
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

	assurance, _, err := utils.Describe("deviceAssurance")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Assurance Schema",
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

	ssh, _, err := utils.Describe("sshFields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSH Certificate Schema",
			err.Error(),
		)
		return
	}

	name := schema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.StringAttribute{
				Optional: true,
			},
			"device_metadata": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	nameList := schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"device_metadata": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}

	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Key Info Schema",
			err.Error(),
		)
		return
	}

	reloadInfo, reloadInfoProps, err := utils.Describe("endpointReloadInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Reload Info Schema",
			err.Error(),
		)
		return
	}

	// The objects can be left empty and will be populated with default values.
	// The plan modifier UseStateForUnknown prevents showing (known after apply)
	// for these.
	resp.Schema = schema.Schema{
		MarkdownDescription: account,

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
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: certInfo,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							utils.MaybeUseStateForUnknown(path.Root("certificate").AtName("x509"), x509Private, computed),
						},
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
					"ssh": schema.SingleNestedAttribute{
						MarkdownDescription: ssh,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"key_id":     name,
							"principals": nameList,
						},
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certInfoProps["duration"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							// If unset the duration will default to 24h.
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"authority_id": schema.StringAttribute{
						MarkdownDescription: certInfoProps["authorityId"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
					"crt_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["crtFile"],
						Optional:            true,
					},
					"key_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["keyFile"],
						Optional:            true,
					},
					"root_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["rootFile"],
						Optional:            true,
					},
					"uid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["uid"],
						Optional:            true,
					},
					"gid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["gid"],
						Optional:            true,
					},
					"mode": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["mode"],
						Optional:            true,
					},
				},
			},
			"key": schema.SingleNestedAttribute{
				// TOOD still?
				// This object is not required by the API but a default object
				// will always be returned with format, type and protection set to
				// "DEFAULT". To avoid "inconsistent result after apply" errors
				// require these fields to be set explicitly in terraform.
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					// The key will always be returned and have default values set if not
					// supplied by the client. This prevents showing (known after apply) in the
					// plan.
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: keyInfo,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: keyInfoProps["type"],
					},
					"format": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: keyInfoProps["format"],
					},
					"pub_file": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: keyInfoProps["pubFile"],
					},
					"protection": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: keyInfoProps["protection"],
					},
				},
			},
			"reload": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: reloadInfo,
				PlanModifiers: []planmodifier.Object{
					// Reload will always be returned and have default values set if not
					// supplied by the client. This prevents showing (known after apply) in the
					// plan.
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: reloadInfoProps["method"],
					},
					"pid_file": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: reloadInfoProps["pidFile"],
					},
					"signal": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: reloadInfoProps["signal"],
					},
					"unit_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: reloadInfoProps["unitName"],
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: policy,
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"assurance": schema.ListAttribute{
						MarkdownDescription: assurance,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"ownership": schema.ListAttribute{
						MarkdownDescription: policyProps["deviceOwnership"],
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
			"wifi": schema.SingleNestedAttribute{
				MarkdownDescription: wifi,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: wifiProps["autojoin"],
						Optional:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: wifiProps["caChain"],
						Computed:            true,
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: wifiProps["externalRadiusServer"],
						Optional:            true,
					},
					"hidden": schema.BoolAttribute{
						MarkdownDescription: wifiProps["hidden"],
						Optional:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: wifiProps["networkAccessServerIP"],
						Optional:            true,
					},
					"ssid": schema.StringAttribute{
						MarkdownDescription: wifiProps["ssid"],
						Required:            true,
					},
				},
			},
			"vpn": schema.SingleNestedAttribute{
				MarkdownDescription: vpn,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: vpnProps["autojoin"],
						Optional:            true,
					},
					"connection_type": schema.StringAttribute{
						MarkdownDescription: vpnProps["connectionType"],
						Required:            true,
					},
					"remote_address": schema.StringAttribute{
						MarkdownDescription: vpnProps["remoteAddress"],
						Required:            true,
					},
					"vendor": schema.StringAttribute{
						MarkdownDescription: vpnProps["vendor"],
						Optional:            true,
					},
					"ike": schema.SingleNestedAttribute{
						MarkdownDescription: ike,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: ikeProps["caChain"],
								Optional:            true,
							},
							"eap": schema.BoolAttribute{
								MarkdownDescription: ikeProps["eap"],
								Optional:            true,
							},
							"remote_id": schema.StringAttribute{
								MarkdownDescription: ikeProps["remoteID"],
								Optional:            true,
							},
						},
					},
				},
			},
			"ethernet": schema.SingleNestedAttribute{
				MarkdownDescription: ethernet,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["autojoin"],
						Optional:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: ethernetProps["caChain"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["externalRadiusServer"],
						Optional:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: ethernetProps["networkAccessServerIP"],
						Optional:            true,
					},
				},
			},
			"browser": schema.SingleNestedAttribute{
				MarkdownDescription: browser,
				Optional:            true,
				Attributes:          map[string]schema.Attribute{},
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
	state := &AccountModel{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := state.ID.ValueString()
	if accountID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Account Request",
			"Account ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetAccount(ctx, accountID, &v20250101.GetAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read account %q: %v", accountID, err),
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
			fmt.Sprintf("Request %q received status %d reading account %s: %s", reqID, httpResp.StatusCode, accountID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20250101.Account{}
	/*
		if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
			resp.Diagnostics.AddError(
				"Smallstep API Client Error",
				fmt.Sprintf("Failed to unmarshal account %s: %v", accountID, err),
			)
			return
		}
	*/
	// TODO remove
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, account); err != nil {
		panic(err)
	}
	// println("Read: " + string(body))

	remote, d := accountFromAPI(ctx, account, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &AccountModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PostAccounts(ctx, &v20250101.PostAccountsParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create account: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20250101.Account{}
	/*
		if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
			resp.Diagnostics.AddError(
				"Smallstep API Client Error",
				fmt.Sprintf("Failed to unmarshal accont: %v", err),
			)
			return
		}
	*/
	// TODO remove
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, account); err != nil {
		panic(err)
	}
	// println("Create: " + string(body))

	model, diags := accountFromAPI(ctx, account, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)

	// Track whether state holds a computed value. Required for our
	// MaybeUseStateForUnknown object plan modifier.
	x509 := types.Object{}
	req.Config.GetAttribute(ctx, path.Root("certificate").AtName("x509"), &x509)
	if x509.IsNull() {
		diags = resp.Private.SetKey(ctx, x509Private, computed)
		resp.Diagnostics.Append(diags...)
	} else {
		diags = resp.Private.SetKey(ctx, x509Private, nil)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &AccountModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	accountID := plan.ID.ValueString()

	account, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, err := r.client.PutAccount(ctx, accountID, &v20250101.PutAccountParams{}, *account)
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
			fmt.Sprintf("Request %q received status %d updating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account = &v20250101.Account{}
	/*
		if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
			resp.Diagnostics.AddError(
				"Smallstep API Client Error",
				fmt.Sprintf("Failed to parse account update response: %v", err),
			)
			return
		}
	*/
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		panic(err)
	}
	// println(string(body))
	if err := json.Unmarshal(body, account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse account update response: %v", err),
		)
		return
	}

	model, diags := accountFromAPI(ctx, account, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)

	// Track whether state holds a computed value. Required for our
	// MaybeUseStateForUnknown object plan modifier.
	x509 := types.Object{}
	req.Config.GetAttribute(ctx, path.Root("certificate").AtName("x509"), &x509)
	if x509.IsNull() {
		diags = resp.Private.SetKey(ctx, x509Private, computed)
		resp.Diagnostics.Append(diags...)
	} else {
		diags = resp.Private.SetKey(ctx, x509Private, nil)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &AccountModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := state.ID.ValueString()
	if accountID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Account Request",
			"Account ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteAccount(ctx, accountID, &v20250101.DeleteAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete account: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
