package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSource = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *v20260501.Client
}

func (ds *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = typeName
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	proxy, props, err := utils.DescribeV20260501("proxy")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Proxy Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: proxy,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"remote_address": schema.StringAttribute{
				MarkdownDescription: props["remoteAddress"],
				Computed:            true,
			},
			"credentials": schema.ListAttribute{
				MarkdownDescription: props["credentials"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"match_addresses": schema.ListAttribute{
				MarkdownDescription: props["matchAddresses"],
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (ds *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*clientset.Clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = clients.V20260501
}

func (ds *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config *Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxyID := config.ID.ValueString()
	if proxyID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Proxy Request",
			"Proxy ID is required",
		)
		return
	}

	httpResp, err := ds.client.GetProxy(ctx, proxyID, &v20260501.GetProxyParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read proxy: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading proxy: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	proxy := &v20260501.Proxy{}
	if err := json.NewDecoder(httpResp.Body).Decode(proxy); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal proxy: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, proxy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
