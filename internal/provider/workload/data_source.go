package workload

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func NewCertificateInfoDataSourceSchema() (schema.SingleNestedAttribute, error) {
	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		return schema.SingleNestedAttribute{}, err
	}

	out := schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: certInfo,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: certInfoProps["type"],
				Optional:            true,
			},
			"duration": schema.StringAttribute{
				MarkdownDescription: certInfoProps["duration"],
				Optional:            true,
			},
			"crt_file": schema.StringAttribute{
				MarkdownDescription: certInfoProps["crtFile"],
				Optional:            true,
			},
			"key_file": schema.StringAttribute{
				MarkdownDescription: certInfoProps["keyFile"],
				Optional:            true,
			},
			"root_file": schema.StringAttribute{
				MarkdownDescription: certInfoProps["rootFile"],
				Optional:            true,
			},
			"uid": schema.Int64Attribute{
				MarkdownDescription: certInfoProps["uid"],
				Optional:            true,
			},
			"gid": schema.Int64Attribute{
				MarkdownDescription: certInfoProps["gid"],
				Optional:            true,
			},
			"mode": schema.Int64Attribute{
				MarkdownDescription: certInfoProps["mode"],
				Optional:            true,
			},
		},
	}

	return out, nil
}

func NewCertificateDataDataSourceSchema() (schema.SingleNestedAttribute, error) {
	certData, _, err := utils.Describe("x509Fields")
	if err != nil {
		return schema.SingleNestedAttribute{}, err
	}

	out := schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: certData,
		Attributes: map[string]schema.Attribute{
			"common_name": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.StringAttribute{
						Optional: true,
					},
					"device_metadata": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"sans": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"organization": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"organizational_unit": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"locality": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"country": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"province": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"street_address": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"postal_code": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"static": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"device_metadata": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
		},
	}
	return out, nil
}

func NewKeyInfoDataSourceSchema() (schema.SingleNestedAttribute, error) {
	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		return schema.SingleNestedAttribute{}, err
	}

	out := schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: keyInfo,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["type"],
			},
			"format": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["format"],
			},
			"pub_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["pubFile"],
			},
			"protection": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: keyInfoProps["protection"],
			},
		},
	}

	return out, nil
}

func NewReloadInfoDataSourceSchema() (schema.SingleNestedAttribute, error) {
	reloadInfo, reloadInfoProps, err := utils.Describe("endpointReloadInfo")
	if err != nil {
		return schema.SingleNestedAttribute{}, err
	}

	out := schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: reloadInfo,
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: reloadInfoProps["method"],
			},
			"pid_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: reloadInfoProps["pidFile"],
			},
			"signal": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: reloadInfoProps["signal"],
			},
			"unit_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: reloadInfoProps["unitName"],
			},
		},
	}

	return out, nil
}
