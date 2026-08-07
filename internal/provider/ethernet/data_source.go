package ethernet

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
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
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

	clients, ok := req.ProviderData.(*clientset.Clients)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = clients.V20250101
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ethernet, props, err := utils.Describe("ethernet")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Ethernet Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: ethernet,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"radius_server_ca": schema.StringAttribute{
				MarkdownDescription: props["radiusServerCA"],
				Computed:            true,
			},
			"autojoin": schema.BoolAttribute{
				MarkdownDescription: props["autojoin"],
				Computed:            true,
			},
			"credentials": schema.SetAttribute{
				MarkdownDescription: props["credentials"],
				ElementType:         types.StringType,
				Computed:            true,
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

	httpResp, err := ds.client.GetEthernet(ctx, id, &v20250101.GetEthernetParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read ethernet %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading ethernet %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	ethernet := &v20250101.Ethernet{}
	if err := json.NewDecoder(httpResp.Body).Decode(ethernet); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal ethernet %s: %v", id, err),
		)
		return
	}

	remote := FromAPI(ctx, ethernet, &resp.Diagnostics, req.Config)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
