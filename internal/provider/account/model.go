package account

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

const name = "smallstep_account"

type AccountModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Certificate types.Object `tfsdk:"certificate"`
	Key         types.Object `tfsdk:"key"`
	Reload      types.Object `tfsdk:"reload"`
	Policy      types.Object `tfsdk:"policy"`
	Browser     types.Object `tfsdk:"browser"`
	Ethernet    types.Object `tfsdk:"ethernet"`
	VPN         types.Object `tfsdk:"vpn"`
	WiFi        types.Object `tfsdk:"wifi"`
}

type CertificateModel struct {
	CrtFile     types.String `tfsdk:"crt_file"`
	KeyFile     types.String `tfsdk:"key_file"`
	RootFile    types.String `tfsdk:"root_file"`
	Duration    types.String `tfsdk:"duration"`
	GID         types.Int64  `tfsdk:"gid"`
	UID         types.Int64  `tfsdk:"uid"`
	Mode        types.Int64  `tfsdk:"mode"`
	AuthorityID types.String `tfsdk:"authority_id"`
	X509        types.Object `tfsdk:"x509"`
	SSH         types.Object `tfsdk:"ssh"`
}

var certificateAttributes = map[string]attr.Type{
	"crt_file":     types.StringType,
	"key_file":     types.StringType,
	"root_file":    types.StringType,
	"duration":     types.StringType,
	"gid":          types.Int64Type,
	"uid":          types.Int64Type,
	"mode":         types.Int64Type,
	"authority_id": types.StringType,
	"x509":         types.ObjectType{AttrTypes: x509Attributes},
	"ssh":          types.ObjectType{AttrTypes: sshAttributes},
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
}

type SSHModel struct {
	KeyID      types.Object `tfsdk:"key_id"`
	Principals types.Object `tfsdk:"principals"`
}

var sshAttributes = map[string]attr.Type{
	"key_id":     types.ObjectType{AttrTypes: certificateFieldAttributes},
	"principals": types.ObjectType{AttrTypes: certificateFieldListAttributes},
}

type KeyModel struct {
	Format     types.String `tfsdk:"format"`
	PubFile    types.String `tfsdk:"pub_file"`
	Type       types.String `tfsdk:"type"`
	Protection types.String `tfsdk:"protection"`
}

var keyAttributes = map[string]attr.Type{
	"format":     types.StringType,
	"pub_file":   types.StringType,
	"type":       types.StringType,
	"protection": types.StringType,
}

type ReloadModel struct {
	Method   types.String `tfsdk:"method"`
	PIDFile  types.String `tfsdk:"pid_file"`
	Signal   types.Int64  `tfsdk:"signal"`
	UnitName types.String `tfsdk:"unit_name"`
}

var reloadAttributes = map[string]attr.Type{
	"method":    types.StringType,
	"pid_file":  types.StringType,
	"signal":    types.Int64Type,
	"unit_name": types.StringType,
}

type PolicyModel struct {
	Assurance types.List `tfsdk:"assurance"`
	OS        types.List `tfsdk:"os"`
	Ownership types.List `tfsdk:"ownership"`
	Source    types.List `tfsdk:"source"`
	Tags      types.List `tfsdk:"tags"`
}

var policyAttributes = map[string]attr.Type{
	"assurance": types.ListType{ElemType: types.StringType},
	"os":        types.ListType{ElemType: types.StringType},
	"ownership": types.ListType{ElemType: types.StringType},
	"source":    types.ListType{ElemType: types.StringType},
	"tags":      types.ListType{ElemType: types.StringType},
}

type BrowserModel struct{}

var browserAttributes = map[string]attr.Type{}

