package managed_radius

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

const name = "smallstep_managed_radius"

type ManagedRadiusModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	NASIPs          types.List   `tfsdk:"nas_ips"`
	ClientCA        types.String `tfsdk:"client_ca"`
	ReplyAttributes types.List   `tfsdk:"reply_attributes"`

	ServerCA       types.String `tfsdk:"server_ca"`
	ServerIP       types.String `tfsdk:"server_ip"`
	ServerPort     types.String `tfsdk:"server_port"`
	ServerHostname types.String `tfsdk:"server_hostname"`
}

type ReplyAttributeModel struct {
	Name                 types.String `tfsdk:"name"`
	Value                types.String `tfsdk:"value"`
	ValueFromCertificate types.String `tfsdk:"value_from_certificate"`
}

var replyAttributeTypes = map[string]attr.Type{
	"name":                   types.StringType,
	"value":                  types.StringType,
	"value_from_certificate": types.StringType,
}

func (model *ManagedRadiusModel) ToAPI(ctx context.Context, diags *diag.Diagnostics) v20250101.ManagedRadius {
	var nasIPs []string
	ds := model.NASIPs.ElementsAs(ctx, &nasIPs, false)
	diags.Append(ds...)

	replyAttrModels := make([]ReplyAttributeModel, 0, len(model.ReplyAttributes.Elements()))
	ds = model.ReplyAttributes.ElementsAs(ctx, &replyAttrModels, false)
	diags.Append(ds...)

	replyAttrs := make([]v20250101.ReplyAttribute, len(replyAttrModels))
	for i, ra := range replyAttrModels {
		replyAttrs[i] = v20250101.ReplyAttribute{
			Name:                 ra.Name.ValueString(),
			Value:                ra.Value.ValueStringPointer(),
			ValueFromCertificate: ra.ValueFromCertificate.ValueStringPointer(),
		}
	}

	return v20250101.ManagedRadius{
		Id:              model.ID.ValueStringPointer(),
		Name:            model.Name.ValueString(),
		NasIPs:          nasIPs,
		ClientCA:        model.ClientCA.ValueString(),
		ReplyAttributes: &replyAttrs,
	}
}

func fromAPI(ctx context.Context, diags *diag.Diagnostics, radius *v20250101.ManagedRadius, state utils.AttributeGetter) ManagedRadiusModel {
	nasIPs, ds := types.ListValueFrom(ctx, types.StringType, radius.NasIPs)
	diags.Append(ds...)

	var replyAttributes types.List
	if radius.ReplyAttributes == nil {
		// Optional lists require special handling because the API returns nil for both null list
		// and lists with length 0 and terraform throws an error if we switch between them.
		listFromState := types.List{}
		ds := state.GetAttribute(ctx, path.Root("reply_attributes"), &listFromState)
		diags.Append(ds...)
		if listFromState.IsNull() || len(listFromState.Elements()) == 0 {
			replyAttributes = listFromState
		} else {
			replyAttributes, ds = types.ListValue(types.ObjectType{AttrTypes: replyAttributeTypes}, nil)
			diags.Append(ds...)
		}
	} else {
		replyAttributeValues := make([]attr.Value, len(utils.Deref(radius.ReplyAttributes)))
		for i, ra := range utils.Deref(radius.ReplyAttributes) {
			obj, ds := basetypes.NewObjectValue(replyAttributeTypes, map[string]attr.Value{
				"name":                   types.StringValue(ra.Name),
				"value":                  types.StringPointerValue(ra.Value),
				"value_from_certificate": types.StringPointerValue(ra.ValueFromCertificate),
			})
			diags.Append(ds...)
			replyAttributeValues[i] = obj
		}
		replyAttributes, ds = types.ListValue(types.ObjectType{AttrTypes: replyAttributeTypes}, replyAttributeValues)
		diags.Append(ds...)
	}

	return ManagedRadiusModel{
		ID:              types.StringPointerValue(radius.Id),
		Name:            types.StringValue(radius.Name),
		NASIPs:          nasIPs,
		ClientCA:        types.StringValue(radius.ClientCA),
		ReplyAttributes: replyAttributes,

		ServerCA:       types.StringPointerValue(radius.ServerCA),
		ServerHostname: types.StringPointerValue(radius.ServerHostname),
		ServerIP:       types.StringPointerValue(radius.ServerIP),
		ServerPort:     types.StringPointerValue(radius.ServerPort),
	}
}

func FromAPI(ctx context.Context, m ManagedRadiusModel) (types.Object, diag.Diagnostics) {
	return basetypes.NewObjectValueFrom(ctx, map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"nas_ips": types.ListType{
			ElemType: types.StringType,
		},
		"client_ca": types.StringType,
		"reply_attributes": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: replyAttributeTypes,
			},
		},
		"server_ca":       types.StringType,
		"server_ip":       types.StringType,
		"server_port":     types.StringType,
		"server_hostname": types.StringType,
	}, m)
}
