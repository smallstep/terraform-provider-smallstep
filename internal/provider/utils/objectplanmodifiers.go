package utils

import (
	"bytes"
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseNullForUnknown returns a plan modifier that works exactly like the builtin
// UseStateForUnknown except that null states are used instead of ignored.
func MaybeUseStateForUnknown(path path.Path, key string, val []byte) planmodifier.Object {
	return maybeUseStateForUnknownModifier{
		path: path,
		key:  key,
		val:  val,
	}
}

// maybeUseStateForUnknownModifier implements the plan modifier.
type maybeUseStateForUnknownModifier struct {
	path path.Path
	key  string
	val  []byte
}

// Description returns a human-readable description of the plan modifier.
func (m maybeUseStateForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m maybeUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyObject implements the plan modification logic.
func (m maybeUseStateForUnknownModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.Path.Equal(m.path) {
		v, diags := req.Private.GetKey(ctx, m.key)
		resp.Diagnostics.Append(diags...)
		if bytes.Equal(v, m.val) {
			resp.PlanValue = req.StateValue
		}
	}
}
