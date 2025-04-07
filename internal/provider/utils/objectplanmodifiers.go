package utils

import (
	"bytes"
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MaybeUseStateForUnknown uses state for unknown if and only if the specified
// key in private state matches the specified value.
// UseStateForUnknown except that null states are used instead of ignored.
func MaybeUseStateForUnknown(key string, val []byte) planmodifier.Object {
	return maybeUseStateForUnknownModifier{
		key: key,
		val: val,
	}
}

// maybeUseStateForUnknownModifier implements the plan modifier.
type maybeUseStateForUnknownModifier struct {
	key string
	val []byte
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
	// NOTE (areed) Copied from terraform's UseStateForUnknown modifier but I don't know if it's needed.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	v, diags := req.Private.GetKey(ctx, m.key)
	resp.Diagnostics.Append(diags...)
	if bytes.Equal(v, m.val) {
		resp.PlanValue = req.StateValue
	}
}

// NullWhen uses the supplied null value for the plan when the attribute at the
// specified path is set. This was created to solve a problem in smallstep_account
// resources where at most one of certificate.x509 and certificate.ssh can be set.
// Because x509 is computed, if ssh is set and applied then subsequent plans will
// shown "known after apply" for x509. This modifier prevents that because x509
// will not return a computed value when ssh is set.
func NullWhen(path path.Path, val basetypes.ObjectValue) planmodifier.Object {
	return nullWhen{
		path: path,
		val:  val,
	}
}

// nullWhen implements the plan modifier.
type nullWhen struct {
	path path.Path
	val  basetypes.ObjectValue
}

// Description returns a human-readable description of the plan modifier.
func (m nullWhen) Description(_ context.Context) string {
	return "A null object will be used when some other attribute is set."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m nullWhen) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyObject implements the plan modification logic.
func (m nullWhen) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// NOTE (areed) Copied from terraform's UseStateForUnknown modifier but I don't know if it's needed.
	if req.ConfigValue.IsUnknown() {
		return
	}

	other := &types.Object{}
	ds := req.Plan.GetAttribute(ctx, m.path, other)
	resp.Diagnostics.Append(ds...)

	if other.IsNull() || other.IsUnknown() {
		return
	}

	resp.PlanValue = m.val
}
