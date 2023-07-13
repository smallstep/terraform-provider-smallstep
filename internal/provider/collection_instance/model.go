package collection_instance

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const instanceTypeName = "smallstep_collection_instance"

type Model struct {
	CollectionSlug types.String `tfsdk:"collection_slug"`
	ID             types.String `tfsdk:"id"`
	Data           types.String `tfsdk:"data"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func fromAPI(ctx context.Context, slug string, instance *v20230301.CollectionInstance, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		CollectionSlug: types.StringValue(slug),
		ID:             types.StringValue(instance.Id),
		CreatedAt:      types.StringValue(instance.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:      types.StringValue(instance.UpdatedAt.Format(time.RFC3339)),
	}

	apiDataJSON, err := json.Marshal(instance.Data)
	if err != nil {
		diags.AddError("Marshal Instance Data", err.Error())
		return nil, diags
	}

	dataFromState := types.String{}
	d := state.GetAttribute(ctx, path.Root("data"), &dataFromState)
	diags = append(diags, d...)
	if utils.IsJSONEqual(dataFromState.ValueString(), string(apiDataJSON)) {
		model.Data = dataFromState
	} else {
		model.Data = types.StringValue(string(apiDataJSON))
	}

	return model, diags
}
func toAPI(model *Model) (*v20230301.PutCollectionInstanceJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	instance := &v20230301.PutCollectionInstanceJSONRequestBody{}

	if err := json.Unmarshal([]byte(model.Data.ValueString()), &instance.Data); err != nil {
		diags.AddError("Parse Instance Data", err.Error())
	}

	return instance, diags
}
