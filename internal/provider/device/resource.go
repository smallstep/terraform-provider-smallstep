package device

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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// Delete
		return
	}
	// Attributes that may either be set by API or synced from an MDM are marked as
	// optional and computed. When one of these attributes is removed from the
	// config the plan keeps the old value instead of setting it to null. We have
	// to modify the plan and change the attribute to null.
	optionalComputedStrings := []path.Path{
		path.Root("display_id"),
		path.Root("display_name"),
		path.Root("serial"),
		path.Root("os"),
		path.Root("ownership"),
	}
	for _, p := range optionalComputedStrings {
		config := types.String{}
		diags := req.Config.GetAttribute(ctx, p, &config)
		resp.Diagnostics.Append(diags...)
		if !config.IsNull() {
			continue
		}

		resp.Plan.SetAttribute(ctx, p, config)
	}

	email := types.String{}
	diags := req.Config.GetAttribute(ctx, path.Root("user").AtName("email"), &email)
	resp.Diagnostics.Append(diags...)
	if email.IsNull() {
		user := basetypes.NewObjectNull(userAttrTypes)
		diags = resp.Plan.SetAttribute(ctx, path.Root("user"), user)
		resp.Diagnostics.Append(diags...)
	}

	tags := types.Set{}
	diags = req.Config.GetAttribute(ctx, path.Root("tags"), &tags)
	resp.Diagnostics.Append(diags...)
	if tags.IsNull() {
		diags = resp.Plan.SetAttribute(ctx, path.Root("tags"), tags)
		resp.Diagnostics.Append(diags...)
	}

	metadata := types.Map{}
	diags = req.Config.GetAttribute(ctx, path.Root("metadata"), &metadata)
	resp.Diagnostics.Append(diags...)
	if metadata.IsNull() || len(metadata.Elements()) == 0 {
		diags = resp.Plan.SetAttribute(ctx, path.Root("metadata"), metadata)
		resp.Diagnostics.Append(diags...)
	}
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
				MarkdownDescription: props["serial"],
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
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 256),
						},
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

	patch := v20250101.DevicePatch{
		DisplayId:   &v20250101.DevicePatch_DisplayId{},
		DisplayName: &v20250101.DevicePatch_DisplayName{},
		Metadata:    &v20250101.DevicePatch_Metadata{},
		Os:          &v20250101.DevicePatch_Os{},
		Ownership:   &v20250101.DevicePatch_Ownership{},
		Serial:      &v20250101.DevicePatch_Serial{},
		Tags:        &v20250101.DevicePatch_Tags{},
		User:        &v20250101.DeviceUserPatch{},
	}

	if resource.DisplayId == nil {
		err := patch.DisplayId.FromDevicePatchDisplayId1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset display_id", err.Error())
		}
	} else {
		err := patch.DisplayId.FromDeviceDisplayId(*resource.DisplayId)
		if err != nil {
			diags.AddError("prepare device patch: set display_id", err.Error())
		}
	}

	if resource.DisplayName == nil {
		err := patch.DisplayName.FromDevicePatchDisplayName1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset display_name", err.Error())
		}
	} else {
		err := patch.DisplayName.FromDeviceDisplayName(*resource.DisplayName)
		if err != nil {
			diags.AddError("prepare device patch: set display_name", err.Error())
		}
	}

	if resource.Metadata == nil {
		err := patch.Metadata.FromDevicePatchMetadata1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset metadata", err.Error())
		}
	} else {
		err := patch.Metadata.FromDeviceMetadata(*resource.Metadata)
		if err != nil {
			diags.AddError("prepare device patch: unset metadata", err.Error())
		}
	}

	if resource.Os == nil {
		err := patch.Os.FromDevicePatchOs1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset os", err.Error())
		}
	} else {
		err := patch.Os.FromDeviceOS(*resource.Os)
		if err != nil {
			diags.AddError("prepare device patch: set os", err.Error())
		}
	}

	if resource.Ownership == nil {
		err := patch.Ownership.FromDevicePatchOwnership1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset ownership", err.Error())
		}
	} else {
		err := patch.Ownership.FromDeviceOwnership(*resource.Ownership)
		if err != nil {
			diags.AddError("prepare device patch: set ownership", err.Error())
		}
	}

	if resource.Serial == nil {
		err := patch.Serial.FromDevicePatchSerial1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset serial", err.Error())
		}
	} else {
		err := patch.Serial.FromDeviceSerial(*resource.Serial)
		if err != nil {
			diags.AddError("prepare device patch: set serial", err.Error())
		}
	}

	if resource.Tags == nil {
		err := patch.Tags.FromDevicePatchTags1(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset tags", err.Error())
		}
	} else {
		err := patch.Tags.FromDeviceTags(*resource.Tags)
		if err != nil {
			diags.AddError("prepare device patch: set tags", err.Error())
		}
	}

	if resource.User == nil || resource.User.Email == "" {
		err := patch.User.Email.FromDeviceUserPatchEmail0(nil)
		if err != nil {
			diags.AddError("prepare device patch: unset user email", err.Error())
		}
	} else {
		err := patch.User.Email.FromDeviceUserPatchEmail1(resource.User.Email)
		if err != nil {
			diags.AddError("prepare device patch: unset user email", err.Error())
		}
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
