package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"go.step.sm/crypto/jose"
)

// type name for both resources and data sources
const provisionerTypeName = "smallstep_provisioner"

type Model struct {
	ID              types.String          `tfsdk:"id"`
	AuthorityID     types.String          `tfsdk:"authority_id"`
	Name            types.String          `tfsdk:"name"`
	Type            types.String          `tfsdk:"type"`
	CreatedAt       types.String          `tfsdk:"created_at"`
	Claims          *ClaimsModel          `tfsdk:"claims"`
	Options         *OptionsModel         `tfsdk:"options"`
	JWK             *JWKModel             `tfsdk:"jwk"`
	OIDC            *OIDCModel            `tfsdk:"oidc"`
	ACME            *ACMEModel            `tfsdk:"acme"`
	ACMEAttestation *ACMEAttestationModel `tfsdk:"acme_attestation"`
	X5C             *X5CModel             `tfsdk:"x5c"`
}

type OptionsModel struct {
	X509 *TemplateModel `tfsdk:"x509"`
	SSH  *TemplateModel `tfsdk:"ssh"`
}

type TemplateModel struct {
	Template     types.String `tfsdk:"template"`
	TemplateData types.String `tfsdk:"template_data"`
}

type ClaimsModel struct {
	DisableRenewal             types.Bool   `tfsdk:"disable_renewal"`
	AllowRenewalAfterExpiry    types.Bool   `tfsdk:"allow_renewal_after_expiry"`
	EnableSSHCA                types.Bool   `tfsdk:"enable_ssh_ca"`
	MinTLSCertDuration         types.String `tfsdk:"min_tls_cert_duration"`
	MaxTLSCertDuration         types.String `tfsdk:"max_tls_cert_duration"`
	DefaultTLSCertDuration     types.String `tfsdk:"default_tls_cert_duration"`
	MinUserSSHCertDuration     types.String `tfsdk:"min_user_ssh_cert_duration"`
	MaxUserSSHCertDuration     types.String `tfsdk:"max_user_ssh_cert_duration"`
	DefaultUserSSHCertDuration types.String `tfsdk:"default_user_ssh_cert_duration"`
	MinHostSSHCertDuration     types.String `tfsdk:"min_host_ssh_cert_duration"`
	MaxHostSSHCertDuration     types.String `tfsdk:"max_host_ssh_cert_duration"`
	DefaultHostSSHCertDuration types.String `tfsdk:"default_host_ssh_cert_duration"`
}

func (claims ClaimsModel) isEmpty() bool {
	switch {
	case !claims.DisableRenewal.IsNull():
		return false
	case !claims.AllowRenewalAfterExpiry.IsNull():
		return false
	case !claims.EnableSSHCA.IsNull():
		return false
	case !claims.MinTLSCertDuration.IsNull():
		return false
	case !claims.MaxTLSCertDuration.IsNull():
		return false
	case !claims.DefaultTLSCertDuration.IsNull():
		return false
	case !claims.MinUserSSHCertDuration.IsNull():
		return false
	case !claims.MaxUserSSHCertDuration.IsNull():
		return false
	case !claims.DefaultUserSSHCertDuration.IsNull():
		return false
	case !claims.MinHostSSHCertDuration.IsNull():
		return false
	case !claims.MaxHostSSHCertDuration.IsNull():
		return false
	case !claims.DefaultHostSSHCertDuration.IsNull():
		return false
	}

	return true
}

type JWKModel struct {
	Key          types.String `tfsdk:"key"`
	EncryptedKey types.String `tfsdk:"encrypted_key"`
}

type OIDCModel struct {
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ConfigurationEndpoint types.String `tfsdk:"configuration_endpoint"`
	Admins                types.Set    `tfsdk:"admins"`
	Domains               types.Set    `tfsdk:"domains"`
	Groups                types.Set    `tfsdk:"groups"`
	ListenAddress         types.String `tfsdk:"listen_address"`
	TenantID              types.String `tfsdk:"tenant_id"`
}

type ACMEModel struct {
	Challenges types.Set  `tfsdk:"challenges"`
	ForceCN    types.Bool `tfsdk:"force_cn"`
	RequireEAB types.Bool `tfsdk:"require_eab"`
}

type ACMEAttestationModel struct {
	AttestationFormats types.Set  `tfsdk:"attestation_formats"`
	AttestationRoots   types.Set  `tfsdk:"attestation_roots"`
	ForceCN            types.Bool `tfsdk:"force_cn"`
	RequireEAB         types.Bool `tfsdk:"require_eab"`
}

