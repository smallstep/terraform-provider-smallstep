package strategy

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/credential"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/policy"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/browser"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/ethernet"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/relay"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/ssh"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/sso"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/vpn"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/wifi"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type StrategyModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Credential types.Object `tfsdk:"credential"`
	Policy     types.Object `tfsdk:"policy"`
	Browser    types.Object `tfsdk:"browser"`
	Ethernet   types.Object `tfsdk:"ethernet"`
	Relay      types.Object `tfsdk:"relay"`
	SSH        types.Object `tfsdk:"ssh"`
	SSO        types.Object `tfsdk:"sso"`
	VPN        types.Object `tfsdk:"vpn"`
	WiFi       types.Object `tfsdk:"wifi"`
}

func fromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (*StrategyModel, diag.Diagnostics) {
	var (
		diags       diag.Diagnostics
		browserObj  = basetypes.NewObjectNull(browser.Attributes)
		ethernetObj = basetypes.NewObjectNull(ethernet.Attributes)
		relayObj    = basetypes.NewObjectNull(relay.Attributes)
		sshObj      = basetypes.NewObjectNull(ssh.Attributes)
		ssoObj      = basetypes.NewObjectNull(sso.Attributes)
		vpnObj      = basetypes.NewObjectNull(vpn.Attributes)
		wifiObj     = basetypes.NewObjectNull(wifi.Attributes)
	)

	credentialObj, ds := credential.FromAPI(ctx, &strategy.EndpointConfiguration, state, path.Root("credential"))
	diags.Append(ds...)

	policyObj, ds := policy.FromAPI(ctx, strategy.EndpointConfiguration.Policy, state, path.Root("policy"))
	diags.Append(ds...)

	switch strategy.Kind {
	case v20250101.Browser:
		conf, err := strategy.Configuration.AsStrategyBrowserMutualTLSConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy Browser Parse Error", err.Error())
			return nil, diags
		}
		browserObj, ds = browser.FromAPI(ctx, &conf, state, path.Root("browser"))
		diags.Append(ds...)
	case v20250101.Ethernet:
		conf, err := strategy.Configuration.AsStrategyLANConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy Ethernet Parse Error", err.Error())
			return nil, diags
		}
		ethernetObj, ds = ethernet.FromAPI(ctx, &conf, state, path.Root("ethernet"))
		diags.Append(ds...)
	case v20250101.Relay:
		conf, err := strategy.Configuration.AsStrategyNetworkRelayConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy Relay Parse Error", err.Error())
			return nil, diags
		}
		relayObj, ds = relay.FromAPI(ctx, &conf, state, path.Root("relay"))
		diags.Append(ds...)
	case v20250101.Ssh:
		conf, err := strategy.Configuration.AsStrategySSHConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy SSH Parse Error", err.Error())
			return nil, diags
		}
		sshObj, ds = ssh.FromAPI(ctx, &conf, state, path.Root("ssh"))
		diags.Append(ds...)
	case v20250101.Sso:
		conf, err := strategy.Configuration.AsStrategySSOConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy SSO Parse Error", err.Error())
			return nil, diags
		}
		ssoObj, ds = sso.FromAPI(ctx, &conf, state, path.Root("sso"))
		diags.Append(ds...)
	case v20250101.Vpn:
		conf, err := strategy.Configuration.AsStrategyVPNConfig()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy VPN Parse Error", err.Error())
			return nil, diags
		}
		vpnObj, ds = vpn.FromAPI(ctx, &conf, state, path.Root("vpn"))
		diags.Append(ds...)
	case v20250101.Wifi:
		details, err := strategy.Details.AsStrategyWLAN()
		if err != nil && !isNullError(err) {
			diags.AddError("Strategy WiFi Parse Error", err.Error())
			return nil, diags
		}
		wifiObj, ds = wifi.FromAPI(ctx, &details, state, path.Root("wifi"))
		diags.Append(ds...)
	default:
		diags.AddError("Unsupported Strategy Kind", string(strategy.Kind))
		return nil, diags
	}

	return &StrategyModel{
		ID:         types.StringValue(strategy.Id),
		Name:       types.StringValue(strategy.Name),
		Credential: credentialObj,
		Policy:     policyObj,
		Browser:    browserObj,
		Ethernet:   ethernetObj,
		Relay:      relayObj,
		SSH:        sshObj,
		SSO:        ssoObj,
		VPN:        vpnObj,
		WiFi:       wifiObj,
	}, diags
}

func toAPI(ctx context.Context, model *StrategyModel) (*v20250101.ProtectionStrategy, diag.Diagnostics) {
	var diags, ds diag.Diagnostics

	strategy := &v20250101.ProtectionStrategy{
		Id:                    model.ID.ValueString(),
		Name:                  model.Name.ValueString(),
		EndpointConfiguration: v20250101.EndpointConfiguration{},
	}

	credential, ds := new(credential.Model).ToAPI(ctx, model.Credential)
	diags.Append(ds...)
	if credential != nil {
		strategy.EndpointConfiguration.CertificateInfo = credential.CertificateInfo
		strategy.EndpointConfiguration.KeyInfo = credential.KeyInfo
	}

	policy, ds := new(policy.Model).ToAPI(ctx, model.Policy)
	diags.Append(ds...)
	if policy != nil {
		strategy.EndpointConfiguration.Policy = policy
	}

	switch {
	case !model.Browser.IsNull():
		strategy.Kind = v20250101.Browser
		conf, ds := new(browser.Model).ToAPI(ctx, model.Browser)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyBrowserMutualTLSConfig(conf); err != nil {
			diags.AddError("Account Browser Configuration Error", err.Error())
		}
	case !model.Ethernet.IsNull():
		strategy.Kind = v20250101.Ethernet
		conf, ds := new(ethernet.Model).ToAPI(ctx, model.Ethernet)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyLANConfig(conf); err != nil {
			diags.AddError("Account Ethernet Configuration Error", err.Error())
		}
	case !model.Relay.IsNull():
		strategy.Kind = v20250101.Relay
		conf, ds := new(relay.Model).ToAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyNetworkRelayConfig(conf); err != nil {
			diags.AddError("Account Relay Configuration Error", err.Error())
		}
	case !model.SSH.IsNull():
		strategy.Kind = v20250101.Ssh
		conf, ds := new(ssh.Model).ToAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategySSHConfig(conf); err != nil {
			diags.AddError("Account SSH Configuration Error", err.Error())
		}
	case !model.SSO.IsNull():
		strategy.Kind = v20250101.Sso
		conf, ds := new(sso.Model).ToAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategySSOConfig(conf); err != nil {
			diags.AddError("Account SSO Configuration Error", err.Error())
		}
	case !model.VPN.IsNull():
		strategy.Kind = v20250101.Vpn
		conf, ds := new(vpn.Model).ToAPI(ctx, model.VPN)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyVPNConfig(conf); err != nil {
			diags.AddError("Account VPN Configuration Error", err.Error())
		}
	case !model.WiFi.IsNull():
		strategy.Kind = v20250101.Wifi
		conf, ds := new(wifi.Model).ToAPI(ctx, model.WiFi)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyWLANConfig(conf); err != nil {
			diags.AddError("Account WiFi Configuration Error", err.Error())
		}
	}

	return strategy, diags
}

func isNullError(err error) bool {
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return se.Offset == 0
	}
	return false
}
