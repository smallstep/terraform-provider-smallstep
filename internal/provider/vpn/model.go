// Package vpn implements smallstep_vpn.
package vpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_vpn"

type VpnModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	RemoteAddress  types.String `tfsdk:"remote_address"`
	Autojoin       types.Bool   `tfsdk:"autojoin"`
	Vendor         types.String `tfsdk:"vendor"`
	Ike            types.Object `tfsdk:"ike"`
	Credentials    types.Set    `tfsdk:"credentials"`
}

type IkeV2ConfigModel struct {
	CaChain  types.String `tfsdk:"ca_chain"`
	Eap      types.Bool   `tfsdk:"eap"`
	RemoteID types.String `tfsdk:"remote_id"`
}

var ikeV2ConfigAttrTypes = map[string]attr.Type{
	"ca_chain":  types.StringType,
	"eap":       types.BoolType,
	"remote_id": types.StringType,
}

func (model *VpnModel) ToAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.Vpn {
	vpn := &v20250101.Vpn{
		Name:           model.Name.ValueStringPointer(),
		ConnectionType: v20250101.VpnType(model.ConnectionType.ValueString()),
		RemoteAddress:  model.RemoteAddress.ValueString(),
		Autojoin:       model.Autojoin.ValueBoolPointer(),
	}

	if !model.Vendor.IsNull() && !model.Vendor.IsUnknown() {
		vendor := v20250101.VpnVendor(model.Vendor.ValueString())
		vpn.Vendor = &vendor
	}

	if !model.Ike.IsNull() && !model.Ike.IsUnknown() {
		var ikeModel IkeV2ConfigModel
		d := model.Ike.As(ctx, &ikeModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			vpn.Ike = &v20250101.IkeV2Config{
				CaChain:  ikeModel.CaChain.ValueString(),
				Eap:      ikeModel.Eap.ValueBoolPointer(),
				RemoteID: ikeModel.RemoteID.ValueStringPointer(),
			}
		}
	}

	if len(model.Credentials.Elements()) > 0 {
		var credentials []string
		d := model.Credentials.ElementsAs(ctx, &credentials, false)
		diags.Append(d...)
		vpn.Credentials = credentials
	}

	return vpn
}

func FromAPI(ctx context.Context, vpn *v20250101.Vpn, diags *diag.Diagnostics, state utils.AttributeGetter) *VpnModel {
	model := &VpnModel{
		ID:             types.StringPointerValue(vpn.Id),
		ConnectionType: types.StringValue(string(vpn.ConnectionType)),
		RemoteAddress:  types.StringValue(vpn.RemoteAddress),
	}

	name, d := utils.ToOptionalString(ctx, vpn.Name, state, path.Root("name"))
	diags.Append(d...)
	model.Name = name

	autojoin, d := utils.ToOptionalBool(ctx, vpn.Autojoin, state, path.Root("autojoin"))
	diags.Append(d...)
	model.Autojoin = autojoin

	vendor, d := utils.ToOptionalString(ctx, (*string)(vpn.Vendor), state, path.Root("vendor"))
	diags.Append(d...)
	model.Vendor = vendor

	if vpn.Ike != nil {
		ikeModel := &IkeV2ConfigModel{
			CaChain: types.StringValue(vpn.Ike.CaChain),
		}

		eap, d := utils.ToOptionalBool(ctx, vpn.Ike.Eap, state, path.Root("ike").AtName("eap"))
		diags.Append(d...)
		ikeModel.Eap = eap

		remoteID, d := utils.ToOptionalString(ctx, vpn.Ike.RemoteID, state, path.Root("ike").AtName("remote_id"))
		diags.Append(d...)
		ikeModel.RemoteID = remoteID

		ikeObj, d := types.ObjectValueFrom(ctx, ikeV2ConfigAttrTypes, ikeModel)
		diags.Append(d...)
		model.Ike = ikeObj
	} else {
		model.Ike = types.ObjectNull(ikeV2ConfigAttrTypes)
	}

	credentials, d := utils.ToOptionalSet(ctx, &vpn.Credentials, state, path.Root("credentials"))
	diags.Append(d...)
	model.Credentials = credentials

	return model
}
