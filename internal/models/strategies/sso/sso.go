package sso

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
	TrustedRoots types.String `tfsdk:"trusted_roots"`
	RedirectURI  types.String `tfsdk:"redirect_uri"`
}

var Attributes = map[string]attr.Type{
	"trusted_roots": types.StringType,
	"redirect_uri":  types.StringType,
}

func FromAPI(ctx context.Context, conf *v20250101.StrategySSOConfig, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	p := path.Root("sso")

	trustedRoots, ds := utils.ToOptionalString(ctx, &conf.TrustRoots, state, p.AtName("trusted_roots"))
	diags.Append(ds...)

	redirectURI, ds := utils.ToOptionalString(ctx, &conf.RedirectUri, state, p.AtName("redirect_uri"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"trusted_roots": trustedRoots,
		"redirect_uri":  redirectURI,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategySSOConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return v20250101.StrategySSOConfig{
		TrustRoots:  m.TrustedRoots.ValueString(),
		RedirectUri: m.RedirectURI.ValueString(),
	}, diags
}
