package ethernet

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
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
	CAChain               types.String `tfsdk:"ca_chain"`
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	Radius                types.Object `tfsdk:"radius"`
}

var Attributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ca_chain":                 types.StringType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
	"radius":                   types.ObjectType{AttrTypes: radiusAttributes},
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyLAN, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var networkAccessServerIP types.String
	ds := state.GetAttribute(ctx, root.AtName("network_access_server_ip"), &networkAccessServerIP)
	diags.Append(ds...)

	var caChain types.String
	ds = state.GetAttribute(ctx, root.AtName("ca_chain"), &caChain)
	diags.Append(ds...)

	var autojoin types.Bool
	ds = state.GetAttribute(ctx, root.AtName("autojoin"), &autojoin)
	diags.Append(ds...)

	var externalRadiusServer types.Bool
	ds = state.GetAttribute(ctx, root.AtName("external_radius_server"), &externalRadiusServer)
	diags.Append(ds...)

	radius, ds := radiusFromAPI(ctx, conf.Radius, state, root.AtName("radius"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIP,
		"ca_chain":                 caChain,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
		"radius":                   radius,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyLANConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.StrategyLANConfig{
		NetworkAccessServerIP: m.NetworkAccessServerIP.ValueStringPointer(),
		CaChain:               m.CAChain.ValueStringPointer(),
		Autojoin:              m.Autojoin.ValueBoolPointer(),
		ExternalRadiusServer:  m.ExternalRadiusServer.ValueBoolPointer(),
	}, diags
}

//nolint:unused // keep it here for the moment
type radiusServerModel struct {
	CAChain     types.String `json:"ca_chain"`
	IPAddresses types.List   `json:"ip_addresses"`
}

//nolint:unused // keep it here for the moment
var radiusServerAttributes = map[string]attr.Type{
	"ca_chain":     types.StringType,
	"ip_addresses": types.ListType{ElemType: types.StringType},
}

//nolint:unused // keep it here for the moment
func radiusServerFromAPI(ctx context.Context, conf *v20250101.RadiusServer, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(radiusServerAttributes), diags
	}

	caChain, ds := utils.ToOptionalString(ctx, &conf.CaChain, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	ipAddresses, ds := utils.ToOptionalList(ctx, &conf.IpAddresses, state, root.AtName("match_addresses"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"ca_chain":     caChain,
		"ip_addresses": ipAddresses,
	})
	diags.Append(ds...)

	return obj, diags
}

//nolint:unused // keep it here for the moment
func (m *radiusServerModel) toAPI(ctx context.Context, obj types.Object) (v20250101.RadiusServer, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.RadiusServer{
		CaChain:     m.CAChain.ValueString(),
		IpAddresses: utils.ToStringList[string](m.IPAddresses),
	}, diags
}

var radiusAttributes = map[string]attr.Type{
	"ca_chain":     types.StringType,
	"ip_addresses": types.ListType{ElemType: types.StringType},
}

func radiusFromAPI(ctx context.Context, r v20250101.RadiusServer, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	caChain, ds := utils.ToOptionalString(ctx, &r.CaChain, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	ipAddresses, ds := utils.ToOptionalList(ctx, &r.IpAddresses, state, root.AtName("ip_addresses"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(radiusAttributes, map[string]attr.Value{
		"ca_chain":     caChain,
		"ip_addresses": ipAddresses,
	})
	diags.Append(ds...)

	return obj, diags
}
