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

	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *v20250101.Client
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

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = client
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	credential, props, err := utils.Describe("credential")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Schema",
			err.Error(),
		)
		return
	}

	cert, certProps, err := utils.Describe("credentialCertificate")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Certificate Schema",
			err.Error(),
		)
		return
	}

	policy, policyProps, err := utils.Describe("policyMatchCriteria")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Policy Schema",
			err.Error(),
		)
		return
	}

	files, filesProps, err := utils.Describe("credentialFiles")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Files Schema",
			err.Error(),
		)
		return
	}

	x509, _, err := utils.Describe("x509Fields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	_, certFieldProps, err := utils.Describe("certificateField")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Certificate Field Schema",
			err.Error(),
		)
		return
	}

	_, certFieldListProps, err := utils.Describe("certificateFieldList")
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
		},
	}

	key, keyProps, err := utils.Describe("credentialKey")
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
						},
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certProps["duration"],
						Computed:            true,
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

	httpResp, err := ds.client.GetCredential(ctx, id, &v20250101.GetCredentialParams{})
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

	credential := &v20250101.Credential{}
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
