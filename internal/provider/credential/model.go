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
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_credential"

type CredentialModel struct {
	ID             types.String `tfsdk:"id"`
	Slug           types.String `tfsdk:"slug"`
	ManagementMode types.String `tfsdk:"management_mode"`
	Certificate    types.Object `tfsdk:"certificate"`
	Key            types.Object `tfsdk:"key"`
	Policy         types.Object `tfsdk:"policy"`
	Files          types.Object `tfsdk:"files"`
}

type CertificateModel struct {
	AuthorityID types.String `tfsdk:"authority_id"`
	Duration    types.String `tfsdk:"duration"`
	X509        types.Object `tfsdk:"x509"`
	NamePolicy  types.Object `tfsdk:"name_policy"`
}

var certificateAttributes = map[string]attr.Type{
	"authority_id": types.StringType,
	"duration":     types.StringType,
	"x509":         types.ObjectType{AttrTypes: x509Attributes},
	"name_policy":  types.ObjectType{AttrTypes: namePolicyAttributes},
}

type X509Model struct {
	CommonName         types.Object `tfsdk:"common_name"`
	SANs               types.Object `tfsdk:"sans"`
	Organization       types.Object `tfsdk:"organization"`
	OrganizationalUnit types.Object `tfsdk:"organizational_unit"`
	Locality           types.Object `tfsdk:"locality"`
	Province           types.Object `tfsdk:"province"`
	StreetAddress      types.Object `tfsdk:"street_address"`
	PostalCode         types.Object `tfsdk:"postal_code"`
	Country            types.Object `tfsdk:"country"`
	GivenName          types.Object `tfsdk:"given_name"`
	SerialNumber       types.Object `tfsdk:"serial_number"`
	Surname            types.Object `tfsdk:"surname"`
	TypedSans          types.Object `tfsdk:"typed_sans"`
	ExtendedKeyUsage   types.List   `tfsdk:"extended_key_usage"`
	CustomExtensions   types.List   `tfsdk:"custom_extensions"`
}

var x509Attributes = map[string]attr.Type{
	"common_name":         types.ObjectType{AttrTypes: certificateFieldAttributes},
	"sans":                types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"organization":        types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"organizational_unit": types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"locality":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"province":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"street_address":      types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"postal_code":         types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"country":             types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"given_name":          types.ObjectType{AttrTypes: certificateFieldAttributes},
	"serial_number":       types.ObjectType{AttrTypes: certificateFieldAttributes},
	"surname":             types.ObjectType{AttrTypes: certificateFieldAttributes},
	"typed_sans":          types.ObjectType{AttrTypes: typedSansAttributes},
	"extended_key_usage":  types.ListType{ElemType: types.StringType},
	"custom_extensions":   types.ListType{ElemType: types.ObjectType{AttrTypes: customExtensionAttributes}},
}

type TypedSANsModel struct {
	DnsNames           types.Object `tfsdk:"dns_names"`
	IpAddresses        types.Object `tfsdk:"ip_addresses"`
	EmailAddresses     types.Object `tfsdk:"email_addresses"`
	Uris               types.Object `tfsdk:"uris"`
	UserPrincipalNames types.Object `tfsdk:"user_principal_names"`
}

// isEmpty reports whether every child CertificateFieldList is either null or
// itself empty (e.g. `dns_names = { static = [] }`), matching the cases where
// the API drops the entire typed_sans struct.
func (m *TypedSANsModel) isEmpty(ctx context.Context) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	for _, obj := range []types.Object{m.DnsNames, m.IpAddresses, m.EmailAddresses, m.Uris, m.UserPrincipalNames} {
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		child := &CertificateFieldListModel{}
		diags.Append(obj.As(ctx, child, basetypes.ObjectAsOptions{})...)
		if !child.isEmpty() {
			return false, diags
		}
	}

	return true, diags
}

var typedSansAttributes = map[string]attr.Type{
	"dns_names":            types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"ip_addresses":         types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"email_addresses":      types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"uris":                 types.ObjectType{AttrTypes: certificateFieldListAttributes},
	"user_principal_names": types.ObjectType{AttrTypes: certificateFieldListAttributes},
}

