package certinfo

import (
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/sshinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/x509info"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func NewResourceSchema() (resourceschema.SingleNestedAttribute, error) {
	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		return resourceschema.SingleNestedAttribute{}, err
	}

	return resourceschema.SingleNestedAttribute{
		MarkdownDescription: certInfo,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]resourceschema.Attribute{
			"x509": x509info.NewResourceSchema(),
			"ssh":  sshinfo.NewResourceSchema(),
			"duration": resourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["duration"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					// If unset the duration will default to 24h.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"authority_id": resourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["authorityID"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"crt_file": resourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["crtFile"],
				Optional:            true,
			},
			"key_file": resourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["keyFile"],
				Optional:            true,
			},
			"root_file": resourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["rootFile"],
				Optional:            true,
			},
			"uid": resourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["uid"],
				Optional:            true,
			},
			"gid": resourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["gid"],
				Optional:            true,
			},
			"mode": resourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["mode"],
				Optional:            true,
			},
		},
	}, nil
}

func NewDataSourceSchema() (datasourceschema.SingleNestedAttribute, error) {
	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		return datasourceschema.SingleNestedAttribute{}, err
	}

	return datasourceschema.SingleNestedAttribute{
		MarkdownDescription: certInfo,
		Computed:            true,
		Attributes: map[string]datasourceschema.Attribute{
			"x509": x509info.NewDataSourceSchema(),
			"ssh":  sshinfo.NewDataSourceSchema(),
			"duration": datasourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["duration"],
				Computed:            true,
			},
			"authority_id": datasourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["authorityID"],
				Computed:            true,
			},
			"crt_file": datasourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["crtFile"],
				Computed:            true,
			},
			"key_file": datasourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["keyFile"],
				Computed:            true,
			},
			"root_file": datasourceschema.StringAttribute{
				MarkdownDescription: certInfoProps["rootFile"],
				Computed:            true,
			},
			"uid": datasourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["uid"],
				Computed:            true,
			},
			"gid": datasourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["gid"],
				Computed:            true,
			},
			"mode": datasourceschema.Int64Attribute{
				MarkdownDescription: certInfoProps["mode"],
				Computed:            true,
			},
		},
	}, nil
}
