package device

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
	device, props, err := utils.Describe("device")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Schema",
			err.Error(),
		)
		return
	}

	deviceUser, userProps, err := utils.Describe("deviceUser")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device User Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: device,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permanent_identifier": schema.StringAttribute{
				MarkdownDescription: props["permanentIdentifier"],
				Required:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: props["permanent_identifier"],
				Optional:            true,
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["displayName"],
				Optional:            true,
				Computed:            true,
			},
			"display_id": schema.StringAttribute{
				MarkdownDescription: props["displayId"],
				Optional:            true,
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: props["os"],
				Optional:            true,
				Computed:            true,
			},
			"ownership": schema.StringAttribute{
				MarkdownDescription: props["ownership"],
				Optional:            true,
				Computed:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: props["metadata"],
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"tags": schema.SetAttribute{
				MarkdownDescription: props["tags"],
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"user": schema.SingleNestedAttribute{
				MarkdownDescription: deviceUser,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{
						MarkdownDescription: userProps["displayName"],
						Optional:            true,
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: userProps["email"],
						Required:            true,
					},
				},
			},
			"connected": schema.BoolAttribute{
				MarkdownDescription: props["connected"],
				Computed:            true,
			},
			"high_assurance": schema.BoolAttribute{
				MarkdownDescription: props["highAssurance"],
				Computed:            true,
			},
			"enrolled_at": schema.StringAttribute{
				MarkdownDescription: props["enrolledAt"],
				Computed:            true,
			},
			"approved_at": schema.StringAttribute{
				MarkdownDescription: props["approvedAt"],
				Computed:            true,
			},
			"last_seen": schema.StringAttribute{
				MarkdownDescription: props["lastSeen"],
				Computed:            true,
			},
		},
	}
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
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := state.ID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Device Request",
			"Device ID is required",
		)
		return
	}

	httpResp, err := r.client.GetDevice(ctx, deviceID, &v20250101.GetDeviceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device %q: %v", state.ID.ValueString(), err),
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
			fmt.Sprintf("Request %q received status %d reading device %s: %s", reqID, httpResp.StatusCode, deviceID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	device := &v20250101.Device{}
	if err := json.NewDecoder(httpResp.Body).Decode(device); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device %s: %v", deviceID, err),
		)
		return
	}

	remote, d := fromAPI(ctx, device, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
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

	httpResp, err := a.client.PostDevices(ctx, &v20250101.PostDevicesParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create device: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating device: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	device := &v20250101.Device{}
	if err := json.NewDecoder(httpResp.Body).Decode(device); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, device, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &Model{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	var deviceID string
	req.State.GetAttribute(ctx, path.Root("id"), &deviceID)

	resource, diags := toAPI(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	patch := v20250101.DevicePatch{}
	var remove []v20250101.DevicePatchRemove

	if resource.DisplayId == nil {
		remove = append(remove, v20250101.DisplayId)
	} else {
		patch.DisplayId = resource.DisplayId
	}

	if resource.DisplayName == nil {
		remove = append(remove, v20250101.DisplayName)
	} else {
		patch.DisplayName = resource.DisplayName
	}

	if resource.Metadata == nil {
		remove = append(remove, v20250101.Metadata)
	} else {
		patch.Metadata = resource.Metadata
	}

	if resource.Os == nil {
		remove = append(remove, v20250101.Os)
	} else {
		patch.Os = resource.Os
	}

	if resource.Ownership == nil {
		remove = append(remove, v20250101.Ownership)
	} else {
		patch.Ownership = resource.Ownership
	}

	if resource.Serial == nil {
		remove = append(remove, v20250101.Serial)
	} else {
		patch.Serial = resource.Serial
	}

	if resource.Tags == nil {
		remove = append(remove, v20250101.Tags)
	} else {
		patch.Tags = resource.Tags
	}

	if resource.User == nil || resource.User.Email == "" {
		remove = append(remove, v20250101.UserEmail)
	} else {
		patch.UserEmail = &resource.User.Email
	}

	httpResp, err := r.client.PatchDevice(ctx, deviceID, &v20250101.PatchDeviceParams{}, patch)
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
			fmt.Sprintf("Request %q received status %d updating device: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	device := &v20250101.Device{}
	if err := json.NewDecoder(httpResp.Body).Decode(device); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse device update response: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, device, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &Model{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := state.ID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Device Request",
			"Device ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteDevice(ctx, deviceID, &v20250101.DeleteDeviceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete device: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating device: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
