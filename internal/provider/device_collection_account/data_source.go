package device_collection_account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/workload"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct {
	client *v20231101.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = name
}

// Configure adds the Smallstep API client to the data source.
func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20231101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20231101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	dca, props, err := utils.Describe("deviceCollectionAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Schema",
			err.Error(),
		)
		return
	}

	certInfo, err := workload.NewCertificateInfoDataSourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	certData, err := workload.NewCertificateDataDataSourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
	}

	keyInfo, err := workload.NewKeyInfoDataSourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	reloadInfo, err := workload.NewReloadInfoDataSourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: dca,

		Attributes: map[string]schema.Attribute{
			"device_collection_slug": schema.StringAttribute{
				MarkdownDescription: props["deviceCollectionSlug"],
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: props["accountID"],
				Optional:            true,
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authorityID"],
				Optional:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["displayName"],
				Optional:            true,
			},
			"certificate_info": certInfo,
			"reload_info":      reloadInfo,
			"certificate_data": certData,
			"key_info":         keyInfo,
		},
	}
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var slug string
	var dcSlug string

	ds := req.Config.GetAttribute(ctx, path.Root("slug"), &slug)
	resp.Diagnostics.Append(ds...)
	ds = req.Config.GetAttribute(ctx, path.Root("device_collection_slug"), &dcSlug)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDeviceCollectionAccount(ctx, dcSlug, slug, &v20231101.GetDeviceCollectionAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device collection account %q: %v", slug, err),
		)
		return
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		resp.Diagnostics.AddError(
			"Device Collection Account Not Found",
			fmt.Sprintf("Device collection account %q data source not found", slug),
		)
		return
	default:
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading device collection account %q: %s", reqID, httpResp.StatusCode, slug, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	dca := &v20231101.DeviceCollectionAccount{}
	if err := json.NewDecoder(httpResp.Body).Decode(dca); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device collection account %s: %v", slug, err),
		)
		return
	}

	remote, ds := fromAPI(ctx, dca, dcSlug, req.Config)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}
