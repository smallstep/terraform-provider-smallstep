package utils

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IsStringEqual func(string, string) bool

// ToEqualString checks if a remote string is equivalent to a string in state
// and if so uses the string from state. This is used to avoid "inconsistent
// result after apply" errors for JSON and duration strings.
func ToEqualString(ctx context.Context, remote *string, state AttributeGetter, p path.Path, isEqual IsStringEqual) (types.String, diag.Diagnostics) {
	stringFromState := types.String{}
	diags := state.GetAttribute(ctx, p, &stringFromState)
	if diags.HasError() {
		return types.String{}, diags
	}

	stateIsEmpty := stringFromState.IsNull() || stringFromState.ValueString() == ""
	remoteIsEmpty := remote == nil || *remote == ""

	switch {
	case remoteIsEmpty && stateIsEmpty:
		return stringFromState, diags
	case remote == nil:
		return types.StringNull(), diags
	case isEqual(*remote, stringFromState.ValueString()):
		return stringFromState, diags
	default:
		return types.StringValue(*remote), diags
	}
}

func IsJSONEqual(a, b string) bool {
	if a == b {
		return true
	}
	var aVal, bVal any
	if err := json.Unmarshal([]byte(a), &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(aVal, bVal)
}

func IsDurationEqual(a, b string) bool {
	if a == b {
		return true
	}
	aDuration, err := time.ParseDuration(a)
	if err != nil {
		return false
	}
	bDuration, err := time.ParseDuration(b)
	if err != nil {
		return false
	}

	return aDuration == bDuration
}
