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

// Secrets and the collection_slug are never returned in read operations and
// won't be available in state for data source so those fields will always be
// empty even if the webhook was created them. The data source schema must not
// document those fields to avoid confusion, but the type cannot have fields not
// found in the schema. So for webhooks a separate model is needed for resource
// and data source.
type DataModel struct {
	ID                   types.String `tfsdk:"id"`
	AuthorityID          types.String `tfsdk:"authority_id"`
	ProvisionerID        types.String `tfsdk:"provisioner_id"`
	Name                 types.String `tfsdk:"name"`
	Kind                 types.String `tfsdk:"kind"`
	CertType             types.String `tfsdk:"cert_type"`
	ServerType           types.String `tfsdk:"server_type"`
	URL                  types.String `tfsdk:"url"`
	DisableTLSClientAuth types.Bool   `tfsdk:"disable_tls_client_auth"`
}

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource implements data.smallstep_provisioner_webhook
type DataSource struct {
	client *v20230301.Client
}

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = typeName
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
	var config DataModel

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

	remote := &DataModel{
		AuthorityID:          types.StringValue(authorityID),
		ProvisionerID:        types.StringValue(provisionerID),
		ID:                   types.StringValue(utils.Deref(webhook.Id)),
		Name:                 types.StringValue(webhook.Name),
		Kind:                 types.StringValue(string(webhook.Kind)),
		CertType:             types.StringValue(string(webhook.CertType)),
		ServerType:           types.StringValue(string(webhook.ServerType)),
		URL:                  types.StringValue(utils.Deref(webhook.Url)),
		DisableTLSClientAuth: types.BoolValue(utils.Deref(webhook.DisableTLSClientAuth)),
	}

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
			"disable_tls_client_auth": schema.BoolAttribute{
				MarkdownDescription: props["disableTLSClientAuth"],
				Computed:            true,
			},
		},
	}
}
