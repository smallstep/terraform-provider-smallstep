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
	TrustedRoots     types.String `tfsdk:"trusted_roots"`
	RedirectURI      types.String `tfsdk:"redirect_uri"`
	Client           types.Object `tfsdk:"client"`
	IdentityProvider types.Object `tfsdk:"identity_provider"`
}

var Attributes = map[string]attr.Type{
	"trusted_roots":     types.StringType,
	"redirect_uri":      types.StringType,
	"client":            types.ObjectType{AttrTypes: clientAttributes},
	"identity_provider": types.ObjectType{AttrTypes: identityProviderAttributes},
}

func FromAPI(ctx context.Context, conf *v20250101.StrategySSO, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	var trustedRoots types.String
	ds := state.GetAttribute(ctx, root.AtName("trusted_roots"), &trustedRoots)
	diags.Append(ds...)

	var redirectURI types.String
	ds = state.GetAttribute(ctx, root.AtName("redirect_uri"), &redirectURI)
	diags.Append(ds...)

	client, ds := clientFromAPI(ctx, conf.Client, state, root.AtName("client"))
	diags.Append(ds...)

	identityProvider, ds := identityProviderFromAPI(ctx, conf.IdentityProvider, state, root.AtName("identity_provider"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"trusted_roots":     trustedRoots,
		"redirect_uri":      redirectURI,
		"client":            client,
		"identity_provider": identityProvider,
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

var clientAttributes = map[string]attr.Type{
	"id":           types.StringType,
	"redirect_uri": types.StringType,
	"secret":       types.StringType,
}

func clientFromAPI(ctx context.Context, c v20250101.DeviceIdentityProviderClient, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, ds := utils.ToOptionalString(ctx, &c.Id, state, root.AtName("id"))
	diags.Append(ds...)

	redirectURI, ds := utils.ToOptionalString(ctx, &c.RedirectUri, state, root.AtName("redirect_uri"))
	diags.Append(ds...)

	secret, ds := utils.ToOptionalString(ctx, c.Secret, state, root.AtName("secret"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(clientAttributes, map[string]attr.Value{
		"id":           id,
		"redirect_uri": redirectURI,
		"secret":       secret,
	})
	diags.Append(ds...)

	return obj, diags
}

var identityProviderAttributes = map[string]attr.Type{
	"authorize_endpoint": types.StringType,
	"jwks_endpoint":      types.StringType,
	"trust_roots":        types.StringType,
}

func identityProviderFromAPI(ctx context.Context, ip v20250101.DeviceIdentityProvider, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	authorizedEndpoint, ds := utils.ToOptionalString(ctx, &ip.AuthorizeEndpoint, state, root.AtName("authorize_endpoint"))
	diags.Append(ds...)

	jwksEndpoint, ds := utils.ToOptionalString(ctx, &ip.JwksEndpoint, state, root.AtName("jwks_endpoint"))
	diags.Append(ds...)

	trustedRoots, ds := utils.ToOptionalString(ctx, &ip.TrustRoots, state, root.AtName("trust_roots"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(identityProviderAttributes, map[string]attr.Value{
		"authorize_endpoint": authorizedEndpoint,
		"jwks_endpoint":      jwksEndpoint,
		"trust_roots":        trustedRoots,
	})
	diags.Append(ds...)

	return obj, diags
}
