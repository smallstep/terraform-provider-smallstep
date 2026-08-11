package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
)

const typeName = "smallstep_workload"

var probeAttrTypes = map[string]attr.Type{
	"target":      types.StringType,
	"protocol":    types.StringType,
	"crt_file":    types.StringType,
	"key_file":    types.StringType,
	"root_file":   types.StringType,
	"server_name": types.StringType,
}

var credentialAttrTypes = map[string]attr.Type{
	"credential_id": types.StringType,
	"probes": types.ListType{
		ElemType: types.ObjectType{AttrTypes: probeAttrTypes},
	},
}

var hookAttrTypes = map[string]attr.Type{
	"before":   types.ListType{ElemType: types.StringType},
	"after":    types.ListType{ElemType: types.StringType},
	"on_error": types.ListType{ElemType: types.StringType},
	"shell":    types.StringType,
}

var hooksAttrTypes = map[string]attr.Type{
	"sign":  types.ObjectType{AttrTypes: hookAttrTypes},
	"renew": types.ObjectType{AttrTypes: hookAttrTypes},
}

var reloadInfoAttrTypes = map[string]attr.Type{
	"method":    types.StringType,
	"pid_file":  types.StringType,
	"signal":    types.Int64Type,
	"unit_name": types.StringType,
}

type ProbeModel struct {
	Target     types.String `tfsdk:"target"`
	Protocol   types.String `tfsdk:"protocol"`
	CrtFile    types.String `tfsdk:"crt_file"`
	KeyFile    types.String `tfsdk:"key_file"`
	RootFile   types.String `tfsdk:"root_file"`
	ServerName types.String `tfsdk:"server_name"`
}

type CredentialModel struct {
	CredentialID types.String `tfsdk:"credential_id"`
	Probes       types.List   `tfsdk:"probes"`
}

type HookModel struct {
	Before  types.List   `tfsdk:"before"`
	After   types.List   `tfsdk:"after"`
	OnError types.List   `tfsdk:"on_error"`
	Shell   types.String `tfsdk:"shell"`
}

type HooksModel struct {
	Sign  types.Object `tfsdk:"sign"`
	Renew types.Object `tfsdk:"renew"`
}

type ReloadInfoModel struct {
	Method   types.String `tfsdk:"method"`
	PidFile  types.String `tfsdk:"pid_file"`
	Signal   types.Int64  `tfsdk:"signal"`
	UnitName types.String `tfsdk:"unit_name"`
}

type Model struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	WorkloadType types.String `tfsdk:"workload_type"`
	Credentials  types.List   `tfsdk:"credentials"`
	Hooks        types.Object `tfsdk:"hooks"`
	ReloadInfo   types.Object `tfsdk:"reload_info"`
}

