package x509info

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/certfield"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type Model struct {
	CommonName         types.String `tfsdk:"common_name"`
	SANs               types.Object `tfsdk:"sans"`
	Organization       types.Object `tfsdk:"organization"`
	OrganizationalUnit types.Object `tfsdk:"organizational_unit"`
	Locality           types.Object `tfsdk:"locality"`
	Province           types.Object `tfsdk:"province"`
	StreetAddress      types.Object `tfsdk:"street_address"`
	PostalCode         types.Object `tfsdk:"postal_code"`
	Country            types.Object `tfsdk:"country"`
}

var Attributes = map[string]attr.Type{
	"common_name":         types.ObjectType{AttrTypes: certfield.Attributes},
	"sans":                types.ObjectType{AttrTypes: certfield.ListAttributes},
	"organization":        types.ObjectType{AttrTypes: certfield.ListAttributes},
	"organizational_unit": types.ObjectType{AttrTypes: certfield.ListAttributes},
	"locality":            types.ObjectType{AttrTypes: certfield.ListAttributes},
	"province":            types.ObjectType{AttrTypes: certfield.ListAttributes},
	"street_address":      types.ObjectType{AttrTypes: certfield.ListAttributes},
	"postal_code":         types.ObjectType{AttrTypes: certfield.ListAttributes},
	"country":             types.ObjectType{AttrTypes: certfield.ListAttributes},
}

func FromAPI(ctx context.Context, details *v20250101.EndpointCertificateInfo_Details, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if details == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"common_name":         nil,
		"sans":                nil,
		"organization":        nil,
		"organizational_unit": nil,
		"locality":            nil,
		"province":            nil,
		"street_address":      nil,
		"postal_code":         nil,
		"country":             nil,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.EndpointCertificateInfo_Details, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.EndpointCertificateInfo_Details{}, diags
}
