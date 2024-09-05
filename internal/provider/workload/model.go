package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_workload"

type Model struct {
	WorkloadType         types.String `tfsdk:"workload_type"`
	DisplayName          types.String `tfsdk:"display_name"`
	Slug                 types.String `tfsdk:"slug"`
	DeviceCollectionSlug types.String `tfsdk:"device_collection_slug"`
	AuthorityID          types.String `tfsdk:"authority_id"`
	CertificateInfo      types.Object `tfsdk:"certificate_info"`
	KeyInfo              types.Object `tfsdk:"key_info"`
	ReloadInfo           types.Object `tfsdk:"reload_info"`
	Hooks                types.Object `tfsdk:"hooks"`
	CertificateData      types.Object `tfsdk:"certificate_data"`
}

type CertificateField struct {
	Static         types.String `tfsdk:"static"`
	DeviceMetadata types.String `tfsdk:"device_metadata"`
}

var certificateFieldType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"static":          types.StringType,
		"device_metadata": types.StringType,
	},
}

type CertificateFieldList struct {
	Static         types.List `tfsdk:"static"`
	DeviceMetadata types.List `tfsdk:"device_metadata"`
}

var certificateFieldListType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"static": types.ListType{
			ElemType: types.StringType,
		},
		"device_metadata": types.ListType{
			ElemType: types.StringType,
		},
	},
}

type CertificateDataModel struct {
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

var certificateDataType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"common_name":         certificateFieldType,
		"sans":                certificateFieldListType,
		"organization":        certificateFieldListType,
		"organizational_unit": certificateFieldListType,
		"locality":            certificateFieldListType,
		"province":            certificateFieldListType,
		"street_address":      certificateFieldListType,
		"postal_code":         certificateFieldListType,
		"country":             certificateFieldListType,
	},
}

