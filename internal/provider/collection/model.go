package collection

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const collectionTypeName = "smallstep_collection"

type Model struct {
	Slug          types.String `tfsdk:"slug"`
	DisplayName   types.String `tfsdk:"display_name"`
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	SchemaURI     types.String `tfsdk:"schema_uri"`
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	ID types.String `tfsdk:"id"`
}

func fromAPI(ctx context.Context, collection *v20231101.Collection, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	displayName, d := utils.ToOptionalString(ctx, &collection.DisplayName, state, path.Root("display_name"))
	diags.Append(d...)

	schemaURI, d := utils.ToOptionalString(ctx, collection.SchemaURI, state, path.Root("schema_uri"))
	diags.Append(d...)

	model := &Model{
		Slug:          types.StringValue(collection.Slug),
		InstanceCount: types.Int64Value(int64(collection.InstanceCount)),
		DisplayName:   displayName,
		SchemaURI:     schemaURI,
		CreatedAt:     types.StringValue(collection.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:     types.StringValue(collection.UpdatedAt.Format(time.RFC3339)),
		ID:            types.StringValue(collection.Slug),
	}

	return model, diags
}

func toAPINew(model *Model) *v20231101.NewCollection {
	return &v20231101.NewCollection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueStringPointer(),
		SchemaURI:   model.SchemaURI.ValueStringPointer(),
	}
}

func toAPI(model *Model) *v20231101.Collection {
	return &v20231101.Collection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueString(),
		SchemaURI:   model.SchemaURI.ValueStringPointer(),
	}
}