func fromAPI(ctx context.Context, workload *v20260501.Workload) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	credentialModels := make([]CredentialModel, len(workload.Credentials))
	for i, cred := range workload.Credentials {
		var probes types.List
		if cred.Probes != nil && len(*cred.Probes) > 0 {
			probeModels := make([]ProbeModel, len(*cred.Probes))
			for j, p := range *cred.Probes {
				probeModels[j] = ProbeModel{
					Target:     types.StringValue(p.Target),
					Protocol:   types.StringValue(string(p.Protocol)),
					CrtFile:    types.StringPointerValue(p.CrtFile),
					KeyFile:    types.StringPointerValue(p.KeyFile),
					RootFile:   types.StringPointerValue(p.RootFile),
					ServerName: types.StringPointerValue(p.ServerName),
				}
			}
			probesVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: probeAttrTypes}, probeModels)
			diags.Append(d...)
			probes = probesVal
		} else {
			probes = types.ListNull(types.ObjectType{AttrTypes: probeAttrTypes})
		}

		credentialModels[i] = CredentialModel{
			CredentialID: types.StringValue(cred.CredentialId),
			Probes:       probes,
		}
	}

	credentials, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: credentialAttrTypes}, credentialModels)
	diags.Append(d...)

	var hooks types.Object
	if workload.Hooks != nil {
		var sign, renew types.Object

		if workload.Hooks.Sign != nil {
			signModel := HookModel{
				Before:  hookSliceToList(ctx, workload.Hooks.Sign.Before, &diags),
				After:   hookSliceToList(ctx, workload.Hooks.Sign.After, &diags),
				OnError: hookSliceToList(ctx, workload.Hooks.Sign.OnError, &diags),
				Shell:   types.StringPointerValue(workload.Hooks.Sign.Shell),
			}
			signVal, d := types.ObjectValueFrom(ctx, hookAttrTypes, signModel)
			diags.Append(d...)
			sign = signVal
		} else {
			sign = types.ObjectNull(hookAttrTypes)
		}

		if workload.Hooks.Renew != nil {
			renewModel := HookModel{
				Before:  hookSliceToList(ctx, workload.Hooks.Renew.Before, &diags),
				After:   hookSliceToList(ctx, workload.Hooks.Renew.After, &diags),
				OnError: hookSliceToList(ctx, workload.Hooks.Renew.OnError, &diags),
				Shell:   types.StringPointerValue(workload.Hooks.Renew.Shell),
			}
			renewVal, d := types.ObjectValueFrom(ctx, hookAttrTypes, renewModel)
			diags.Append(d...)
			renew = renewVal
		} else {
			renew = types.ObjectNull(hookAttrTypes)
		}

		hooksModel := HooksModel{Sign: sign, Renew: renew}
		hooksVal, d := types.ObjectValueFrom(ctx, hooksAttrTypes, hooksModel)
		diags.Append(d...)
		hooks = hooksVal
	} else {
		hooks = types.ObjectNull(hooksAttrTypes)
	}

	var reloadInfo types.Object
	if workload.ReloadInfo != nil {
		var signal types.Int64
		if workload.ReloadInfo.Signal != nil {
			signal = types.Int64Value(int64(*workload.ReloadInfo.Signal))
		} else {
			signal = types.Int64Null()
		}

		reloadModel := ReloadInfoModel{
			Method:   types.StringValue(string(workload.ReloadInfo.Method)),
			PidFile:  types.StringPointerValue(workload.ReloadInfo.PidFile),
			Signal:   signal,
			UnitName: types.StringPointerValue(workload.ReloadInfo.UnitName),
		}
		reloadVal, d := types.ObjectValueFrom(ctx, reloadInfoAttrTypes, reloadModel)
		diags.Append(d...)
		reloadInfo = reloadVal
	} else {
		reloadInfo = types.ObjectNull(reloadInfoAttrTypes)
	}

	return &Model{
		ID:           types.StringPointerValue(workload.Id),
		Name:         types.StringPointerValue(workload.Name),
		WorkloadType: types.StringPointerValue(workload.WorkloadType),
		Credentials:  credentials,
		Hooks:        hooks,
		ReloadInfo:   reloadInfo,
	}, diags
}

func hookSliceToList(ctx context.Context, strs *[]string, diags *diag.Diagnostics) types.List {
	if strs == nil || len(*strs) == 0 {
		return types.ListNull(types.StringType)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, *strs)
	diags.Append(d...)
	return list
}

