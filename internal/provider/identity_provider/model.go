package identity_provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
)

const idp_name = "smallstep_identity_provider"

type IdentityProviderModel struct {
	TrustRoots        types.String `tfsdk:"trust_roots"`
	Issuer            types.String `tfsdk:"issuer"`
	AuthorizeEndpoint types.String `tfsdk:"authorize_endpoint"`
	TokenEndpoint     types.String `tfsdk:"token_endpoint"`
	JWKSEndpoint      types.String `tfsdk:"jwks_endpoint"`
}

func idpToAPI(model *IdentityProviderModel) v20250101.IdentityProvider {
	return v20250101.IdentityProvider{
		TrustRoots: model.TrustRoots.ValueString(),
	}
}

func idpFromAPI(idp *v20250101.IdentityProvider) IdentityProviderModel {
	return IdentityProviderModel{
		TrustRoots:        types.StringValue(idp.TrustRoots),
		Issuer:            types.StringPointerValue(idp.Issuer),
		AuthorizeEndpoint: types.StringPointerValue(idp.AuthorizeEndpoint),
		TokenEndpoint:     types.StringPointerValue(idp.TokenEndpoint),
		JWKSEndpoint:      types.StringPointerValue(idp.JwksEndpoint),
	}
}
