package agent_configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_agent_configuration"

type Model struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Provisioner     types.String `tfsdk:"provisioner_name"`
	AuthorityID     types.String `tfsdk:"authority_id"`
	AttestationSlug types.String `tfsdk:"attestation_slug"`
}

func fromAPI(ctx context.Context, ac *v20230301.AgentConfiguration, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		ID:          types.StringValue(utils.Deref(ac.Id)),
		Name:        types.StringValue(ac.Name),
		Provisioner: types.StringValue(ac.Provisioner),
		AuthorityID: types.StringValue(ac.AuthorityID),
	}

	attestationSlug, d := utils.ToOptionalString(ctx, ac.AttestationSlug, state, path.Root("attestation_slug"))
	diags = append(diags, d...)
	model.AttestationSlug = attestationSlug

	return model, diags
}

func toAPI(model *Model) *v20230301.AgentConfiguration {
	return &v20230301.AgentConfiguration{
		Id:              model.ID.ValueStringPointer(),
		Name:            model.Name.ValueString(),
		AuthorityID:     model.AuthorityID.ValueString(),
		Provisioner:     model.Provisioner.ValueString(),
		AttestationSlug: model.AttestationSlug.ValueStringPointer(),
	}
}