func (m *Model) toAPI(ctx context.Context) (*v20260501.Workload, diag.Diagnostics) {
	var diags diag.Diagnostics

	var credentialModels []CredentialModel
	d := m.Credentials.ElementsAs(ctx, &credentialModels, false)
	diags.Append(d...)

	credentials := make([]v20260501.WorkloadCredential, len(credentialModels))
	for i, cm := range credentialModels {
		var probes *[]v20260501.Probe
		if !cm.Probes.IsNull() && !cm.Probes.IsUnknown() {
			var probeModels []ProbeModel
			d := cm.Probes.ElementsAs(ctx, &probeModels, false)
			diags.Append(d...)

			probeSlice := make([]v20260501.Probe, len(probeModels))
			for j, pm := range probeModels {
				probeSlice[j] = v20260501.Probe{
					Target:     pm.Target.ValueString(),
					Protocol:   v20260501.ProbeProtocol(pm.Protocol.ValueString()),
					CrtFile:    pm.CrtFile.ValueStringPointer(),
					KeyFile:    pm.KeyFile.ValueStringPointer(),
					RootFile:   pm.RootFile.ValueStringPointer(),
					ServerName: pm.ServerName.ValueStringPointer(),
				}
			}
			probes = &probeSlice
		}

		credentials[i] = v20260501.WorkloadCredential{
			CredentialId: cm.CredentialID.ValueString(),
			Probes:       probes,
		}
	}

	var hooks *v20260501.EndpointHooks
	if !m.Hooks.IsNull() && !m.Hooks.IsUnknown() {
		var hooksModel HooksModel
		d := m.Hooks.As(ctx, &hooksModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)

		hooks = &v20260501.EndpointHooks{}

		if !hooksModel.Sign.IsNull() && !hooksModel.Sign.IsUnknown() {
			var signModel HookModel
			d := hooksModel.Sign.As(ctx, &signModel, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			hooks.Sign = hookModelToAPI(ctx, &signModel, &diags)
		}

		if !hooksModel.Renew.IsNull() && !hooksModel.Renew.IsUnknown() {
			var renewModel HookModel
			d := hooksModel.Renew.As(ctx, &renewModel, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			hooks.Renew = hookModelToAPI(ctx, &renewModel, &diags)
		}
	}

	var reloadInfo *v20260501.EndpointReloadInfo
	if !m.ReloadInfo.IsNull() && !m.ReloadInfo.IsUnknown() {
		var reloadModel ReloadInfoModel
		d := m.ReloadInfo.As(ctx, &reloadModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)

		var signal *int
		if !reloadModel.Signal.IsNull() && !reloadModel.Signal.IsUnknown() {
			sigVal := int(reloadModel.Signal.ValueInt64())
			signal = &sigVal
		}

		reloadInfo = &v20260501.EndpointReloadInfo{
			Method:   v20260501.EndpointReloadInfoMethod(reloadModel.Method.ValueString()),
			PidFile:  reloadModel.PidFile.ValueStringPointer(),
			Signal:   signal,
			UnitName: reloadModel.UnitName.ValueStringPointer(),
		}
	}

	var id, name, workloadType *string
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id = new(m.ID.ValueString())
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		name = new(m.Name.ValueString())
	}
	if !m.WorkloadType.IsNull() && !m.WorkloadType.IsUnknown() {
		workloadType = new(m.WorkloadType.ValueString())
	}

	return &v20260501.Workload{
		Id:           id,
		Name:         name,
		WorkloadType: workloadType,
		Credentials:  credentials,
		Hooks:        hooks,
		ReloadInfo:   reloadInfo,
	}, diags
}

func hookModelToAPI(ctx context.Context, model *HookModel, diags *diag.Diagnostics) *v20260501.EndpointHook {
	hook := &v20260501.EndpointHook{}

	if !model.Before.IsNull() && !model.Before.IsUnknown() {
		var before []string
		d := model.Before.ElementsAs(ctx, &before, false)
		diags.Append(d...)
		hook.Before = &before
	}

	if !model.After.IsNull() && !model.After.IsUnknown() {
		var after []string
		d := model.After.ElementsAs(ctx, &after, false)
		diags.Append(d...)
		hook.After = &after
	}

	if !model.OnError.IsNull() && !model.OnError.IsUnknown() {
		var onError []string
		d := model.OnError.ElementsAs(ctx, &onError, false)
		diags.Append(d...)
		hook.OnError = &onError
	}

	if !model.Shell.IsNull() && !model.Shell.IsUnknown() {
		hook.Shell = new(model.Shell.ValueString())
	}

	return hook
}
