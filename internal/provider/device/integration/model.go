package integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/integrations/intune"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/integrations/jamf"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type Model struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Jamf   types.Object `tfsdk:"jamf"`
	Intune types.Object `tfsdk:"intune"`
}

func fromAPI(ctx context.Context, integration *v20250101.DeviceInventoryIntegration, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var (
		diags, ds diag.Diagnostics
		name      = basetypes.NewStringNull()
		jamfObj   = basetypes.NewObjectNull(jamf.Attributes)
		intuneObj = basetypes.NewObjectNull(intune.Attributes)
	)

	if integration.Name != nil {
		name = types.StringValue(*integration.Name)
	}

	switch integration.Kind {
	case v20250101.DeviceInventoryIntegrationKindJamf:
		conf, err := integration.Configuration.AsJamfInventoryIntegration()
		if err != nil {
			diags.AddError("Device Inventory Integration Jamf Parse Error", err.Error())
			return nil, diags
		}

		jamfObj, ds = jamf.FromAPI(ctx, &conf, state, path.Root("jamf"))
		diags.Append(ds...)
	case v20250101.DeviceInventoryIntegrationKindIntune:
		conf, err := integration.Configuration.AsIntuneInventoryIntegration()
		if err != nil {
			diags.AddError("Device Inventory Integration Intune Parse Error", err.Error())
			return nil, diags
		}

		intuneObj, ds = intune.FromAPI(ctx, &conf, state, path.Root("intune"))
		diags.Append(ds...)
	default:
		diags.AddError("unsupported device inventory integration kind", string(integration.Kind))
		return nil, diags
	}

	return &Model{
		ID:     types.StringValue(integration.Id),
		Name:   name,
		Jamf:   jamfObj,
		Intune: intuneObj,
	}, diags
}

func toAPI(ctx context.Context, model *Model) (*v20250101.DeviceInventoryIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	integration := &v20250101.DeviceInventoryIntegration{
		Id:            model.ID.ValueString(),
		Name:          model.Name.ValueStringPointer(),
		Configuration: v20250101.DeviceInventoryIntegration_Configuration{},
	}

	switch {
	case !model.Jamf.IsNull():
		integration.Kind = v20250101.DeviceInventoryIntegrationKindJamf
		conf, ds := new(jamf.Model).ToAPI(ctx, model.Jamf)
		diags.Append(ds...)

		if err := integration.Configuration.FromJamfInventoryIntegration(conf); err != nil {
			diags.AddError("Device Inventory Integration Configuration Error", err.Error())
		}
	case !model.Intune.IsNull():
		integration.Kind = v20250101.DeviceInventoryIntegrationKindIntune
		conf, ds := new(intune.Model).ToAPI(ctx, model.Jamf)
		diags.Append(ds...)

		if err := integration.Configuration.FromIntuneInventoryIntegration(conf); err != nil {
			diags.AddError("Device Inventory Integration Configuration Error", err.Error())
		}
	}

	return integration, diags
}
