package managed_configuration

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

// DataSource implements data.smallstep_managed_configuration
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
	var config Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ID.ValueString()

	httpResp, err := a.client.GetManagedConfiguration(ctx, id, &v20230301.GetManagedConfigurationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read managed configuration %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading managed configuration %q: %s", httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	ac := &v20230301.ManagedConfiguration{}
	if err := json.NewDecoder(httpResp.Body).Decode(ac); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal managed configuration %q: %v", id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, ac, req.Config)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read managed configuration %q data source", id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, props, err := utils.Describe("managedConfiguration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	me, meProps, err := utils.Describe("managedEndpoint")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	x509, x509Props, err := utils.Describe("endpointX509CertificateData")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	ssh, sshProps, err := utils.Describe("endpointSSHCertificateData")
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
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"agent_configuration_id": schema.StringAttribute{
				MarkdownDescription: props["agentConfigurationID"],
				Computed:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: props["hostID"],
				Computed:            true,
			},
			"managed_endpoints": schema.SetNestedAttribute{
				MarkdownDescription: me,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: meProps["id"],
							Computed:            true,
						},
						"endpoint_configuration_id": schema.StringAttribute{
							MarkdownDescription: meProps["endpointConfigurationID"],
							Computed:            true,
						},
						"x509_certificate_data": schema.SingleNestedAttribute{
							MarkdownDescription: x509,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"common_name": schema.StringAttribute{
									MarkdownDescription: x509Props["commonName"],
									Computed:            true,
								},
								"sans": schema.SetAttribute{
									ElementType:         types.StringType,
									MarkdownDescription: x509Props["sans"],
									Computed:            true,
								},
							},
						},
						"ssh_certificate_data": schema.SingleNestedAttribute{
							MarkdownDescription: ssh,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"key_id": schema.StringAttribute{
									MarkdownDescription: sshProps["keyID"],
									Computed:            true,
								},
								"principals": schema.SetAttribute{
									ElementType:         types.StringType,
									MarkdownDescription: sshProps["principals"],
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}
