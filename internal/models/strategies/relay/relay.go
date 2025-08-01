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
	MatchDomains types.List `tfsdk:"match_domains"`
	Regions      types.List `tfsdk:"regions"`
}

var Attributes = map[string]attr.Type{
	"match_domains": types.ListType{ElemType: types.StringType},
	"regions":       types.ListType{ElemType: types.StringType},
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyNetworkRelay, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	matchDomains, ds := utils.ToOptionalList(ctx, &conf.MatchDomains, state, root.AtName("match_domains"))
	diags.Append(ds...)

	regions, ds := utils.ToOptionalList(ctx, &conf.Regions, state, root.AtName("regions"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"match_domains": matchDomains,
		"regions":       regions,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyNetworkRelay, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyNetworkRelay{
		MatchDomains: utils.ToStringList[string](m.MatchDomains),
		Regions:      utils.ToStringList[v20250101.StrategyNetworkRelayRegions](m.Regions),
	}, diags
}
