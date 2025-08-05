package sshinfo

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
	KeyID      types.Object `tfsdk:"key_id"`
	Principals types.Object `tfsdk:"principals"`
}

var Attributes = map[string]attr.Type{
	"key_id":     types.ObjectType{AttrTypes: certfield.Attributes},
	"principals": types.ObjectType{AttrTypes: certfield.ListAttributes},
}

func FromAPI(ctx context.Context, details *v20250101.EndpointCertificateInfo_Details, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if details == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	fields, err := details.AsSshFields()
	if err != nil {
		diags.AddError("SSH Parse Error", err.Error())
		return basetypes.NewObjectNull(Attributes), diags
	}

	keyID, ds := certfield.FromAPI(ctx, fields.KeyId, state, root.AtName("key_id"))
	diags.Append(ds...)

	principals, ds := certfield.FromAPI(ctx, fields.KeyId, state, root.AtName("principals"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"key_id":     keyID,
		"principals": principals,
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

	keyID, ds := new(certfield.Model).ToAPI(ctx, m.KeyID)
	diags.Append(ds...)

	principals, ds := new(certfield.ListModel).ToAPI(ctx, m.Principals)
	diags.Append(ds...)

	details := &v20250101.EndpointCertificateInfo_Details{}
	if err := details.FromSshFields(v20250101.SshFields{
		KeyId:      keyID,
		Principals: principals,
	}); err != nil {
		diags.AddError("X509 Decode Error", err.Error())
	}

	return &v20250101.EndpointCertificateInfo_Details{}, diags
}
