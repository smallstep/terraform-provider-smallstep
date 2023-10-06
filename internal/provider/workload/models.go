package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type CertificateInfoModel struct {
	Type     types.String `tfsdk:"type"`
	CrtFile  types.String `tfsdk:"crt_file"`
	KeyFile  types.String `tfsdk:"key_file"`
	RootFile types.String `tfsdk:"root_file"`
	Duration types.String `tfsdk:"duration"`
	GID      types.Int64  `tfsdk:"gid"`
	UID      types.Int64  `tfsdk:"uid"`
	Mode     types.Int64  `tfsdk:"mode"`
}

func (ci CertificateInfoModel) ToAPI() v20230301.EndpointCertificateInfo {
	d := ci.Duration.ValueStringPointer()
	// duration defaults to 24h if not set, which means the schema must
	// mark it as both computed and optional. With the computed flag it
	// will be uknown rather than null if not set. The ValueStringPointer()
	// method returns the empty string for unknown values, but passing an
	// empty string to the API results in a 400 since it's not a valid
	// duration.
	if ci.Duration.IsUnknown() {
		d = nil
	}
	return v20230301.EndpointCertificateInfo{
		Type:     v20230301.EndpointCertificateInfoType(ci.Type.ValueString()),
		Duration: d,
		CrtFile:  ci.CrtFile.ValueStringPointer(),
		KeyFile:  ci.KeyFile.ValueStringPointer(),
		RootFile: ci.RootFile.ValueStringPointer(),
		Uid:      utils.ToIntPointer(ci.UID.ValueInt64Pointer()),
		Gid:      utils.ToIntPointer(ci.GID.ValueInt64Pointer()),
		Mode:     utils.ToIntPointer(ci.Mode.ValueInt64Pointer()),
	}
}

type HookModel struct {
	Shell   types.String `tfsdk:"shell"`
	Before  types.List   `tfsdk:"before"`
	After   types.List   `tfsdk:"after"`
	OnError types.List   `tfsdk:"on_error"`
}

var HookObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"shell": types.StringType,
		"before": types.ListType{
			ElemType: types.StringType,
		},
		"after": types.ListType{
			ElemType: types.StringType,
		},
		"on_error": types.ListType{
			ElemType: types.StringType,
		},
	},
}

func (h *HookModel) ToAPI(ctx context.Context) (*v20230301.EndpointHook, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	var before *[]string
	var after *[]string
	var onError *[]string

	if !h.Before.IsNull() {
		diags.Append(h.Before.ElementsAs(ctx, &before, false)...)
	}
	if !h.After.IsNull() {
		diags.Append(h.After.ElementsAs(ctx, &after, false)...)
	}
	if !h.OnError.IsNull() {
		diags.Append(h.OnError.ElementsAs(ctx, &onError, false)...)
	}

	return &v20230301.EndpointHook{
		Shell:   h.Shell.ValueStringPointer(),
		Before:  before,
		After:   after,
		OnError: onError,
	}, diags
}

type HooksModel struct {
	Sign  types.Object `tfsdk:"sign"`
	Renew types.Object `tfsdk:"renew"`
}

var HooksObjectType = map[string]attr.Type{
	"sign":  HookObjectType,
	"renew": HookObjectType,
}

func HookToAPI(ctx context.Context, hook types.Object) (*v20230301.EndpointHook, diag.Diagnostics) {
	hookModel := &HookModel{}
	diags := hook.As(ctx, &hookModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})

	h, d := hookModel.ToAPI(ctx)
	diags.Append(d...)

	return h, diags
}

func (h *HooksModel) ToAPI(ctx context.Context) (*v20230301.EndpointHooks, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	sign, d := HookToAPI(ctx, h.Sign)
	diags.Append(d...)

	renew, d := HookToAPI(ctx, h.Renew)
	diags.Append(d...)

	return &v20230301.EndpointHooks{
		Sign:  sign,
		Renew: renew,
	}, diags
}

type KeyInfoModel struct {
	Format  types.String `tfsdk:"format"`
	PubFile types.String `tfsdk:"pub_file"`
	Type    types.String `tfsdk:"type"`
}

func (ki *KeyInfoModel) ToAPI() *v20230301.EndpointKeyInfo {
	if ki == nil {
		return nil
	}

	return &v20230301.EndpointKeyInfo{
		Format:  utils.ToStringPointer[string, v20230301.EndpointKeyInfoFormat](ki.Format.ValueStringPointer()),
		PubFile: ki.PubFile.ValueStringPointer(),
		Type:    utils.ToStringPointer[string, v20230301.EndpointKeyInfoType](ki.Type.ValueStringPointer()),
	}
}

type ReloadInfoModel struct {
	Method   types.String `tfsdk:"method"`
	PIDFile  types.String `tfsdk:"pid_file"`
	Signal   types.Int64  `tfsdk:"signal"`
	UnitName types.String `tfsdk:"unit_name"`
}

var ReloadInfoType = map[string]attr.Type{
	"method":    types.StringType,
	"pid_file":  types.StringType,
	"signal":    types.Int64Type,
	"unit_name": types.StringType,
}

func (ri *ReloadInfoModel) ToAPI() *v20230301.EndpointReloadInfo {
	if ri == nil {
		return nil
	}

	return &v20230301.EndpointReloadInfo{
		Method:   v20230301.EndpointReloadInfoMethod(ri.Method.ValueString()),
		PidFile:  ri.PIDFile.ValueStringPointer(),
		Signal:   utils.ToIntPointer(ri.Signal.ValueInt64Pointer()),
		UnitName: ri.UnitName.ValueStringPointer(),
	}
}

func HookFromAPI(ctx context.Context, hook *v20230301.EndpointHook, hookPath path.Path, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if hook == nil {
		return basetypes.NewObjectNull(HookObjectType.AttrTypes), diags
	}

	shell, d := utils.ToOptionalString(ctx, hook.Shell, state, hookPath.AtName("shell"))
	diags = append(diags, d...)

	before, d := utils.ToOptionalList(ctx, hook.Before, state, hookPath.AtName("before"))
	diags = append(diags, d...)

	after, d := utils.ToOptionalList(ctx, hook.After, state, hookPath.AtName("after"))
	diags = append(diags, d...)

	onError, d := utils.ToOptionalList(ctx, hook.OnError, state, hookPath.AtName("on_error"))
	diags = append(diags, d...)

	obj, d := basetypes.NewObjectValue(HookObjectType.AttrTypes, map[string]attr.Value{
		"shell":    shell,
		"before":   before,
		"after":    after,
		"on_error": onError,
	})
	diags.Append(d...)

	return obj, diags
}
