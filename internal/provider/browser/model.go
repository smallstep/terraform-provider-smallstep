// Package browser implements smallstep_browser.
package browser

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_browser"

type BrowserModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MatchAddresses types.List   `tfsdk:"match_addresses"`
	Credentials    types.Set    `tfsdk:"credentials"`
}

func (model *BrowserModel) ToAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.Browser {
	browser := &v20250101.Browser{}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		browser.Name = model.Name.ValueStringPointer()
	}

	var matchAddresses []string
	d := model.MatchAddresses.ElementsAs(ctx, &matchAddresses, false)
	diags.Append(d...)
	browser.MatchAddresses = matchAddresses

	if len(model.Credentials.Elements()) > 0 {
		var credentials []string
		d = model.Credentials.ElementsAs(ctx, &credentials, false)
		diags.Append(d...)
		browser.Credentials = credentials
	}

	return browser
}

func FromAPI(ctx context.Context, browser *v20250101.Browser, diags *diag.Diagnostics, state utils.AttributeGetter) *BrowserModel {
	model := &BrowserModel{
		ID:   types.StringPointerValue(browser.Id),
		Name: types.StringPointerValue(browser.Name),
	}

	name, d := utils.ToOptionalString(ctx, browser.Name, state, path.Root("name"))
	diags.Append(d...)
	model.Name = name

	matchAddresses, d := utils.ToOptionalList(ctx, &browser.MatchAddresses, state, path.Root("match_addresses"))
	diags.Append(d...)
	model.MatchAddresses = matchAddresses

	credentials, d := utils.ToOptionalSet(ctx, &browser.Credentials, state, path.Root("credentials"))
	diags.Append(d...)
	model.Credentials = credentials

	return model
}
