package sso_integration

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

// DataSourceModel is used only by the data source and doesn't include UI-only fields
type DataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	RedirectURI         types.String `tfsdk:"redirect_uri"`
	LifecycleFailureURI types.String `tfsdk:"lifecycle_failure_uri"`
	Secret              types.String `tfsdk:"secret"`
}

func (ds *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = typeName
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	integration, props, err := utils.DescribeV20260501("ssoIntegration")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI SSO Integration Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: integration,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"redirect_uri": schema.StringAttribute{
				MarkdownDescription: props["redirectURI"],
				Computed:            true,
			},
			"lifecycle_failure_uri": schema.StringAttribute{
				MarkdownDescription: props["lifecycleFailureURI"],
				Computed:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: props["secret"],
				Computed:            true,
				Sensitive:           true,
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
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	ds.client = clients.V20260501
}

func (ds *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config *DataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integrationID := config.ID.ValueString()
	if integrationID == "" {
		resp.Diagnostics.AddError("Invalid Read SSO Integration Request", "SSO Integration ID is required")
		return
	}

	httpResp, err := ds.client.GetSsoIntegration(ctx, integrationID, &v20260501.GetSsoIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to read sso integration: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading sso integration: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	integration := &v20260501.SsoIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(integration); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal sso integration: %v", err))
		return
	}

	model := &DataSourceModel{
		ID:                  types.StringPointerValue(integration.Id),
		RedirectURI:         types.StringValue(integration.RedirectURI),
		LifecycleFailureURI: types.StringPointerValue(integration.LifecycleFailureURI),
		Secret:              types.StringNull(),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
