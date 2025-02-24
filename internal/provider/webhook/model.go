package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_provisioner_webhook"

type Model struct {
	ID                   types.String    `tfsdk:"id"`
	AuthorityID          types.String    `tfsdk:"authority_id"`
	ProvisionerID        types.String    `tfsdk:"provisioner_id"`
	Name                 types.String    `tfsdk:"name"`
	Kind                 types.String    `tfsdk:"kind"`
	CertType             types.String    `tfsdk:"cert_type"`
	ServerType           types.String    `tfsdk:"server_type"`
	URL                  types.String    `tfsdk:"url"`
	BearerToken          types.String    `tfsdk:"bearer_token"`
	BasicAuth            *BasicAuthModel `tfsdk:"basic_auth"`
	DisableTLSClientAuth types.Bool      `tfsdk:"disable_tls_client_auth"`
	CollectionSlug       types.String    `tfsdk:"collection_slug"`
	Secret               types.String    `tfsdk:"secret"`
}

type BasicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func fromAPI(ctx context.Context, webhook *v20250101.ProvisionerWebhook, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := &Model{
		ID:         types.StringValue(utils.Deref(webhook.Id)),
		Name:       types.StringValue(webhook.Name),
		Kind:       types.StringValue(string(webhook.Kind)),
		CertType:   types.StringValue(string(webhook.CertType)),
		ServerType: types.StringValue(string(webhook.ServerType)),
		URL:        types.StringValue(utils.Deref(webhook.Url)),
	}

	// secret is only set on the first response to a new webhook and is only set
	// for EXTERNAL webhooks. If it's nil in the API response use state for
	// external and null for hosted webhooks.
	if webhook.Secret == nil {
		if webhook.ServerType == v20250101.EXTERNAL {
			secretFromState := types.String{}
			d := state.GetAttribute(ctx, path.Root("secret"), &secretFromState)
			diags = append(diags, d...)
			data.Secret = secretFromState
		} else {
			data.Secret = types.StringNull()
		}
	} else {
		data.Secret = types.StringValue(utils.Deref(webhook.Secret))
	}

	// Currently the API never returns collection slug so always use state
	collectionSlugFromState := types.String{}
	d := state.GetAttribute(ctx, path.Root("collection_slug"), &collectionSlugFromState)
	diags = append(diags, d...)
	data.CollectionSlug = collectionSlugFromState

	// bearer token and basic auth are never set in API responses.
	// Always use state.
	bearerTokenFromState := types.String{}
	d = state.GetAttribute(ctx, path.Root("bearer_token"), &bearerTokenFromState)
	diags = append(diags, d...)
	data.BearerToken = bearerTokenFromState

	basic := &BasicAuthModel{}
	d = state.GetAttribute(ctx, path.Root("basic_auth"), &basic)
	diags = append(diags, d...)
	data.BasicAuth = basic

	disableTLSClientAuth, d := utils.ToOptionalBool(ctx, webhook.DisableTLSClientAuth, state, path.Root("disable_tls_client_auth"))
	diags = append(diags, d...)
	data.DisableTLSClientAuth = disableTLSClientAuth

	return data, diags
}

func toAPI(model *Model) *v20250101.ProvisionerWebhook {
	webhook := &v20250101.ProvisionerWebhook{
		Id:                   model.ID.ValueStringPointer(),
		Name:                 model.Name.ValueString(),
		BearerToken:          model.BearerToken.ValueStringPointer(),
		CertType:             v20250101.ProvisionerWebhookCertType(model.CertType.ValueString()),
		DisableTLSClientAuth: model.DisableTLSClientAuth.ValueBoolPointer(),
		CollectionSlug:       model.CollectionSlug.ValueStringPointer(),
		Kind:                 v20250101.ProvisionerWebhookKind(model.Kind.ValueString()),
		Secret:               model.Secret.ValueStringPointer(),
		ServerType:           v20250101.ProvisionerWebhookServerType(model.ServerType.ValueString()),
		Url:                  model.URL.ValueStringPointer(),
	}

	if model.BasicAuth != nil {
		webhook.BasicAuth = &v20250101.BasicAuth{
			Username: model.BasicAuth.Username.ValueString(),
			Password: model.BasicAuth.Password.ValueString(),
		}
	}

	return webhook
}
