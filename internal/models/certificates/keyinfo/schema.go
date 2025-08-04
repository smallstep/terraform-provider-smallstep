package keyinfo

import (
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func NewResourceSchema() (resourceschema.SingleNestedAttribute, error) {
	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		return resourceschema.SingleNestedAttribute{}, err
	}

	return resourceschema.SingleNestedAttribute{
		// This object is not required by the API but a default object
		// will always be returned with format, type and protection set to
		// "DEFAULT". To avoid "inconsistent result after apply" errors
		// require these fields to be set explicitly in terraform.
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.Object{
			// The key will always be returned and have default values set if not
			// supplied by the client. This prevents showing (known after apply) in the
			// plan.
			objectplanmodifier.UseStateForUnknown(),
		},
		MarkdownDescription: keyInfo,
		Attributes: map[string]resourceschema.Attribute{
			"type": resourceschema.StringAttribute{
				Required:            true,
				MarkdownDescription: keyInfoProps["type"],
			},
			"format": resourceschema.StringAttribute{
				Required:            true,
				MarkdownDescription: keyInfoProps["format"],
			},
			"pub_file": resourceschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["pubFile"],
			},
			"protection": resourceschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["protection"],
			},
		},
	}, nil
}

func NewDataSourceSchema() (datasourceschema.SingleNestedAttribute, error) {
	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		return datasourceschema.SingleNestedAttribute{}, err
	}

	return datasourceschema.SingleNestedAttribute{
		Computed:            true,
		MarkdownDescription: keyInfo,
		Attributes: map[string]datasourceschema.Attribute{
			"type": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: keyInfoProps["type"],
			},
			"format": datasourceschema.StringAttribute{
				Required:            true,
				MarkdownDescription: keyInfoProps["format"],
			},
			"pub_file": datasourceschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["pubFile"],
			},
			"protection": datasourceschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["protection"],
			},
		},
	}, nil
}
