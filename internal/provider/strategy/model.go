package strategy

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

type StrategyModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Browser  types.Object `tfsdk:"browser"`
	Ethernet types.Object `tfsdk:"ethernet"`
	Relay    types.Object `tfsdk:"relay"`
	SSH      types.Object `tfsdk:"ssh"`
	SSO      types.Object `tfsdk:"sso"`
	VPN      types.Object `tfsdk:"vpn"`
	WiFi     types.Object `tfsdk:"wifi"`
}

type BrowserMutualTLSModel struct {
	MatchAddresses types.List `tfsdk:"match_addresses"`
}

var browserMutualTLSAttributes = map[string]attr.Type{
	"match_addresses": types.ListType{ElemType: types.StringType},
}

func (m *BrowserMutualTLSModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategyBrowserMutualTLS, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyBrowserMutualTLS{
		MatchAddresses: toList[string](m.MatchAddresses),
	}, diags
}

type LANModel struct {
	NetworkAccessServerIP types.String `tfsdk:"NetworkAccessServerIP"`
	CAChain               types.String `tfsdk:"ca_chain"`
	Autojoin              types.Bool   `tfsdk;"autojoin"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
}

var lanAttributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ca_chain":                 types.StringType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
}

func (m *LANModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategyLAN, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyLAN{
		NetworkAccessServerIP: m.NetworkAccessServerIP.ValueStringPointer(),
		CaChain:               m.CAChain.ValueStringPointer(),
		Autojoin:              m.Autojoin.ValueBoolPointer(),
		ExternalRadiusServer:  m.ExternalRadiusServer.ValueBoolPointer(),
	}, diags
}

type NetworkRelayModel struct {
	MatchDomains types.List `tfsdk:"match_domains"`
	Regions      types.List `tfsdk:"regions"`
}

var networkRelayAttributes = map[string]attr.Type{
	"match_domains": types.ListType{ElemType: types.StringType},
	"regions":       types.ListType{ElemType: types.StringType},
}

func (m *NetworkRelayModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategyNetworkRelay, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyNetworkRelay{
		MatchDomains: toList[string](m.MatchDomains),
		Regions:      toList[v20250101.StrategyNetworkRelayRegions](m.Regions),
	}, diags
}

type SSHModel struct{}

var sshAttributes = map[string]attr.Type{}

func (m *SSHModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategySSH, diag.Diagnostics) {
	var diags diag.Diagnostics
	return v20250101.StrategySSH{}, diags
}

type SSOModel struct {
	TrustedRoots types.String `tfsdk:"trusted_roots"`
	RedirectURI  types.String `tfsdk:"redirect_uri"`
}

var ssoAttributes = map[string]attr.Type{
	"trusted_roots": types.StringType,
	"redirect_uri":  types.StringType,
}

func (m *SSOModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategySSO, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategySSO{
		TrustRoots:  m.TrustedRoots.ValueString(),
		RedirectUri: m.RedirectURI.ValueString(),
	}, diags
}

type VPNModel struct {
	ConnectionType types.String `tfsdk:"connection_type"`
	Vendor         types.String `tfsdk:"vendor"`
	RemoteAddress  types.String `tfsdk:"remote_address"`
	IKEV2Config    types.Object `tfsdk:"ike"`
	Autojoin       types.Bool   `tfsdk:"autojoin"`
}

var vpnAttributes = map[string]attr.Type{
	"connection_type": types.StringType,
	"vendor":          types.StringType,
	"remote_address":  types.StringType,
	"ike":             types.ObjectType{AttrTypes: ikeAttributes},
	"autojoin":        types.BoolType,
}

func (m *VPNModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategyVPN, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyVPN{
		ConnectionType: v20250101.VpnAccountConnectionType(m.ConnectionType.ValueString()),
		Vendor:         utils.ToStringPointer[v20250101.VpnAccountVendor](m.Vendor.ValueStringPointer()),
		RemoteAddress:  m.RemoteAddress.ValueString(),
		// IKEV2Configm: ,
		Autojoin: m.Autojoin.ValueBoolPointer(),
	}, diags
}

type IKEModel struct {
	CAChain  types.String `tfsdk:"ca_chain"`
	RemoteID types.String `tfsdk:"remote_id"`
	EAP      types.Bool   `tfsdk:"eap"`
}

var ikeAttributes = map[string]attr.Type{
	"ca_chain":  types.StringType,
	"remote_id": types.StringType,
	"eap":       types.BoolType,
}

func (m *IKEModel) toAPI(ctx context.Context, obj types.Object) (v20250101.IkeV2Config, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.IkeV2Config{
		CaChain:  m.CAChain.ValueStringPointer(),
		RemoteID: m.RemoteID.ValueStringPointer(),
		Eap:      m.EAP.ValueBoolPointer(),
	}, diags
}

type WLANModel struct {
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
	SSID                  types.String `tfsdk:"ssid"`
	CAChain               types.String `tfsdk:"ca_chain"`
	Hidden                types.Bool   `tfsdk:"hidden"`
	Autojoin              types.Bool   `tfsdk;"autojoin"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
}

