package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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

	dcSlug := state.DeviceCollectionSlug.ValueString()
	workloadSlug := state.Slug.ValueString()

	httpResp, err := r.client.GetWorkload(ctx, dcSlug, workloadSlug, &v20230301.GetWorkloadParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read workload %q: %v", workloadSlug, err),
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
			fmt.Sprintf("Received status %d reading workload %q: %s", httpResp.StatusCode, workloadSlug, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20230301.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal workload %q: %v", workloadSlug, err),
		)
		return
	}

	remote, d := fromAPI(ctx, workload, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read workload %q resource", workloadSlug))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_info"), remote.CertificateInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_info"), remote.KeyInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hooks"), remote.Hooks)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("reload_info"), remote.ReloadInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), remote.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workload_type"), remote.WorkloadType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), remote.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), remote.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_metadata_key_sans"), remote.DeviceMetadataKeySANs)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_sans"), remote.StaticSANs)...)
	// Not returned from API. Use state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_collection_slug"), state.DeviceCollectionSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), state.AdminEmails)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	workload, props, err := utils.Describe("workload")

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
		MarkdownDescription: workload,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal use only.",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["displayName"],
				Required:            true,
			},
			"workload_type": schema.StringAttribute{
				MarkdownDescription: props["workloadType"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device_collection_slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the device collection the workload will be added to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_emails": schema.SetAttribute{
				MarkdownDescription: props["adminEmails"],
				ElementType:         types.StringType,
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"static_sans": schema.SetAttribute{
				MarkdownDescription: props["staticSANs"],
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"device_metadata_key_sans": schema.SetAttribute{
				MarkdownDescription: props["deviceMetadataKeySANs"],
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
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
					"unit_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: reloadInfoProps["unitName"],
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
						Computed:            true,
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
	slug := plan.Slug.ValueString()

	httpResp, err := a.client.PutWorkload(ctx, dcSlug, slug, &v20230301.PutWorkloadParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d creating workload %q: %s", httpResp.StatusCode, plan.DisplayName.ValueString(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20230301.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, workload, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create workload %q resource", plan.DisplayName.ValueString()))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_info"), state.CertificateInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_info"), state.KeyInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hooks"), state.Hooks)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("reload_info"), state.ReloadInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), state.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workload_type"), state.WorkloadType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), state.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), state.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_metadata_key_sans"), state.DeviceMetadataKeySANs)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_sans"), state.StaticSANs)...)
	// Not returned by the API. Use plan.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_collection_slug"), plan.DeviceCollectionSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), plan.AdminEmails)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	slug := plan.Slug.ValueString()

	httpResp, err := r.client.PutWorkload(ctx, dcSlug, slug, &v20230301.PutWorkloadParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to update workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d updating workload %q: %s", httpResp.StatusCode, plan.DisplayName.ValueString(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20230301.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, workload, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("update workload %q resource", plan.DisplayName.ValueString()))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_info"), state.CertificateInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_info"), state.KeyInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hooks"), state.Hooks)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("reload_info"), state.ReloadInfo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), state.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workload_type"), state.WorkloadType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), state.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), state.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_metadata_key_sans"), state.DeviceMetadataKeySANs)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_sans"), state.StaticSANs)...)
	// Not returned by the API. Use plan.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_collection_slug"), plan.DeviceCollectionSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), plan.AdminEmails)...)
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &Model{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dcSlug := state.DeviceCollectionSlug.ValueString()
	workloadSlug := state.Slug.ValueString()

	httpResp, err := a.client.DeleteWorkload(ctx, dcSlug, workloadSlug, &v20230301.DeleteWorkloadParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete workload %q: %v", workloadSlug, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNoContent {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d deleting workload %q: %s", httpResp.StatusCode, workloadSlug, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}
