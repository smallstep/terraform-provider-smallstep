package credential

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestCredentialResourceSlugPlanModifiers(t *testing.T) {
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	NewResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.Empty(t, schemaResp.Diagnostics)

	slug, ok := schemaResp.Schema.Attributes["slug"].(schema.StringAttribute)
	require.True(t, ok)
	nullResource := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)

	tests := map[string]struct {
		state           types.String
		plan            types.String
		requiresReplace bool
	}{
		"changed": {
			state:           types.StringValue("example-old"),
			plan:            types.StringValue("example-new"),
			requiresReplace: true,
		},
		"unchanged": {
			state: types.StringValue("example-old"),
			plan:  types.StringValue("example-old"),
		},
		"create": {
			state: types.StringNull(),
			plan:  types.StringValue("example-new"),
		},
		"create with unknown slug": {
			state: types.StringNull(),
			plan:  types.StringUnknown(),
		},
		"update with unknown slug": {
			// RequiresReplace treats an unknown plan as different from known state.
			state:           types.StringValue("example-old"),
			plan:            types.StringUnknown(),
			requiresReplace: true,
		},
		"destroy": {
			state: types.StringValue("example-old"),
			plan:  types.StringNull(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := tfsdk.State{Schema: schemaResp.Schema, Raw: nullResource}
			plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: nullResource}
			if !tt.state.IsNull() {
				require.Empty(t, state.SetAttribute(ctx, path.Root("slug"), tt.state))
			}
			if !tt.plan.IsNull() {
				require.Empty(t, plan.SetAttribute(ctx, path.Root("slug"), tt.plan))
			}

			req := planmodifier.StringRequest{
				Path:           path.Root("slug"),
				PathExpression: path.MatchRoot("slug"),
				Config:         tfsdk.Config{Schema: schemaResp.Schema, Raw: plan.Raw},
				ConfigValue:    tt.plan,
				State:          state,
				StateValue:     tt.state,
				Plan:           plan,
				PlanValue:      tt.plan,
			}
			var requiresReplace bool
			for _, modifier := range slug.PlanModifiers {
				resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
				modifier.PlanModifyString(ctx, req, &resp)
				require.Empty(t, resp.Diagnostics)
				requiresReplace = requiresReplace || resp.RequiresReplace
				req.PlanValue = resp.PlanValue
			}

			require.Equal(t, tt.requiresReplace, requiresReplace)
			require.True(t, tt.plan.Equal(req.PlanValue), "slug plan value must be preserved")
		})
	}
}
