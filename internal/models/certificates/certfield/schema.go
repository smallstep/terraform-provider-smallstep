package certfield

import (
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewResourceSchema() resourceschema.SingleNestedAttribute {
	return resourceschema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]resourceschema.Attribute{
			"static": resourceschema.StringAttribute{
				Optional: true,
			},
			"device_metadata": resourceschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func NewDataSourceSchema() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]datasourceschema.Attribute{
			"static": datasourceschema.StringAttribute{
				Optional: true,
			},
			"device_metadata": datasourceschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func NewListResourceSchema() resourceschema.SingleNestedAttribute {
	return resourceschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]resourceschema.Attribute{
			"static": resourceschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"device_metadata": resourceschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func NewListDataSourceSchema() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]datasourceschema.Attribute{
			"static": datasourceschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"device_metadata": datasourceschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}
