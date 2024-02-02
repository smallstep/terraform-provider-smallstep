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
	ID                   types.String          `tfsdk:"id"`
	WorkloadType         types.String          `tfsdk:"workload_type"`
	DisplayName          types.String          `tfsdk:"display_name"`
	Slug                 types.String          `tfsdk:"slug"`
	DeviceCollectionSlug types.String          `tfsdk:"device_collection_slug"`
	CertificateInfo      *CertificateInfoModel `tfsdk:"certificate_info"`
	KeyInfo              *KeyInfoModel         `tfsdk:"key_info"`
	// ReloadInfo and Hooks are optional. Need to use Object type to support
	// the "unknown" state.
	ReloadInfo      types.Object `tfsdk:"reload_info"`
	Hooks           types.Object `tfsdk:"hooks"`
	AdminEmails     types.Set    `tfsdk:"admin_emails"`
	CertificateData types.Object `tfsdk:"certificate_data"`
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
	// PreserveSubject    types.Bool   `tfsdk:"preserve_subject"`
	// PreseveSANs        types.Bool   `tfsdk:"preserve_sans"`
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
		// "preserve_subject":    types.BoolType,
		// "preserve_sans":       types.BoolType,
	},
}

func fromAPI(ctx context.Context, workload *v20231101.Workload, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	ciDuration, d := utils.ToEqualString(ctx, workload.CertificateInfo.Duration, state, path.Root("certificate_info").AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	ciCrtFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.CrtFile, state, path.Root("certificate_info").AtName("crt_file"))
	diags = append(diags, d...)

	ciKeyFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.KeyFile, state, path.Root("certificate_info").AtName("key_file"))
	diags = append(diags, d...)

	ciRootFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.RootFile, state, path.Root("certificate_info").AtName("root_file"))
	diags = append(diags, d...)

	ciUID, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Uid, state, path.Root("certificate_info").AtName("uid"))
	diags = append(diags, d...)

	ciGID, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Gid, state, path.Root("certificate_info").AtName("gid"))
	diags = append(diags, d...)

	ciMode, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Mode, state, path.Root("certificate_info").AtName("mode"))
	diags = append(diags, d...)

	model := &Model{
		ID:           types.StringValue(workload.Slug),
		Slug:         types.StringValue(workload.Slug),
		WorkloadType: types.StringValue(workload.WorkloadType),
		DisplayName:  types.StringValue(workload.DisplayName),
		CertificateInfo: &CertificateInfoModel{
			Type:     types.StringValue(string(workload.CertificateInfo.Type)),
			Duration: ciDuration,
			CrtFile:  ciCrtFile,
			KeyFile:  ciKeyFile,
			RootFile: ciRootFile,
			UID:      ciUID,
			GID:      ciGID,
			Mode:     ciMode,
		},
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

	if workload.KeyInfo != nil {
		format, d := utils.ToOptionalString(ctx, workload.KeyInfo.Format, state, path.Root("key_info").AtName("format"))
		diags = append(diags, d...)

		pubFile, d := utils.ToOptionalString(ctx, workload.KeyInfo.PubFile, state, path.Root("key_info").AtName("pub_file"))
		diags = append(diags, d...)

		typ, d := utils.ToOptionalString(ctx, workload.KeyInfo.Type, state, path.Root("key_info").AtName("type"))
		diags = append(diags, d...)

		protection, d := utils.ToOptionalString(ctx, workload.KeyInfo.Protection, state, path.Root("key_info").AtName("protection"))
		diags = append(diags, d...)

		model.KeyInfo = &KeyInfoModel{
			Format:     format,
			PubFile:    pubFile,
			Type:       typ,
			Protection: protection,
		}
	}

	if workload.ReloadInfo != nil {
		pidFile, d := utils.ToOptionalString(ctx, workload.ReloadInfo.PidFile, state, path.Root("reload_info").AtName("method"))
		diags.Append(d...)

		signal, d := utils.ToOptionalInt(ctx, workload.ReloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
		diags.Append(d...)

		unitName, d := utils.ToOptionalString(ctx, workload.ReloadInfo.UnitName, state, path.Root("reload_info").AtName("unit_name"))
		diags.Append(d...)

		reloadInfoObject, d := basetypes.NewObjectValue(ReloadInfoType, map[string]attr.Value{
			"method":    types.StringValue(string(workload.ReloadInfo.Method)),
			"pid_file":  pidFile,
			"signal":    signal,
			"unit_name": unitName,
		})
		diags.Append(d...)
		model.ReloadInfo = reloadInfoObject
	} else {
		model.ReloadInfo = basetypes.NewObjectNull(ReloadInfoType)
	}

	x509Fields, err := workload.AsX509Fields()
	if err != nil {
		diags.AddError("parse workload response", err.Error())
		return nil, diags
	}

	cn := basetypes.NewObjectNull(certificateFieldType.AttrTypes)
	if x509Fields.CommonName != nil {
		cfStatic, err := x509Fields.CommonName.AsCertificateFieldStatic()
		if err != nil {
			diags.AddError("parse static common name", err.Error())
			return nil, diags
		}
		cfDeviceMetadata, err := x509Fields.CommonName.AsCertificateFieldDeviceMetadata()
		if err != nil {
			diags.AddError("parse device metadata common name", err.Error())
			return nil, diags
		}
		static, d := utils.ToOptionalString(ctx, &cfStatic.Static, state, path.Root("certificate_data").AtName("common_name").AtName("static"))
		diags.Append(d...)
		dm, d := utils.ToOptionalString(ctx, &cfDeviceMetadata.DeviceMetadata, state, path.Root("certificate_data").AtName("common_name").AtName("device_metadata"))
		cn, d = basetypes.NewObjectValue(certificateFieldType.AttrTypes, map[string]attr.Value{
			"static":          static,
			"device_metadata": dm,
		})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
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

	/*
		preserveSubject, d := utils.ToOptionalBool(ctx, x509Fields.PreserveSubject, state, path.Root("certificate_data").AtName("preserve_subject"))
		diags.Append(d...)

		preserveSANs, d := utils.ToOptionalBool(ctx, x509Fields.PreserveSANs, state, path.Root("certificate_data").AtName("preserve_sans"))
		diags.Append(d...)
	*/

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
		// "preserve_subject":    preserveSubject,
		// "preserve_sans":       preserveSANs,
	})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	model.CertificateData = certData

	return model, diags
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

	var adminEmails []string
	diags.Append(model.AdminEmails.ElementsAs(ctx, &adminEmails, false)...)

	ci := model.CertificateInfo.ToAPI()

	workload := &v20231101.Workload{
		DisplayName:     model.DisplayName.ValueString(),
		WorkloadType:    model.WorkloadType.ValueString(),
		Slug:            model.Slug.ValueString(),
		CertificateInfo: &ci,
		KeyInfo:         model.KeyInfo.ToAPI(),
		Hooks:           hooks,
		ReloadInfo:      reloadInfo.ToAPI(),
		AdminEmails:     &adminEmails,
	}

	certDataModel := &CertificateDataModel{}
	diags = model.CertificateData.As(ctx, &certDataModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
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
				return nil, diags
			}
		} else if cn.DeviceMetadata.ValueString() != "" {
			err := cnField.FromCertificateFieldDeviceMetadata(v20231101.CertificateFieldDeviceMetadata{
				DeviceMetadata: cn.DeviceMetadata.ValueString(),
			})
			if err != nil {
				diags.AddError("workload device metadata common name", err.Error())
				return nil, diags
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

	streetCFL, d := toCertificateFieldList(ctx, street)
	diags.Append(d...)

	err := workload.FromX509Fields(v20231101.X509Fields{
		CommonName:         cnField,
		Country:            countryCFL,
		Locality:           localityCFL,
		Organization:       orgCFL,
		OrganizationalUnit: ouCFL,
		PostalCode:         postalCFL,
		Province:           provinceCFL,
		StreetAddress:      streetCFL,
	})
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
		return basetypes.NewObjectNull(HooksObjectType), diags
	}

	static, d := utils.ToOptionalList(ctx, cfl.Static, state, p.AtName("static"))
	diags.Append(d...)

	deviceMetadata, d := utils.ToOptionalList(ctx, cfl.DeviceMetadata, state, p.AtName("device_metadata"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(HooksObjectType, map[string]attr.Value{
		"static":          static,
		"device_metadata": deviceMetadata,
	})
	diags.Append(d...)

	return obj, diags
}
