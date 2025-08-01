package certfield

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
	Static         types.String `tfsdk:"static"`
	DeviceMetadata types.String `tfsdk:"device_metadata"`
}

var Attributes = map[string]attr.Type{
	"static":          types.StringType,
	"device_metadata": types.StringType,
}

func FromAPI(ctx context.Context, cf *v20250101.CertificateField, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if cf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	static, ds := utils.ToOptionalString(ctx, cf.Static, state, root.AtName("static"))
	diags.Append(ds...)

	deviceMetadata, ds := utils.ToOptionalString(ctx, cf.DeviceMetadata, state, root.AtName("device_metadata"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"static":          static,
		"device_metadata": deviceMetadata,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.CertificateField, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.CertificateField{
		Static:         m.Static.ValueStringPointer(),
		DeviceMetadata: m.DeviceMetadata.ValueStringPointer(),
	}, diags
}

type ListModel struct {
	Static         types.List `tfsdk:"static"`
	DeviceMetadata types.List `tfsdk:"device_metadata"`
}

var ListAttributes = map[string]attr.Type{
	"static":          types.ListType{ElemType: types.StringType},
	"device_metadata": types.ListType{ElemType: types.StringType},
}

func ListFromAPI(ctx context.Context, cf *v20250101.CertificateFieldList, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if cf == nil {
		return basetypes.NewObjectNull(ListAttributes), diags
	}

	static, ds := utils.ToOptionalList(ctx, cf.Static, state, root.AtName("static"))
	diags.Append(ds...)

	deviceMetadata, ds := utils.ToOptionalList(ctx, cf.DeviceMetadata, state, root.AtName("device_metadata"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ListAttributes, map[string]attr.Value{
		"static":          static,
		"device_metadata": deviceMetadata,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *ListModel) ToAPI(ctx context.Context, obj types.Object) (*v20250101.CertificateFieldList, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.CertificateFieldList{
		Static:         utils.ToStringListPointer[string](m.Static),
		DeviceMetadata: utils.ToStringListPointer[string](m.DeviceMetadata),
	}, diags
}
