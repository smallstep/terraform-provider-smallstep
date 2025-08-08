package browser

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
	MatchAddresses types.List `tfsdk:"match_addresses"`
}

var Attributes = map[string]attr.Type{
	"match_addresses": types.ListType{ElemType: types.StringType},
}

func FromAPI(ctx context.Context, conf *v20250101.StrategyBrowserMutualTLS, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var matchAddresses types.List
	ds := state.GetAttribute(ctx, root.AtName("match_addresses"), &matchAddresses)
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"match_addresses": matchAddresses,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategyBrowserMutualTLSConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.StrategyBrowserMutualTLSConfig{
		MatchAddresses: utils.ToStringList[string](m.MatchAddresses),
	}, diags
}
