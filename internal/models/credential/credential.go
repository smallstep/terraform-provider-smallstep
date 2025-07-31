package credential

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificateinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/keyinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type Model struct {
	CertificateInfo types.Object `tfsdk:"certificate_info"`
	KeyInfo         types.Object `tfsdk:"key_info"`
}

var Attributes = map[string]attr.Type{
	"certificate_info": types.ObjectType{AttrTypes: certificateinfo.Attributes},
	"key_info":         types.ObjectType{AttrTypes: keyinfo.Attributes},
}

func FromAPI(ctx context.Context, credential *v20250101.CredentialConfigurationRequest, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if credential == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	certificateInfo, ds := certificateinfo.FromAPI(ctx, credential.CertificateInfo, state, root.AtName("certificate_info"))
	diags.Append(ds...)

	keyInfo, ds := keyinfo.FromAPI(ctx, credential.KeyInfo, state, root.AtName("key_info"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"certificate_info": certificateInfo,
		"key_info":         keyInfo,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.CredentialConfigurationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() {
		return nil, diags
	}

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	ci, ds := new(certificateinfo.Model).ToAPI(ctx, m.CertificateInfo)
	diags.Append(ds...)

	ki, ds := new(keyinfo.Model).ToAPI(ctx, m.KeyInfo)
	diags.Append(ds...)

	return &v20250101.CredentialConfigurationRequest{
		CertificateInfo: ci,
		KeyInfo:         ki,
	}, diags
}