type X5CModel struct {
	Roots types.Set `tfsdk:"roots"`
}

func toAPI(ctx context.Context, m *Model) (*v20230301.Provisioner, error) {
	p := &v20230301.Provisioner{
		Id:   m.ID.ValueStringPointer(),
		Name: m.Name.ValueString(),
		Type: v20230301.ProvisionerType(m.Type.ValueString()),
	}
	if m.Claims != nil {
		p.Claims = &v20230301.ProvisionerClaims{
			DisableRenewal:             m.Claims.DisableRenewal.ValueBoolPointer(),
			AllowRenewalAfterExpiry:    m.Claims.AllowRenewalAfterExpiry.ValueBoolPointer(),
			EnableSSHCA:                m.Claims.EnableSSHCA.ValueBoolPointer(),
			MinTLSCertDuration:         m.Claims.MinTLSCertDuration.ValueStringPointer(),
			MaxTLSCertDuration:         m.Claims.MaxTLSCertDuration.ValueStringPointer(),
			DefaultTLSCertDuration:     m.Claims.DefaultTLSCertDuration.ValueStringPointer(),
			MinUserSSHCertDuration:     m.Claims.MinUserSSHCertDuration.ValueStringPointer(),
			MaxUserSSHCertDuration:     m.Claims.MaxUserSSHCertDuration.ValueStringPointer(),
			DefaultUserSSHCertDuration: m.Claims.DefaultUserSSHCertDuration.ValueStringPointer(),
			MinHostSSHCertDuration:     m.Claims.MinHostSSHCertDuration.ValueStringPointer(),
			MaxHostSSHCertDuration:     m.Claims.MaxHostSSHCertDuration.ValueStringPointer(),
			DefaultHostSSHCertDuration: m.Claims.DefaultHostSSHCertDuration.ValueStringPointer(),
		}
	}

	if m.Options != nil {
		p.Options = &v20230301.ProvisionerOptions{}
		if m.Options.X509 != nil {
			p.Options.X509 = &v20230301.X509Options{
				Template: m.Options.X509.Template.ValueStringPointer(),
			}
			if !m.Options.X509.TemplateData.IsNull() {
				var tmplData any
				err := json.Unmarshal([]byte(m.Options.X509.TemplateData.ValueString()), &tmplData)
				if err != nil {
					return nil, err
				}
				p.Options.X509.TemplateData = &tmplData
			}
		}
		if m.Options.SSH != nil {
			p.Options.Ssh = &v20230301.SshOptions{
				Template: m.Options.SSH.Template.ValueStringPointer(),
			}
			if !m.Options.SSH.TemplateData.IsNull() {
				var tmplData any
				err := json.Unmarshal([]byte(m.Options.SSH.TemplateData.ValueString()), &tmplData)
				if err != nil {
					return nil, err
				}
				p.Options.Ssh.TemplateData = &tmplData
			}
		}
	}

	switch {
	case m.JWK != nil:
		ek := m.JWK.EncryptedKey.ValueString()
		jwk := v20230301.JwkProvisioner{
			Key:          map[string]any{},
			EncryptedKey: &ek,
		}

		// validate the public key
		pubKey := &jose.JSONWebKey{}
		if err := json.Unmarshal([]byte(m.JWK.Key.ValueString()), pubKey); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(m.JWK.Key.ValueString()), &jwk.Key); err != nil {
			return nil, err
		}

		if err := p.FromJwkProvisioner(jwk); err != nil {
			return nil, err
		}
	case m.OIDC != nil:
		oidc := v20230301.OidcProvisioner{
			ClientID:              m.OIDC.ClientID.ValueString(),
			ClientSecret:          m.OIDC.ClientSecret.ValueString(),
			ConfigurationEndpoint: m.OIDC.ConfigurationEndpoint.ValueString(),
		}
		if !m.OIDC.Admins.IsNull() {
			diagnostics := m.OIDC.Admins.ElementsAs(ctx, &oidc.Admins, false)
			if err := utils.DiagnosticsToErr(diagnostics); err != nil {
				return nil, err
			}
		}
		if !m.OIDC.Domains.IsNull() {
			diagnostics := m.OIDC.Domains.ElementsAs(ctx, &oidc.Domains, false)
			if err := utils.DiagnosticsToErr(diagnostics); err != nil {
				return nil, err
			}
		}
		if !m.OIDC.Groups.IsNull() {
			diagnostics := m.OIDC.Groups.ElementsAs(ctx, &oidc.Groups, false)
			if err := utils.DiagnosticsToErr(diagnostics); err != nil {
				return nil, err
			}
		}
		if !m.OIDC.ListenAddress.IsNull() {
			oidc.ListenAddress = m.OIDC.ListenAddress.ValueStringPointer()
		}
		if !m.OIDC.TenantID.IsNull() {
			oidc.TenantID = m.OIDC.TenantID.ValueStringPointer()
		}
		if err := p.FromOidcProvisioner(oidc); err != nil {
			return nil, err
		}
	case m.ACME != nil:
		acme := v20230301.AcmeProvisioner{
			ForceCN:    m.ACME.ForceCN.ValueBoolPointer(),
			RequireEAB: m.ACME.RequireEAB.ValueBool(),
		}

		diagnostics := m.ACME.Challenges.ElementsAs(ctx, &acme.Challenges, false)
		if err := utils.DiagnosticsToErr(diagnostics); err != nil {
			return nil, err
		}

		if err := p.FromAcmeProvisioner(acme); err != nil {
			return nil, err
		}
	case m.ACMEAttestation != nil:
		attest := v20230301.AcmeAttestationProvisioner{
			ForceCN:    m.ACMEAttestation.ForceCN.ValueBoolPointer(),
			RequireEAB: m.ACMEAttestation.RequireEAB.ValueBoolPointer(),
		}

		diagnostics := m.ACMEAttestation.AttestationFormats.ElementsAs(ctx, &attest.AttestationFormats, false)
		if err := utils.DiagnosticsToErr(diagnostics); err != nil {
			return nil, err
		}

		if !m.ACMEAttestation.AttestationRoots.IsNull() {
			diagnostics = m.ACMEAttestation.AttestationRoots.ElementsAs(ctx, &attest.AttestationRoots, false)
			if err := utils.DiagnosticsToErr(diagnostics); err != nil {
				return nil, err
			}
		}

		if err := p.FromAcmeAttestationProvisioner(attest); err != nil {
			return nil, err
		}
	case m.X5C != nil:
		x5c := v20230301.X5cProvisioner{}
		diagnostics := m.X5C.Roots.ElementsAs(ctx, &x5c.Roots, false)
		if err := utils.DiagnosticsToErr(diagnostics); err != nil {
			return nil, err
		}

		if err := p.FromX5cProvisioner(x5c); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func fromAPI(ctx context.Context, provisioner *v20230301.Provisioner, authorityID string, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := &Model{
		ID:          types.StringValue(utils.Deref(provisioner.Id)),
		AuthorityID: types.StringValue(authorityID),
		Name:        types.StringValue(provisioner.Name),
		Type:        types.StringValue(string(provisioner.Type)),
	}
	if provisioner.CreatedAt != nil {
		data.CreatedAt = types.StringValue((*provisioner.CreatedAt).Format(time.RFC3339))
	}

	if provisioner.Claims != nil {
		disableRenewal, d := utils.ToOptionalBool(ctx, provisioner.Claims.DisableRenewal, state, path.Root("claims").AtName("disable_renewal"))
		diags = append(diags, d...)

		allowRenewalAfterExpiry, d := utils.ToOptionalBool(ctx, provisioner.Claims.AllowRenewalAfterExpiry, state, path.Root("claims").AtName("allow_renewal_after_expiry"))
		diags = append(diags, d...)

		enableSSHCA, d := utils.ToOptionalBool(ctx, provisioner.Claims.EnableSSHCA, state, path.Root("claims").AtName("enable_ssh_ca"))
		diags = append(diags, d...)

		minTLSCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MinTLSCertDuration, state, path.Root("claims").AtName("min_tls_cert_duration"))
		diags = append(diags, d...)

		maxTLSCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MaxTLSCertDuration, state, path.Root("claims").AtName("max_tls_cert_duration"))
		diags = append(diags, d...)

		defaultTLSCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.DefaultTLSCertDuration, state, path.Root("claims").AtName("default_tls_cert_duration"))
		diags = append(diags, d...)

		minUserSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MinUserSSHCertDuration, state, path.Root("claims").AtName("min_user_ssh_cert_duration"))
		diags = append(diags, d...)

		maxUserSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MaxUserSSHCertDuration, state, path.Root("claims").AtName("max_user_ssh_cert_duration"))
		diags = append(diags, d...)

		defaultUserSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.DefaultUserSSHCertDuration, state, path.Root("claims").AtName("default_user_ssh_cert_duration"))
		diags = append(diags, d...)

		minHostSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MinHostSSHCertDuration, state, path.Root("claims").AtName("min_host_ssh_cert_duration"))
		diags = append(diags, d...)

		maxHostSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.MaxHostSSHCertDuration, state, path.Root("claims").AtName("max_host_ssh_cert_duration"))
		diags = append(diags, d...)

		defaultHostSSHCertDuration, d := utils.ToOptionalString(ctx, provisioner.Claims.DefaultHostSSHCertDuration, state, path.Root("claims").AtName("default_host_ssh_cert_duration"))
		diags = append(diags, d...)

		data.Claims = &ClaimsModel{
			DisableRenewal:             disableRenewal,
			AllowRenewalAfterExpiry:    allowRenewalAfterExpiry,
			EnableSSHCA:                enableSSHCA,
			MinTLSCertDuration:         minTLSCertDuration,
			MaxTLSCertDuration:         maxTLSCertDuration,
			DefaultTLSCertDuration:     defaultTLSCertDuration,
			MinUserSSHCertDuration:     minUserSSHCertDuration,
			MaxUserSSHCertDuration:     maxUserSSHCertDuration,
			DefaultUserSSHCertDuration: defaultUserSSHCertDuration,
			MinHostSSHCertDuration:     minHostSSHCertDuration,
			MaxHostSSHCertDuration:     maxHostSSHCertDuration,
			DefaultHostSSHCertDuration: defaultHostSSHCertDuration,
		}
	}

	if provisioner.Options != nil {
		data.Options = &OptionsModel{}

		if provisioner.Options.X509 != nil {
			data.Options.X509 = &TemplateModel{
				Template: types.StringPointerValue(provisioner.Options.X509.Template),
			}
			if provisioner.Options.X509.TemplateData != nil {
				tmplData, err := json.Marshal(provisioner.Options.X509.TemplateData)
				if err != nil {
					diags.AddError(
						"Serialize X509 Template Data",
						err.Error(),
					)
					return nil, diags
				}
				data.Options.X509.TemplateData = types.StringValue(string(tmplData))
			}
		}
		if provisioner.Options.Ssh != nil {
			data.Options.SSH = &TemplateModel{
				Template: types.StringPointerValue(provisioner.Options.Ssh.Template),
			}
			if provisioner.Options.Ssh.TemplateData != nil {
				tmplData, err := json.Marshal(provisioner.Options.Ssh.TemplateData)
				if err != nil {
					diags.AddError(
						"Serialize SSH Template Data",
						err.Error(),
					)
					return nil, diags
				}
				data.Options.SSH.TemplateData = types.StringValue(string(tmplData))
			}
		}
	}

	switch provisioner.Type {
	case v20230301.JWK:
		jwk, err := provisioner.AsJwkProvisioner()
		if err != nil {
			diags.AddError(
				"Parse JWK Provisioner",
				fmt.Sprintf("provisioner %s: %v", data.Name.ValueString(), err),
			)
			return nil, diags
		}
		pubKeyJSON, err := json.Marshal(jwk.Key)
		if err != nil {
			diags.AddError(
				"Serialize JWK Key",
				fmt.Sprintf("Failed to serialize JWK public key for provisioner %s: %v", data.Name.ValueString(), err),
			)
		}
		data.JWK = &JWKModel{
			Key: types.StringValue(string(pubKeyJSON)),
		}
		encryptedKey, diags := utils.ToOptionalString(ctx, jwk.EncryptedKey, state, path.Root("jwk").AtName("encrypted_key"))
		if diags.HasError() {
			return nil, diags
		}
		data.JWK.EncryptedKey = encryptedKey

	case v20230301.OIDC:
		oidc, err := provisioner.AsOidcProvisioner()
		if err != nil {
			diags.AddError(
				"Parse OIDC Provisioner",
				fmt.Sprintf("provisioner %s: %v", data.Name.ValueString(), err),
			)
			return nil, diags
		}

		data.OIDC = &OIDCModel{
			ClientID:              types.StringValue(oidc.ClientID),
			ConfigurationEndpoint: types.StringValue(oidc.ConfigurationEndpoint),
		}

		secret, diags := utils.ToOptionalString(ctx, &oidc.ClientSecret, state, path.Root("oidc").AtName("client_secret"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.ClientSecret = secret

		admins, diags := utils.ToOptionalSet(ctx, oidc.Admins, state, path.Root("oidc").AtName("admins"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.Admins = admins

		domains, diags := utils.ToOptionalSet(ctx, oidc.Domains, state, path.Root("oidc").AtName("domains"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.Domains = domains

		groups, diags := utils.ToOptionalSet(ctx, oidc.Groups, state, path.Root("oidc").AtName("groups"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.Groups = groups

		listen, diags := utils.ToOptionalString(ctx, oidc.ListenAddress, state, path.Root("oidc").AtName("listen_address"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.ListenAddress = listen

		tenantID, diags := utils.ToOptionalString(ctx, oidc.TenantID, state, path.Root("oidc").AtName("tenant_id"))
		if diags.HasError() {
			return nil, diags
		}
		data.OIDC.TenantID = tenantID

	case v20230301.ACME:
		acme, err := provisioner.AsAcmeProvisioner()
		if err != nil {
			diags.AddError(
				"Parse ACME Provisioner",
				fmt.Sprintf("provisioner %s: %v", data.Name.ValueString(), err),
			)
			return nil, diags
		}
		data.ACME = &ACMEModel{
			RequireEAB: types.BoolValue(acme.RequireEAB),
			ForceCN:    types.BoolPointerValue(acme.ForceCN),
		}

		var challenges []attr.Value
		for _, challenge := range acme.Challenges {
			challenges = append(challenges, types.StringValue(string(challenge)))
		}
		challengesSet, diags := types.SetValue(types.StringType, challenges)
		if diags.HasError() {
			return nil, diags
		}
		data.ACME.Challenges = challengesSet

	case v20230301.ACMEATTESTATION:
		attest, err := provisioner.AsAcmeAttestationProvisioner()
		if err != nil {
			diags.AddError(
				"Parse ACME Attestation Provisioner",
				fmt.Sprintf("provisioner %s: %v", data.Name.ValueString(), err),
			)
			return nil, diags
		}
		data.ACMEAttestation = &ACMEAttestationModel{
			RequireEAB: types.BoolPointerValue(attest.RequireEAB),
			ForceCN:    types.BoolPointerValue(attest.ForceCN),
		}

		var attestationFormats []attr.Value
		for _, format := range attest.AttestationFormats {
			attestationFormats = append(attestationFormats, types.StringValue(string(format)))
		}
		formatsSet, diags := types.SetValue(types.StringType, attestationFormats)
		if diags.HasError() {
			return nil, diags
		}
		data.ACMEAttestation.AttestationFormats = formatsSet

		if attest.AttestationRoots != nil {
			var roots []attr.Value
			for _, root := range *attest.AttestationRoots {
				roots = append(roots, types.StringValue(root))
			}
			rootsSet, diags := types.SetValue(types.StringType, roots)
			if diags.HasError() {
				return nil, diags
			}
			data.ACMEAttestation.AttestationRoots = rootsSet
		} else {
			data.ACMEAttestation.AttestationRoots = types.SetNull(types.StringType)
		}

		attestationRoots, diags := utils.ToOptionalSet(ctx, attest.AttestationRoots, state, path.Root("acme_attestation").AtName("attestation_roots"))
		if diags.HasError() {
			return nil, diags
		}
		data.ACMEAttestation.AttestationRoots = attestationRoots

	case v20230301.X5C:
		x5c, err := provisioner.AsX5cProvisioner()
		if err != nil {
			diags.AddError(
				"Parse X5C Provisioner",
				fmt.Sprintf("provisioner %s: %v", data.Name.ValueString(), err),
			)
			return nil, diags
		}

		var roots []attr.Value
		for _, root := range x5c.Roots {
			roots = append(roots, types.StringValue(root))
		}
		rootsSet, diags := types.SetValue(types.StringType, roots)
		if diags.HasError() {
			return nil, diags
		}
		data.X5C = &X5CModel{
			Roots: rootsSet,
		}

	default:
		diags.AddError(
			"Smallstep Invalid Provisioner",
			fmt.Sprintf("Unknown type for provisioner %s", data.Name.ValueString()),
		)
		return nil, diags
	}

	return data, diags
}
