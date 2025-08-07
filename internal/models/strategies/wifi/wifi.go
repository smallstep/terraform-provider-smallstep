package wifi

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
	SSID                  types.String `tfsdk:"ssid"`
	CAChain               types.String `tfsdk:"ca_chain"`
	Hidden                types.Bool   `tfsdk:"hidden"`
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	Network               types.Object `tfsdk:"network"`
	Radius                types.Object `tfsdk:"radius"`
}

var Attributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ssid":                     types.StringType,
	"ca_chain":                 types.StringType,
	"hidden":                   types.BoolType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
	"network":                  types.ObjectType{AttrTypes: networkAttributes},
	"radius":                   types.ObjectType{AttrTypes: radiusAttributes},
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyWLAN, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var networkAccessServerIP types.String
	ds := state.GetAttribute(ctx, root.AtName("network_access_server_ip"), &networkAccessServerIP)
	diags.Append(ds...)

	var ssid types.String
	ds = state.GetAttribute(ctx, root.AtName("ssid"), &ssid)
	diags.Append(ds...)

	var caChain types.String
	ds = state.GetAttribute(ctx, root.AtName("ca_chain"), &caChain)
	diags.Append(ds...)

	var hidden types.Bool
	ds = state.GetAttribute(ctx, root.AtName("hidden"), &hidden)
	diags.Append(ds...)

	var autojoin types.Bool
	ds = state.GetAttribute(ctx, root.AtName("autojoin"), &autojoin)
	diags.Append(ds...)

	var externalRadiusServer types.Bool
	ds = state.GetAttribute(ctx, root.AtName("external_radius_server"), &externalRadiusServer)
	diags.Append(ds...)

	network, ds := networkFromAPI(ctx, conf.Network, state, root.AtName("network"))
	diags.Append(ds...)

	radius, ds := radiusFromAPI(ctx, conf.Radius, state, root.AtName("radius"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIP,
		"ssid":                     ssid,
		"ca_chain":                 caChain,
		"hidden":                   hidden,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
		"network":                  network,
		"radius":                   radius,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyWLANConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.StrategyWLANConfig{
		NetworkAccessServerIP: m.NetworkAccessServerIP.ValueStringPointer(),
		Ssid:                  m.SSID.ValueString(),
		CaChain:               m.CAChain.ValueStringPointer(),
		Hidden:                m.Hidden.ValueBoolPointer(),
		Autojoin:              m.Autojoin.ValueBoolPointer(),
		ExternalRadiusServer:  m.ExternalRadiusServer.ValueBoolPointer(),
	}, diags
}

var networkAttributes = map[string]attr.Type{
	"ssid":     types.StringType,
	"hidden":   types.BoolType,
	"autojoin": types.BoolType,
}

func networkFromAPI(ctx context.Context, n v20250101.WirelessNetwork, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	ssid, ds := utils.ToOptionalString(ctx, &n.Ssid, state, root.AtName("ssid"))
	diags.Append(ds...)

	hidden, ds := utils.ToOptionalBool(ctx, &n.Hidden, state, root.AtName("hidden"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, &n.Autojoin, state, root.AtName("autojoin"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(networkAttributes, map[string]attr.Value{
		"ssid":     ssid,
		"hidden":   hidden,
		"autojoin": autojoin,
	})
	diags.Append(ds...)

	return obj, diags
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
