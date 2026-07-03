package credential

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_credential"

type CredentialModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Certificate types.Object `tfsdk:"certificate"`
	Key         types.Object `tfsdk:"key"`
	Policy      types.Object `tfsdk:"policy"`
	Files       types.Object `tfsdk:"files"`
}

type CertificateModel struct {
	AuthorityID types.String `tfsdk:"authority_id"`
	Duration    types.String `tfsdk:"duration"`
	X509        types.Object `tfsdk:"x509"`
}

var certificateAttributes = map[string]attr.Type{
	"authority_id": types.StringType,
	"duration":     types.StringType,
	"x509":         types.ObjectType{AttrTypes: x509Attributes},
}

type X509Model struct {
	CommonName         types.Object `tfsdk:"common_name"`
	SANs               types.Object `tfsdk:"sans"`
	TypedSANs          types.Object `tfsdk:"typed_sans"`
	Organization       types.Object `tfsdk:"organization"`
	OrganizationalUnit types.Object `tfsdk:"organizational_unit"`
	Locality           types.Object `tfsdk:"locality"`
	Province           types.Object `tfsdk:"province"`
	StreetAddress      types.Object `tfsdk:"street_address"`
	PostalCode         types.Object `tfsdk:"postal_code"`
	Country            types.Object `tfsdk:"country"`
	SerialNumber       types.Object `tfsdk:"serial_number"`
	GivenName          types.Object `tfsdk:"given_name"`
	Surname            types.Object `tfsdk:"surname"`
	ExtendedKeyUsage   types.List   `tfsdk:"extended_key_usage"`
	CustomExtensions   types.List   `tfsdk:"custom_extensions"`
}

var x509Attributes = map[string]attr.Type{
	"common_name":         types.ObjectType{AttrTypes: certificateFieldAttributes},
	"sans":                types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"typed_sans":          types.ObjectType{AttrTypes: typedSANsAttributes},
	"organization":        types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"organizational_unit": types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"locality":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"province":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"street_address":      types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"postal_code":         types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"country":             types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"serial_number":       types.ObjectType{AttrTypes: certificateFieldAttributes},
	"given_name":          types.ObjectType{AttrTypes: certificateFieldAttributes},
	"surname":             types.ObjectType{AttrTypes: certificateFieldAttributes},
	"extended_key_usage":  types.ListType{ElemType: types.StringType},
	"custom_extensions":   types.ListType{ElemType: types.ObjectType{AttrTypes: customExtensionAttributes}},
}

type TypedSANsModel struct {
	DNSNames           types.Object `tfsdk:"dns_names"`
	IPAddresses        types.Object `tfsdk:"ip_addresses"`
	EmailAddresses     types.Object `tfsdk:"email_addresses"`
	URIs               types.Object `tfsdk:"uris"`
	UserPrincipalNames types.Object `tfsdk:"user_principal_names"`
}

var typedSANsAttributes = map[string]attr.Type{
	"dns_names":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"ip_addresses":         types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"email_addresses":      types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"uris":                 types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"user_principal_names": types.ObjectType{AttrTypes: certificateFieldListAttributes},
}

type CustomExtensionModel struct {
	OID      types.String `tfsdk:"oid"`
	Critical types.Bool   `tfsdk:"critical"`
	Value    types.String `tfsdk:"value"`
}

var customExtensionAttributes = map[string]attr.Type{
	"oid":      types.StringType,
	"critical": types.BoolType,
	"value":    types.StringType,
}

type KeyModel struct {
	Type       types.String `tfsdk:"type"`
	Protection types.String `tfsdk:"protection"`
	PubFile    types.String `tfsdk:"pub_file"`
}

var keyAttributes = map[string]attr.Type{
	"type":       types.StringType,
	"protection": types.StringType,
	"pub_file":   types.StringType,
}

type FilesModel struct {
	RootFile  types.String `tfsdk:"root_file"`
	CrtFile   types.String `tfsdk:"crt_file"`
	KeyFile   types.String `tfsdk:"key_file"`
	KeyFormat types.String `tfsdk:"key_format"`
	UID       types.Int64  `tfsdk:"uid"`
	GID       types.Int64  `tfsdk:"gid"`
	Mode      types.Int64  `tfsdk:"mode"`
}

