package relay

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_relay"

type Model struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Hostname            types.String `tfsdk:"hostname"`
	CaChain             types.String `tfsdk:"ca_chain"`
	AllowedTargets      types.List   `tfsdk:"allowed_targets"`
	IssuingAuthorityID  types.String `tfsdk:"issuing_authority_id"`
}

func fromAPI(ctx context.Context, relay *v20260501.Relay) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	var allowedTargets types.List
	if relay.AllowedTargets != nil {
		allowedTargets, _ = types.ListValueFrom(ctx, types.StringType, *relay.AllowedTargets)
	} else {
		allowedTargets = types.ListNull(types.StringType)
	}

	return &Model{
		ID:                 types.StringPointerValue(relay.Id),
		Name:               types.StringValue(relay.Name),
		Hostname:           types.StringValue(relay.Hostname),
		CaChain:            types.StringValue(relay.CaChain),
		AllowedTargets:     allowedTargets,
		IssuingAuthorityID: types.StringPointerValue(relay.IssuingAuthorityID),
	}, diags
}

func (m *Model) toAPI(ctx context.Context) (*v20260501.Relay, diag.Diagnostics) {
	var diags diag.Diagnostics

	var allowedTargets *[]string
	if !m.AllowedTargets.IsNull() && !m.AllowedTargets.IsUnknown() {
		at := []string{}
		_ = m.AllowedTargets.ElementsAs(ctx, &at, false)
		allowedTargets = &at
	}

	return &v20260501.Relay{
		Id:                 utils.Ref(m.ID.ValueString()),
		Name:               m.Name.ValueString(),
		Hostname:           m.Hostname.ValueString(),
		CaChain:            m.CaChain.ValueString(),
		AllowedTargets:     allowedTargets,
		IssuingAuthorityID: utils.Ref(m.IssuingAuthorityID.ValueString()),
	}, diags
}