type CustomExtensionModel struct {
	Oid      types.String `tfsdk:"oid"`
	Critical types.Bool   `tfsdk:"critical"`
	Value    types.String `tfsdk:"value"`
}

var customExtensionAttributes = map[string]attr.Type{
	"oid":      types.StringType,
	"critical": types.BoolType,
	"value":    types.StringType,
}

type X509NamesModel struct {
	CommonNames types.List `tfsdk:"common_names"`
	Dns         types.List `tfsdk:"dns"`
	Emails      types.List `tfsdk:"emails"`
	Ips         types.List `tfsdk:"ips"`
	Uris        types.List `tfsdk:"uris"`
}

func (m *X509NamesModel) isEmpty() bool {
	switch {
	case len(m.CommonNames.Elements()) > 0:
		return false
	case len(m.Dns.Elements()) > 0:
		return false
	case len(m.Emails.Elements()) > 0:
		return false
	case len(m.Ips.Elements()) > 0:
		return false
	case len(m.Uris.Elements()) > 0:
		return false
	}
	return true
}

var x509NamesAttributes = map[string]attr.Type{
	"common_names": types.ListType{ElemType: types.StringType},
	"dns":          types.ListType{ElemType: types.StringType},
	"emails":       types.ListType{ElemType: types.StringType},
	"ips":          types.ListType{ElemType: types.StringType},
	"uris":         types.ListType{ElemType: types.StringType},
}

type NamePolicyModel struct {
	Allow              types.Object `tfsdk:"allow"`
	Deny               types.Object `tfsdk:"deny"`
	AllowWildcardNames types.Bool   `tfsdk:"allow_wildcard_names"`
}

func (m *NamePolicyModel) isEmpty() bool {
	switch {
	case !m.Allow.IsNull():
		return false
	case !m.Deny.IsNull():
		return false
	case m.AllowWildcardNames.ValueBool():
		return false
	}
	return true
}

var namePolicyAttributes = map[string]attr.Type{
	"allow":                types.ObjectType{AttrTypes: x509NamesAttributes},
	"deny":                 types.ObjectType{AttrTypes: x509NamesAttributes},
	"allow_wildcard_names": types.BoolType,
}

type KeyModel struct {
	Type          types.String `tfsdk:"type"`
	Protection    types.String `tfsdk:"protection"`
	PubFile       types.String `tfsdk:"pub_file"`
	Compatibility types.String `tfsdk:"compatibility"`
	Store         types.String `tfsdk:"store"`
}

