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

type Model struct {
	ConnectionType types.String `tfsdk:"connection_type"`
	Vendor         types.String `tfsdk:"vendor"`
	RemoteAddress  types.String `tfsdk:"remote_address"`
	IKEV2Config    types.Object `tfsdk:"ike"`
	Autojoin       types.Bool   `tfsdk:"autojoin"`
}

var Attributes = map[string]attr.Type{
	"connection_type": types.StringType,
	"vendor":          types.StringType,
	"remote_address":  types.StringType,
	"ike":             types.ObjectType{AttrTypes: ikeAttributes},
	"autojoin":        types.BoolType,
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyVPN, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	connectionType, ds := utils.ToOptionalString(ctx, &conf.ConnectionType, state, root.AtName("connection_type"))
	diags.Append(ds...)

	vendor, ds := utils.ToOptionalString(ctx, conf.Vendor, state, root.AtName("vendor"))
	diags.Append(ds...)

	remoteAddress, ds := utils.ToOptionalString(ctx, &conf.RemoteAddress, state, root.AtName("remote_address"))
	diags.Append(ds...)

	ike, ds := ikeFromAPI(ctx, conf.Ike, state, root.AtName("ike"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, conf.Autojoin, state, root.AtName("connection_type"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"connection_type": connectionType,
		"vendor":          vendor,
		"remote_address":  remoteAddress,
		"ike":             ike,
		"autojoin":        autojoin,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyVPN, diag.Diagnostics) {
	var (
		diags diag.Diagnostics
		ike   *v20250101.IkeV2Config
	)

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	if !m.IKEV2Config.IsNull() && !m.IKEV2Config.IsUnknown() {
		ike, ds = new(ikeModel).toAPI(ctx, m.IKEV2Config)
		diags.Append(ds...)
	}

	return v20250101.StrategyVPN{
		ConnectionType: v20250101.VpnAccountConnectionType(m.ConnectionType.ValueString()),
		Vendor:         utils.ToStringPointer[v20250101.VpnAccountVendor](m.Vendor.ValueStringPointer()),
		RemoteAddress:  m.RemoteAddress.ValueString(),
		Ike:            ike,
		Autojoin:       m.Autojoin.ValueBoolPointer(),
	}, diags
}

type ikeModel struct {
	CAChain  types.String `tfsdk:"ca_chain"`
	RemoteID types.String `tfsdk:"remote_id"`
	EAP      types.Bool   `tfsdk:"eap"`
}

var ikeAttributes = map[string]attr.Type{
	"ca_chain":  types.StringType,
	"remote_id": types.StringType,
	"eap":       types.BoolType,
}

func ikeFromAPI(ctx context.Context, ike *v20250101.IkeV2Config, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ike == nil {
		return basetypes.NewObjectNull(ikeAttributes), diags
	}

	caChain, ds := utils.ToOptionalString(ctx, ike.CaChain, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	remoteID, ds := utils.ToOptionalString(ctx, ike.RemoteID, state, root.AtName("remote_id"))
	diags.Append(ds...)

	eap, ds := utils.ToOptionalBool(ctx, ike.Eap, state, root.AtName("eap"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ikeAttributes, map[string]attr.Value{
		"ca_chain":  caChain,
		"remote_id": remoteID,
		"eap":       eap,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *ikeModel) toAPI(ctx context.Context, obj types.Object) (*v20250101.IkeV2Config, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.IkeV2Config{
		CaChain:  m.CAChain.ValueStringPointer(),
		RemoteID: m.RemoteID.ValueStringPointer(),
		Eap:      m.EAP.ValueBoolPointer(),
	}, diags
}
