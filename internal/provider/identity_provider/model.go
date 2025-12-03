package identity_provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
)

const idp_name = "smallstep_identity_provider"
const client_name = "smallstep_identity_provider_client"

type IdentityProviderModel struct {
	TrustRoots        types.String `tfsdk:"trust_roots"`
	Issuer            types.String `tfsdk:"issuer"`
	AuthorizeEndpoint types.String `tfsdk:"authorize_endpoint"`
	TokenEndpoint     types.String `tfsdk:"token_endpoint"`
	JWKSEndpoint      types.String `tfsdk:"jwks_endpoint"`
}

type ClientModel struct {
	ID              types.String `tfsdk:"id"`
	RedirectURI     types.String `tfsdk:"redirect_uri"`
	Secret          types.String `tfsdk:"secret"`
	StoreSecret     types.Bool   `tfsdk:"store_secret"`
	WriteSecretFile types.String `tfsdk:"write_secret_file"`
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

func clientToAPI(model *ClientModel) v20250101.IdpClient {
	return v20250101.IdpClient{
		Id:          model.ID.ValueStringPointer(),
		RedirectURI: model.RedirectURI.ValueString(),
	}
}

func clientFromAPI(client *v20250101.IdpClient) ClientModel {
	return ClientModel{
		RedirectURI: types.StringValue(client.RedirectURI),
		ID:          types.StringPointerValue(client.Id),
		Secret:      types.StringPointerValue(client.Secret),
	}
}
