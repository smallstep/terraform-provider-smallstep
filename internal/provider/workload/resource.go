package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
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
	workload, workloadProps, err := utils.DescribeV20260501("workload")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Workload Schema", err.Error())
		return
	}

	_, credentialProps, err := utils.DescribeV20260501("workloadCredential")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Workload Credential Schema", err.Error())
		return
	}

	_, probeProps, err := utils.DescribeV20260501("probe")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Probe Schema", err.Error())
		return
	}

	_, hooksProps, err := utils.DescribeV20260501("endpointHooks")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Hooks Schema", err.Error())
		return
	}

	_, hookProps, err := utils.DescribeV20260501("endpointHook")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Hook Schema", err.Error())
		return
	}

	_, reloadInfoProps, err := utils.DescribeV20260501("endpointReloadInfo")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Reload Info Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: workload,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: workloadProps["id"],
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: workloadProps["name"],
				Optional:            true,
			},
			"workload_type": schema.StringAttribute{
				MarkdownDescription: workloadProps["workloadType"],
				Optional:            true,
			},
			"credentials": schema.ListNestedAttribute{
				MarkdownDescription: workloadProps["credentials"],
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"credential_id": schema.StringAttribute{
							MarkdownDescription: credentialProps["credentialId"],
							Required:            true,
						},
						"probes": schema.ListNestedAttribute{
							MarkdownDescription: credentialProps["probes"],
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"target": schema.StringAttribute{
										MarkdownDescription: probeProps["target"],
										Required:            true,
									},
									"protocol": schema.StringAttribute{
										MarkdownDescription: probeProps["protocol"],
										Required:            true,
										Validators: []validator.String{
											stringvalidator.OneOf("TLS", "SSH"),
										},
									},
									"crt_file": schema.StringAttribute{
										MarkdownDescription: probeProps["crtFile"],
										Optional:            true,
									},
									"key_file": schema.StringAttribute{
										MarkdownDescription: probeProps["keyFile"],
										Optional:            true,
									},
									"root_file": schema.StringAttribute{
										MarkdownDescription: probeProps["rootFile"],
										Optional:            true,
									},
									"server_name": schema.StringAttribute{
										MarkdownDescription: probeProps["serverName"],
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
			"hooks": schema.SingleNestedAttribute{
				MarkdownDescription: hooksProps["hooks"],
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"sign": schema.SingleNestedAttribute{
						MarkdownDescription: hooksProps["sign"],
						Optional:            true,
						Attributes: map[string]schema.Attribute{
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
							"shell": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: hookProps["shell"],
							},
						},
					},
					"renew": schema.SingleNestedAttribute{
						MarkdownDescription: hooksProps["renew"],
						Optional:            true,
						Attributes: map[string]schema.Attribute{
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
							"shell": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: hookProps["shell"],
							},
						},
					},
				},
			},
			"reload_info": schema.SingleNestedAttribute{
				MarkdownDescription: reloadInfoProps["reloadInfo"],
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["method"],
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("AUTOMATIC", "CUSTOM", "SIGNAL", "DBUS", "PLATFORM"),
						},
					},
					"pid_file": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["pidFile"],
						Optional:            true,
					},
					"signal": schema.Int64Attribute{
						MarkdownDescription: reloadInfoProps["signal"],
						Optional:            true,
					},
					"unit_name": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["unitName"],
						Optional:            true,
					},
				},
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
	var plan *Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiWorkload, d := plan.toAPI(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostWorkloads(ctx, &v20260501.PostWorkloadsParams{}, *apiWorkload)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to create workload: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating workload: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	workload := &v20260501.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal workload: %v", err))
		return
	}

	model, d := fromAPI(ctx, workload)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workloadID := state.ID.ValueString()
	if workloadID == "" {
		resp.Diagnostics.AddError("Invalid Read Workload Request", "Workload ID is required")
		return
	}

	httpResp, err := r.client.GetWorkload(ctx, workloadID, &v20260501.GetWorkloadParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to read workload: %v", err))
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
			fmt.Sprintf("Request %q received status %d reading workload: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	workload := &v20260501.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal workload: %v", err))
		return
	}

	model, d := fromAPI(ctx, workload)
	resp.Diagnostics.Append(d...)
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

	workloadID := state.ID.ValueString()
	if workloadID == "" {
		resp.Diagnostics.AddError("Invalid Update Workload Request", "Workload ID is required")
		return
	}

	apiWorkload, d := plan.toAPI(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PutWorkload(ctx, workloadID, &v20260501.PutWorkloadParams{}, *apiWorkload)
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to update workload: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating workload: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	workload := &v20260501.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal workload: %v", err))
		return
	}

	model, d := fromAPI(ctx, workload)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workloadID := state.ID.ValueString()
	if workloadID == "" {
		resp.Diagnostics.AddError("Invalid Delete Workload Request", "Workload ID is required")
		return
	}

	httpResp, err := r.client.DeleteWorkload(ctx, workloadID, &v20260501.DeleteWorkloadParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to delete workload: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting workload: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
