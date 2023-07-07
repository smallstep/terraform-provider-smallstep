package collection

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const collectionTypeName = "smallstep_collection"

type Model struct {
	Slug          types.String `tfsdk:"slug"`
	DisplayName   types.String `tfsdk:"display_name"`
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	ID types.String `tfsdk:"id"`
}

func fromAPI(ctx context.Context, collection *v20230301.Collection, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		Slug:          types.StringValue(collection.Slug),
		InstanceCount: types.Int64Value(int64(collection.InstanceCount)),
		DisplayName:   types.StringValue(collection.DisplayName),
		CreatedAt:     types.StringValue(collection.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:     types.StringValue(collection.UpdatedAt.Format(time.RFC3339)),
		ID:            types.StringValue(collection.Slug),
	}

	return model, diags
}

func toAPI(model *Model) *v20230301.NewCollection {
	return &v20230301.NewCollection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueStringPointer(),
	}
}
