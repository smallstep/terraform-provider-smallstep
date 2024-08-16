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
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.Resource = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20231101.Client
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

	client, ok := req.ProviderData.(*v20231101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20231101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	httpResp, err := r.client.GetWorkload(ctx, dcSlug, workloadSlug, &v20231101.GetWorkloadParams{})
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
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading workload %q: %s", reqID, httpResp.StatusCode, workloadSlug, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20231101.Workload{}
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_data"), remote.CertificateData)...)
	// Not returned from API. Use state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_collection_slug"), state.DeviceCollectionSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), state.AdminEmails)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	workload, props, err := utils.Describe("workload")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	certInfo, err := NewCertificateInfoResourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	certData, err := NewCertificateDataResourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
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

	keyInfo, err := NewKeyInfoResourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	reloadInfo, err := NewReloadInfoResourceSchema()
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
			"key_info":         keyInfo,
			"reload_info":      reloadInfo,
			"certificate_data": certData,
			"certificate_info": certInfo,
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

	httpResp, err := a.client.PutWorkload(ctx, dcSlug, slug, &v20231101.PutWorkloadParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating workload %q: %s", reqID, httpResp.StatusCode, plan.DisplayName.ValueString(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20231101.Workload{}
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_data"), state.CertificateData)...)
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

	httpResp, err := r.client.PutWorkload(ctx, dcSlug, slug, &v20231101.PutWorkloadParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to update workload %q: %v", plan.DisplayName.ValueString(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating workload %q: %s", reqID, httpResp.StatusCode, plan.DisplayName.ValueString(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	workload := &v20231101.Workload{}
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_data"), state.CertificateData)...)
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

	httpResp, err := a.client.DeleteWorkload(ctx, dcSlug, workloadSlug, &v20231101.DeleteWorkloadParams{})
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
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting workload %q: %s", reqID, httpResp.StatusCode, workloadSlug, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}
