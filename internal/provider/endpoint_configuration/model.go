package endpoint_configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Hooks           *HooksModel           `tfsdk:"hooks"`
	KeyInfo         *KeyInfoModel         `tfsdk:"key_info"`
	ReloadInfo      *ReloadInfoModel      `tfsdk:"reload_info"`
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

func (h *HookModel) toAPI(ctx context.Context) (*v20230301.EndpointHook, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	var before *[]string
	var after *[]string
	var onError *[]string

	diags.Append(h.Before.ElementsAs(ctx, &before, false)...)
	diags.Append(h.After.ElementsAs(ctx, &after, false)...)
	diags.Append(h.OnError.ElementsAs(ctx, &onError, false)...)

	return &v20230301.EndpointHook{
		Shell:   h.Shell.ValueStringPointer(),
		Before:  before,
		After:   after,
		OnError: onError,
	}, diags
}

type HooksModel struct {
	Sign  *HookModel `tfsdk:"sign"`
	Renew *HookModel `tfsdk:"renew"`
}

func (h *HooksModel) toAPI(ctx context.Context) (*v20230301.EndpointHooks, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h == nil {
		return nil, diags
	}

	sign, d := h.Sign.toAPI(ctx)
	diags.Append(d...)

	renew, d := h.Renew.toAPI(ctx)
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
		diags = append(diags, d...)

		renew, d := hookFromAPI(ctx, ec.Hooks.Renew, path.Root("hooks").AtName("renew"), state)
		diags = append(diags, d...)

		model.Hooks = &HooksModel{
			Sign:  sign,
			Renew: renew,
		}
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
		diags = append(diags, d...)

		signal, d := utils.ToOptionalInt(ctx, ec.ReloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
		diags = append(diags, d...)

		model.ReloadInfo = &ReloadInfoModel{
			Method:  types.StringValue(string(ec.ReloadInfo.Method)),
			PIDFile: pidFile,
			Signal:  signal,
		}
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.EndpointConfiguration, diag.Diagnostics) {
	hooks, diags := model.Hooks.toAPI(ctx)

	return &v20230301.EndpointConfiguration{
		Name:            model.Name.ValueString(),
		Kind:            v20230301.EndpointConfigurationKind(model.Kind.ValueString()),
		AuthorityID:     model.AuthorityID.ValueString(),
		Provisioner:     model.Provisioner.ValueString(),
		CertificateInfo: model.CertificateInfo.toAPI(),
		Hooks:           hooks,
		KeyInfo:         model.KeyInfo.toAPI(),
		ReloadInfo:      model.ReloadInfo.toAPI(),
	}, diags
}

func hookFromAPI(ctx context.Context, hook *v20230301.EndpointHook, hookPath path.Path, state utils.AttributeGetter) (*HookModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if hook == nil {
		return nil, diags
	}

	shell, d := utils.ToOptionalString(ctx, hook.Shell, state, hookPath.AtName("shell"))
	diags = append(diags, d...)

	before, d := utils.ToOptionalList(ctx, hook.Before, state, hookPath.AtName("before"))
	diags = append(diags, d...)

	after, d := utils.ToOptionalList(ctx, hook.After, state, hookPath.AtName("after"))
	diags = append(diags, d...)

	onError, d := utils.ToOptionalList(ctx, hook.OnError, state, hookPath.AtName("on_error"))
	diags = append(diags, d...)

	hm := &HookModel{
		Shell:   shell,
		Before:  before,
		After:   after,
		OnError: onError,
	}
	return hm, diags
}
