// Package wifi implements smallstep_wifi.
package wifi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_wifi"

type WifiModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	SSID               types.String `tfsdk:"ssid"`
	RadiusServerCA     types.String `tfsdk:"radius_server_ca"`
	RadiusServerDomain types.String `tfsdk:"radius_server_domain"`
	Autojoin           types.Bool   `tfsdk:"autojoin"`
	Hidden             types.Bool   `tfsdk:"hidden"`
	Credentials        types.Set    `tfsdk:"credentials"`
}

func (model *WifiModel) ToAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.Wifi {
	wifi := &v20250101.Wifi{
		Name:               model.Name.ValueStringPointer(),
		Ssid:               model.SSID.ValueString(),
		RadiusServerCA:     model.RadiusServerCA.ValueString(),
		RadiusServerDomain: model.RadiusServerDomain.ValueStringPointer(),
		Autojoin:           model.Autojoin.ValueBoolPointer(),
		Hidden:             model.Hidden.ValueBoolPointer(),
	}

	if len(model.Credentials.Elements()) > 0 {
		var credentials []string
		d := model.Credentials.ElementsAs(ctx, &credentials, false)
		diags.Append(d...)
		wifi.Credentials = credentials
	}

	return wifi
}

func FromAPI(ctx context.Context, wifi *v20250101.Wifi, diags *diag.Diagnostics, state utils.AttributeGetter) *WifiModel {
	model := &WifiModel{
		ID:             types.StringPointerValue(wifi.Id),
		SSID:           types.StringValue(wifi.Ssid),
		RadiusServerCA: types.StringValue(wifi.RadiusServerCA),
	}

	name, d := utils.ToOptionalString(ctx, wifi.Name, state, path.Root("name"))
	diags.Append(d...)
	model.Name = name

	radiusServerDomain, d := utils.ToOptionalString(ctx, wifi.RadiusServerDomain, state, path.Root("radius_server_domain"))
	diags.Append(d...)
	model.RadiusServerDomain = radiusServerDomain

	autojoin, d := utils.ToOptionalBool(ctx, wifi.Autojoin, state, path.Root("autojoin"))
	diags.Append(d...)
	model.Autojoin = autojoin

	hidden, d := utils.ToOptionalBool(ctx, wifi.Hidden, state, path.Root("hidden"))
	diags.Append(d...)
	model.Hidden = hidden

	credentials, d := utils.ToOptionalSet(ctx, &wifi.Credentials, state, path.Root("credentials"))
	diags.Append(d...)
	model.Credentials = credentials

	return model
}