type EthernetModel struct {
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	CAChain               types.String `tfsdk:"ca_chain"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
}

var ethernetAttributes = map[string]attr.Type{
	"autojoin":                 types.BoolType,
	"ca_chain":                 types.StringType,
	"external_radius_server":   types.BoolType,
	"network_access_server_ip": types.StringType,
}

type VPNModel struct {
	Autojoin       types.Bool   `tfsdk:"autojoin"`
	ConnectionType types.String `tfsdk:"connection_type"`
	IKE            types.Object `tfsdk:"ike"`
	RemoteAddress  types.String `tfsdk:"remote_address"`
	Vendor         types.String `tfsdk:"vendor"`
}

var vpnAttributes = map[string]attr.Type{
	"autojoin":        types.BoolType,
	"connection_type": types.StringType,
	"remote_address":  types.StringType,
	"vendor":          types.StringType,
	"ike":             types.ObjectType{AttrTypes: ikeAttributes},
}

type IKEModel struct {
	CAChain  types.String `tfsdk:"ca_chain"`
	EAP      types.Bool   `tfsdk:"eap"`
	RemoteID types.String `tfsdk:"remote_id"`
}

var ikeAttributes = map[string]attr.Type{
	"ca_chain":  types.StringType,
	"eap":       types.BoolType,
	"remote_id": types.StringType,
}

type WiFiModel struct {
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	CAChain               types.String `tfsdk:"ca_chain"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	Hidden                types.Bool   `tfsdk:"hidden"`
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
	SSID                  types.String `tfsdk:"ssid"`
}

