// Package ethernet implements smallstep_ethernet resource and data source.
package ethernet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_ethernet"

type EthernetModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	RadiusServerCA types.String `tfsdk:"radius_server_ca"`
	Autojoin       types.Bool   `tfsdk:"autojoin"`
	Credentials    types.Set    `tfsdk:"credentials"`
}

func (model *EthernetModel) ToAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.Ethernet {
	ethernet := &v20250101.Ethernet{
		Name:           model.Name.ValueStringPointer(),
		RadiusServerCA: model.RadiusServerCA.ValueString(),
		Autojoin:       model.Autojoin.ValueBoolPointer(),
	}

	if len(model.Credentials.Elements()) > 0 {
		var credentials []string
		d := model.Credentials.ElementsAs(ctx, &credentials, false)
		diags.Append(d...)
		ethernet.Credentials = credentials
	}

	return ethernet
}

func FromAPI(ctx context.Context, ethernet *v20250101.Ethernet, diags *diag.Diagnostics, state utils.AttributeGetter) *EthernetModel {
	model := &EthernetModel{
		ID:             types.StringPointerValue(ethernet.Id),
		RadiusServerCA: types.StringValue(ethernet.RadiusServerCA),
	}

	name, d := utils.ToOptionalString(ctx, ethernet.Name, state, path.Root("name"))
	diags.Append(d...)
	model.Name = name

	autojoin, d := utils.ToOptionalBool(ctx, ethernet.Autojoin, state, path.Root("autojoin"))
	diags.Append(d...)
	model.Autojoin = autojoin

	credentials, d := utils.ToOptionalSet(ctx, &ethernet.Credentials, state, path.Root("credentials"))
	diags.Append(d...)
	model.Credentials = credentials

	return model
}
