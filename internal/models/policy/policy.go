package policy

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
	Assurance types.List `tfsdk:"assurance"`
	OS        types.List `tfsdk:"os"`
	Ownership types.List `tfsdk:"ownership"`
	Source    types.List `tfsdk:"source"`
	Tags      types.List `tfsdk:"tags"`
}

var Attributes = map[string]attr.Type{
	"assurance": types.ListType{ElemType: types.StringType},
	"os":        types.ListType{ElemType: types.StringType},
	"ownership": types.ListType{ElemType: types.StringType},
	"source":    types.ListType{ElemType: types.StringType},
	"tags":      types.ListType{ElemType: types.StringType},
}

func FromAPI(ctx context.Context, p *v20250101.PolicyMatchCriteria, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if p == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	assurance, ds := utils.ToOptionalList(ctx, p.Assurance, state, root.AtName("assurance"))
	diags.Append(ds...)

	os, ds := utils.ToOptionalList(ctx, p.OperatingSystem, state, root.AtName("os"))
	diags.Append(ds...)

	ownership, ds := utils.ToOptionalList(ctx, p.Ownership, state, root.AtName("ownership"))
	diags.Append(ds...)

	source, ds := utils.ToOptionalList(ctx, p.Source, state, root.AtName("source"))
	diags.Append(ds...)

	tags, ds := utils.ToOptionalList(ctx, p.Tags, state, root.AtName("tags"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"assurance": assurance,
		"os":        os,
		"ownership": ownership,
		"source":    source,
		"tags":      tags,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.PolicyMatchCriteria, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() {
		return nil, diags
	}

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	return &v20250101.PolicyMatchCriteria{
		Assurance:       utils.ToStringListPointer[v20250101.DeviceAssurance](m.Assurance),
		OperatingSystem: utils.ToStringListPointer[v20250101.DeviceOS](m.OS),
		Ownership:       utils.ToStringListPointer[v20250101.DeviceOwnership](m.Ownership),
		Source:          utils.ToStringListPointer[v20250101.DeviceDiscoverySource](m.Source),
		Tags:            utils.ToStringListPointer[string](m.Tags),
	}, diags
}
