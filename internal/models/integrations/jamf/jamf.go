package jamf

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
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TenantURL    types.String `tfsdk:"tenant_url"`
}

var Attributes = map[string]attr.Type{
	"client_id":     types.StringType,
	"client_secret": types.StringType,
	"tenant_url":    types.StringType,
}

func FromAPI(ctx context.Context, jamf *v20250101.JamfInventoryIntegration, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if jamf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	clientID, ds := utils.ToOptionalString(ctx, jamf.ClientId, state, root.AtName("client_id"))
	diags.Append(ds...)

	clientSecret, ds := utils.ToOptionalString(ctx, jamf.ClientId, state, root.AtName("client_secret"))
	diags.Append(ds...)

	tenantURL, ds := utils.ToOptionalString(ctx, jamf.ClientId, state, root.AtName("tenant_url"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"tenant_url":    tenantURL,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.JamfInventoryIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.JamfInventoryIntegration{
		ClientId:     m.ClientID.ValueStringPointer(),
		ClientSecret: m.ClientSecret.ValueStringPointer(),
		TenantUrl:    m.TenantURL.ValueString(),
	}, diags
}