func (f *FilesModel) isEmpty() bool {
	switch {
	case f.RootFile.ValueString() != "":
		return false
	case f.CrtFile.ValueString() != "":
		return false
	case f.KeyFile.ValueString() != "":
		return false
	case f.KeyFormat.ValueString() != "":
		return false
	case f.UID.ValueInt64() != 0:
		return false
	case f.GID.ValueInt64() != 0:
		return false
	case f.Mode.ValueInt64() != 0:
		return false
	}
	return true
}

var filesAttributes = map[string]attr.Type{
	"root_file":  types.StringType,
	"crt_file":   types.StringType,
	"key_file":   types.StringType,
	"key_format": types.StringType,
	"uid":        types.Int64Type,
	"gid":        types.Int64Type,
	"mode":       types.Int64Type,
}

type PolicyModel struct {
	Assurance types.List `tfsdk:"assurance"`
	OS        types.List `tfsdk:"os"`
	Ownership types.List `tfsdk:"ownership"`
	Source    types.List `tfsdk:"source"`
	Tags      types.List `tfsdk:"tags"`
}

func (p *PolicyModel) isEmpty() bool {
	switch {
	case len(p.Assurance.Elements()) > 0:
		return false
	case len(p.OS.Elements()) > 0:
		return false
	case len(p.Ownership.Elements()) > 0:
		return false
	case len(p.Source.Elements()) > 0:
		return false
	case len(p.Tags.Elements()) > 0:
		return false
	}
	return true
}

var policyAttributes = map[string]attr.Type{
	"assurance": types.ListType{ElemType: types.StringType},
	"os":        types.ListType{ElemType: types.StringType},
	"ownership": types.ListType{ElemType: types.StringType},
	"source":    types.ListType{ElemType: types.StringType},
	"tags":      types.ListType{ElemType: types.StringType},
}

type CertificateFieldModel struct {
	Static         types.String `tfsdk:"static"`
	DeviceMetadata types.String `tfsdk:"device_metadata"`
}

var certificateFieldAttributes = map[string]attr.Type{
	"static":          types.StringType,
	"device_metadata": types.StringType,
}

type CertificateFieldListModel struct {
	Static         types.List `tfsdk:"static"`
	DeviceMetadata types.List `tfsdk:"device_metadata"`
}

var certificateFieldListAttributes = map[string]attr.Type{
	"static":          types.ListType{ElemType: types.StringType},
	"device_metadata": types.ListType{ElemType: types.StringType},
}

func (k *KeyModel) toAPI() v20250101.CredentialKey {
	return v20250101.CredentialKey{
		Type:       (*v20250101.CredentialKeyType)(k.Type.ValueStringPointer()),
		Protection: (*v20250101.CredentialKeyProtection)(k.Protection.ValueStringPointer()),
		PubFile:    k.PubFile.ValueStringPointer(),
	}
}

