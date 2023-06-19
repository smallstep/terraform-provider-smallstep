package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"go.step.sm/crypto/jose"
)

// type name for both resources and data sources
const provisionerTypeName = "smallstep_provisioner"

type Model struct {
	ID          types.String  `tfsdk:"id"`
	AuthorityID types.String  `tfsdk:"authority_id"`
	Name        types.String  `tfsdk:"name"`
	Type        types.String  `tfsdk:"type"`
	CreatedAt   types.String  `tfsdk:"created_at"`
	Claims      *ClaimsModel  `tfsdk:"claims"`
	Options     *OptionsModel `tfsdk:"options"`
	JWK         *JWKModel     `tfsdk:"jwk"`
	OIDC        *OIDCModel    `tfsdk:"oidc"`
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
	Admins                types.List   `tfsdk:"admins"`
	Domains               types.List   `tfsdk:"domains"`
	Groups                types.List   `tfsdk:"groups"`
	ListenAddress         types.String `tfsdk:"listen_address"`
	TenantID              types.String `tfsdk:"tenant_id"`
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
	}

	return p, nil
}

func fromAPI(provisioner *v20230301.Provisioner, authorityID string) (*Model, diag.Diagnostics) {
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
		data.Claims = &ClaimsModel{
			DisableRenewal:             types.BoolPointerValue(provisioner.Claims.DisableRenewal),
			AllowRenewalAfterExpiry:    types.BoolPointerValue(provisioner.Claims.AllowRenewalAfterExpiry),
			EnableSSHCA:                types.BoolPointerValue(provisioner.Claims.EnableSSHCA),
			MinTLSCertDuration:         types.StringPointerValue(provisioner.Claims.MinTLSCertDuration),
			MaxTLSCertDuration:         types.StringPointerValue(provisioner.Claims.MaxTLSCertDuration),
			DefaultTLSCertDuration:     types.StringPointerValue(provisioner.Claims.DefaultTLSCertDuration),
			MinUserSSHCertDuration:     types.StringPointerValue(provisioner.Claims.MinUserSSHCertDuration),
			MaxUserSSHCertDuration:     types.StringPointerValue(provisioner.Claims.MaxUserSSHCertDuration),
			DefaultUserSSHCertDuration: types.StringPointerValue(provisioner.Claims.DefaultUserSSHCertDuration),
			MinHostSSHCertDuration:     types.StringPointerValue(provisioner.Claims.MinHostSSHCertDuration),
			MaxHostSSHCertDuration:     types.StringPointerValue(provisioner.Claims.MaxHostSSHCertDuration),
			DefaultHostSSHCertDuration: types.StringPointerValue(provisioner.Claims.DefaultHostSSHCertDuration),
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
			Key:          types.StringValue(string(pubKeyJSON)),
			EncryptedKey: types.StringValue(utils.Deref(jwk.EncryptedKey)),
		}

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
			ClientSecret:          types.StringValue(oidc.ClientSecret),
			ConfigurationEndpoint: types.StringValue(oidc.ConfigurationEndpoint),
		}

		if oidc.Admins != nil {
			var admins []attr.Value
			for _, admin := range *oidc.Admins {
				admins = append(admins, types.StringValue(admin))
			}
			adminsList, listDiags := types.ListValue(types.StringType, admins)
			if listDiags.HasError() {
				return nil, listDiags
			}
			data.OIDC.Admins = adminsList
		} else {
			data.OIDC.Admins = types.ListNull(types.StringType)
		}

		if oidc.Domains != nil {
			var domains []attr.Value
			for _, domain := range *oidc.Domains {
				domains = append(domains, types.StringValue(domain))
			}
			domainsList, listDiags := types.ListValue(types.StringType, domains)
			if listDiags.HasError() {
				return nil, listDiags
			}
			data.OIDC.Domains = domainsList
		} else {
			data.OIDC.Domains = types.ListNull(types.StringType)
		}

		if oidc.Groups != nil {
			var groups []attr.Value
			for _, group := range *oidc.Groups {
				groups = append(groups, types.StringValue(group))
			}
			groupsList, listDiags := types.ListValue(types.StringType, groups)
			if listDiags.HasError() {
				return nil, listDiags
			}
			data.OIDC.Groups = groupsList
		} else {
			data.OIDC.Groups = types.ListNull(types.StringType)
		}

		if oidc.ListenAddress != nil {
			data.OIDC.ListenAddress = types.StringValue(*oidc.ListenAddress)
		}

		if oidc.TenantID != nil {
			data.OIDC.TenantID = types.StringValue(*oidc.TenantID)
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
