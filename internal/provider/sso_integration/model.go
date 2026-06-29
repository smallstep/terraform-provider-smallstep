package sso_integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_sso_integration"

type Model struct {
	ID                  types.String `tfsdk:"id"`
	RedirectURI         types.String `tfsdk:"redirect_uri"`
	LifecycleFailureURI types.String `tfsdk:"lifecycle_failure_uri"`
	Secret              types.String `tfsdk:"secret"`
	StoreSecret         types.Bool   `tfsdk:"store_secret"`
	WriteSecretFile     types.String `tfsdk:"write_secret_file"`
}

func fromAPI(ctx context.Context, integration *v20260501.SsoIntegration) (*Model, diag.Diagnostics) {
	return &Model{
		ID:                  types.StringPointerValue(integration.Id),
		RedirectURI:         types.StringValue(integration.RedirectURI),
		LifecycleFailureURI: types.StringPointerValue(integration.LifecycleFailureURI),
		Secret:              types.StringNull(),
		StoreSecret:         types.BoolNull(),
		WriteSecretFile:     types.StringNull(),
	}, diag.Diagnostics{}
}

func (m *Model) toAPI(ctx context.Context) (*v20260501.SsoIntegration, diag.Diagnostics) {
	return &v20260501.SsoIntegration{
		Id:                  utils.Ref(m.ID.ValueString()),
		RedirectURI:         m.RedirectURI.ValueString(),
		LifecycleFailureURI: utils.Ref(m.LifecycleFailureURI.ValueString()),
		Secret:              utils.Ref(m.Secret.ValueString()),
	}, diag.Diagnostics{}
}