var keyAttributes = map[string]attr.Type{
	"type":          types.StringType,
	"protection":    types.StringType,
	"pub_file":      types.StringType,
	"compatibility": types.StringType,
	"store":         types.StringType,
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

func (c *CertificateFieldModel) isEmpty() bool {
	switch {
	case c.Static.ValueString() != "":
		return false
	case c.DeviceMetadata.ValueString() != "":
		return false
	}
	return true
}

var certificateFieldAttributes = map[string]attr.Type{
	"static":          types.StringType,
	"device_metadata": types.StringType,
}

type CertificateFieldListModel struct {
	Static                   types.List `tfsdk:"static"`
	DeviceMetadata           types.List `tfsdk:"device_metadata"`
	InsecureIncludeRequested types.Bool `tfsdk:"insecure_include_requested"`
}

func (c *CertificateFieldListModel) isEmpty() bool {
	switch {
	case len(c.Static.Elements()) > 0:
		return false
	case len(c.DeviceMetadata.Elements()) > 0:
		return false
	case c.InsecureIncludeRequested.ValueBool():
		return false
	}
	return true
}

var certificateFieldListAttributes = map[string]attr.Type{
	"static":                     types.ListType{ElemType: types.StringType},
	"device_metadata":            types.ListType{ElemType: types.StringType},
	"insecure_include_requested": types.BoolType,
}

func (k *KeyModel) toAPI() v20260501.CredentialKey {
	return v20260501.CredentialKey{
		Type:          (*v20260501.CredentialKeyType)(k.Type.ValueStringPointer()),
		Protection:    (*v20260501.CredentialKeyProtection)(k.Protection.ValueStringPointer()),
		PubFile:       k.PubFile.ValueStringPointer(),
		Compatibility: (*v20260501.CredentialKeyCompatibility)(k.Compatibility.ValueStringPointer()),
		Store:         (*v20260501.CredentialKeyStore)(k.Store.ValueStringPointer()),
	}
}

func (m *CertificateModel) toAPI(ctx context.Context, diags *diag.Diagnostics) v20260501.CredentialCertificate {
	cert := v20260501.CredentialCertificate{
		Type:        v20260501.CredentialCertificateTypeX509,
		AuthorityID: m.AuthorityID.ValueString(),
	}

	cert.Duration = m.Duration.ValueStringPointer()
	cert.NamePolicy = asNamePolicy(ctx, diags, m.NamePolicy)

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

func asX509Names(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20260501.X509Names {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &X509NamesModel{}
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)

	names := &v20260501.X509Names{}
	diags.Append(model.CommonNames.ElementsAs(ctx, &names.CommonNames, false)...)
	diags.Append(model.Dns.ElementsAs(ctx, &names.Dns, false)...)
	diags.Append(model.Emails.ElementsAs(ctx, &names.Emails, false)...)
	diags.Append(model.Ips.ElementsAs(ctx, &names.Ips, false)...)
	diags.Append(model.Uris.ElementsAs(ctx, &names.Uris, false)...)

	return names
}

func asNamePolicy(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20260501.X509NamePolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &NamePolicyModel{}
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)

	return &v20260501.X509NamePolicy{
		Allow:              asX509Names(ctx, diags, model.Allow),
		Deny:               asX509Names(ctx, diags, model.Deny),
		AllowWildcardNames: model.AllowWildcardNames.ValueBoolPointer(),
	}
}

func asTypedSans(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20260501.X509TypedSANs {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &TypedSANsModel{}
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)

	return &v20260501.X509TypedSANs{
		DnsNames:           asCertificateFieldList(ctx, diags, model.DnsNames),
		IpAddresses:        asCertificateFieldList(ctx, diags, model.IpAddresses),
		EmailAddresses:     asCertificateFieldList(ctx, diags, model.EmailAddresses),
		Uris:               asCertificateFieldList(ctx, diags, model.Uris),
		UserPrincipalNames: asCertificateFieldList(ctx, diags, model.UserPrincipalNames),
	}
}

func asExtendedKeyUsage(ctx context.Context, diags *diag.Diagnostics, list types.List) *[]v20260501.X509ExtendedKeyUsage {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var usage []v20260501.X509ExtendedKeyUsage
	diags.Append(list.ElementsAs(ctx, &usage, false)...)

	return &usage
}

func asCustomExtensions(ctx context.Context, diags *diag.Diagnostics, list types.List) *[]v20260501.X509CustomExtension {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var models []CustomExtensionModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)

	extensions := make([]v20260501.X509CustomExtension, len(models))
	for i, m := range models {
		value, err := base64.StdEncoding.DecodeString(m.Value.ValueString())
		if err != nil {
			diags.AddError("Decode Custom Extension Value", fmt.Sprintf("custom_extensions[%d].value must be base64-encoded: %s", i, err.Error()))
			continue
		}

		extensions[i] = v20260501.X509CustomExtension{
			Oid:      m.Oid.ValueString(),
			Critical: m.Critical.ValueBoolPointer(),
			Value:    value,
		}
	}

	return &extensions
}

func (m *FilesModel) toAPI() *v20260501.CredentialFiles {
	if m == nil {
		return nil
	}

	return &v20260501.CredentialFiles{
		RootFile:  m.RootFile.ValueStringPointer(),
		CrtFile:   m.CrtFile.ValueStringPointer(),
		KeyFile:   m.KeyFile.ValueStringPointer(),
		KeyFormat: (*v20260501.CredentialFilesKeyFormat)(m.KeyFormat.ValueStringPointer()),
		Uid:       utils.ToIntPointer(m.UID.ValueInt64Pointer()),
		Gid:       utils.ToIntPointer(m.GID.ValueInt64Pointer()),
		Mode:      utils.ToIntPointer(m.Mode.ValueInt64Pointer()),
	}
}

