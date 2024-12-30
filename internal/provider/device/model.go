package device

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

// type name for both resources and data sources
const typeName = "smallstep_device"

type Model struct {
	ID                  types.String `tfsdk:"id"`
	PermanentIdentifier types.String `tfsdk:"permanent_identifier"`
	DisplayName         types.String `tfsdk:"display_name"`
	DisplayID           types.String `tfsdk:"display_id"`
	Serial              types.String `tfsdk:"serial"`
	OS                  types.String `tfsdk:"os"`
	Ownership           types.String `tfsdk:"ownership"`
	User                types.Object `tfsdk:"user"`
	Tags                types.Set    `tfsdk:"tags"`
	Metadata            types.Map    `tfsdk:"metadata"`
	ApprovedAt          types.String `tfsdk:"approved_at"`
	EnrolledAt          types.String `tfsdk:"enrolled_at"`
	LastSeen            types.String `tfsdk:"last_seen"`
	Connected           types.Bool   `tfsdk:"connected"`
	HighAssurance       types.Bool   `tfsdk:"high_assurance"`
}

type UserModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	Email       types.String `tfsdk:"email"`
}

func (user *UserModel) AsAPI(ctx context.Context) (*v20250101.DeviceUser, diag.Diagnostics) {
	d := diag.Diagnostics{}

	if user == nil {
		return nil, d
	}

	return &v20250101.DeviceUser{
		DisplayName: user.DisplayName.ValueStringPointer(),
		Email:       user.Email.ValueString(),
	}, d
}

var userAttrTypes = map[string]attr.Type{
	"display_name": types.StringType,
	"email":        types.StringType,
}

func fromAPI(ctx context.Context, device *v20250101.Device, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		ID:                  types.StringValue(device.Id),
		PermanentIdentifier: types.StringValue(device.PermanentIdentifier),
		Connected:           types.BoolValue(device.Connected),
		HighAssurance:       types.BoolValue(device.HighAssurance),
	}

	displayName, d := utils.ToOptionalString(ctx, device.DisplayName, state, path.Root("display_name"))
	diags.Append(d...)
	model.DisplayName = displayName

	displayID, d := utils.ToOptionalString(ctx, device.DisplayId, state, path.Root("display_id"))
	diags.Append(d...)
	model.DisplayID = displayID

	serial, d := utils.ToOptionalString(ctx, device.Serial, state, path.Root("serial"))
	diags.Append(d...)
	model.Serial = serial

	os, d := utils.ToOptionalString(ctx, device.Os, state, path.Root("os"))
	diags.Append(d...)
	model.OS = os

	ownership, d := utils.ToOptionalString(ctx, device.Ownership, state, path.Root("ownership"))
	diags.Append(d...)
	model.Ownership = ownership

	// user
	if device.User != nil {
		userDisplayName, d := utils.ToOptionalString(ctx, device.User.DisplayName, state, path.Root("user").AtName("display_name"))
		diags.Append(d...)

		user, d := basetypes.NewObjectValue(userAttrTypes, map[string]attr.Value{
			"display_name": userDisplayName,
			"email":        types.StringValue(device.User.Email),
		})
		diags.Append(d...)
		model.User = user
	} else {
		model.User = basetypes.NewObjectNull(userAttrTypes)
	}

	// model.Tags
	tags, d := utils.ToOptionalSet(ctx, device.Tags, state, path.Root("tags"))
	diags.Append(d...)
	model.Tags = tags

	if device.Metadata != nil {
		meta := map[string]attr.Value{}
		for k, v := range *device.Metadata {
			s, ok := v.(string)
			if !ok {
				diags.AddError("Invalid device metadata", k+" is not a string")
				continue
			}
			meta[k] = types.StringValue(s)
		}
		metadata, d := types.MapValue(types.StringType, meta)
		diags.Append(d...)
		model.Metadata = metadata
	} else {
		// see comment in utils/optional.go
		metadataFromState := types.Map{}
		d := state.GetAttribute(ctx, path.Root("metadata"), &metadataFromState)
		diags.Append(d...)
		switch {
		case metadataFromState.IsUnknown():
			model.Metadata = types.MapNull(types.StringType)
		case metadataFromState.IsNull():
			model.Metadata = metadataFromState
		case len(metadataFromState.Elements()) == 0:
			model.Metadata = metadataFromState
		default:
			model.Metadata = types.MapNull(types.StringType)
		}
	}

	if device.ApprovedAt != nil {
		model.ApprovedAt = types.StringValue((*device.ApprovedAt).Format(time.RFC3339))
	} else {
		model.ApprovedAt = types.StringNull()
	}
	if device.EnrolledAt != nil {
		model.EnrolledAt = types.StringValue((*device.EnrolledAt).Format(time.RFC3339))
	} else {
		model.EnrolledAt = types.StringNull()
	}
	if device.LastSeen != nil {
		model.LastSeen = types.StringValue((*device.LastSeen).Format(time.RFC3339))
	} else {
		model.LastSeen = types.StringNull()
	}

	return model, diags
}

func toAPI(ctx context.Context, m *Model) (*v20250101.DeviceRequest, diag.Diagnostics) {
	// user
	diags := diag.Diagnostics{}

	d := &v20250101.DeviceRequest{
		PermanentIdentifier: m.PermanentIdentifier.ValueString(),
	}

	if !m.DisplayID.IsNull() && !m.DisplayID.IsUnknown() {
		d.DisplayId = m.DisplayID.ValueStringPointer()
	}
	if !m.DisplayName.IsNull() && !m.DisplayName.IsUnknown() {
		d.DisplayName = m.DisplayName.ValueStringPointer()
	}
	if !m.Serial.IsNull() && !m.Serial.IsUnknown() {
		d.Serial = m.Serial.ValueStringPointer()
	}
	if !m.OS.IsNull() && !m.OS.IsUnknown() {
		d.Os = utils.Ref(v20250101.DeviceOS(m.OS.ValueString()))
	}
	if !m.Ownership.IsNull() && !m.Ownership.IsUnknown() {
		d.Ownership = utils.Ref(v20250101.DeviceOwnership(m.Ownership.ValueString()))
	}
	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		meta := map[string]types.String{}
		diag := m.Metadata.ElementsAs(ctx, &meta, false)
		diags.Append(diag...)

		if len(meta) > 0 {
			metadata := map[string]any{}
			for k, v := range meta {
				metadata[k] = v.ValueString()
			}
			d.Metadata = &metadata
		}
	}
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		var tags []string
		diag := m.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(diag...)
		d.Tags = &tags
	}
	if !m.User.IsNull() && !m.User.IsUnknown() {
		user := &UserModel{}
		diag := m.User.As(ctx, &user, basetypes.ObjectAsOptions{
			UnhandledUnknownAsEmpty: true,
		})
		diags.Append(diag...)
		d.User = &v20250101.DeviceUser{
			DisplayName: user.DisplayName.ValueStringPointer(),
			Email:       user.Email.ValueString(),
		}
	}

	return d, nil
}