var wifiAttributes = map[string]attr.Type{
	"autojoin":                 types.BoolType,
	"ca_chain":                 types.StringType,
	"external_radius_server":   types.BoolType,
	"hidden":                   types.BoolType,
	"network_access_server_ip": types.StringType,
	"ssid":                     types.StringType,
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

func (k *KeyModel) toAPI() *v20250101.EndpointKeyInfo {
	if k == nil {
		return nil
	}

	return &v20250101.EndpointKeyInfo{
		Format:     utils.ToStringPointer[v20250101.EndpointKeyInfoFormat](k.Format.ValueStringPointer()),
		PubFile:    k.PubFile.ValueStringPointer(),
		Type:       utils.ToStringPointer[v20250101.EndpointKeyInfoType](k.Type.ValueStringPointer()),
		Protection: utils.ToStringPointer[v20250101.EndpointKeyInfoProtection](k.Protection.ValueStringPointer()),
	}
}

func (r *ReloadModel) toAPI() *v20250101.EndpointReloadInfo {
	if r == nil {
		return nil
	}

	return &v20250101.EndpointReloadInfo{
		Method:   v20250101.EndpointReloadInfoMethod(r.Method.ValueString()),
		PidFile:  r.PIDFile.ValueStringPointer(),
		Signal:   utils.ToIntPointer(r.Signal.ValueInt64Pointer()),
		UnitName: r.UnitName.ValueStringPointer(),
	}
}

func (c *CertificateModel) toAPI(ctx context.Context) (*v20250101.EndpointCertificateInfo, diag.Diagnostics) {
	var diags diag.Diagnostics

	if c == nil {
		return nil, diags
	}

	eci := &v20250101.EndpointCertificateInfo{
		CrtFile:     c.CrtFile.ValueStringPointer(),
		KeyFile:     c.KeyFile.ValueStringPointer(),
		RootFile:    c.RootFile.ValueStringPointer(),
		Uid:         utils.ToIntPointer(c.UID.ValueInt64Pointer()),
		Gid:         utils.ToIntPointer(c.GID.ValueInt64Pointer()),
		Mode:        utils.ToIntPointer(c.Mode.ValueInt64Pointer()),
		Type:        "X509",
		AuthorityID: c.AuthorityID.ValueStringPointer(),
	}

	// duration defaults to 24h if not set, which means the schema must
	// mark it as both computed and optional. With the computed flag it
	// will be unknown rather than null if not set. The ValueStringPointer()
	// method returns the empty string for unknown values, but passing an
	// empty string to the API results in a 400 since it's not a valid
	// duration.
	if !c.Duration.IsUnknown() {
		eci.Duration = c.Duration.ValueStringPointer()
	}

	if !c.X509.IsNull() && !c.X509.IsUnknown() {
		x509 := &X509Model{}
		ds := c.X509.As(ctx, &x509, basetypes.ObjectAsOptions{})
		diags.Append(ds...)
		eci.Type = "X509"
		x, ds := x509.toAPI(ctx)
		diags.Append(ds...)
		eci.Details = &v20250101.EndpointCertificateInfo_Details{}
		if err := eci.Details.FromX509Fields(*x); err != nil {
			diags.AddError("Format X509 Attributes", err.Error())
		}
	}

	if !c.SSH.IsNull() && !c.SSH.IsUnknown() {
		ssh := &SSHModel{}
		ds := c.SSH.As(ctx, &ssh, basetypes.ObjectAsOptions{})
		diags.Append(ds...)
		eci.Type = "SSH_USER"
		s, ds := ssh.toAPI(ctx)
		diags.Append(ds...)
		eci.Details = &v20250101.EndpointCertificateInfo_Details{}
		if err := eci.Details.FromSshFields(*s); err != nil {
			diags.AddError("Format SSH Attributes", err.Error())
		}
	}

	return eci, diags
}

func (ssh SSHModel) toAPI(ctx context.Context) (*v20250101.SshFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	keyID, ds := asCertificateField(ctx, ssh.KeyID)
	diags.Append(ds...)

	principals, ds := asCertificateFieldList(ctx, ssh.Principals)
	diags.Append(ds...)

	return &v20250101.SshFields{
		KeyId:      keyID,
		Principals: principals,
	}, diags
}

func (x509 X509Model) toAPI(ctx context.Context) (*v20250101.X509Fields, diag.Diagnostics) {
	var diags diag.Diagnostics

	cn, ds := asCertificateField(ctx, x509.CommonName)
	diags.Append(ds...)

	sans, ds := asCertificateFieldList(ctx, x509.SANs)
	diags.Append(ds...)

	org, ds := asCertificateFieldList(ctx, x509.Organization)
	diags.Append(ds...)

	ou, ds := asCertificateFieldList(ctx, x509.OrganizationalUnit)
	diags.Append(ds...)

	locality, ds := asCertificateFieldList(ctx, x509.Locality)
	diags.Append(ds...)

	province, ds := asCertificateFieldList(ctx, x509.Province)
	diags.Append(ds...)

	street, ds := asCertificateFieldList(ctx, x509.StreetAddress)
	diags.Append(ds...)

	postal, ds := asCertificateFieldList(ctx, x509.PostalCode)
	diags.Append(ds...)

	country, ds := asCertificateFieldList(ctx, x509.Country)
	diags.Append(ds...)

	return &v20250101.X509Fields{
		CommonName:         cn,
		Sans:               sans,
		Country:            country,
		Locality:           locality,
		Organization:       org,
		OrganizationalUnit: ou,
		PostalCode:         postal,
		Province:           province,
		StreetAddress:      street,
	}, diags
}

func (p *PolicyModel) toAPI(ctx context.Context) (*v20250101.PolicyMatchCriteria, diag.Diagnostics) {
	var diags diag.Diagnostics

	if p == nil {
		return nil, diags
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

	return policy, diags
}

func (cf CertificateFieldModel) toAPI() *v20250101.CertificateField {
	return &v20250101.CertificateField{
		Static:         cf.Static.ValueStringPointer(),
		DeviceMetadata: cf.DeviceMetadata.ValueStringPointer(),
	}
}

func (cfl CertificateFieldListModel) toAPI(ctx context.Context) (*v20250101.CertificateFieldList, diag.Diagnostics) {
	var diags diag.Diagnostics

	var static *[]string
	var deviceMetadata *[]string

	if !cfl.Static.IsNull() {
		diags.Append(cfl.Static.ElementsAs(ctx, &static, false)...)
	}
	if !cfl.DeviceMetadata.IsNull() {
		diags.Append(cfl.DeviceMetadata.ElementsAs(ctx, &deviceMetadata, false)...)
	}

	return &v20250101.CertificateFieldList{
		Static:         static,
		DeviceMetadata: deviceMetadata,
	}, diags
}

func asCertificateFieldList(ctx context.Context, obj types.Object) (*v20250101.CertificateFieldList, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	model := &CertificateFieldListModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	cfl, ds := model.toAPI(ctx)
	diags.Append(ds...)

	return cfl, diags
}

func asCertificateField(ctx context.Context, obj types.Object) (*v20250101.CertificateField, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	model := &CertificateFieldModel{}
	ds := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return model.toAPI(), diags
}

func toAPI(ctx context.Context, model *AccountModel) (*v20250101.Account, diag.Diagnostics) {
	var diags diag.Diagnostics

	account := &v20250101.Account{
		Id:            model.ID.ValueString(),
		Name:          model.Name.ValueString(),
		Configuration: &v20250101.Account_Configuration{},
	}

	cert := &CertificateModel{}
	ds := model.Certificate.As(ctx, &cert, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)
	certInfo, ds := cert.toAPI(ctx)
	diags.Append(ds...)
	account.CertificateInfo = certInfo

	key := &KeyModel{}
	ds = model.Key.As(ctx, &key, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)
	account.KeyInfo = key.toAPI()

	reload := &ReloadModel{}
	ds = model.Reload.As(ctx, &reload, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)
	account.ReloadInfo = reload.toAPI()

	policy := &PolicyModel{}
	ds = model.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)
	p, ds := policy.toAPI(ctx)
	diags.Append(ds...)
	account.Policy = p

	switch {
	case !model.WiFi.IsNull():
		t := v20250101.AccountTypeWifi
		account.Type = &t

		wifi := &WiFiModel{}
		ds := model.WiFi.As(ctx, wifi, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.Configuration.FromWifiAccount(v20250101.WifiAccount{
			Autojoin:              wifi.Autojoin.ValueBoolPointer(),
			CaChain:               wifi.CAChain.ValueStringPointer(),
			ExternalRadiusServer:  wifi.ExternalRadiusServer.ValueBoolPointer(),
			Hidden:                wifi.Hidden.ValueBoolPointer(),
			NetworkAccessServerIP: wifi.NetworkAccessServerIP.ValueStringPointer(),
			Ssid:                  wifi.SSID.ValueString(),
		})
		if err != nil {
			diags.AddError("Account WiFi Configuration Error", err.Error())
		}

	case !model.VPN.IsNull():
		typ := v20250101.AccountTypeVpn
		account.Type = &typ

		vpn := &VPNModel{}
		ds := model.VPN.As(ctx, vpn, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		vpnAccount := v20250101.VpnAccount{
			Autojoin:       vpn.Autojoin.ValueBoolPointer(),
			ConnectionType: v20250101.VpnType(vpn.ConnectionType.ValueString()),
			RemoteAddress:  vpn.RemoteAddress.ValueString(),
			Vendor:         utils.ToStringPointer[v20250101.VpnVendor](vpn.Vendor.ValueStringPointer()),
		}

		if !vpn.IKE.IsNull() && !vpn.IKE.IsUnknown() {
			ike := &IKEModel{}
			ds := vpn.IKE.As(ctx, ike, basetypes.ObjectAsOptions{})
			diags.Append(ds...)
			vpnAccount.Ike = &v20250101.IkeV2Config{
				CaChain:  ike.CAChain.ValueString(),
				Eap:      ike.EAP.ValueBoolPointer(),
				RemoteID: ike.RemoteID.ValueStringPointer(),
			}
		}

		err := account.Configuration.FromVpnAccount(vpnAccount)
		if err != nil {
			diags.AddError("Account VPN Configuration Error", err.Error())
		}

	case !model.Browser.IsNull():
		typ := v20250101.AccountTypeBrowser
		account.Type = &typ

		browser := &BrowserModel{}
		ds := model.Browser.As(ctx, browser, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.Configuration.FromBrowserAccount(v20250101.BrowserAccount{})
		if err != nil {
			diags.AddError("Account Browser Configuration Error", err.Error())
		}

	case !model.Ethernet.IsNull():
		typ := v20250101.AccountTypeEthernet
		account.Type = &typ

		ethernet := &EthernetModel{}
		ds := model.Ethernet.As(ctx, ethernet, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.Configuration.FromEthernetAccount(v20250101.EthernetAccount{
			Autojoin:              ethernet.Autojoin.ValueBoolPointer(),
			CaChain:               ethernet.CAChain.ValueStringPointer(),
			ExternalRadiusServer:  ethernet.ExternalRadiusServer.ValueBoolPointer(),
			NetworkAccessServerIP: ethernet.NetworkAccessServerIP.ValueStringPointer(),
		})
		if err != nil {
			diags.AddError("Account Ethernet Configuration Error", err.Error())
		}
	}

	return account, diags
}

func accountFromAPI(ctx context.Context, account *v20250101.Account, state utils.AttributeGetter) (*AccountModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	cert, d := certificateObjectFromAPI(ctx, account.CertificateInfo, state)
	diags.Append(d...)

	key, d := keyObjectFromAPI(ctx, account.KeyInfo, state)
	diags.Append(d...)

	reload, d := reloadObjectFromAPI(ctx, account.ReloadInfo, state)
	diags.Append(d...)

	policy, d := policyObjectFromAPI(ctx, account.Policy, state)
	diags.Append(d...)

	browser, ethernet, vpn, wifi, d := configurationObjectsFromAPI(ctx, account, state)
	diags.Append(d...)

	model := &AccountModel{
		ID:          types.StringValue(account.Id),
		Name:        types.StringValue(account.Name),
		Certificate: cert,
		Key:         key,
		Reload:      reload,
		Policy:      policy,
		Browser:     browser,
		Ethernet:    ethernet,
		VPN:         vpn,
		WiFi:        wifi,
	}

	return model, diags
}

func certificateObjectFromAPI(ctx context.Context, certInfo *v20250101.EndpointCertificateInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	p := path.Root("certificate")

	dur, d := utils.ToEqualString(ctx, certInfo.Duration, state, p.AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	crtFile, d := utils.ToOptionalString(ctx, certInfo.CrtFile, state, p.AtName("crt_file"))
	diags = append(diags, d...)

	keyFile, d := utils.ToOptionalString(ctx, certInfo.KeyFile, state, p.AtName("key_file"))
	diags = append(diags, d...)

	rootFile, d := utils.ToOptionalString(ctx, certInfo.RootFile, state, p.AtName("root_file"))
	diags = append(diags, d...)

	uid, d := utils.ToOptionalInt(ctx, certInfo.Uid, state, p.AtName("uid"))
	diags = append(diags, d...)

	gid, d := utils.ToOptionalInt(ctx, certInfo.Gid, state, p.AtName("gid"))
	diags = append(diags, d...)

	mode, d := utils.ToOptionalInt(ctx, certInfo.Mode, state, p.AtName("mode"))
	diags = append(diags, d...)

	authorityID, d := utils.ToOptionalString(ctx, certInfo.AuthorityID, state, p.AtName("authority_id"))
	diags = append(diags, d...)

	x509Obj := basetypes.NewObjectNull(x509Attributes)
	sshObj := basetypes.NewObjectNull(sshAttributes)

	switch certInfo.Type {
	case v20250101.EndpointCertificateInfoTypeX509:
		x509, err := certInfo.Details.AsX509Fields()
		if err != nil {
			diags.AddError("Parse certificate x509 attributes", err.Error())
		} else {
			x509Obj, d = x509ObjectFromAPI(ctx, x509, state)
			diags.Append(d...)
		}
	case v20250101.EndpointCertificateInfoTypeSSHUSER:
		ssh, err := certInfo.Details.AsSshFields()
		if err != nil {
			diags.AddError("Parse certificate ssh attributes", err.Error())
		} else {
			sshObj, d = sshObjectFromAPI(ctx, ssh, state)
			diags.Append(d...)
		}
	}

	out, d := basetypes.NewObjectValue(certificateAttributes, map[string]attr.Value{
		"crt_file":     crtFile,
		"key_file":     keyFile,
		"root_file":    rootFile,
		"duration":     dur,
		"gid":          gid,
		"uid":          uid,
		"mode":         mode,
		"ssh":          sshObj,
		"x509":         x509Obj,
		"authority_id": authorityID,
	})
	diags.Append(d...)

	return out, diags
}

func policyObjectFromAPI(ctx context.Context, policy *v20250101.PolicyMatchCriteria, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if policy == nil {
		// We got back nil policy from the API, but if we supplied empty policy, i.e.
		// "policy = {}", then we will get "inconsistent result after apply" error.
		// To avoid this we look up the object from state and if it was empty use
		// that rather than a null object.
		obj := &PolicyModel{}
		d := state.GetAttribute(ctx, path.Root("policy"), &obj)
		diags.Append(d...)
		if obj == nil {
			return basetypes.NewObjectNull(policyAttributes), diags
		}
		wasEmpty := false
		if obj.Assurance.IsNull() && obj.OS.IsNull() && obj.Ownership.IsNull() && obj.Source.IsNull() && obj.Tags.IsNull() {
			wasEmpty = true
		}
		if !wasEmpty {
			return basetypes.NewObjectNull(policyAttributes), diags
		}
		return basetypes.NewObjectValue(policyAttributes, map[string]attr.Value{
			"assurance": types.ListNull(types.StringType),
			"os":        types.ListNull(types.StringType),
			"ownership": types.ListNull(types.StringType),
			"source":    types.ListNull(types.StringType),
			"tags":      types.ListNull(types.StringType),
		})
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

	return obj, d
}

func certificateFieldObjectFromAPI(ctx context.Context, cf *v20250101.CertificateField, state utils.AttributeGetter, p path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if cf == nil {
		return basetypes.NewObjectNull(certificateFieldAttributes), diags
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

	return obj, diags
}

func certificateFieldListObjectFromAPI(ctx context.Context, cfl *v20250101.CertificateFieldList, state utils.AttributeGetter, p path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if cfl == nil {
		return basetypes.NewObjectNull(certificateFieldListAttributes), diags
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

	return obj, diags
}

func sshObjectFromAPI(ctx context.Context, ssh v20250101.SshFields, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	p := path.Root("certificate").AtName("ssh")

	keyID, d := certificateFieldObjectFromAPI(ctx, ssh.KeyId, state, p.AtName("key_id"))
	diags.Append(d...)

	principals, d := certificateFieldListObjectFromAPI(ctx, ssh.Principals, state, p.AtName("principals"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(sshAttributes, map[string]attr.Value{
		"key_id":     keyID,
		"principals": principals,
	})
	diags.Append(d...)

	return obj, diags
}

func x509ObjectFromAPI(ctx context.Context, x509 v20250101.X509Fields, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	p := path.Root("certificate").AtName("x509")

	cn, d := certificateFieldObjectFromAPI(ctx, x509.CommonName, state, p.AtName("common_name"))
	diags.Append(d...)

	sans, d := certificateFieldListObjectFromAPI(ctx, x509.Sans, state, p.AtName("sans"))
	diags.Append(d...)

	org, d := certificateFieldListObjectFromAPI(ctx, x509.Organization, state, p.AtName("organization"))
	diags.Append(d...)

	ou, d := certificateFieldListObjectFromAPI(ctx, x509.OrganizationalUnit, state, p.AtName("organizational_unit"))
	diags.Append(d...)

	locality, d := certificateFieldListObjectFromAPI(ctx, x509.Locality, state, p.AtName("locality"))
	diags.Append(d...)

	province, d := certificateFieldListObjectFromAPI(ctx, x509.Province, state, p.AtName("province"))
	diags.Append(d...)

	street, d := certificateFieldListObjectFromAPI(ctx, x509.StreetAddress, state, p.AtName("street_address"))
	diags.Append(d...)

	postal, d := certificateFieldListObjectFromAPI(ctx, x509.PostalCode, state, p.AtName("postal_code"))
	diags.Append(d...)

	country, d := certificateFieldListObjectFromAPI(ctx, x509.Country, state, p.AtName("country"))
	diags.Append(d...)

	obj, d := basetypes.NewObjectValue(x509Attributes, map[string]attr.Value{
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

	return obj, diags
}

func keyObjectFromAPI(ctx context.Context, key *v20250101.EndpointKeyInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if key == nil {
		return basetypes.NewObjectNull(keyAttributes), diags
	}

	format, ds := utils.ToOptionalString(ctx, key.Format, state, path.Root("key").AtName("format"))
	diags.Append(ds...)

	pubFile, ds := utils.ToOptionalString(ctx, key.PubFile, state, path.Root("key").AtName("pub_file"))
	diags.Append(ds...)

	typ, ds := utils.ToOptionalString(ctx, key.Type, state, path.Root("key").AtName("type"))
	diags.Append(ds...)

	protection, ds := utils.ToOptionalString(ctx, key.Protection, state, path.Root("key").AtName("protection"))
	diags.Append(ds...)

	out, ds := basetypes.NewObjectValue(keyAttributes, map[string]attr.Value{
		"format":     format,
		"pub_file":   pubFile,
		"type":       typ,
		"protection": protection,
	})
	diags.Append(ds...)

	return out, diags
}

func reloadObjectFromAPI(ctx context.Context, reloadInfo *v20250101.EndpointReloadInfo, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	pidFile, d := utils.ToOptionalString(ctx, reloadInfo.PidFile, state, path.Root("reload").AtName("method"))
	diags.Append(d...)

	signal, d := utils.ToOptionalInt(ctx, reloadInfo.Signal, state, path.Root("reload").AtName("signal"))
	diags.Append(d...)

	unitName, d := utils.ToOptionalString(ctx, reloadInfo.UnitName, state, path.Root("reload").AtName("unit_name"))
	diags.Append(d...)

	out, d := basetypes.NewObjectValue(reloadAttributes, map[string]attr.Value{
		"method":    types.StringValue(string(reloadInfo.Method)),
		"pid_file":  pidFile,
		"signal":    signal,
		"unit_name": unitName,
	})
	diags.Append(d...)

	return out, diags
}

// Returns (browser, ethernet, wifi, vpn) objects
func configurationObjectsFromAPI(ctx context.Context, account *v20250101.Account, state utils.AttributeGetter) (types.Object, types.Object, types.Object, types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	browserObj := basetypes.NewObjectNull(browserAttributes)
	ethernetObj := basetypes.NewObjectNull(ethernetAttributes)
	vpnObj := basetypes.NewObjectNull(vpnAttributes)
	wifiObj := basetypes.NewObjectNull(wifiAttributes)

	if account.Type == nil || *account.Type == "" || account.Configuration == nil {
		return browserObj, ethernetObj, vpnObj, wifiObj, diags
	}

	switch *account.Type {

	case v20250101.AccountTypeBrowser:
		_, err := account.Configuration.AsBrowserAccount()
		if err != nil {
			diags.AddError("Account Browser Parse Error", err.Error())
			break
		}

		obj, ds := basetypes.NewObjectValue(browserAttributes, map[string]attr.Value{})
		diags.Append(ds...)
		browserObj = obj

	case v20250101.AccountTypeEthernet:
		ethernet, err := account.Configuration.AsEthernetAccount()
		if err != nil {
			diags.AddError("Account Ethernet Parse Error", err.Error())
			break
		}

		autojoin, ds := utils.ToOptionalBool(ctx, ethernet.Autojoin, state, path.Root("ethernet").AtName("autojoin"))
		diags.Append(ds...)

		caChain, ds := utils.ToOptionalString(ctx, ethernet.CaChain, state, path.Root("ethernet").AtName("ca_chain"))
		diags.Append(ds...)

		externalRadiusServer, ds := utils.ToOptionalBool(ctx, ethernet.ExternalRadiusServer, state, path.Root("ethernet").AtName("external_radius_server"))
		diags.Append(ds...)

		nasIP, ds := utils.ToOptionalString(ctx, ethernet.NetworkAccessServerIP, state, path.Root("ethernet").AtName("network_access_server_ip"))
		diags.Append(ds...)

		ethernetObj, ds = basetypes.NewObjectValue(ethernetAttributes, map[string]attr.Value{
			"autojoin":                 autojoin,
			"ca_chain":                 caChain,
			"external_radius_server":   externalRadiusServer,
			"network_access_server_ip": nasIP,
		})
		diags.Append(ds...)

	case v20250101.AccountTypeVpn:
		vpn, err := account.Configuration.AsVpnAccount()
		if err != nil {
			diags.AddError("Account VPN Parse Error", err.Error())
			break
		}

		autojoin, ds := utils.ToOptionalBool(ctx, vpn.Autojoin, state, path.Root("vpn").AtName("autojoin"))
		diags.Append(ds...)

		vendor, ds := utils.ToOptionalString(ctx, vpn.Vendor, state, path.Root("vpn").AtName("vendor"))
		diags.Append(ds...)

		ike, ds := ikeObjectFromAPI(ctx, vpn.Ike, state)
		diags.Append(ds...)

		vpnObj, ds = basetypes.NewObjectValue(vpnAttributes, map[string]attr.Value{
			"autojoin":        autojoin,
			"connection_type": types.StringValue(string(vpn.ConnectionType)),
			"remote_address":  types.StringValue(vpn.RemoteAddress),
			"vendor":          vendor,
			"ike":             ike,
		})
		diags.Append(ds...)

	case v20250101.AccountTypeWifi:
		wifi, err := account.Configuration.AsWifiAccount()
		if err != nil {
			diags.AddError("Account WiFi Parse Error", err.Error())
			break
		}

		autojoin, ds := utils.ToOptionalBool(ctx, wifi.Autojoin, state, path.Root("wifi").AtName("autojoin"))
		diags.Append(ds...)

		caChain, ds := utils.ToOptionalString(ctx, wifi.CaChain, state, path.Root("wifi").AtName("ca_chain"))
		diags.Append(ds...)

		externalRadiusServer, ds := utils.ToOptionalBool(ctx, wifi.ExternalRadiusServer, state, path.Root("wifi").AtName("external_radius_server"))
		diags.Append(ds...)

		hidden, ds := utils.ToOptionalBool(ctx, wifi.Hidden, state, path.Root("wifi").AtName("hidden"))
		diags.Append(ds...)

		nasIP, ds := utils.ToOptionalString(ctx, wifi.NetworkAccessServerIP, state, path.Root("wifi").AtName("network_access_server_ip"))
		diags.Append(ds...)

		wifiObj, ds = basetypes.NewObjectValue(wifiAttributes, map[string]attr.Value{
			"autojoin":                 autojoin,
			"ca_chain":                 caChain,
			"external_radius_server":   externalRadiusServer,
			"hidden":                   hidden,
			"network_access_server_ip": nasIP,
			"ssid":                     types.StringValue(wifi.Ssid),
		})
		diags.Append(ds...)
	}

	return browserObj, ethernetObj, vpnObj, wifiObj, diags
}

func ikeObjectFromAPI(ctx context.Context, ike *v20250101.IkeV2Config, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ike == nil {
		return basetypes.NewObjectNull(ikeAttributes), diags
	}

	p := path.Root("vpn").AtName("ike")

	caChain, ds := utils.ToOptionalString(ctx, &ike.CaChain, state, p.AtName("ca_chain"))
	diags.Append(ds...)

	eap, ds := utils.ToOptionalBool(ctx, ike.Eap, state, p.AtName("eap"))
	diags.Append(ds...)

	remoteID, ds := utils.ToOptionalString(ctx, ike.RemoteID, state, p.AtName("remote_id"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ikeAttributes, map[string]attr.Value{
		"ca_chain":  caChain,
		"eap":       eap,
		"remote_id": remoteID,
	})
	diags.Append(ds...)

	return obj, diags
}
