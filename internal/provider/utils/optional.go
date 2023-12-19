package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These helpers handle conversion from API types to terraform types for
// optional fields. The challenge with optional fields is that they may be
// either null or empty in terraform config, but the API returns nil for
// both.  In this case use the value from state to avoid "inconsistent result
// after apply" errors.
// https://github.com/hashicorp/terraform/blob/main/docs/resource-instance-change-lifecycle.md#planresourcechange

func ToOptionalSet(ctx context.Context, remote *[]string, priorState AttributeGetter, p path.Path) (types.Set, diag.Diagnostics) {
	if remote == nil || len(*remote) == 0 {
		setFromState := types.Set{}
		diags := priorState.GetAttribute(ctx, p, &setFromState)
		if diags.HasError() {
			return types.Set{}, diags
		}
		if setFromState.IsNull() || len(setFromState.Elements()) == 0 {
			return setFromState, diags
		}
	}

	if remote == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	values := make([]attr.Value, len(*remote))
	for i, s := range *remote {
		values[i] = types.StringValue(s)
	}
	return types.SetValue(types.StringType, values)
}

func ToOptionalList(ctx context.Context, remote *[]string, priorState AttributeGetter, p path.Path) (types.List, diag.Diagnostics) {
	if remote == nil || len(*remote) == 0 {
		listFromState := types.List{}
		diags := priorState.GetAttribute(ctx, p, &listFromState)
		if diags.HasError() {
			return types.List{}, diags
		}
		if listFromState.IsNull() || len(listFromState.Elements()) == 0 {
			return listFromState, diags
		}
	}

	if remote == nil {
		return types.ListNull(types.StringType), diag.Diagnostics{}
	}

	values := make([]attr.Value, len(*remote))
	for i, s := range *remote {
		values[i] = types.StringValue(s)
	}
	return types.ListValue(types.StringType, values)
}

func ToOptionalString[S ~string](ctx context.Context, remote *S, priorState AttributeGetter, p path.Path) (types.String, diag.Diagnostics) {
	if remote == nil || *remote == "" {
		stringFromState := types.String{}
		diags := priorState.GetAttribute(ctx, p, &stringFromState)
		if diags.HasError() {
			return types.String{}, diags
		}
		if stringFromState.IsNull() || stringFromState.ValueString() == "" {
			return stringFromState, diags
		}
	}

	if remote == nil {
		return types.StringNull(), diag.Diagnostics{}
	}

	return types.StringValue(string(*remote)), diag.Diagnostics{}
}

func ToOptionalBool(ctx context.Context, remote *bool, priorState AttributeGetter, p path.Path) (types.Bool, diag.Diagnostics) {
	if remote == nil || *remote == false {
		boolFromState := types.Bool{}
		diags := priorState.GetAttribute(ctx, p, &boolFromState)
		if diags.HasError() {
			return types.Bool{}, diags
		}
		if boolFromState.IsNull() || boolFromState.ValueBool() == false {
			return boolFromState, diags
		}
	}

	if remote == nil {
		return types.BoolNull(), diag.Diagnostics{}
	}

	return types.BoolValue(*remote), diag.Diagnostics{}
}

func ToOptionalInt(ctx context.Context, remote *int, priorState AttributeGetter, p path.Path) (types.Int64, diag.Diagnostics) {
	if remote == nil || *remote == 0 {
		intFromState := types.Int64{}
		diags := priorState.GetAttribute(ctx, p, &intFromState)
		if diags.HasError() {
			return types.Int64{}, diags
		}
		if intFromState.IsNull() || intFromState.ValueInt64() == 0 {
			return intFromState, diags
		}
	}

	if remote == nil {
		return types.Int64Null(), diag.Diagnostics{}
	}

	return types.Int64Value(int64(*remote)), diag.Diagnostics{}
}
