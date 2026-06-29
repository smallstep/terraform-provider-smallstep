package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_workload"

func convertHookToMap(ctx context.Context, hook *v20260501.EndpointHook, diags *diag.Diagnostics) map[string]attr.Value {
	hookObj := map[string]attr.Value{}
	beforeVals := convertStringsToValues(hook.Before)
	if beforeVals != nil {
		hookObj["before"] = types.ListValueMust(types.StringType, beforeVals)
	} else {
		hookObj["before"] = types.ListNull(types.StringType)
	}
	afterVals := convertStringsToValues(hook.After)
	if afterVals != nil {
		hookObj["after"] = types.ListValueMust(types.StringType, afterVals)
	} else {
		hookObj["after"] = types.ListNull(types.StringType)
	}
	onErrorVals := convertStringsToValues(hook.OnError)
	if onErrorVals != nil {
		hookObj["on_error"] = types.ListValueMust(types.StringType, onErrorVals)
	} else {
		hookObj["on_error"] = types.ListNull(types.StringType)
	}
	hookObj["shell"] = types.StringPointerValue(hook.Shell)
	return hookObj
}

type Model struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	WorkloadType types.String `tfsdk:"workload_type"`
	Credentials types.List   `tfsdk:"credentials"`
	Hooks       types.Object `tfsdk:"hooks"`
}

func fromAPI(ctx context.Context, workload *v20260501.Workload) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	credentials, d := types.ListValueFrom(ctx, types.StringType, workload.Credentials)
	diags.Append(d...)

	// Convert hooks
	var hooks types.Object
	if workload.Hooks != nil {
		hooksObj := map[string]attr.Value{}

		// Convert sign hook (always include, even if nil)
		if workload.Hooks.Sign != nil {
			signObj := convertHookToMap(ctx, workload.Hooks.Sign, &diags)
			signVal, _ := types.ObjectValue(getHookAttrTypes(), signObj)
			hooksObj["sign"] = signVal
		} else {
			hooksObj["sign"] = types.ObjectNull(getHookAttrTypes())
		}

		// Convert renew hook (always include, even if nil)
		if workload.Hooks.Renew != nil {
			renewObj := convertHookToMap(ctx, workload.Hooks.Renew, &diags)
			renewVal, _ := types.ObjectValue(getHookAttrTypes(), renewObj)
			hooksObj["renew"] = renewVal
		} else {
			hooksObj["renew"] = types.ObjectNull(getHookAttrTypes())
		}

		hooks, _ = types.ObjectValue(getHooksAttrTypes(), hooksObj)
	} else {
		hooks = types.ObjectNull(getHooksAttrTypes())
	}

	return &Model{
		ID:           types.StringPointerValue(workload.Id),
		Name:         types.StringPointerValue(workload.Name),
		WorkloadType: types.StringPointerValue(workload.WorkloadType),
		Credentials:  credentials,
		Hooks:        hooks,
	}, diags
}

func convertStringsToValues(strs *[]string) []attr.Value {
	if strs == nil || len(*strs) == 0 {
		return nil
	}
	result := []attr.Value{}
	for _, s := range *strs {
		result = append(result, types.StringValue(s))
	}
	return result
}

func getHookAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"before":   types.ListType{ElemType: types.StringType},
		"after":    types.ListType{ElemType: types.StringType},
		"on_error": types.ListType{ElemType: types.StringType},
		"shell":    types.StringType,
	}
}

func getHooksAttrTypes() map[string]attr.Type {
	hookType := types.ObjectType{AttrTypes: getHookAttrTypes()}
	return map[string]attr.Type{
		"sign":  hookType,
		"renew": hookType,
	}
}

func (m *Model) toAPI(ctx context.Context) (*v20260501.Workload, diag.Diagnostics) {
	var diags diag.Diagnostics

	var credentials []string
	d := m.Credentials.ElementsAs(ctx, &credentials, false)
	diags.Append(d...)

	var hooks *v20260501.EndpointHooks
	if !m.Hooks.IsNull() && !m.Hooks.IsUnknown() {
		hooks = &v20260501.EndpointHooks{}

		hooksObj := m.Hooks

		// Extract sign and renew from the object
		signVal := hooksObj.Attributes()["sign"]
		if signVal != nil && !signVal.IsNull() && !signVal.IsUnknown() {
			signObj := signVal.(types.Object)
			hooks.Sign = extractHook(ctx, signObj, &diags)
		}

		renewVal := hooksObj.Attributes()["renew"]
		if renewVal != nil && !renewVal.IsNull() && !renewVal.IsUnknown() {
			renewObj := renewVal.(types.Object)
			hooks.Renew = extractHook(ctx, renewObj, &diags)
		}
	}

	return &v20260501.Workload{
		Id:           utils.Ref(m.ID.ValueString()),
		Name:         utils.Ref(m.Name.ValueString()),
		WorkloadType: utils.Ref(m.WorkloadType.ValueString()),
		Credentials:  credentials,
		Hooks:        hooks,
	}, diags
}

func extractHook(ctx context.Context, hookObj types.Object, diags *diag.Diagnostics) *v20260501.EndpointHook {
	hook := &v20260501.EndpointHook{}
	attrs := hookObj.Attributes()

	if beforeVal, ok := attrs["before"]; ok && !beforeVal.IsNull() && !beforeVal.IsUnknown() {
		var before []string
		d := beforeVal.(types.List).ElementsAs(ctx, &before, false)
		diags.Append(d...)
		hook.Before = &before
	}

	if afterVal, ok := attrs["after"]; ok && !afterVal.IsNull() && !afterVal.IsUnknown() {
		var after []string
		d := afterVal.(types.List).ElementsAs(ctx, &after, false)
		diags.Append(d...)
		hook.After = &after
	}

	if onErrorVal, ok := attrs["on_error"]; ok && !onErrorVal.IsNull() && !onErrorVal.IsUnknown() {
		var onError []string
		d := onErrorVal.(types.List).ElementsAs(ctx, &onError, false)
		diags.Append(d...)
		hook.OnError = &onError
	}

	if shellVal, ok := attrs["shell"]; ok && !shellVal.IsNull() && !shellVal.IsUnknown() {
		hook.Shell = utils.Ref(shellVal.(types.String).ValueString())
	}

	return hook
}
