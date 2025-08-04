package ssh

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

type Model struct{}

var Attributes = map[string]attr.Type{}

func FromAPI(ctx context.Context, conf *v20250101.StrategySSHConfig, state utils.AttributeGetter, root path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if conf == nil {
		return basetypes.NewObjectNull(Attributes), diags
	}

	obj, ds := basetypes.NewObjectValue(Attributes, map[string]attr.Value{})
	diags.Append(ds...)

	return obj, diags
}

func (m *Model) ToAPI(ctx context.Context, obj types.Object) (v20250101.StrategySSHConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	return v20250101.StrategySSHConfig{}, diags
}
