package endpoint_configuration

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

// DataSource implements data.smallstep_endpoint_configuration
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

	httpResp, err := a.client.GetEndpointConfiguration(ctx, id, &v20230301.GetEndpointConfigurationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read endpoint configuration %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading endpoint configuration %q: %s", httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	ac := &v20230301.EndpointConfiguration{}
	if err := json.NewDecoder(httpResp.Body).Decode(ac); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal endpoint configuration %q: %v", id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, ac, req.Config)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read endpoint configuration %q data source", id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, props, err := utils.Describe("endpointConfiguration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	_, hookProps, err := utils.Describe("endpointHook")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	hooks, hooksProps, err := utils.Describe("endpointHooks")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	reloadInfo, reloadInfoProps, err := utils.Describe("endpointReloadInfo")
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
			"kind": schema.StringAttribute{
				MarkdownDescription: props["kind"],
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"authority_id": schema.StringAttribute{
				MarkdownDescription: props["authorityID"],
				Computed:            true,
			},
			"provisioner_name": schema.StringAttribute{
				MarkdownDescription: props["provisioner"],
				Computed:            true,
			},
			"key_info": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: keyInfo,
				Attributes: map[string]schema.Attribute{
					"format": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["format"],
					},
					"pub_file": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["pubFile"],
					},
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["type"],
					},
				},
			},
			"reload_info": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: reloadInfo,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["method"],
					},
					"pid_file": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["pidFile"],
					},
					"signal": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["signal"],
					},
				},
			},
			"hooks": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: hooks,
				Attributes: map[string]schema.Attribute{
					"sign": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: hooksProps["sign"],
						Attributes: map[string]schema.Attribute{
							"shell": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: hookProps["shell"],
							},
							"before": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["before"],
							},
							"after": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["after"],
							},
							"on_error": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["onError"],
							},
						},
					},
					"renew": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: hooksProps["renew"],
						Attributes: map[string]schema.Attribute{
							"shell": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: hookProps["shell"],
							},
							"before": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["before"],
							},
							"after": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["after"],
							},
							"on_error": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: hookProps["onError"],
							},
						},
					},
				},
			},
			"certificate_info": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: certInfo,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: certInfoProps["type"],
						Computed:            true,
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certInfoProps["duration"],
						Computed:            true,
					},
					"crt_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["crtFile"],
						Computed:            true,
					},
					"key_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["keyFile"],
						Computed:            true,
					},
					"root_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["rootFile"],
						Computed:            true,
					},
					"uid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["uid"],
						Computed:            true,
					},
					"gid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["gid"],
						Computed:            true,
					},
					"mode": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["mode"],
						Computed:            true,
					},
				},
			},
		},
	}
}
