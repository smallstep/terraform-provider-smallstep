package managed_configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_managed_configuration"

type Model struct {
	ID                   types.String      `tfsdk:"id"`
	Name                 types.String      `tfsdk:"name"`
	AgentConfigurationID types.String      `tfsdk:"agent_configuration_id"`
	HostID               types.String      `tfsdk:"host_id"`
	ManagedEndpoints     []ManagedEndpoint `tfsdk:"managed_endpoints"`
}

func (model Model) toAPI(ctx context.Context) (v20230301.ManagedConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	mc := v20230301.ManagedConfiguration{
		Id:                   model.ID.ValueStringPointer(),
		Name:                 model.Name.ValueString(),
		AgentConfigurationID: model.AgentConfigurationID.ValueString(),
		HostID:               model.HostID.ValueStringPointer(),
	}

	for _, me := range model.ManagedEndpoints {
		me, d := me.toAPI(ctx)
		diags.Append(d...)
		mc.ManagedEndpoints = append(mc.ManagedEndpoints, me)
	}

	return mc, diags
}

type ManagedEndpoint struct {
	ID                      types.String         `tfsdk:"id"`
	EndpointConfigurationID types.String         `tfsdk:"endpoint_configuration_id"`
	SSHCertificateData      *SSHCertificateData  `tfsdk:"ssh_certificate_data"`
	X509CertificateData     *X509CertificateData `tfsdk:"x509_certificate_data"`
}

func (me *ManagedEndpoint) toAPI(ctx context.Context) (v20230301.ManagedEndpoint, diag.Diagnostics) {
	x509, diags := me.X509CertificateData.toAPI(ctx)
	ssh, d := me.SSHCertificateData.toAPI(ctx)
	diags.Append(d...)

	return v20230301.ManagedEndpoint{
		Id:                      me.ID.ValueStringPointer(),
		EndpointConfigurationID: me.EndpointConfigurationID.ValueString(),
		X509CertificateData:     x509,
		SshCertificateData:      ssh,
	}, diags
}

type SSHCertificateData struct {
	KeyID      types.String `tfsdk:"key_id"`
	Principals types.Set    `tfsdk:"principals"`
}

func (ssh *SSHCertificateData) toAPI(ctx context.Context) (*v20230301.EndpointSSHCertificateData, diag.Diagnostics) {
	if ssh == nil {
		return nil, diag.Diagnostics{}
	}
	var principals []string
	diags := ssh.Principals.ElementsAs(ctx, &principals, false)

	return &v20230301.EndpointSSHCertificateData{
		KeyID:      ssh.KeyID.ValueString(),
		Principals: principals,
	}, diags
}

type X509CertificateData struct {
	CommonName types.String `tfsdk:"common_name"`
	SANs       types.Set    `tfsdk:"sans"`
}

func (x509 *X509CertificateData) toAPI(ctx context.Context) (*v20230301.EndpointX509CertificateData, diag.Diagnostics) {
	if x509 == nil {
		return nil, diag.Diagnostics{}
	}

	var sans []string
	diags := x509.SANs.ElementsAs(ctx, &sans, false)

	return &v20230301.EndpointX509CertificateData{
		CommonName: x509.CommonName.ValueString(),
		Sans:       sans,
	}, diags
}

func fromAPI(ctx context.Context, mc *v20230301.ManagedConfiguration, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		ID:                   types.StringValue(utils.Deref(mc.Id)),
		Name:                 types.StringValue(mc.Name),
		AgentConfigurationID: types.StringValue(mc.AgentConfigurationID),
		HostID:               types.StringValue(utils.Deref(mc.HostID)),
	}

	for i, me := range mc.ManagedEndpoints {
		p := path.Root("managed_endpoints").AtListIndex(i)
		ssh, d := fromSSHCertificateData(ctx, me.SshCertificateData, state, p)
		diags.Append(d...)

		x509, d := fromX509CertificateData(ctx, me.X509CertificateData, state, p)
		diags.Append(d...)

		model.ManagedEndpoints = append(model.ManagedEndpoints, ManagedEndpoint{
			ID:                      types.StringValue(utils.Deref(me.Id)),
			EndpointConfigurationID: types.StringValue(me.EndpointConfigurationID),
			SSHCertificateData:      ssh,
			X509CertificateData:     x509,
		})
	}

	return model, diags
}

func fromSSHCertificateData(ctx context.Context, ssh *v20230301.EndpointSSHCertificateData, state utils.AttributeGetter, p path.Path) (*SSHCertificateData, diag.Diagnostics) {
	if ssh == nil {
		return nil, diag.Diagnostics{}
	}

	values := make([]attr.Value, len(ssh.Principals))
	for i, s := range ssh.Principals {
		values[i] = types.StringValue(s)
	}
	principals, diags := types.SetValue(types.StringType, values)

	return &SSHCertificateData{
		KeyID:      types.StringValue(ssh.KeyID),
		Principals: principals,
	}, diags
}

func fromX509CertificateData(ctx context.Context, x509 *v20230301.EndpointX509CertificateData, state utils.AttributeGetter, p path.Path) (*X509CertificateData, diag.Diagnostics) {
	if x509 == nil {
		return nil, diag.Diagnostics{}
	}

	values := make([]attr.Value, len(x509.Sans))
	for i, s := range x509.Sans {
		values[i] = types.StringValue(s)
	}
	sans, diags := types.SetValue(types.StringType, values)

	return &X509CertificateData{
		CommonName: types.StringValue(x509.CommonName),
		SANs:       sans,
	}, diags
}
