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
	NetworkAccessServerIP types.String `tfsdk:"NetworkAccessServerIP"`
	CAChain               types.String `tfsdk:"ca_chain"`
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
}

var Attributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ca_chain":                 types.StringType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyLANConfig, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	networkAccessServerIP, ds := utils.ToOptionalString(ctx, conf.NetworkAccessServerIP, state, root.AtName("network_access_server_ip"))
	diags.Append(ds...)

	caChain, ds := utils.ToOptionalString(ctx, conf.CaChain, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, conf.Autojoin, state, root.AtName("autojoin"))
	diags.Append(ds...)

	externalRadiusServer, ds := utils.ToOptionalBool(ctx, conf.ExternalRadiusServer, state, root.AtName("external_radius_server"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIP,
		"ca_chain":                 caChain,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
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

type radiusServerModel struct {
	CAChain     types.String `json:"ca_chain"`
	IPAddresses types.List   `json:"ip_addresses"`
}

var radiusServerAttributes = map[string]attr.Type{
	"ca_chain":     types.StringType,
	"ip_addresses": types.ListType{ElemType: types.StringType},
}

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
