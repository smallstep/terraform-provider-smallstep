package endpoint_configuration

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

const typeName = "smallstep_endpoint_configuration"

type Model struct {
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	AuthorityID     types.String          `tfsdk:"authority_id"`
	Provisioner     types.String          `tfsdk:"provisioner_name"`
	Kind            types.String          `tfsdk:"kind"`
	CertificateInfo *CertificateInfoModel `tfsdk:"certificate_info"`
	KeyInfo         *KeyInfoModel         `tfsdk:"key_info"`
	ReloadInfo      types.Object          `tfsdk:"reload_info"`
	Hooks           types.Object          `tfsdk:"hooks"`
}

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

func (ci CertificateInfoModel) toAPI() v20230301.EndpointCertificateInfo {
	return v20230301.EndpointCertificateInfo{
		Type:     v20230301.EndpointCertificateInfoType(ci.Type.ValueString()),
		Duration: ci.Duration.ValueStringPointer(),
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

var hookObjectType = types.ObjectType{
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

func (h *HookModel) toAPI(ctx context.Context) (*v20230301.EndpointHook, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	var before *[]string
	var after *[]string
	var onError *[]string

	// TODO do I still need double pointer with this null check?
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

var hooksObjectType = map[string]attr.Type{
	"sign":  hookObjectType,
	"renew": hookObjectType,
}

func hookToAPI(ctx context.Context, hook types.Object) (*v20230301.EndpointHook, diag.Diagnostics) {
	hookModel := &HookModel{}
	diags := hook.As(ctx, hookModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})

	h, d := hookModel.toAPI(ctx)
	diags.Append(d...)

	return h, diags
}

func (h *HooksModel) toAPI(ctx context.Context) (*v20230301.EndpointHooks, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	sign, d := hookToAPI(ctx, h.Sign)
	diags.Append(d...)

	renew, d := hookToAPI(ctx, h.Renew)
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

func (ki *KeyInfoModel) toAPI() *v20230301.EndpointKeyInfo {
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
	Method  types.String `tfsdk:"method"`
	PIDFile types.String `tfsdk:"pid_file"`
	Signal  types.Int64  `tfsdk:"signal"`
}

var reloadInfoType = map[string]attr.Type{
	"method":   types.StringType,
	"pid_file": types.StringType,
	"signal":   types.Int64Type,
}

func (ri *ReloadInfoModel) toAPI() *v20230301.EndpointReloadInfo {
	if ri == nil {
		return nil
	}

	return &v20230301.EndpointReloadInfo{
		Method:  v20230301.EndpointReloadInfoMethod(ri.Method.ValueString()),
		PidFile: ri.PIDFile.ValueStringPointer(),
		Signal:  utils.ToIntPointer(ri.Signal.ValueInt64Pointer()),
	}
}

func fromAPI(ctx context.Context, ec *v20230301.EndpointConfiguration, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	ciDuration, d := utils.ToEqualString(ctx, ec.CertificateInfo.Duration, state, path.Root("certificate_info").AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	ciCrtFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.CrtFile, state, path.Root("certificate_info").AtName("crt_file"))
	diags = append(diags, d...)

	ciKeyFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.KeyFile, state, path.Root("certificate_info").AtName("key_file"))
	diags = append(diags, d...)

	ciRootFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.RootFile, state, path.Root("certificate_info").AtName("root_file"))
	diags = append(diags, d...)

	ciUID, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Uid, state, path.Root("certificate_info").AtName("uid"))
	diags = append(diags, d...)

	ciGID, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Gid, state, path.Root("certificate_info").AtName("gid"))
	diags = append(diags, d...)

	ciMode, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Mode, state, path.Root("certificate_info").AtName("mode"))
	diags = append(diags, d...)

	model := &Model{
		ID:          types.StringValue(utils.Deref(ec.Id)),
		Name:        types.StringValue(ec.Name),
		Kind:        types.StringValue(string(ec.Kind)),
		AuthorityID: types.StringValue(ec.AuthorityID),
		Provisioner: types.StringValue(ec.Provisioner),
		CertificateInfo: &CertificateInfoModel{
			Type:     types.StringValue(string(ec.CertificateInfo.Type)),
			Duration: ciDuration,
			CrtFile:  ciCrtFile,
			KeyFile:  ciKeyFile,
			RootFile: ciRootFile,
			UID:      ciUID,
			GID:      ciGID,
			Mode:     ciMode,
		},
	}

	if ec.Hooks != nil {
		sign, d := hookFromAPI(ctx, ec.Hooks.Sign, path.Root("hooks").AtName("sign"), state)
		diags.Append(d...)

		renew, d := hookFromAPI(ctx, ec.Hooks.Renew, path.Root("hooks").AtName("renew"), state)
		diags.Append(d...)

		hooksObj, d := basetypes.NewObjectValue(hooksObjectType, map[string]attr.Value{
			"sign":  sign,
			"renew": renew,
		})
		diags.Append(d...)

		model.Hooks = hooksObj
	} else {
		model.Hooks = basetypes.NewObjectNull(hooksObjectType)
	}

	if ec.KeyInfo != nil {
		format, d := utils.ToOptionalString(ctx, ec.KeyInfo.Format, state, path.Root("key_info").AtName("format"))
		diags = append(diags, d...)

		pubFile, d := utils.ToOptionalString(ctx, ec.KeyInfo.PubFile, state, path.Root("key_info").AtName("pub_file"))
		diags = append(diags, d...)

		typ, d := utils.ToOptionalString(ctx, ec.KeyInfo.Type, state, path.Root("key_info").AtName("type"))
		diags = append(diags, d...)

		model.KeyInfo = &KeyInfoModel{
			Format:  format,
			PubFile: pubFile,
			Type:    typ,
		}
	}

	if ec.ReloadInfo != nil {
		pidFile, d := utils.ToOptionalString(ctx, ec.ReloadInfo.PidFile, state, path.Root("reload_info").AtName("method"))
		diags.Append(d...)

		signal, d := utils.ToOptionalInt(ctx, ec.ReloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
		diags.Append(d...)

		reloadInfoObject, d := basetypes.NewObjectValue(reloadInfoType, map[string]attr.Value{
			"method":   types.StringValue(string(ec.ReloadInfo.Method)),
			"pid_file": pidFile,
			"signal":   signal,
		})
		diags.Append(d...)
		model.ReloadInfo = reloadInfoObject
	} else {
		model.ReloadInfo = basetypes.NewObjectNull(reloadInfoType)
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.EndpointConfiguration, diag.Diagnostics) {
	hooksModel := &HooksModel{}
	diags := model.Hooks.As(ctx, &hooksModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	hooks, d := hooksModel.toAPI(ctx)
	diags.Append(d...)

	reloadInfo := &ReloadInfoModel{}
	d = model.ReloadInfo.As(ctx, &reloadInfo, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)

	return &v20230301.EndpointConfiguration{
		Name:            model.Name.ValueString(),
		Kind:            v20230301.EndpointConfigurationKind(model.Kind.ValueString()),
		AuthorityID:     model.AuthorityID.ValueString(),
		Provisioner:     model.Provisioner.ValueString(),
		CertificateInfo: model.CertificateInfo.toAPI(),
		KeyInfo:         model.KeyInfo.toAPI(),
		Hooks:           hooks,
		ReloadInfo:      reloadInfo.toAPI(),
	}, diags
}

func hookFromAPI(ctx context.Context, hook *v20230301.EndpointHook, hookPath path.Path, state utils.AttributeGetter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if hook == nil {
		return basetypes.NewObjectNull(hookObjectType.AttrTypes), diags
	}

	shell, d := utils.ToOptionalString(ctx, hook.Shell, state, hookPath.AtName("shell"))
	diags = append(diags, d...)

	before, d := utils.ToOptionalList(ctx, hook.Before, state, hookPath.AtName("before"))
	diags = append(diags, d...)

	after, d := utils.ToOptionalList(ctx, hook.After, state, hookPath.AtName("after"))
	diags = append(diags, d...)

	onError, d := utils.ToOptionalList(ctx, hook.OnError, state, hookPath.AtName("on_error"))
	diags = append(diags, d...)

	obj, d := basetypes.NewObjectValue(hookObjectType.AttrTypes, map[string]attr.Value{
		"shell":    shell,
		"before":   before,
		"after":    after,
		"on_error": onError,
	})
	diags.Append(d...)

	return obj, diags
}
