package device_collection

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
	resp.TypeName = collectionTypeName
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

	httpResp, err := r.client.GetDeviceCollection(ctx, state.Slug.ValueString(), &v20231101.GetDeviceCollectionParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device-collection %s: %v", state.Slug.String(), err),
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
			fmt.Sprintf("Request %q received status %d reading device-collection %s: %s", reqID, httpResp.StatusCode, state.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.DeviceCollection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device-collection %s: %v", state.Slug.String(), err),
		)
		return
	}

	remote, d := fromAPI(ctx, collection, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read device-collection %q resource", collection.Slug))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), remote.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), remote.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("authority_id"), remote.AuthorityID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_type"), remote.DeviceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws_vm"), remote.AWSDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("azure_vm"), state.AzureDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gcp_vm"), remote.GCPDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tpm"), state.TPMDevice)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("deviceCollection")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	aws, awsProps, err := utils.Describe("awsVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	gcp, gcpProps, err := utils.Describe("gcpVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	azure, azureProps, err := utils.Describe("azureVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	tpm, tpmProps, err := utils.Describe("tpm")
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
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authorityID"],
				Required:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Required:            true,
			},
			"device_type": schema.StringAttribute{
				MarkdownDescription: props["deviceType"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_vm": schema.SingleNestedAttribute{
				MarkdownDescription: aws,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"accounts": schema.SetAttribute{
						MarkdownDescription: awsProps["accounts"],
						ElementType:         types.StringType,
						Required:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableCustomSANs"],
						Optional:            true,
					},
				},
			},
			"azure_vm": schema.SingleNestedAttribute{
				MarkdownDescription: azure,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"resource_groups": schema.SetAttribute{
						MarkdownDescription: azureProps["resourceGroups"],
						ElementType:         types.StringType,
						Required:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: azureProps["disableCustomSANs"],
						Optional:            true,
					},
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: azureProps["tenantID"],
						Required:            true,
					},
					"audience": schema.StringAttribute{
						MarkdownDescription: azureProps["audience"],
						Optional:            true,
					},
				},
			},
			"gcp_vm": schema.SingleNestedAttribute{
				MarkdownDescription: gcp,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"service_accounts": schema.SetAttribute{
						MarkdownDescription: gcpProps["serviceAccounts"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"project_ids": schema.SetAttribute{
						MarkdownDescription: gcpProps["projectIDs"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: gcpProps["disableCustomSANs"],
						Optional:            true,
					},
				},
			},
			"tpm": schema.SingleNestedAttribute{
				MarkdownDescription: tpm,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"attestor_roots": schema.StringAttribute{
						MarkdownDescription: tpmProps["attestorRoots"],
						Optional:            true,
					},
					"attestor_intermediates": schema.StringAttribute{
						MarkdownDescription: tpmProps["attestorIntermediates"],
						Optional:            true,
					},
					"force_cn": schema.BoolAttribute{
						MarkdownDescription: tpmProps["forceCN"],
						Optional:            true,
					},
					"require_eab": schema.BoolAttribute{
						MarkdownDescription: tpmProps["requireEAB"],
						Optional:            true,
					},
				},
			},
		},
	}
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, d := toAPI(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PutDeviceCollection(ctx, reqBody.Slug, &v20231101.PutDeviceCollectionParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create device collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating device collection %q: %s", reqID, httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.DeviceCollection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %q: %v", plan.Slug.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, collection, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create collection %q resource", plan.Slug.ValueString()))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), state.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), state.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("authority_id"), state.AuthorityID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_type"), state.DeviceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws_vm"), state.AWSDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("azure_vm"), state.AzureDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gcp_vm"), state.GCPDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tpm"), state.TPMDevice)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, d := toAPI(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PutDeviceCollection(ctx, reqBody.Slug, &v20231101.PutDeviceCollectionParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create device collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating device collection %q: %s", reqID, httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.DeviceCollection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %q: %v", plan.Slug.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, collection, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create collection %q resource", plan.Slug.ValueString()))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), state.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), state.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("authority_id"), state.AuthorityID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_type"), state.DeviceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws_vm"), state.AWSDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("azure_vm"), state.AzureDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gcp_vm"), state.GCPDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tpm"), state.TPMDevice)...)
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan Model

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.DeleteDeviceCollection(ctx, plan.Slug.ValueString(), &v20231101.DeleteDeviceCollectionParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete device collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting device collection %q: %s", reqID, httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}
