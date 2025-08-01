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

func FromAPI(ctx context.Context, conf *v20250101.StrategyLAN, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
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

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyLAN, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyLAN{
		NetworkAccessServerIP: m.NetworkAccessServerIP.ValueStringPointer(),
		CaChain:               m.CAChain.ValueStringPointer(),
		Autojoin:              m.Autojoin.ValueBoolPointer(),
		ExternalRadiusServer:  m.ExternalRadiusServer.ValueBoolPointer(),
	}, diags
}
