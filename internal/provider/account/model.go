package account

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

const accountTypeName = "smallstep_account"

type Model struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	WiFi     types.Object `tfsdk:"wifi"`
	VPN      types.Object `tfsdk:"vpn"`
	Browser  types.Object `tfsdk:"browser"`
	Ethernet types.Object `tfsdk:"ethernet"`
}

type WiFiModel struct {
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	CAChain               types.String `tfsdk:"ca_chain"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	Hidden                types.Bool   `tfsdk:"hidden"`
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
	SSID                  types.String `tfsdk:"ssid"`
}

var wifiAttrTypes = map[string]attr.Type{
	"autojoin":                 types.BoolType,
	"ca_chain":                 types.StringType,
	"external_radius_server":   types.BoolType,
	"hidden":                   types.BoolType,
	"network_access_server_ip": types.StringType,
	"ssid":                     types.StringType,
}

type BrowserModel struct{}

var browserAttrTypes = map[string]attr.Type{}

type VPNModel struct {
	Autojoin       types.Bool   `tfsdk:"autojoin"`
	ConnectionType types.String `tfsdk:"connection_type"`
	IKE            types.Object `tfsdk:"ike"`
	RemoteAddress  types.String `tfsdk:"remote_address"`
	Vendor         types.String `tfsdk:"vendor"`
}

var ikeAttrTypes = map[string]attr.Type{
	"ca_chain":  types.StringType,
	"eap":       types.BoolType,
	"remote_id": types.StringType,
}

var vpnAttrTypes = map[string]attr.Type{
	"autojoin":        types.BoolType,
	"connection_type": types.StringType,
	"remote_address":  types.StringType,
	"vendor":          types.StringType,
	"ike": types.ObjectType{
		AttrTypes: ikeAttrTypes,
	},
}

type IKEModel struct {
	CAChain  types.String `tfsdk:"ca_chain"`
	EAP      types.Bool   `tfsdk:"eap"`
	RemoteID types.String `tfsdk:"remote_id"`
}

type EthernetModel struct {
	Autojoin              types.Bool   `tfsdk:"autojoin"`
	CAChain               types.String `tfsdk:"ca_chain"`
	ExternalRadiusServer  types.Bool   `tfsdk:"external_radius_server"`
	NetworkAccessServerIP types.String `tfsdk:"network_access_server_ip"`
}

var ethernetAttrTypes = map[string]attr.Type{
	"autojoin":                 types.BoolType,
	"ca_chain":                 types.StringType,
	"external_radius_server":   types.BoolType,
	"network_access_server_ip": types.StringType,
}

func toAPI(ctx context.Context, model *Model) (*v20231101.Account, diag.Diagnostics) {
	var diags diag.Diagnostics

	account := &v20231101.Account{
		Id:   model.ID.ValueStringPointer(),
		Name: model.Name.ValueString(),
		Type: v20231101.AccountType(model.Name.ValueString()),
	}

	// TODO(areed) validate no more than one is set
	switch {
	case !model.WiFi.IsNull():
		account.Type = v20231101.Wifi

		wifi := &WiFiModel{}
		ds := model.WiFi.As(ctx, wifi, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.FromWifiAccount(v20231101.WifiAccount{
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
		account.Type = v20231101.Vpn

		vpn := &VPNModel{}
		ds := model.VPN.As(ctx, vpn, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		vpnAccount := v20231101.VpnAccount{
			Autojoin:       vpn.Autojoin.ValueBoolPointer(),
			ConnectionType: v20231101.VpnAccountConnectionType(vpn.ConnectionType.ValueString()),
			RemoteAddress:  vpn.RemoteAddress.ValueString(),
			Vendor:         utils.ToStringPointer[v20231101.VpnAccountVendor](vpn.Vendor.ValueStringPointer()),
		}

		if !vpn.IKE.IsNull() && !vpn.IKE.IsUnknown() {
			ike := &IKEModel{}
			ds := vpn.IKE.As(ctx, ike, basetypes.ObjectAsOptions{})
			diags.Append(ds...)
			vpnAccount.Ike = &v20231101.IkeV2Config{
				CaChain:  ike.CAChain.ValueStringPointer(),
				Eap:      ike.EAP.ValueBoolPointer(),
				RemoteID: ike.RemoteID.ValueStringPointer(),
			}
		}

		err := account.FromVpnAccount(vpnAccount)
		if err != nil {
			diags.AddError("Account VPN Configuration Error", err.Error())
		}

	case !model.Browser.IsNull():
		account.Type = v20231101.Browser

		browser := &BrowserModel{}
		ds := model.Browser.As(ctx, browser, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.FromBrowserAccount(v20231101.BrowserAccount{})
		if err != nil {
			diags.AddError("Account Browser Configuration Error", err.Error())
		}

	case !model.Ethernet.IsNull():
		account.Type = v20231101.Ethernet

		ethernet := &EthernetModel{}
		ds := model.Ethernet.As(ctx, ethernet, basetypes.ObjectAsOptions{})
		diags.Append(ds...)

		err := account.FromEthernetAccount(v20231101.EthernetAccount{
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

func fromAPI(ctx context.Context, account *v20231101.Account, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, ds := utils.ToOptionalString(ctx, account.Id, state, path.Root("id"))
	diags.Append(ds...)

	model := &Model{
		ID:   id,
		Name: types.StringValue(account.Name),
	}

	switch account.Type {
	case v20231101.Wifi:
		wifi, err := account.AsWifiAccount()
		if err != nil {
			diags.AddError("Account Wifi Parse Error", err.Error())
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

		obj, ds := basetypes.NewObjectValue(wifiAttrTypes, map[string]attr.Value{
			"autojoin":                 autojoin,
			"ca_chain":                 caChain,
			"external_radius_server":   externalRadiusServer,
			"hidden":                   hidden,
			"network_access_server_ip": nasIP,
			"ssid":                     types.StringValue(wifi.Ssid),
		})
		diags.Append(ds...)

		model.WiFi = obj
		model.VPN = basetypes.NewObjectNull(vpnAttrTypes)
		model.Browser = basetypes.NewObjectNull(browserAttrTypes)
		model.Ethernet = basetypes.NewObjectNull(ethernetAttrTypes)

	case v20231101.Vpn:
		vpn, err := account.AsVpnAccount()
		if err != nil {
			diags.AddError("Account VPN Parse Error", err.Error())
			break
		}

		autojoin, ds := utils.ToOptionalBool(ctx, vpn.Autojoin, state, path.Root("vpn").AtName("autojoin"))
		diags.Append(ds...)

		vendor, ds := utils.ToOptionalString(ctx, vpn.Vendor, state, path.Root("vpn").AtName("vendor"))
		diags.Append(ds...)

		ike, ds := ikeFromAPI(ctx, vpn.Ike, path.Root("vpn").AtName("ike"), state)
		diags.Append(ds...)

		obj, ds := basetypes.NewObjectValue(vpnAttrTypes, map[string]attr.Value{
			"autojoin":        autojoin,
			"connection_type": types.StringValue(string(vpn.ConnectionType)),
			"remote_address":  types.StringValue(vpn.RemoteAddress),
			"vendor":          vendor,
			"ike":             ike,
		})
		diags.Append(ds...)

		model.VPN = obj
		model.WiFi = basetypes.NewObjectNull(wifiAttrTypes)
		model.Browser = basetypes.NewObjectNull(browserAttrTypes)
		model.Ethernet = basetypes.NewObjectNull(ethernetAttrTypes)

	case v20231101.Browser:
		_, err := account.AsBrowserAccount()
		if err != nil {
			diags.AddError("Account Browser Parse Error", err.Error())
			break
		}

		obj, ds := basetypes.NewObjectValue(browserAttrTypes, map[string]attr.Value{})
		diags.Append(ds...)

		model.Browser = obj
		model.WiFi = basetypes.NewObjectNull(wifiAttrTypes)
		model.VPN = basetypes.NewObjectNull(vpnAttrTypes)
		model.Ethernet = basetypes.NewObjectNull(ethernetAttrTypes)

	case v20231101.Ethernet:
		ethernet, err := account.AsEthernetAccount()
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

		obj, ds := basetypes.NewObjectValue(ethernetAttrTypes, map[string]attr.Value{
			"autojoin":                 autojoin,
			"ca_chain":                 caChain,
			"external_radius_server":   externalRadiusServer,
			"network_access_server_ip": nasIP,
		})
		diags.Append(ds...)

		model.Ethernet = obj
		model.WiFi = basetypes.NewObjectNull(wifiAttrTypes)
		model.VPN = basetypes.NewObjectNull(vpnAttrTypes)
		model.Browser = basetypes.NewObjectNull(browserAttrTypes)
	}

	return model, diags
}

func ikeFromAPI(ctx context.Context, ike *v20231101.IkeV2Config, ikePath path.Path, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ike == nil {
		return basetypes.NewObjectNull(ikeAttrTypes), diags
	}

	caChain, ds := utils.ToOptionalString(ctx, ike.CaChain, state, ikePath.AtName("ca_chain"))
	diags.Append(ds...)

	eap, ds := utils.ToOptionalBool(ctx, ike.Eap, state, ikePath.AtName("eap"))
	diags.Append(ds...)

	remoteID, ds := utils.ToOptionalString(ctx, ike.RemoteID, state, ikePath.AtName("remote_id"))
	diags.Append(ds...)

	obj, ds := basetypes.NewObjectValue(ikeAttrTypes, map[string]attr.Value{
		"ca_chain":  caChain,
		"eap":       eap,
		"remote_id": remoteID,
	})
	diags.Append(ds...)

	return obj, diags
}
