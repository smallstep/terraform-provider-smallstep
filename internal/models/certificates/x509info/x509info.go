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
	CommonName         types.Object `tfsdk:"common_name"`
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

	fields, err := details.AsX509Fields()
	if err != nil {
		diags.AddError("SSH Parse Error", err.Error())
		return basetypes.NewObjectNull(Attributes), diags
	}

	commonName, ds := certfield.FromAPI(ctx, fields.CommonName, state, root.AtName("common_name"))
	diags.Append(ds...)

	sans, ds := certfield.ListFromAPI(ctx, fields.Sans, state, root.AtName("sans"))
	diags.Append(ds...)

	organization, ds := certfield.ListFromAPI(ctx, fields.Organization, state, root.AtName("organization"))
	diags.Append(ds...)

	organizationalUnit, ds := certfield.ListFromAPI(ctx, fields.OrganizationalUnit, state, root.AtName("organizational_unit"))
	diags.Append(ds...)

	locality, ds := certfield.ListFromAPI(ctx, fields.Locality, state, root.AtName("locality"))
	diags.Append(ds...)

	province, ds := certfield.ListFromAPI(ctx, fields.Province, state, root.AtName("province"))
	diags.Append(ds...)

	streetAddress, ds := certfield.ListFromAPI(ctx, fields.StreetAddress, state, root.AtName("street_address"))
	diags.Append(ds...)

	postalCode, ds := certfield.ListFromAPI(ctx, fields.PostalCode, state, root.AtName("postal_code"))
	diags.Append(ds...)

	country, ds := certfield.ListFromAPI(ctx, fields.Country, state, root.AtName("country"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"common_name":         commonName,
		"sans":                sans,
		"organization":        organization,
		"organizational_unit": organizationalUnit,
		"locality":            locality,
		"province":            province,
		"street_address":      streetAddress,
		"postal_code":         postalCode,
		"country":             country,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.EndpointCertificateInfo_Details, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	commonName, ds := new(certfield.Model).ToAPI(ctx, m.CommonName)
	diags.Append(ds...)

	sans, ds := new(certfield.ListModel).ToAPI(ctx, m.SANs)
	diags.Append(ds...)

	organization, ds := new(certfield.ListModel).ToAPI(ctx, m.Organization)
	diags.Append(ds...)

	organizationalUnit, ds := new(certfield.ListModel).ToAPI(ctx, m.OrganizationalUnit)
	diags.Append(ds...)

	locality, ds := new(certfield.ListModel).ToAPI(ctx, m.Locality)
	diags.Append(ds...)

	province, ds := new(certfield.ListModel).ToAPI(ctx, m.Province)
	diags.Append(ds...)

	streetAddress, ds := new(certfield.ListModel).ToAPI(ctx, m.StreetAddress)
	diags.Append(ds...)

	postalCode, ds := new(certfield.ListModel).ToAPI(ctx, m.PostalCode)
	diags.Append(ds...)

	country, ds := new(certfield.ListModel).ToAPI(ctx, m.Country)
	diags.Append(ds...)

	details := &v20250101.EndpointCertificateInfo_Details{}
	if err := details.FromX509Fields(v20250101.X509Fields{
		CommonName:         commonName,
		Sans:               sans,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
		StreetAddress:      streetAddress,
		PostalCode:         postalCode,
		Country:            country,
	}); err != nil {
		diags.AddError("X509 Decode Error", err.Error())
	}
	return details, diags
}
