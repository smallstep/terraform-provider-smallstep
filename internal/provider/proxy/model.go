package proxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_proxy"

type Model struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	RemoteAddress   types.String `tfsdk:"remote_address"`
	Credentials     types.List   `tfsdk:"credentials"`
	MatchAddresses  types.List   `tfsdk:"match_addresses"`
}

func fromAPI(ctx context.Context, proxy *v20260501.Proxy) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	credentials, d := types.ListValueFrom(ctx, types.StringType, proxy.Credentials)
	diags.Append(d...)

	var matchAddresses types.List
	if proxy.MatchAddresses != nil {
		matchAddresses, d = types.ListValueFrom(ctx, types.StringType, *proxy.MatchAddresses)
	} else {
		matchAddresses = types.ListNull(types.StringType)
	}
	diags.Append(d...)

	return &Model{
		ID:             types.StringPointerValue(proxy.Id),
		Name:           types.StringPointerValue(proxy.Name),
		RemoteAddress:  types.StringValue(proxy.RemoteAddress),
		Credentials:    credentials,
		MatchAddresses: matchAddresses,
	}, diags
}

func (m *Model) toAPI(ctx context.Context) (*v20260501.Proxy, diag.Diagnostics) {
	var diags diag.Diagnostics

	var credentials []string
	d := m.Credentials.ElementsAs(ctx, &credentials, false)
	diags.Append(d...)

	var matchAddresses *[]string
	if !m.MatchAddresses.IsNull() && !m.MatchAddresses.IsUnknown() {
		ma := []string{}
		d := m.MatchAddresses.ElementsAs(ctx, &ma, false)
		diags.Append(d...)
		matchAddresses = &ma
	}

	return &v20260501.Proxy{
		Id:             utils.Ref(m.ID.ValueString()),
		Name:           utils.Ref(m.Name.ValueString()),
		RemoteAddress:  m.RemoteAddress.ValueString(),
		Credentials:    credentials,
		MatchAddresses: matchAddresses,
	}, diags
}
