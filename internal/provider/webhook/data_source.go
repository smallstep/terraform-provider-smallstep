package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource implements data.smallstep_provisioner_webhook
type DataSource struct {
	client *v20230301.Client
}

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = webhookTypeName
}

// Configure adds the Smallstep API client to the data source.
func (a *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20230301.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20230301.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.client = client
}

func (a *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	provisionerID := config.ProvisionerID.ValueString()
	authorityID := config.AuthorityID.ValueString()
	idOrName := config.ID.ValueString()
	if config.ID.IsNull() {
		idOrName = config.Name.ValueString()
	}
	httpResp, err := a.client.GetWebhook(ctx, authorityID, provisionerID, idOrName, &v20230301.GetWebhookParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read webhook %s: %v", config.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading webhook %s: %s", httpResp.StatusCode, config.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	webhook := &v20230301.ProvisionerWebhook{}
	if err := json.NewDecoder(httpResp.Body).Decode(webhook); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal webhook %s: %v", config.ID.String(), err),
		)
		return
	}

	remote, d := fromAPI(ctx, webhook, req.Config)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	remote.AuthorityID = types.StringValue(authorityID)
	remote.ProvisionerID = types.StringValue(provisionerID)

	tflog.Trace(ctx, fmt.Sprintf("read webhook %q data source", idOrName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, props, err := utils.Describe("provisionerWebhook")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}
	basicAuth, basicAuthProps, err := utils.Describe("basicAuth")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Optional:            true,
				Computed:            true,
			},
			"provisioner_id": schema.StringAttribute{
				MarkdownDescription: props["provisioner_id"],
				Required:            true,
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authority_id"],
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: props["kind"],
				Computed:            true,
			},
			"cert_type": schema.StringAttribute{
				MarkdownDescription: props["certType"],
				Computed:            true,
			},
			"server_type": schema.StringAttribute{
				MarkdownDescription: props["serverType"],
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: props["url"],
				Computed:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: props["secret"],
				Computed:            true,
			},
			"collection_slug": schema.StringAttribute{
				MarkdownDescription: props["collectionSlug"],
				Computed:            true,
			},
			"disable_tls_client_auth": schema.BoolAttribute{
				MarkdownDescription: props["disableTLSClientAuth"],
				Computed:            true,
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: props["bearerToken"],
				Computed:            true,
			},
			"basic_auth": schema.SingleNestedAttribute{
				MarkdownDescription: basicAuth,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: basicAuthProps["username"],
						Computed:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: basicAuthProps["password"],
						Computed:            true,
					},
				},
			},
		},
	}
}
