package intune

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
	AppID           types.String `tfsdk:"app_id"`
	AppSecret       types.String `tfsdk:"app_secret"`
	AzureTenantName types.String `tfsdk:"azure_tenant_name"`
}

var Attributes = map[string]attr.Type{
	"app_id":            types.StringType,
	"app_secret":        types.StringType,
	"azure_tenant_name": types.StringType,
}

func FromAPI(ctx context.Context, intune *v20250101.IntuneInventoryIntegration, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if intune == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	appID, ds := utils.ToOptionalString(ctx, &intune.AppId, state, root.AtName("app_id"))
	diags.Append(ds...)

	appSecret, ds := utils.ToOptionalString(ctx, &intune.AppSecret, state, root.AtName("app_secret"))
	diags.Append(ds...)

	tenantName, ds := utils.ToOptionalString(ctx, &intune.AzureTenantName, state, root.AtName("azure_tenant_name"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"app_id":            appID,
		"app_secret":        appSecret,
		"azure_tenant_name": tenantName,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.IntuneInventoryIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.IntuneInventoryIntegration{
		AppId:           m.AppID.ValueString(),
		AppSecret:       m.AppSecret.ValueString(),
		AzureTenantName: m.AzureTenantName.ValueString(),
	}, diags
}
