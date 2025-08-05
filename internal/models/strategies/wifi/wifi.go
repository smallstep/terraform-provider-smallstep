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
}

var Attributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ssid":                     types.StringType,
	"ca_chain":                 types.StringType,
	"hidden":                   types.BoolType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyWLAN, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var netIP string
	if len(conf.Radius.IpAddresses) > 0 {
		netIP = conf.Radius.IpAddresses[0]
	}
	networkAccessServerIP, ds := utils.ToOptionalString(ctx, &netIP, state, root.AtName("network_access_server_ip"))
	diags.Append(ds...)

	hidden, ds := utils.ToOptionalBool(ctx, &conf.Network.Hidden, state, root.AtName("hidden"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, &conf.Network.Autojoin, state, root.AtName("autojoin"))
	diags.Append(ds...)

	var external bool
	externalRadiusServer, ds := utils.ToOptionalBool(ctx, &external, state, root.AtName("external_radius_server"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIP,
		"ssid":                     types.StringValue(conf.Network.Ssid),
		"ca_chain":                 types.StringValue(conf.Radius.CaChain),
		"hidden":                   hidden,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
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
