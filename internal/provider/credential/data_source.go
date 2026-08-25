package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *v20260501.Client
}

func (ds *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = name
}

// Configure adds the Smallstep API client to the data source.
func (ds *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*clientset.Clients)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = clients.V20260501
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	credential, props, err := utils.DescribeV20260501("credential")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Schema",
			err.Error(),
		)
		return
	}

	cert, certProps, err := utils.DescribeV20260501("credentialCertificate")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Certificate Schema",
			err.Error(),
		)
		return
	}

	policy, policyProps, err := utils.DescribeV20260501("policyMatchCriteria")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Policy Schema",
			err.Error(),
		)
		return
	}

	files, filesProps, err := utils.DescribeV20260501("credentialFiles")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Files Schema",
			err.Error(),
		)
		return
	}

	x509, x509Props, err := utils.DescribeV20260501("x509Fields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	typedSans, _, err := utils.DescribeV20260501("x509TypedSANs")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Typed SANs Schema",
			err.Error(),
		)
		return
	}

	customExtension, customExtensionProps, err := utils.DescribeV20260501("x509CustomExtension")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Custom Extension Schema",
			err.Error(),
		)
		return
	}

	namePolicy, namePolicyProps, err := utils.DescribeV20260501("x509NamePolicy")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Name Policy Schema",
			err.Error(),
		)
		return
	}

	_, x509NamesProps, err := utils.DescribeV20260501("x509Names")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Names Schema",
			err.Error(),
		)
		return
	}

	_, certFieldProps, err := utils.DescribeV20260501("certificateField")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Certificate Field Schema",
			err.Error(),
		)
		return
	}

	_, certFieldListProps, err := utils.DescribeV20260501("certificateFieldList")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Certificate Field List Schema",
			err.Error(),
		)
		return
	}

	name := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.StringAttribute{
				MarkdownDescription: certFieldProps["static"],
				Computed:            true,
			},
			"device_metadata": schema.StringAttribute{
				MarkdownDescription: certFieldProps["deviceMetadata"],
				Computed:            true,
			},
		},
	}

	nameList := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.ListAttribute{
				MarkdownDescription: certFieldListProps["static"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"device_metadata": schema.ListAttribute{
				MarkdownDescription: certFieldListProps["deviceMetadata"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"insecure_include_requested": schema.BoolAttribute{
				MarkdownDescription: certFieldListProps["insecureIncludeRequested"],
				Computed:            true,
			},
		},
	}

	x509Names := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"common_names": schema.ListAttribute{
				MarkdownDescription: x509NamesProps["commonNames"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"dns": schema.ListAttribute{
				MarkdownDescription: x509NamesProps["dns"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"emails": schema.ListAttribute{
				MarkdownDescription: x509NamesProps["emails"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"ips": schema.ListAttribute{
				MarkdownDescription: x509NamesProps["ips"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"uris": schema.ListAttribute{
				MarkdownDescription: x509NamesProps["uris"],
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}

	key, keyProps, err := utils.DescribeV20260501("credentialKey")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Key Info Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: credential,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Computed:            true,
			},
			"management_mode": schema.StringAttribute{
				MarkdownDescription: props["managementMode"],
				Computed:            true,
			},
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: cert,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"common_name":         name,
							"sans":                nameList,
							"organization":        nameList,
							"organizational_unit": nameList,
							"locality":            nameList,
							"country":             nameList,
							"province":            nameList,
							"street_address":      nameList,
							"postal_code":         nameList,
							"given_name":          name,
							"serial_number":       name,
							"surname":             name,
							"typed_sans": schema.SingleNestedAttribute{
								MarkdownDescription: typedSans,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"dns_names":            nameList,
									"ip_addresses":         nameList,
									"email_addresses":      nameList,
									"uris":                 nameList,
									"user_principal_names": nameList,
								},
							},
							"extended_key_usage": schema.ListAttribute{
								MarkdownDescription: x509Props["extendedKeyUsage"],
								ElementType:         types.StringType,
								Computed:            true,
							},
							"custom_extensions": schema.ListNestedAttribute{
								MarkdownDescription: customExtension,
								Computed:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"oid": schema.StringAttribute{
											MarkdownDescription: customExtensionProps["oid"],
											Computed:            true,
										},
										"critical": schema.BoolAttribute{
											MarkdownDescription: customExtensionProps["critical"],
											Computed:            true,
										},
										"value": schema.StringAttribute{
											MarkdownDescription: customExtensionProps["value"],
											Computed:            true,
										},
									},
								},
							},
						},
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certProps["duration"],
						Computed:            true,
					},
					"name_policy": schema.SingleNestedAttribute{
						MarkdownDescription: namePolicy,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"allow": x509Names,
							"deny":  x509Names,
							"allow_wildcard_names": schema.BoolAttribute{
								MarkdownDescription: namePolicyProps["allowWildcardNames"],
								Computed:            true,
							},
						},
					},
					"authority_id": schema.StringAttribute{
						MarkdownDescription: certProps["authorityID"],
						Computed:            true,
					},
				},
			},
			"key": schema.SingleNestedAttribute{
				MarkdownDescription: key,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyProps["type"],
					},
					"pub_file": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyProps["pubFile"],
					},
					"protection": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyProps["protection"],
					},
					"compatibility": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyProps["compatibility"],
					},
					"store": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyProps["store"],
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: policy,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"assurance": schema.ListAttribute{
						MarkdownDescription: policyProps["assurance"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"ownership": schema.ListAttribute{
						MarkdownDescription: policyProps["ownership"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"os": schema.ListAttribute{
						MarkdownDescription: policyProps["operatingSystem"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"source": schema.ListAttribute{
						MarkdownDescription: policyProps["source"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: policyProps["tags"],
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
			},
			"files": schema.SingleNestedAttribute{
				MarkdownDescription: files,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"root_file": schema.StringAttribute{
						MarkdownDescription: filesProps["rootFile"],
						Computed:            true,
					},
					"crt_file": schema.StringAttribute{
						MarkdownDescription: filesProps["crtFile"],
						Computed:            true,
					},
					"key_file": schema.StringAttribute{
						MarkdownDescription: filesProps["keyFile"],
						Computed:            true,
					},
					"key_format": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: filesProps["keyFormat"],
					},
					"uid": schema.Int64Attribute{
						MarkdownDescription: filesProps["uid"],
						Computed:            true,
					},
					"gid": schema.Int64Attribute{
						MarkdownDescription: filesProps["gid"],
						Computed:            true,
					},
					"mode": schema.Int64Attribute{
						MarkdownDescription: filesProps["mode"],
						Computed:            true,
					},
				},
			},
		},
	}
}

func (ds *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id string
	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := ds.client.GetCredential(ctx, id, &v20260501.GetCredentialParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read credential %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading credential %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	credential := &v20260501.Credential{}
	if err := json.NewDecoder(httpResp.Body).Decode(credential); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal credential %s: %v", id, err),
		)
		return
	}

	remote := fromAPI(ctx, &resp.Diagnostics, credential, req.Config)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
