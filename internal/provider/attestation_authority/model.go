package attestation_authority

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_attestation_authority"

type Model struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Root                  types.String `tfsdk:"root"`
	Slug                  types.String `tfsdk:"slug"`
	AttestorIntermediates types.String `tfsdk:"attestor_intermediates"`
	AttestorRoots         types.String `tfsdk:"attestor_roots"`
	CreatedAt             types.String `tfsdk:"created_at"`
}

func fromAPI(ctx context.Context, aa *v20230301.AttestationAuthority, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		ID:            types.StringValue(utils.Deref(aa.Id)),
		Name:          types.StringValue(aa.Name),
		Root:          types.StringValue(utils.Deref(aa.Root)),
		Slug:          types.StringValue(utils.Deref(aa.Slug)),
		AttestorRoots: types.StringValue(aa.AttestorRoots),
		CreatedAt:     types.StringValue(aa.CreatedAt.Format(time.RFC3339)),
	}

	attestorIntermediates, d := utils.ToOptionalString(ctx, aa.AttestorIntermediates, state, path.Root("attestor_intermediates"))
	diags = append(diags, d...)
	model.AttestorIntermediates = attestorIntermediates

	return model, diags
}

func toAPI(model *Model) *v20230301.AttestationAuthority {
	return &v20230301.AttestationAuthority{
		Name:                  model.Name.ValueString(),
		AttestorRoots:         model.AttestorRoots.ValueString(),
		AttestorIntermediates: model.AttestorIntermediates.ValueStringPointer(),
	}
}