func (certDataModel CertificateDataModel) ToAPI(ctx context.Context) (v20231101.X509Fields, diag.Diagnostics) {
	var diags diag.Diagnostics

	cn := &CertificateField{}
	diags = certDataModel.CommonName.As(ctx, &cn, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	sans := &CertificateFieldList{}
	diags = certDataModel.SANs.As(ctx, &sans, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	org := &CertificateFieldList{}
	diags = certDataModel.Organization.As(ctx, &org, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	ou := &CertificateFieldList{}
	diags = certDataModel.OrganizationalUnit.As(ctx, &ou, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	locality := &CertificateFieldList{}
	diags = certDataModel.Locality.As(ctx, &locality, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	province := &CertificateFieldList{}
	diags = certDataModel.Province.As(ctx, &province, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	street := &CertificateFieldList{}
	diags = certDataModel.StreetAddress.As(ctx, &street, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	postal := &CertificateFieldList{}
	diags = certDataModel.PostalCode.As(ctx, &postal, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	country := &CertificateFieldList{}
	diags = certDataModel.Country.As(ctx, &country, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})

	cnField := &v20231101.CertificateField{}
	if cn != nil {
		if cn.Static.ValueString() != "" {
			err := cnField.FromCertificateFieldStatic(v20231101.CertificateFieldStatic{
				Static: cn.Static.ValueString(),
			})
			if err != nil {
				diags.AddError("workload static common name", err.Error())
				return v20231101.X509Fields{}, diags
			}
		} else if cn.DeviceMetadata.ValueString() != "" {
			err := cnField.FromCertificateFieldDeviceMetadata(v20231101.CertificateFieldDeviceMetadata{
				DeviceMetadata: cn.DeviceMetadata.ValueString(),
			})
			if err != nil {
				diags.AddError("workload device metadata common name", err.Error())
				return v20231101.X509Fields{}, diags
			}
		}
	}

	countryCFL, d := toCertificateFieldList(ctx, country)
	diags.Append(d...)

	localityCFL, d := toCertificateFieldList(ctx, locality)
	diags.Append(d...)

	orgCFL, d := toCertificateFieldList(ctx, org)
	diags.Append(d...)

	ouCFL, d := toCertificateFieldList(ctx, ou)
	diags.Append(d...)

	postalCFL, d := toCertificateFieldList(ctx, postal)
	diags.Append(d...)

	provinceCFL, d := toCertificateFieldList(ctx, province)
	diags.Append(d...)

	sansCFL, d := toCertificateFieldList(ctx, sans)
	diags.Append(d...)

	streetCFL, d := toCertificateFieldList(ctx, street)
	diags.Append(d...)

	return v20231101.X509Fields{
		CommonName:         cnField,
		Sans:               sansCFL,
		Country:            countryCFL,
		Locality:           localityCFL,
		Organization:       orgCFL,
		OrganizationalUnit: ouCFL,
		PostalCode:         postalCFL,
		Province:           provinceCFL,
		StreetAddress:      streetCFL,
	}, diags
}

func fromAPI(ctx context.Context, workload *v20231101.Workload, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	certInfo, ds := CertInfoFromAPI(ctx, workload.CertificateInfo, state)
	diags.Append(ds...)

	keyInfo, ds := KeyInfoFromAPI(ctx, workload.KeyInfo, state)
	diags.Append(ds...)

	reloadInfo, ds := ReloadInfoFromAPI(ctx, workload.ReloadInfo, state)
	diags.Append(ds...)

	workloadType, d := utils.ToOptionalString(ctx, workload.WorkloadType, state, path.Root("workload_type"))
	diags.Append(d...)

	model := &Model{
		Slug:            types.StringValue(workload.Slug),
		WorkloadType:    workloadType,
		DisplayName:     types.StringValue(workload.DisplayName),
		AuthorityID:     types.StringValue(workload.AuthorityID),
		CertificateInfo: certInfo,
		KeyInfo:         keyInfo,
		ReloadInfo:      reloadInfo,
	}

	if workload.Hooks != nil {
		sign, d := HookFromAPI(ctx, workload.Hooks.Sign, path.Root("hooks").AtName("sign"), state)
		diags.Append(d...)

		renew, d := HookFromAPI(ctx, workload.Hooks.Renew, path.Root("hooks").AtName("renew"), state)
		diags.Append(d...)

		hooksObj, d := basetypes.NewObjectValue(HooksObjectType, map[string]attr.Value{
			"sign":  sign,
			"renew": renew,
		})
		diags.Append(d...)

		model.Hooks = hooksObj
	} else {
		model.Hooks = basetypes.NewObjectNull(HooksObjectType)
	}

	x509Fields, err := workload.AsX509Fields()
	if err != nil {
		diags.AddError("parse workload response", err.Error())
		return nil, diags
	}

	certData, ds := CertDataFromAPI(ctx, x509Fields, state)
	diags.Append(ds...)

	model.CertificateData = certData

	return model, diags
}

func ReloadInfoFromAPI(ctx context.Context, reloadInfo *v20231101.EndpointReloadInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	pidFile, d := utils.ToOptionalString(ctx, reloadInfo.PidFile, state, path.Root("reload_info").AtName("method"))
	diags.Append(d...)

	signal, d := utils.ToOptionalInt(ctx, reloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
	diags.Append(d...)

	unitName, d := utils.ToOptionalString(ctx, reloadInfo.UnitName, state, path.Root("reload_info").AtName("unit_name"))
	diags.Append(d...)

	out, d := basetypes.NewObjectValue(reloadInfoAttrTypes, map[string]attr.Value{
		"method":    types.StringValue(string(reloadInfo.Method)),
		"pid_file":  pidFile,
		"signal":    signal,
		"unit_name": unitName,
	})
	diags.Append(d...)

	return out, diags
}

func KeyInfoFromAPI(ctx context.Context, keyInfo *v20231101.EndpointKeyInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if keyInfo == nil {
		return basetypes.NewObjectNull(keyInfoAttrTypes), diags
	}

	format, ds := utils.ToOptionalString(ctx, keyInfo.Format, state, path.Root("key_info").AtName("format"))
	diags.Append(ds...)

	pubFile, ds := utils.ToOptionalString(ctx, keyInfo.PubFile, state, path.Root("key_info").AtName("pub_file"))
	diags.Append(ds...)

	typ, ds := utils.ToOptionalString(ctx, keyInfo.Type, state, path.Root("key_info").AtName("type"))
	diags.Append(ds...)

	protection, ds := utils.ToOptionalString(ctx, keyInfo.Protection, state, path.Root("key_info").AtName("protection"))
	diags.Append(ds...)

	out, ds := basetypes.NewObjectValue(keyInfoAttrTypes, map[string]attr.Value{
		"format":     format,
		"pub_file":   pubFile,
		"type":       typ,
		"protection": protection,
	})
	diags.Append(ds...)

	return out, diags
}

func CertInfoFromAPI(ctx context.Context, certInfo *v20231101.EndpointCertificateInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	dur, d := utils.ToEqualString(ctx, certInfo.Duration, state, path.Root("certificate_info").AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	crtFile, d := utils.ToOptionalString(ctx, certInfo.CrtFile, state, path.Root("certificate_info").AtName("crt_file"))
	diags = append(diags, d...)

	keyFile, d := utils.ToOptionalString(ctx, certInfo.KeyFile, state, path.Root("certificate_info").AtName("key_file"))
	diags = append(diags, d...)

	rootFile, d := utils.ToOptionalString(ctx, certInfo.RootFile, state, path.Root("certificate_info").AtName("root_file"))
	diags = append(diags, d...)

	uid, d := utils.ToOptionalInt(ctx, certInfo.Uid, state, path.Root("certificate_info").AtName("uid"))
	diags = append(diags, d...)

	gid, d := utils.ToOptionalInt(ctx, certInfo.Gid, state, path.Root("certificate_info").AtName("gid"))
	diags = append(diags, d...)

	mode, d := utils.ToOptionalInt(ctx, certInfo.Mode, state, path.Root("certificate_info").AtName("mode"))
	diags = append(diags, d...)

	out, d := basetypes.NewObjectValue(certInfoAttrTypes, map[string]attr.Value{
		"type":      types.StringValue(string(certInfo.Type)),
		"crt_file":  crtFile,
		"key_file":  keyFile,
		"root_file": rootFile,
		"duration":  dur,
		"gid":       gid,
		"uid":       uid,
		"mode":      mode,
	})
	diags.Append(d...)

	return out, diags
}

func CertDataFromAPI(ctx context.Context, x509Fields v20231101.X509Fields, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	cn := basetypes.NewObjectNull(certificateFieldType.AttrTypes)
	if x509Fields.CommonName != nil {
		cfStatic, err := x509Fields.CommonName.AsCertificateFieldStatic()
		if err != nil {
			diags.AddError("parse static common name", err.Error())
			return types.Object{}, diags
		}
		cfDeviceMetadata, err := x509Fields.CommonName.AsCertificateFieldDeviceMetadata()
		if err != nil {
			diags.AddError("parse device metadata common name", err.Error())
			return types.Object{}, diags
		}
		static, d := utils.ToOptionalString(ctx, &cfStatic.Static, state, path.Root("certificate_data").AtName("common_name").AtName("static"))
		diags.Append(d...)
		dm, d := utils.ToOptionalString(ctx, &cfDeviceMetadata.DeviceMetadata, state, path.Root("certificate_data").AtName("common_name").AtName("device_metadata"))
		diags.Append(d...)
		cn, d = basetypes.NewObjectValue(certificateFieldType.AttrTypes, map[string]attr.Value{
			"static":          static,
			"device_metadata": dm,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.Object{}, diags
		}
	}

	sans, d := CertificateFieldListFromAPI(ctx, x509Fields.Sans, state, path.Root("certificate_data").AtName("sans"))
	diags.Append(d...)

	org, d := CertificateFieldListFromAPI(ctx, x509Fields.Organization, state, path.Root("certificate_data").AtName("organization"))
	diags.Append(d...)

	ou, d := CertificateFieldListFromAPI(ctx, x509Fields.OrganizationalUnit, state, path.Root("certificate_data").AtName("organizational_unit"))
	diags.Append(d...)

	locality, d := CertificateFieldListFromAPI(ctx, x509Fields.Locality, state, path.Root("certificate_data").AtName("locality"))
	diags.Append(d...)

	province, d := CertificateFieldListFromAPI(ctx, x509Fields.Province, state, path.Root("certificate_data").AtName("province"))
	diags.Append(d...)

	street, d := CertificateFieldListFromAPI(ctx, x509Fields.StreetAddress, state, path.Root("certificate_data").AtName("street_address"))
	diags.Append(d...)

	postal, d := CertificateFieldListFromAPI(ctx, x509Fields.PostalCode, state, path.Root("certificate_data").AtName("postal_code"))
	diags.Append(d...)

	country, d := CertificateFieldListFromAPI(ctx, x509Fields.Country, state, path.Root("certificate_data").AtName("country"))
	diags.Append(d...)

	certData, d := basetypes.NewObjectValue(certificateDataType.AttrTypes, map[string]attr.Value{
		"common_name":         cn,
		"sans":                sans,
		"organization":        org,
		"organizational_unit": ou,
		"locality":            locality,
		"province":            province,
		"street_address":      street,
		"postal_code":         postal,
		"country":             country,
	})
	diags.Append(d...)

	return certData, diags
}

func toAPI(ctx context.Context, model *Model) (*v20231101.Workload, diag.Diagnostics) {
	hooksModel := &HooksModel{}
	diags := model.Hooks.As(ctx, &hooksModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	hooks, d := hooksModel.ToAPI(ctx)
	diags.Append(d...)

	reloadInfo := &ReloadInfoModel{}
	d = model.ReloadInfo.As(ctx, &reloadInfo, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)

	ci := &CertificateInfoModel{}
	d = model.CertificateInfo.As(ctx, &ci, basetypes.ObjectAsOptions{})
	diags.Append(d...)

	keyInfo := &KeyInfoModel{}
	d = model.KeyInfo.As(ctx, &keyInfo, basetypes.ObjectAsOptions{})
	diags.Append(d...)

	workload := &v20231101.Workload{
		DisplayName:     model.DisplayName.ValueString(),
		WorkloadType:    model.WorkloadType.ValueStringPointer(),
		Slug:            model.Slug.ValueString(),
		AuthorityID:     model.AuthorityID.ValueString(),
		CertificateInfo: ci.ToAPI(),
		KeyInfo:         keyInfo.ToAPI(),
		Hooks:           hooks,
		ReloadInfo:      reloadInfo.ToAPI(),
	}

	certDataModel := &CertificateDataModel{}
	diags = model.CertificateData.As(ctx, &certDataModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	x509Fields, ds := certDataModel.ToAPI(ctx)
	diags.Append(ds...)

	err := workload.FromX509Fields(x509Fields)
	if err != nil {
		diags.AddError("workload certificate data", err.Error())
		return nil, diags
	}

	return workload, diags
}

func toCertificateFieldList(ctx context.Context, cfl *CertificateFieldList) (*v20231101.CertificateFieldList, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfl == nil {
		return nil, diags
	}

	var static *[]string
	var deviceMetadata *[]string

	if !cfl.Static.IsNull() {
		diags.Append(cfl.Static.ElementsAs(ctx, &static, false)...)
	}
	if !cfl.DeviceMetadata.IsNull() {
		diags.Append(cfl.DeviceMetadata.ElementsAs(ctx, &deviceMetadata, false)...)
	}

	return &v20231101.CertificateFieldList{
		Static:         static,
		DeviceMetadata: deviceMetadata,
	}, diags
}

func CertificateFieldListFromAPI(ctx context.Context, cfl *v20231101.CertificateFieldList, state utils.AttributeGetter, p path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if cfl == nil {
		return basetypes.NewObjectNull(certificateFieldListType.AttrTypes), diags
	}

	static, d := utils.ToOptionalList(ctx, cfl.Static, state, p.AtName("static"))
	diags.Append(d...)

	deviceMetadata, d := utils.ToOptionalList(ctx, cfl.DeviceMetadata, state, p.AtName("device_metadata"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(certificateFieldListType.AttrTypes, map[string]attr.Value{
		"static":          static,
		"device_metadata": deviceMetadata,
	})
	diags.Append(d...)

	return obj, diags
}
