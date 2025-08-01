package certinfo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/sshinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/x509info"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type Model struct {
	AuthorityID types.String `tfsdk:"authority_id"`
	CrtFile     types.String `tfsdk:"crt_file"`
	KeyFile     types.String `tfsdk:"key_file"`
	RootFile    types.String `tfsdk:"root_file"`
	Duration    types.String `tfsdk:"duration"`
	UID         types.Int32  `tfsdk:"uid"`
	GID         types.Int32  `tfsdk:"gid"`
	Mode        types.Int32  `tfsdk:"mode"`
	X509        types.Object `tfsdk:"x509"`
	SSH         types.Object `tfsdk:"ssh"`
}

var Attributes = map[string]attr.Type{
	"authority_id": types.StringType,
	"crt_file":     types.StringType,
	"key_file":     types.StringType,
	"root_file":    types.StringType,
	"duration":     types.StringType,
	"uid":          types.Int32Type,
	"gid":          types.Int32Type,
	"mode":         types.Int32Type,
	"x509":         types.ObjectType{AttrTypes: x509info.Attributes},
	"ssh":          types.ObjectType{AttrTypes: sshinfo.Attributes},
}

func FromAPI(ctx context.Context, ci *v20250101.EndpointCertificateInfo, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ci == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	authorityID, ds := utils.ToOptionalString(ctx, ci.AuthorityID, state, root.AtName("authority_id"))
	diags.Append(ds...)

	crtFile, ds := utils.ToOptionalString(ctx, ci.CrtFile, state, root.AtName("crt_file"))
	diags.Append(ds...)

	keyFile, ds := utils.ToOptionalString(ctx, ci.KeyFile, state, root.AtName("key_file"))
	diags.Append(ds...)

	rootFile, ds := utils.ToOptionalString(ctx, ci.RootFile, state, root.AtName("root_file"))
	diags.Append(ds...)

	duration, ds := utils.ToOptionalString(ctx, ci.Duration, state, root.AtName("duration"))
	diags.Append(ds...)

	uid, ds := utils.ToOptionalInt(ctx, ci.Uid, state, root.AtName("uid"))
	diags.Append(ds...)

	gid, ds := utils.ToOptionalInt(ctx, ci.Gid, state, root.AtName("gid"))
	diags.Append(ds...)

	mode, ds := utils.ToOptionalInt(ctx, ci.Mode, state, root.AtName("mode"))
	diags.Append(ds...)

	var (
		x509Object = basetypes.NewObjectNull(x509info.Attributes)
		sshObject  = basetypes.NewObjectNull(sshinfo.Attributes)
	)

	switch ci.Type {
	case v20250101.EndpointCertificateInfoTypeX509:
		x509Object, ds = x509info.FromAPI(ctx, ci.Details, state, root.AtName("x509"))
		diags.Append(ds...)
	case v20250101.EndpointCertificateInfoTypeSSHUSER:
		sshObject, ds = sshinfo.FromAPI(ctx, ci.Details, state, root.AtName("ssh"))
		diags.Append(ds...)
	case v20250101.EndpointCertificateInfoTypeSSHHOST:
		sshObject, ds = sshinfo.FromAPI(ctx, ci.Details, state, root.AtName("ssh"))
		diags.Append(ds...)
	default:
		x509Object, ds = x509info.FromAPI(ctx, ci.Details, state, root.AtName("x509"))
		diags.Append(ds...)
	}

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{
		"authority_id": authorityID,
		"crt_file":     crtFile,
		"key_file":     keyFile,
		"root_file":    rootFile,
		"duration":     duration,
		"uid":          uid,
		"gid":          gid,
		"mode":         mode,
		"x509":         x509Object,
		"ssh":          sshObject,
	})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (*v20250101.EndpointCertificateInfo, diag.Diagnostics) {
	var (
		typ     v20250101.EndpointCertificateInfoType
		details *v20250101.EndpointCertificateInfo_Details
		diags   diag.Diagnostics
	)

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	switch {
	case !m.X509.IsNull():
		typ = v20250101.EndpointCertificateInfoTypeX509
	case !m.SSH.IsNull():
		// typ = v20250101.EndpointCertificateInfoTypeSSHHOST
		typ = v20250101.EndpointCertificateInfoTypeSSHUSER
	default:
		typ = v20250101.EndpointCertificateInfoTypeX509
	}

	return &v20250101.EndpointCertificateInfo{
		AuthorityID: m.AuthorityID.ValueStringPointer(),
		CrtFile:     m.CrtFile.ValueStringPointer(),
		KeyFile:     m.KeyFile.ValueStringPointer(),
		RootFile:    m.RootFile.ValueStringPointer(),
		Duration:    m.Duration.ValueStringPointer(),
		Uid:         utils.ToIntPointer(m.UID.ValueInt32Pointer()),
		Gid:         utils.ToIntPointer(m.GID.ValueInt32Pointer()),
		Mode:        utils.ToIntPointer(m.Mode.ValueInt32Pointer()),
		Type:        typ,
		Details:     details,
	}, diags
}
