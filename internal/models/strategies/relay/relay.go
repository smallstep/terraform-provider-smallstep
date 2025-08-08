package relay

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
	MatchDomains    types.List   `tfsdk:"match_domains"`
	Regions         types.List   `tfsdk:"regions"`
	ProxyInsstances types.List   `tfsdk:"proxy_instances"`
	Server          types.Object `tfsdk:"server"`
}

var Attributes = map[string]attr.Type{
	"match_domains":   types.ListType{ElemType: types.StringType},
	"regions":         types.ListType{ElemType: types.StringType},
	"proxy_instances": types.ListType{ElemType: types.ObjectType{AttrTypes: proxyInstanceAttributes}},
	"server":          types.ObjectType{AttrTypes: serverAttributes},
}

func FromAPI(ctx context.Context, r *v20250101.StrategyNetworkRelay, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if r == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var matchDomains types.List
	ds := state.GetAttribute(ctx, root.AtName("match_domains"), &matchDomains)
	diags.Append(ds...)

	var regions types.List
	ds = state.GetAttribute(ctx, root.AtName("regions"), &regions)
	diags.Append(ds...)

	proxyInstances, ds := proxyInstanceListFromAPI(ctx, r.ProxyInstances, state, root.AtName("proxy_instances"))
	diags.Append(ds...)

	server, ds := serverFromAPI(ctx, r.Server, state, root.AtName("server"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"match_domains":   matchDomains,
		"regions":         regions,
		"proxy_instances": proxyInstances,
		"server":          server,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyNetworkRelayConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.StrategyNetworkRelayConfig{
		MatchDomains: utils.ToStringList[string](m.MatchDomains),
		Regions:      utils.ToStringList[v20250101.StrategyNetworkRelayConfigRegions](m.Regions),
	}, diags
}

var proxyInstanceAttributes = map[string]attr.Type{
	"ip_address": types.StringType,
	"region":     types.StringType,
	"status":     types.StringType,
}

func proxyInstanceFromAPI(ctx context.Context, pi v20250101.ProxyInstance, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	ipAddress, ds := utils.ToOptionalString(ctx, &pi.IpAddress, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	region, ds := utils.ToOptionalString(ctx, &pi.Region, state, root.AtName("region"))
	diags.Append(ds...)

	status, ds := utils.ToOptionalString(ctx, &pi.Status, state, root.AtName("status"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(proxyInstanceAttributes, map[string]attr.Value{
		"ip_address": ipAddress,
		"region":     region,
		"status":     status,
	})
	diags.Append(ds...)

	return obj, diags
}

func proxyInstanceListFromAPI(ctx context.Context, pis []v20250101.ProxyInstance, state utils.AttributeGetter, root path.Path) (types.List, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		ds       diag.Diagnostics
		elements = make([]attr.Value, len(pis))
	)
	for i, pi := range pis {
		elements[i], ds = proxyInstanceFromAPI(ctx, pi, state, root.AtListIndex(i))
		diags.Append(ds...)
	}

	obj, ds := basetypes.NewListValue(types.ObjectType{AttrTypes: proxyInstanceAttributes}, elements)
	diags.Append(ds...)

	return obj, diags
}

var serverAttributes = map[string]attr.Type{
	"ca_chain": types.StringType,
	"hostname": types.StringType,
}

func serverFromAPI(ctx context.Context, s v20250101.StrategyNetworkRelayServer, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	caChain, ds := utils.ToOptionalString(ctx, &s.CaChain, state, root.AtName("ca_chain"))
	diags.Append(ds...)

	hostname, ds := utils.ToOptionalString(ctx, &s.Hostname, state, root.AtName("hostname"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(serverAttributes, map[string]attr.Value{
		"ca_chain": caChain,
		"hostname": hostname,
	})
	diags.Append(ds...)

	return obj, diags
}