var wlanAttributes = map[string]attr.Type{
	"network_access_server_ip": types.StringType,
	"ssid":                     types.StringType,
	"ca_chain":                 types.StringType,
	"hidden":                   types.BoolType,
	"autojoin":                 types.BoolType,
	"external_radius_server":   types.BoolType,
}

func (m *WLANModel) toAPI(ctx context.Context, obj types.Object) (v20250101.StrategyWLAN, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds := obj.As(ctx, m, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	return v20250101.StrategyWLAN{
		NetworkAccessServerIP: m.NetworkAccessServerIP.ValueStringPointer(),
		Ssid:                  m.SSID.ValueString(),
		CaChain:               m.CAChain.ValueStringPointer(),
		Hidden:                m.Hidden.ValueBoolPointer(),
		Autojoin:              m.Autojoin.ValueBoolPointer(),
		ExternalRadiusServer:  m.ExternalRadiusServer.ValueBoolPointer(),
	}, diags
}

func toAPI(ctx context.Context, model *StrategyModel) (*v20250101.ProtectionStrategy, diag.Diagnostics) {
	var diags diag.Diagnostics

	strategy := &v20250101.ProtectionStrategy{
		Configuration: v20250101.ProtectionStrategy_Configuration{},
		Credential:    nil,
		Id:            model.ID.ValueString(),
		Name:          model.Name.ValueString(),
		Policy:        nil,
	}

	switch {
	case !model.Browser.IsNull():
		strategy.Kind = v20250101.Browser
		browser, ds := new(BrowserMutualTLSModel).toAPI(ctx, model.Browser)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyBrowserMutualTLS(browser); err != nil {
			diags.AddError("Account Browser Configuration Error", err.Error())
		}
	case !model.Ethernet.IsNull():
		strategy.Kind = v20250101.Ethernet
		lan, ds := new(LANModel).toAPI(ctx, model.Ethernet)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyLAN(lan); err != nil {
			diags.AddError("Account Ethernet Configuration Error", err.Error())
		}
	case !model.Relay.IsNull():
		strategy.Kind = v20250101.Relay
		relay, ds := new(NetworkRelayModel).toAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyNetworkRelay(relay); err != nil {
			diags.AddError("Account Relay Configuration Error", err.Error())
		}
	case !model.SSH.IsNull():
		strategy.Kind = v20250101.Ssh
		ssh, ds := new(SSHModel).toAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategySSH(ssh); err != nil {
			diags.AddError("Account SSH Configuration Error", err.Error())
		}
	case !model.SSO.IsNull():
		strategy.Kind = v20250101.Sso
		sso, ds := new(SSOModel).toAPI(ctx, model.Relay)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategySSO(sso); err != nil {
			diags.AddError("Account SSO Configuration Error", err.Error())
		}
	case !model.VPN.IsNull():
		strategy.Kind = v20250101.Vpn
		vpn, ds := new(VPNModel).toAPI(ctx, model.VPN)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyVPN(vpn); err != nil {
			diags.AddError("Account VPN Configuration Error", err.Error())
		}
	case !model.WiFi.IsNull():
		strategy.Kind = v20250101.Wifi
		wifi, ds := new(WLANModel).toAPI(ctx, model.WiFi)
		diags.Append(ds...)
		if err := strategy.Configuration.FromStrategyWLAN(wifi); err != nil {
			diags.AddError("Account WiFi Configuration Error", err.Error())
		}
	}

	return strategy, diags
}

func fromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (*StrategyModel, diag.Diagnostics) {
	browser := basetypes.NewObjectNull(browserMutualTLSAttributes)
	ethernet := basetypes.NewObjectNull(lanAttributes)
	relay := basetypes.NewObjectNull(networkRelayAttributes)
	ssh := basetypes.NewObjectNull(sshAttributes)
	sso := basetypes.NewObjectNull(ssoAttributes)
	vpn := basetypes.NewObjectNull(vpnAttributes)
	wifi := basetypes.NewObjectNull(wlanAttributes)

	var diags, ds diag.Diagnostics
	switch strategy.Kind {
	case v20250101.Browser:
		browser, ds = browserObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Ethernet:
		ethernet, ds = ethernetObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Relay:
		relay, ds = relayObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Ssh:
		ssh, ds = sshObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Sso:
		sso, ds = ssoObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Vpn:
		vpn, ds = vpnObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	case v20250101.Wifi:
		wifi, ds = wifiObjectFromAPI(ctx, strategy, state)
		diags.Append(ds...)
	default:
		diags.AddError("Unsupported Strategy Kind", string(strategy.Kind))
	}

	return &StrategyModel{
		ID:       types.StringValue(strategy.Id),
		Name:     types.StringValue(strategy.Name),
		Browser:  browser,
		Ethernet: ethernet,
		Relay:    relay,
		SSH:      ssh,
		SSO:      sso,
		VPN:      vpn,
		WiFi:     wifi,
	}, diags
}

func browserObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategyBrowserMutualTLS()
	if err != nil {
		diags.AddError("Strategy Browser Parse Error", err.Error())
		return types.Object{}, diags
	}

	matchAddresses, ds := utils.ToOptionalList(ctx, &conf.MatchAddresses, state, path.Root("browser").AtName("match_addresses"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(browserMutualTLSAttributes, map[string]attr.Value{
		"match_addresses": matchAddresses,
	})
	diags.Append(ds...)

	return obj, diags
}

func ethernetObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategyLAN()
	if err != nil {
		diags.AddError("Strategy Ethernet Parse Error", err.Error())
		return types.Object{}, diags
	}

	networkAccessServerIP, ds := utils.ToOptionalString(ctx, conf.NetworkAccessServerIP, state, path.Root("ethernet").AtName("network_access_server_ip"))
	diags.Append(ds...)

	caChain, ds := utils.ToOptionalString(ctx, conf.CaChain, state, path.Root("ethernet").AtName("ca_chain"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, conf.Autojoin, state, path.Root("ethernet").AtName("autojoin"))
	diags.Append(ds...)

	externalRadiusServer, ds := utils.ToOptionalBool(ctx, conf.ExternalRadiusServer, state, path.Root("ethernet").AtName("external_radius_server"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(lanAttributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIP,
		"ca_chain":                 caChain,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
	})
	diags.Append(ds...)

	return obj, diags
}

func relayObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategyNetworkRelay()
	if err != nil {
		diags.AddError("Strategy Network Relay Parse Error", err.Error())
		return types.Object{}, diags
	}

	p := path.Root("relay")

	matchDomains, ds := utils.ToOptionalList(ctx, &conf.MatchDomains, state, p.AtName("match_domains"))
	diags.Append(ds...)

	regions, ds := utils.ToOptionalList(ctx, &conf.Regions, state, p.AtName("regions"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(networkRelayAttributes, map[string]attr.Value{
		"match_domains": matchDomains,
		"regions":       regions,
	})
	diags.Append(ds...)

	return obj, diags
}

func sshObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	_, err := strategy.Configuration.AsStrategySSH()
	if err != nil {
		diags.AddError("Strategy SSH Parse Error", err.Error())
		return types.Object{}, diags
	}

	obj, ds := basetypes.NewObjectValue(sshAttributes, map[string]attr.Value{})
	diags.Append(ds...)

	return obj, diags
}

func ssoObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategySSO()
	if err != nil {
		diags.AddError("Strategy SSO Parse Error", err.Error())
		return types.Object{}, diags
	}

	p := path.Root("sso")

	trustedRoots, ds := utils.ToOptionalString(ctx, &conf.TrustRoots, state, p.AtName("trusted_roots"))
	diags.Append(ds...)

	redirectUri, ds := utils.ToOptionalString(ctx, &conf.RedirectUri, state, p.AtName("redirect_uri"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ssoAttributes, map[string]attr.Value{
		"trusted_roots": trustedRoots,
		"redirect_uri":  redirectUri,
	})
	diags.Append(ds...)

	return obj, diags
}

func vpnObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategyVPN()
	if err != nil {
		diags.AddError("Strategy VPN Parse Error", err.Error())
		return types.Object{}, diags
	}

	p := path.Root("vpn")

	connectionType, ds := utils.ToOptionalString(ctx, &conf.ConnectionType, state, p.AtName("connection_type"))
	diags.Append(ds...)

	vendor, ds := utils.ToOptionalString(ctx, conf.Vendor, state, p.AtName("vendor"))
	diags.Append(ds...)

	remoteAddress, ds := utils.ToOptionalString(ctx, &conf.RemoteAddress, state, p.AtName("remote_address"))
	diags.Append(ds...)

	ike, ds := ikeObjectFromAPI(ctx, conf.Ike, state)
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, conf.Autojoin, state, p.AtName("connection_type"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(vpnAttributes, map[string]attr.Value{
		"connection_type": connectionType,
		"vendor":          vendor,
		"remote_address":  remoteAddress,
		"ike":             ike,
		"autojoin":        autojoin,
	})
	diags.Append(ds...)

	return obj, diags
}

func ikeObjectFromAPI(ctx context.Context, ike *v20250101.IkeV2Config, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ike == nil {
		return basetypes.NewObjectNull(ikeAttributes), diags
	}

	p := path.Root("vpn").AtName("ike")

	caChain, ds := utils.ToOptionalString(ctx, ike.CaChain, state, p.AtName("ca_chain"))
	diags.Append(ds...)

	remoteID, ds := utils.ToOptionalString(ctx, ike.RemoteID, state, p.AtName("remote_id"))
	diags.Append(ds...)

	eap, ds := utils.ToOptionalBool(ctx, ike.Eap, state, p.AtName("eap"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ikeAttributes, map[string]attr.Value{
		"ca_chain":  caChain,
		"remote_id": remoteID,
		"eap":       eap,
	})
	diags.Append(ds...)

	return obj, diags
}

func wifiObjectFromAPI(ctx context.Context, strategy *v20250101.ProtectionStrategy, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	conf, err := strategy.Configuration.AsStrategyWLAN()
	if err != nil {
		diags.AddError("Strategy WLAN Parse Error", err.Error())
		return types.Object{}, diags
	}

	p := path.Root("wifi")

	networkAccessServerIp, ds := utils.ToOptionalString(ctx, conf.NetworkAccessServerIP, state, p.AtName("network_access_server_ip"))
	diags.Append(ds...)

	ssid, ds := utils.ToOptionalString(ctx, &conf.Ssid, state, p.AtName("ssid"))
	diags.Append(ds...)

	caChain, ds := utils.ToOptionalString(ctx, conf.CaChain, state, p.AtName("ca_chain"))
	diags.Append(ds...)

	hidden, ds := utils.ToOptionalBool(ctx, conf.Hidden, state, p.AtName("hidden"))
	diags.Append(ds...)

	autojoin, ds := utils.ToOptionalBool(ctx, conf.Autojoin, state, p.AtName("autojoin"))
	diags.Append(ds...)

	externalRadiusServer, ds := utils.ToOptionalBool(ctx, conf.ExternalRadiusServer, state, p.AtName("external_radius_server"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(wlanAttributes, map[string]attr.Value{
		"network_access_server_ip": networkAccessServerIp,
		"ssid":                     ssid,
		"ca_chain":                 caChain,
		"hidden":                   hidden,
		"autojoin":                 autojoin,
		"external_radius_server":   externalRadiusServer,
	})
	diags.Append(ds...)

	return obj, diags
}

func toList[T ~string](list types.List) []T {
	elements := list.Elements()
	ret := make([]T, len(elements))
	for i, v := range elements {
		ret[i] = T(v.String())
	}
	return ret
}