func (x509 *X509Model) toAPI(ctx context.Context, diags *diag.Diagnostics) v20260501.X509Fields {
	return v20260501.X509Fields{
		CommonName:         asCertificateField(ctx, diags, x509.CommonName),
		Sans:               asCertificateFieldList(ctx, diags, x509.SANs),
		Country:            asCertificateFieldList(ctx, diags, x509.Country),
		Locality:           asCertificateFieldList(ctx, diags, x509.Locality),
		Organization:       asCertificateFieldList(ctx, diags, x509.Organization),
		OrganizationalUnit: asCertificateFieldList(ctx, diags, x509.OrganizationalUnit),
		PostalCode:         asCertificateFieldList(ctx, diags, x509.PostalCode),
		Province:           asCertificateFieldList(ctx, diags, x509.Province),
		StreetAddress:      asCertificateFieldList(ctx, diags, x509.StreetAddress),
		GivenName:          asCertificateField(ctx, diags, x509.GivenName),
		SerialNumber:       asCertificateField(ctx, diags, x509.SerialNumber),
		Surname:            asCertificateField(ctx, diags, x509.Surname),
		TypedSans:          asTypedSans(ctx, diags, x509.TypedSans),
		ExtendedKeyUsage:   asExtendedKeyUsage(ctx, diags, x509.ExtendedKeyUsage),
		CustomExtensions:   asCustomExtensions(ctx, diags, x509.CustomExtensions),
	}
}

func (p *PolicyModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *v20260501.PolicyMatchCriteria {
	if p == nil {
		return nil
	}

	policy := &v20260501.PolicyMatchCriteria{}

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

func (cf *CertificateFieldModel) toAPI() *v20260501.CertificateField {
	return &v20260501.CertificateField{
		Static:         cf.Static.ValueStringPointer(),
		DeviceMetadata: cf.DeviceMetadata.ValueStringPointer(),
	}
}

func (cfl *CertificateFieldListModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *v20260501.CertificateFieldList {
	var static *[]string
	var deviceMetadata *[]string

	diags.Append(cfl.Static.ElementsAs(ctx, &static, false)...)
	diags.Append(cfl.DeviceMetadata.ElementsAs(ctx, &deviceMetadata, false)...)

	return &v20260501.CertificateFieldList{
		Static:                   static,
		DeviceMetadata:           deviceMetadata,
		InsecureIncludeRequested: cfl.InsecureIncludeRequested.ValueBoolPointer(),
	}
}

func asCertificateFieldList(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20260501.CertificateFieldList {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &CertificateFieldListModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return model.toAPI(ctx, diags)
}

func asCertificateField(ctx context.Context, diags *diag.Diagnostics, obj types.Object) *v20260501.CertificateField {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	model := &CertificateFieldModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return model.toAPI()
}

func toAPI(ctx context.Context, diags *diag.Diagnostics, model *CredentialModel) v20260501.Credential {
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

	var managementMode *v20260501.EndpointManagementMode
	if !model.ManagementMode.IsNull() && !model.ManagementMode.IsUnknown() {
		managementMode = (*v20260501.EndpointManagementMode)(model.ManagementMode.ValueStringPointer())
	}

	return v20260501.Credential{
		Id:             model.ID.ValueStringPointer(),
		Slug:           model.Slug.ValueString(),
		ManagementMode: managementMode,
		Certificate:    cert.toAPI(ctx, diags),
		Key:            key.toAPI(),
		Policy:         policy.toAPI(ctx, diags),
		Files:          files.toAPI(),
	}
}

func fromAPI(ctx context.Context, diags *diag.Diagnostics, credential *v20260501.Credential, state utils.AttributeGetter) CredentialModel {
	managementMode, d := utils.ToOptionalString(ctx, credential.ManagementMode, state, path.Root("management_mode"))
	diags.Append(d...)

	return CredentialModel{
		ID:             types.StringPointerValue(credential.Id),
		Slug:           types.StringValue(credential.Slug),
		ManagementMode: managementMode,
		Certificate:    certificateObjectFromAPI(ctx, diags, credential.Certificate, state),
		Key:            keyObjectFromAPI(ctx, diags, credential.Key, state),
		Policy:         policyObjectFromAPI(ctx, diags, credential.Policy, state),
		Files:          filesObjectFromAPI(ctx, diags, credential.Files, state),
	}
}

func certificateObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cert v20260501.CredentialCertificate, state utils.AttributeGetter) types.Object {
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
		"name_policy":  namePolicyObjectFromAPI(ctx, diags, cert.NamePolicy, state),
	})
	diags.Append(d...)

	return out
}

func namePolicyObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, policy *v20260501.X509NamePolicy, state utils.AttributeGetter) types.Object {
	p := path.Root("certificate").AtName("name_policy")

	if policy == nil || reflect.DeepEqual(policy, new(v20260501.X509NamePolicy)) {
		// See comment in policyObjectFromAPI: users can set a non-null empty
		// name_policy in config, such as `name_policy = {}`, but the API
		// returns nil for all of these.
		obj := &NamePolicyModel{}
		d := state.GetAttribute(ctx, p, &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(namePolicyAttributes)
		}

		if obj.isEmpty() {
			o, d := basetypes.NewObjectValue(namePolicyAttributes, map[string]attr.Value{
				"allow":                obj.Allow,
				"deny":                 obj.Deny,
				"allow_wildcard_names": obj.AllowWildcardNames,
			})
			diags.Append(d...)
			return o
		}

		return basetypes.NewObjectNull(namePolicyAttributes)
	}

	allowWildcard, d := utils.ToOptionalBool(ctx, policy.AllowWildcardNames, state, p.AtName("allow_wildcard_names"))
	diags.Append(d...)

	out, d := basetypes.NewObjectValue(namePolicyAttributes, map[string]attr.Value{
		"allow":                x509NamesObjectFromAPI(ctx, diags, policy.Allow, state, p.AtName("allow")),
		"deny":                 x509NamesObjectFromAPI(ctx, diags, policy.Deny, state, p.AtName("deny")),
		"allow_wildcard_names": allowWildcard,
	})
	diags.Append(d...)

	return out
}

