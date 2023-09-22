package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.Resource = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20230301.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = typeName
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

	id := state.ID.ValueString()

	httpResp, err := r.client.GetEndpointConfiguration(ctx, id, &v20230301.GetEndpointConfigurationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read endpoint configuration %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading endpoint configuration %q: %s", httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	ac := &v20230301.EndpointConfiguration{}
	if err := json.NewDecoder(httpResp.Body).Decode(ac); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal endpoint configuration %q: %v", id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, ac, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read endpoint configuration %q resource", id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("endpointConfiguration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	_, hookProps, err := utils.Describe("endpointHook")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	hooks, hooksProps, err := utils.Describe("endpointHooks")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	reloadInfo, reloadInfoProps, err := utils.Describe("endpointReloadInfo")
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
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: props["kind"],
				Required:            true,
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authorityID"],
				Required:            true,
			},
			"provisioner_name": schema.StringAttribute{
				MarkdownDescription: props["provisioner"],
				Required:            true,
			},
			"key_info": schema.SingleNestedAttribute{
				// This object is not required by the API but a default object
				// will always be returned with both format and type set to
				// "DEFAULT". To avoid "inconsistent result after apply" errors
				// require these fields to be set explicitly in terraform.
				Required:            true,
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
				},
			},
			"reload_info": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: reloadInfo,
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
				},
			},
			"hooks": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: hooks,
				Attributes: map[string]schema.Attribute{
					"sign": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: hooksProps["sign"],
						Attributes: map[string]schema.Attribute{
							"shell": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: hookProps["shell"],
							},
							"before": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["before"],
							},
							"after": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["after"],
							},
							"on_error": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["onError"],
							},
						},
					},
					"renew": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: hooksProps["renew"],
						Attributes: map[string]schema.Attribute{
							"shell": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: hookProps["shell"],
							},
							"before": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["before"],
							},
							"after": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["after"],
							},
							"on_error": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: hookProps["onError"],
							},
						},
					},
				},
			},
			"certificate_info": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: certInfo,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: certInfoProps["type"],
						Required:            true,
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certInfoProps["duration"],
						Optional:            true,
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
		},
	}
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &Model{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dcSlug := plan.DeviceCollectionSlug.ValueString()

	httpResp, err := a.client.PostWorkloads(ctx, dcSlug, &v20230301.PostWorkloadsParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create endpoint configuration %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d creating endpoint configuration %q: %s", httpResp.StatusCode, plan.DisplayName.ValueString(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	ac := &v20230301.EndpointConfiguration{}
	if err := json.NewDecoder(httpResp.Body).Decode(ac); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal endpoint configuration %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, ac, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create endpoint configuration %q resource", plan.DisplayName.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Workloads must be destroyed and re-created.")
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Coming soon.")
}