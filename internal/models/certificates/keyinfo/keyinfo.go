package keyinfo

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
	Type       types.String `tfsdk:"type"`
	Protection types.String `tfsdk:"protection"`
	Format     types.String `tfsdk:"format"`
	PubFile    types.String `tfsdk:"pub_file"`
}

var Attributes = map[string]attr.Type{
	"type":       types.StringType,
	"protection": types.StringType,
	"format":     types.StringType,
	"pub_file":   types.StringType,
}

func FromAPI(ctx context.Context, ci *v20250101.EndpointKeyInfo, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ci == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	typ, ds := utils.ToOptionalString(ctx, ci.Type, state, root.AtName("type"))
	diags.Append(ds...)

	protection, ds := utils.ToOptionalString(ctx, ci.Protection, state, root.AtName("protection"))
	diags.Append(ds...)

	format, ds := utils.ToOptionalString(ctx, ci.Format, state, root.AtName("format"))
	diags.Append(ds...)

	pubFile, ds := utils.ToOptionalString(ctx, ci.PubFile, state, root.AtName("pub_file"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"type":       typ,
		"protection": protection,
		"format":     format,
		"pub_file":   pubFile,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.EndpointKeyInfo, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.EndpointKeyInfo{
		Type:       utils.ToStringPointer[v20250101.EndpointKeyInfoType](m.Type.ValueStringPointer()),
		Protection: utils.ToStringPointer[v20250101.EndpointKeyInfoProtection](m.Protection.ValueStringPointer()),
		Format:     utils.ToStringPointer[v20250101.EndpointKeyInfoFormat](m.Format.ValueStringPointer()),
		PubFile:    m.PubFile.ValueStringPointer(),
	}, diags
}