func x509NamesObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, names *v20260501.X509Names, state utils.AttributeGetter, p path.Path) types.Object {
	if names == nil {
		obj := &X509NamesModel{}
		d := state.GetAttribute(ctx, p, &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(x509NamesAttributes)
		}

		if obj.isEmpty() {
			out, d := basetypes.NewObjectValue(x509NamesAttributes, map[string]attr.Value{
				"common_names": obj.CommonNames,
				"dns":          obj.Dns,
				"emails":       obj.Emails,
				"ips":          obj.Ips,
				"uris":         obj.Uris,
			})
			diags.Append(d...)
			return out
		}

		return basetypes.NewObjectNull(x509NamesAttributes)
	}

	commonNames, d := utils.ToOptionalList(ctx, names.CommonNames, state, p.AtName("common_names"))
	diags.Append(d...)

	dns, d := utils.ToOptionalList(ctx, names.Dns, state, p.AtName("dns"))
	diags.Append(d...)

	emails, d := utils.ToOptionalList(ctx, names.Emails, state, p.AtName("emails"))
	diags.Append(d...)

	ips, d := utils.ToOptionalList(ctx, names.Ips, state, p.AtName("ips"))
	diags.Append(d...)

	uris, d := utils.ToOptionalList(ctx, names.Uris, state, p.AtName("uris"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(x509NamesAttributes, map[string]attr.Value{
		"common_names": commonNames,
		"dns":          dns,
		"emails":       emails,
		"ips":          ips,
		"uris":         uris,
	})
	diags.Append(d...)

	return obj
}

func typedSansObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, sans *v20260501.X509TypedSANs, state utils.AttributeGetter, p path.Path) types.Object {
	if sans == nil {
		obj := &TypedSANsModel{}
		d := state.GetAttribute(ctx, p, &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(typedSansAttributes)
		}

		empty, d := obj.isEmpty(ctx)
		diags.Append(d...)

		if empty {
			out, d := basetypes.NewObjectValue(typedSansAttributes, map[string]attr.Value{
				"dns_names":            obj.DnsNames,
				"ip_addresses":         obj.IpAddresses,
				"email_addresses":      obj.EmailAddresses,
				"uris":                 obj.Uris,
				"user_principal_names": obj.UserPrincipalNames,
			})
			diags.Append(d...)
			return out
		}

		return basetypes.NewObjectNull(typedSansAttributes)
	}

	obj, d := basetypes.NewObjectValue(typedSansAttributes, map[string]attr.Value{
		"dns_names":            certificateFieldListObjectFromAPI(ctx, diags, sans.DnsNames, state, p.AtName("dns_names")),
		"ip_addresses":         certificateFieldListObjectFromAPI(ctx, diags, sans.IpAddresses, state, p.AtName("ip_addresses")),
		"email_addresses":      certificateFieldListObjectFromAPI(ctx, diags, sans.EmailAddresses, state, p.AtName("email_addresses")),
		"uris":                 certificateFieldListObjectFromAPI(ctx, diags, sans.Uris, state, p.AtName("uris")),
		"user_principal_names": certificateFieldListObjectFromAPI(ctx, diags, sans.UserPrincipalNames, state, p.AtName("user_principal_names")),
	})
	diags.Append(d...)

	return obj
}

func customExtensionsListFromAPI(ctx context.Context, diags *diag.Diagnostics, extensions *[]v20260501.X509CustomExtension) types.List {
	listType := types.ObjectType{AttrTypes: customExtensionAttributes}

	if extensions == nil || len(*extensions) == 0 {
		return types.ListNull(listType)
	}

	models := make([]CustomExtensionModel, len(*extensions))
	for i, ext := range *extensions {
		models[i] = CustomExtensionModel{
			Oid: types.StringValue(ext.Oid),
			// The API omits `critical` from its response both when it was never
			// set and when it was explicitly set to false, so nil and false are
			// indistinguishable on read. Both cases have the same correct value.
			Critical: types.BoolValue(ext.Critical != nil && *ext.Critical),
			Value:    types.StringValue(base64.StdEncoding.EncodeToString(ext.Value)),
		}
	}

	list, d := types.ListValueFrom(ctx, listType, models)
	diags.Append(d...)

	return list
}

func filesObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, files *v20260501.CredentialFiles, state utils.AttributeGetter) types.Object {
	p := path.Root("files")

	if files == nil || reflect.DeepEqual(files, new(v20260501.CredentialFiles)) {
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

func policyObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, policy *v20260501.PolicyMatchCriteria, state utils.AttributeGetter) types.Object {
	if policy == nil || reflect.DeepEqual(policy, new(v20260501.PolicyMatchCriteria)) {
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

func certificateFieldObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cf *v20260501.CertificateField, state utils.AttributeGetter, p path.Path) types.Object {
	if cf == nil {
		obj := &CertificateFieldModel{}
		d := state.GetAttribute(ctx, p, &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(certificateFieldAttributes)
		}

		if obj.isEmpty() {
			out, d := basetypes.NewObjectValue(certificateFieldAttributes, map[string]attr.Value{
				"static":          obj.Static,
				"device_metadata": obj.DeviceMetadata,
			})
			diags.Append(d...)
			return out
		}

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

func certificateFieldListObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, cfl *v20260501.CertificateFieldList, state utils.AttributeGetter, p path.Path) types.Object {
	if cfl == nil {
		// The API drops empty CertificateFieldList objects instead of returning them.
		// If Terraform config contains a non-null empty object (e.g. `dns_names = { static = [] }`),
		// returning null causes an "inconsistent result after apply" error. Preserve the
		// applied empty object in that case.

		obj := &CertificateFieldListModel{}
		d := state.GetAttribute(ctx, p, &obj)
		diags.Append(d...)

		if obj == nil {
			return basetypes.NewObjectNull(certificateFieldListAttributes)
		}

		if obj.isEmpty() {
			out, d := basetypes.NewObjectValue(certificateFieldListAttributes, map[string]attr.Value{
				"static":                     obj.Static,
				"device_metadata":            obj.DeviceMetadata,
				"insecure_include_requested": obj.InsecureIncludeRequested,
			})
			diags.Append(d...)
			return out
		}

		return basetypes.NewObjectNull(certificateFieldListAttributes)
	}

	static, d := utils.ToOptionalList(ctx, cfl.Static, state, p.AtName("static"))
	diags.Append(d...)

	deviceMetadata, d := utils.ToOptionalList(ctx, cfl.DeviceMetadata, state, p.AtName("device_metadata"))
	diags.Append(d...)

	insecureIncludeRequested, d := utils.ToOptionalBool(ctx, cfl.InsecureIncludeRequested, state, p.AtName("insecure_include_requested"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(certificateFieldListAttributes, map[string]attr.Value{
		"static":                     static,
		"device_metadata":            deviceMetadata,
		"insecure_include_requested": insecureIncludeRequested,
	})
	diags.Append(d...)

	return obj
}

func x509ObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, x509 v20260501.X509Fields, state utils.AttributeGetter) types.Object {
	p := path.Root("certificate").AtName("x509")

	extendedKeyUsage, d := utils.ToOptionalList(ctx, x509.ExtendedKeyUsage, state, p.AtName("extended_key_usage"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(x509Attributes, map[string]attr.Value{
		"common_name":         certificateFieldObjectFromAPI(ctx, diags, x509.CommonName, state, p.AtName("common_name")),
		"sans":                certificateFieldListObjectFromAPI(ctx, diags, x509.Sans, state, p.AtName("sans")),
		"organization":        certificateFieldListObjectFromAPI(ctx, diags, x509.Organization, state, p.AtName("organization")),
		"organizational_unit": certificateFieldListObjectFromAPI(ctx, diags, x509.OrganizationalUnit, state, p.AtName("organizational_unit")),
		"locality":            certificateFieldListObjectFromAPI(ctx, diags, x509.Locality, state, p.AtName("locality")),
		"province":            certificateFieldListObjectFromAPI(ctx, diags, x509.Province, state, p.AtName("province")),
		"street_address":      certificateFieldListObjectFromAPI(ctx, diags, x509.StreetAddress, state, p.AtName("street_address")),
		"postal_code":         certificateFieldListObjectFromAPI(ctx, diags, x509.PostalCode, state, p.AtName("postal_code")),
		"country":             certificateFieldListObjectFromAPI(ctx, diags, x509.Country, state, p.AtName("country")),
		"given_name":          certificateFieldObjectFromAPI(ctx, diags, x509.GivenName, state, p.AtName("given_name")),
		"serial_number":       certificateFieldObjectFromAPI(ctx, diags, x509.SerialNumber, state, p.AtName("serial_number")),
		"surname":             certificateFieldObjectFromAPI(ctx, diags, x509.Surname, state, p.AtName("surname")),
		"typed_sans":          typedSansObjectFromAPI(ctx, diags, x509.TypedSans, state, p.AtName("typed_sans")),
		"extended_key_usage":  extendedKeyUsage,
		"custom_extensions":   customExtensionsListFromAPI(ctx, diags, x509.CustomExtensions),
	})
	diags.Append(d...)

	return obj
}

func keyObjectFromAPI(ctx context.Context, diags *diag.Diagnostics, key v20260501.CredentialKey, state utils.AttributeGetter) types.Object {
	pubFile, ds := utils.ToOptionalString(ctx, key.PubFile, state, path.Root("key").AtName("pub_file"))
	diags.Append(ds...)

	typ, ds := utils.ToOptionalString(ctx, key.Type, state, path.Root("key").AtName("type"))
	diags.Append(ds...)

	protection, ds := utils.ToOptionalString(ctx, key.Protection, state, path.Root("key").AtName("protection"))
	diags.Append(ds...)

	compatibility, ds := utils.ToOptionalStringWithDefault(ctx, key.Compatibility, v20260501.CredentialKeyCompatibilityDEFAULT, state, path.Root("key").AtName("compatibility"))
	diags.Append(ds...)

	store, ds := utils.ToOptionalStringWithDefault(ctx, key.Store, v20260501.CredentialKeyStoreDEFAULT, state, path.Root("key").AtName("store"))
	diags.Append(ds...)

	out, ds := basetypes.NewObjectValue(keyAttributes, map[string]attr.Value{
		"pub_file":      pubFile,
		"type":          typ,
		"protection":    protection,
		"compatibility": compatibility,
		"store":         store,
	})
	diags.Append(ds...)

	return out
}

func isAttested(keyType types.String) bool {
	return keyType.ValueString() == "HARDWARE_ATTESTED"
}