func (m *CertificateModel) toAPI(ctx context.Context, diags *diag.Diagnostics) v20250101.CredentialCertificate {
	cert := v20250101.CredentialCertificate{
		Type:        v20250101.CredentialCertificateTypeX509,
		AuthorityID: m.AuthorityID.ValueString(),
	}

	if !m.Duration.IsNull() && !m.Duration.IsUnknown() && m.Duration.ValueString() != "" {
		cert.Duration = m.Duration.ValueStringPointer()
	}

	if !m.X509.IsNull() && !m.X509.IsUnknown() {
		x509 := &X509Model{}
		ds := m.X509.As(ctx, &x509, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		x := x509.toAPI(ctx, diags)

		if err := cert.Fields.FromX509Fields(x); err != nil {
			diags.AddError("Format X509 Attributes", err.Error())
		}
	}

	return cert
}

func (m *FilesModel) toAPI() *v20250101.CredentialFiles {
	if m == nil {
		return nil
	}

	return &v20250101.CredentialFiles{
		RootFile:  m.RootFile.ValueStringPointer(),
		CrtFile:   m.CrtFile.ValueStringPointer(),
		KeyFile:   m.KeyFile.ValueStringPointer(),
		KeyFormat: (*v20250101.CredentialFilesKeyFormat)(m.KeyFormat.ValueStringPointer()),
		Uid:       utils.ToIntPointer(m.UID.ValueInt64Pointer()),
		Gid:       utils.ToIntPointer(m.GID.ValueInt64Pointer()),
		Mode:      utils.ToIntPointer(m.Mode.ValueInt64Pointer()),
	}
}

func (x509 *X509Model) toAPI(ctx context.Context, diags *diag.Diagnostics) v20250101.X509Fields {
	fields := v20250101.X509Fields{
		CommonName:         asCertificateField(ctx, diags, x509.CommonName),
		Sans:               asCertificateFieldList(ctx, diags, x509.SANs),
		TypedSans:          asTypedSANs(ctx, diags, x509.TypedSANs),
		Country:            asCertificateFieldList(ctx, diags, x509.Country),
		Locality:           asCertificateFieldList(ctx, diags, x509.Locality),
		Organization:       asCertificateFieldList(ctx, diags, x509.Organization),
		OrganizationalUnit: asCertificateFieldList(ctx, diags, x509.OrganizationalUnit),
		PostalCode:         asCertificateFieldList(ctx, diags, x509.PostalCode),
		Province:           asCertificateFieldList(ctx, diags, x509.Province),
		StreetAddress:      asCertificateFieldList(ctx, diags, x509.StreetAddress),
		SerialNumber:       asCertificateField(ctx, diags, x509.SerialNumber),
		GivenName:          asCertificateField(ctx, diags, x509.GivenName),
		Surname:            asCertificateField(ctx, diags, x509.Surname),
		CustomExtensions:   customExtensionsToAPI(ctx, diags, x509.CustomExtensions),
	}

	if len(x509.ExtendedKeyUsage.Elements()) > 0 {
		diags.Append(x509.ExtendedKeyUsage.ElementsAs(ctx, &fields.ExtendedKeyUsage, false)...)
	}

	return fields
}

func asTypedSANs(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20250101.X509TypedSANs {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &TypedSANsModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return &v20250101.X509TypedSANs{
		DnsNames:           asCertificateFieldList(ctx, diags, model.DNSNames),
		IpAddresses:        asCertificateFieldList(ctx, diags, model.IPAddresses),
		EmailAddresses:     asCertificateFieldList(ctx, diags, model.EmailAddresses),
		Uris:               asCertificateFieldList(ctx, diags, model.URIs),
		UserPrincipalNames: asCertificateFieldList(ctx, diags, model.UserPrincipalNames),
	}
}

func customExtensionsToAPI(ctx context.Context, diags *diag.Diagnostics, list types.List) *[]v20250101.X509CustomExtension {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	models := []CustomExtensionModel{}
	diags.Append(list.ElementsAs(ctx, &models, false)...)

	exts := make([]v20250101.X509CustomExtension, 0, len(models))
	for i, m := range models {
		value, err := base64.StdEncoding.DecodeString(m.Value.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("certificate").AtName("x509").AtName("custom_extensions").AtListIndex(i).AtName("value"),
				"Invalid Custom Extension Value",
				fmt.Sprintf("The value must be base64-encoded: %v", err),
			)
			continue
		}

		exts = append(exts, v20250101.X509CustomExtension{
			Oid:      m.OID.ValueString(),
			Critical: m.Critical.ValueBoolPointer(),
			Value:    value,
		})
	}

	return &exts
}

func (p *PolicyModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.PolicyMatchCriteria {
	if p == nil {
		return nil
	}

	policy := &v20250101.PolicyMatchCriteria{}

	if len(p.Assurance.Elements()) > 0 {
		diags.Append(p.Assurance.ElementsAs(ctx, &policy.Assurance, false)...)
	}
	if len(p.OS.Elements()) > 0 {
		diags.Append(p.OS.ElementsAs(ctx, &policy.OperatingSystem, false)...)
	}
	if len(p.Ownership.Elements()) > 0 {
		diags.Append(p.Ownership.ElementsAs(ctx, &policy.Ownership, false)...)
	}
	if len(p.Source.Elements()) > 0 {
		diags.Append(p.Source.ElementsAs(ctx, &policy.Source, false)...)
	}
	if len(p.Tags.Elements()) > 0 {
		diags.Append(p.Tags.ElementsAs(ctx, &policy.Tags, false)...)
	}

	return policy
}

func (cf *CertificateFieldModel) toAPI() *v20250101.CertificateField {
	return &v20250101.CertificateField{
		Static:         cf.Static.ValueStringPointer(),
		DeviceMetadata: cf.DeviceMetadata.ValueStringPointer(),
	}
}

func (cfl *CertificateFieldListModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *v20250101.CertificateFieldList {
	var static *[]string
	var deviceMetadata *[]string

	diags.Append(cfl.Static.ElementsAs(ctx, &static, false)...)
	diags.Append(cfl.DeviceMetadata.ElementsAs(ctx, &deviceMetadata, false)...)

	return &v20250101.CertificateFieldList{
		Static:         static,
		DeviceMetadata: deviceMetadata,
	}
}

func asCertificateFieldList(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20250101.CertificateFieldList {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &CertificateFieldListModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return model.toAPI(ctx, diags)
}

func asCertificateField(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20250101.CertificateField {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &CertificateFieldModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return model.toAPI()
}

func toAPI(ctx context.Context, diags *diag.Diagnostics, model *CredentialModel) v20250101.Credential {
	cert := CertificateModel{}
	ds := model.Certificate.As(ctx, &cert, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	key := KeyModel{}
	ds = model.Key.As(ctx, &key, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	policy := &PolicyModel{}
	ds = model.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	files := &FilesModel{}
	ds = model.Files.As(ctx, &files, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.Credential{
		Id:          model.ID.ValueStringPointer(),
		Slug:        model.Slug.ValueString(),
		Certificate: cert.toAPI(ctx, diags),
		Key:         key.toAPI(),
		Policy:      policy.toAPI(ctx, diags),
		Files:       files.toAPI(),
	}
}

func fromAPI(ctx context.Context, diags *diag.Diagnostics, credential *v20250101.Credential, state utils.AttributeGetter) CredentialModel {
	return CredentialModel{
		ID:          types.StringPointerValue(credential.Id),
		Slug:        types.StringValue(credential.Slug),
		Certificate: certificateObjectFromAPI(ctx, diags, credential.Certificate, state),
		Key:         keyObjectFromAPI(ctx, diags, credential.Key, state),
		Policy:      policyObjectFromAPI(ctx, diags, credential.Policy, state),
		Files:       filesObjectFromAPI(ctx, diags, credential.Files, state),
	}
}

func certificateObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cert v20250101.CredentialCertificate, state utils.AttributeGetter) types.Object {
	dur, d := utils.ToEqualString(ctx, cert.Duration, state, path.Root("certificate").AtName("duration"), utils.IsDurationEqual)
	diags.Append(d...)

	x509Obj := basetypes.NewObjectNull(x509Attributes)
	x509, err := cert.Fields.AsX509Fields()
	if err != nil {
		diags.AddError("Parse certificate x509 attributes", err.Error())
	} else {
		x509Obj = x509ObjectFromAPI(ctx, diags, x509, state)
	}

	out, d := basetypes.NewObjectValue(certificateAttributes, map[string]attr.Value{
		"duration":     dur,
		"x509":         x509Obj,
		"authority_id": types.StringValue(cert.AuthorityID),
	})
	diags.Append(d...)

	return out
}

func filesObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, files *v20250101.CredentialFiles, state utils.AttributeGetter) types.Object {
	p := path.Root("files")

	if files == nil || reflect.DeepEqual(files, new(v20250101.CredentialFiles)) {
		// See comments in policyObjectFromAPI regarding empty objects.
		obj := &FilesModel{}
		d := state.GetAttribute(ctx, path.Root("files"), &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(filesAttributes)
		}

		if obj.isEmpty() {
			out, d := basetypes.NewObjectValue(filesAttributes, map[string]attr.Value{
				"root_file":  obj.RootFile,
				"crt_file":   obj.CrtFile,
				"key_file":   obj.KeyFile,
				"key_format": obj.KeyFormat,
				"uid":        obj.UID,
				"gid":        obj.GID,
				"mode":       obj.Mode,
			})
			diags.Append(d...)
			return out
		}

		return basetypes.NewObjectNull(policyAttributes)
	}

	rootFile, d := utils.ToOptionalString(ctx, files.RootFile, state, p.AtName("root_file"))
	diags.Append(d...)

	crtFile, d := utils.ToOptionalString(ctx, files.CrtFile, state, p.AtName("crt_file"))
	diags.Append(d...)

	keyFile, d := utils.ToOptionalString(ctx, files.KeyFile, state, p.AtName("key_file"))
	diags.Append(d...)

	format, d := utils.ToOptionalString(ctx, files.KeyFormat, state, p.AtName("key_format"))
	diags.Append(d...)

	uid, d := utils.ToOptionalInt(ctx, files.Uid, state, p.AtName("uid"))
	diags.Append(d...)

	gid, d := utils.ToOptionalInt(ctx, files.Gid, state, p.AtName("gid"))
	diags.Append(d...)

	mode, d := utils.ToOptionalInt(ctx, files.Mode, state, p.AtName("mode"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(filesAttributes, map[string]attr.Value{
		"root_file":  rootFile,
		"crt_file":   crtFile,
		"key_file":   keyFile,
		"key_format": format,
		"uid":        uid,
		"gid":        gid,
		"mode":       mode,
	})
	diags.Append(d...)

	return obj
}

func policyObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, policy *v20250101.PolicyMatchCriteria, state utils.AttributeGetter) types.Object {
	if policy == nil || reflect.DeepEqual(policy, new(v20250101.PolicyMatchCriteria)) {
		// Users can set non-null empty policies in terraform config, such as
		// `policy = {}` or `policy = { assurance = [] }`.
		// The API will return a nil policy object for all of these, but
		// terraform does not interpret these as null. If we set a null object
		// in state after an empty policy was applied then terraform will raise
		// an "inconsistent result after apply" error. To avoid that we have to
		// examine the object that was applied and if it was empty use that.

		obj := &PolicyModel{}
		d := state.GetAttribute(ctx, path.Root("policy"), &obj)
		diags.Append(d...)

		// State had a null object, not an empty object, so we don't have to
		// worry about any inconsistencies.
		if obj == nil {
			return basetypes.NewObjectNull(policyAttributes)
		}

		// State had some empty object that is equivalent to the API's nil policy.
		// We use the object from state to avoid errors.
		if obj.isEmpty() {
			o, d := basetypes.NewObjectValue(policyAttributes, map[string]attr.Value{
				"assurance": obj.Assurance,
				"os":        obj.OS,
				"ownership": obj.Ownership,
				"source":    obj.Source,
				"tags":      obj.Tags,
			})
			diags.Append(d...)
			return o
		}

		// The object in state was neither null nor empty. There is some material
		// inconsistency between state and the API. We return a null object to
		// notify terraform of the discrepancy.
		return basetypes.NewObjectNull(policyAttributes)
	}

	assurance, d := utils.ToOptionalList(ctx, policy.Assurance, state, path.Root("policy").AtName("assurance"))
	diags.Append(d...)

	os, d := utils.ToOptionalList(ctx, policy.OperatingSystem, state, path.Root("policy").AtName("os"))
	diags.Append(d...)

	ownership, d := utils.ToOptionalList(ctx, policy.Ownership, state, path.Root("policy").AtName("ownership"))
	diags.Append(d...)

	source, d := utils.ToOptionalList(ctx, policy.Source, state, path.Root("policy").AtName("source"))
	diags.Append(d...)

	tags, d := utils.ToOptionalList(ctx, policy.Tags, state, path.Root("policy").AtName("tags"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(policyAttributes, map[string]attr.Value{
		"assurance": assurance,
		"os":        os,
		"ownership": ownership,
		"source":    source,
		"tags":      tags,
	})
	diags.Append(d...)

	return obj
}

func certificateFieldObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cf *v20250101.CertificateField, state utils.AttributeGetter, p path.Path) types.Object {
	if cf == nil {
		return basetypes.NewObjectNull(certificateFieldAttributes)
	}

	static, d := utils.ToOptionalString(ctx, cf.Static, state, p.AtName("static"))
	diags.Append(d...)

	dm, d := utils.ToOptionalString(ctx, cf.DeviceMetadata, state, p.AtName("device_metadata"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(certificateFieldAttributes, map[string]attr.Value{
		"static":          static,
		"device_metadata": dm,
	})
	diags.Append(d...)

	return obj
}

func certificateFieldListObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cfl *v20250101.CertificateFieldList, state utils.AttributeGetter, p path.Path) types.Object {
	if cfl == nil {
		return basetypes.NewObjectNull(certificateFieldListAttributes)
	}

	static, d := utils.ToOptionalList(ctx, cfl.Static, state, p.AtName("static"))
	diags.Append(d...)

	deviceMetadata, d := utils.ToOptionalList(ctx, cfl.DeviceMetadata, state, p.AtName("device_metadata"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(certificateFieldListAttributes, map[string]attr.Value{
		"static":          static,
		"device_metadata": deviceMetadata,
	})
	diags.Append(d...)

	return obj
}

func x509ObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, x509 v20250101.X509Fields, state utils.AttributeGetter) types.Object {
	p := path.Root("certificate").AtName("x509")

	eku, d := utils.ToOptionalList(ctx, x509.ExtendedKeyUsage, state, p.AtName("extended_key_usage"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(x509Attributes, map[string]attr.Value{
		"common_name":         certificateFieldObjectFromAPI(ctx, diags, x509.CommonName, state, p.AtName("common_name")),
		"sans":                certificateFieldListObjectFromAPI(ctx, diags, x509.Sans, state, p.AtName("sans")),
		"typed_sans":          typedSANsObjectFromAPI(ctx, diags, x509.TypedSans, state, p.AtName("typed_sans")),
		"organization":        certificateFieldListObjectFromAPI(ctx, diags, x509.Organization, state, p.AtName("organization")),
		"organizational_unit": certificateFieldListObjectFromAPI(ctx, diags, x509.OrganizationalUnit, state, p.AtName("organizational_unit")),
		"locality":            certificateFieldListObjectFromAPI(ctx, diags, x509.Locality, state, p.AtName("locality")),
		"province":            certificateFieldListObjectFromAPI(ctx, diags, x509.Province, state, p.AtName("province")),
		"street_address":      certificateFieldListObjectFromAPI(ctx, diags, x509.StreetAddress, state, p.AtName("street_address")),
		"postal_code":         certificateFieldListObjectFromAPI(ctx, diags, x509.PostalCode, state, p.AtName("postal_code")),
		"country":             certificateFieldListObjectFromAPI(ctx, diags, x509.Country, state, p.AtName("country")),
		"serial_number":       certificateFieldObjectFromAPI(ctx, diags, x509.SerialNumber, state, p.AtName("serial_number")),
		"given_name":          certificateFieldObjectFromAPI(ctx, diags, x509.GivenName, state, p.AtName("given_name")),
		"surname":             certificateFieldObjectFromAPI(ctx, diags, x509.Surname, state, p.AtName("surname")),
		"extended_key_usage":  eku,
		"custom_extensions":   customExtensionsListFromAPI(ctx, diags, x509.CustomExtensions, state, p.AtName("custom_extensions")),
	})
	diags.Append(d...)

	return obj
}

func typedSANsObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, sans *v20250101.X509TypedSANs, state utils.AttributeGetter, p path.Path) types.Object {
	if sans == nil {
		return basetypes.NewObjectNull(typedSANsAttributes)
	}

	obj, d := basetypes.NewObjectValue(typedSANsAttributes, map[string]attr.Value{
		"dns_names":            certificateFieldListObjectFromAPI(ctx, diags, sans.DnsNames, state, p.AtName("dns_names")),
		"ip_addresses":         certificateFieldListObjectFromAPI(ctx, diags, sans.IpAddresses, state, p.AtName("ip_addresses")),
		"email_addresses":      certificateFieldListObjectFromAPI(ctx, diags, sans.EmailAddresses, state, p.AtName("email_addresses")),
		"uris":                 certificateFieldListObjectFromAPI(ctx, diags, sans.Uris, state, p.AtName("uris")),
		"user_principal_names": certificateFieldListObjectFromAPI(ctx, diags, sans.UserPrincipalNames, state, p.AtName("user_principal_names")),
	})
	diags.Append(d...)

	return obj
}

func customExtensionsListFromAPI(ctx context.Context, diags *diag.Diagnostics, exts *[]v20250101.X509CustomExtension, state utils.AttributeGetter, p path.Path) types.List {
	elemType := types.ObjectType{AttrTypes: customExtensionAttributes}

	priorList := types.List{}
	d := state.GetAttribute(ctx, p, &priorList)
	if d.HasError() {
		diags.Append(d...)
		return types.ListNull(elemType)
	}

	priors := []CustomExtensionModel{}
	if !priorList.IsNull() && !priorList.IsUnknown() {
		diags.Append(priorList.ElementsAs(ctx, &priors, false)...)
	}

	if exts == nil || len(*exts) == 0 {
		// The API returns nil for both null and empty lists. Use the prior
		// value to avoid "inconsistent result after apply" errors.
		if !priorList.IsUnknown() && (priorList.IsNull() || len(priorList.Elements()) == 0) {
			return priorList
		}
		if exts == nil {
			return types.ListNull(elemType)
		}
	}

	values := make([]attr.Value, len(*exts))
	for i, ext := range *exts {
		value := types.StringValue(base64.StdEncoding.EncodeToString(ext.Value))
		critical := types.BoolPointerValue(ext.Critical)

		// Prefer equivalent prior element values to avoid "inconsistent
		// result after apply" errors, e.g. a value encoded with line breaks
		// or an omitted critical returned by the API as false.
		if i < len(priors) {
			prior := priors[i]
			if utils.IsBase64Equal(value.ValueString(), prior.Value.ValueString()) {
				value = prior.Value
			}
			remoteCriticalFalse := ext.Critical == nil || !*ext.Critical
			priorCriticalFalse := prior.Critical.IsNull() || !prior.Critical.ValueBool()
			if remoteCriticalFalse && priorCriticalFalse {
				critical = prior.Critical
			}
		}

		obj, d := basetypes.NewObjectValue(customExtensionAttributes, map[string]attr.Value{
			"oid":      types.StringValue(ext.Oid),
			"critical": critical,
			"value":    value,
		})
		diags.Append(d...)
		values[i] = obj
	}

	list, d := types.ListValue(elemType, values)
	diags.Append(d...)

	return list
}

func keyObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, key v20250101.CredentialKey, state utils.AttributeGetter) types.Object {
	pubFile, ds := utils.ToOptionalString(ctx, key.PubFile, state, path.Root("key").AtName("pub_file"))
	diags.Append(ds...)

	typ, ds := utils.ToOptionalString(ctx, key.Type, state, path.Root("key").AtName("type"))
	diags.Append(ds...)

	protection, ds := utils.ToOptionalString(ctx, key.Protection, state, path.Root("key").AtName("protection"))
	diags.Append(ds...)

	out, ds := basetypes.NewObjectValue(keyAttributes, map[string]attr.Value{
		"pub_file":   pubFile,
		"type":       typ,
		"protection": protection,
	})
	diags.Append(ds...)

	return out
}

func isAttested(keyType types.String) bool {
	return keyType.ValueString() == "HARDWARE_ATTESTED"
}
