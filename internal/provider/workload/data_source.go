package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
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
	workload, workloadProps, err := utils.DescribeV20260501("workload")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Workload Schema", err.Error())
		return
	}

	_, credentialProps, err := utils.DescribeV20260501("workloadCredential")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Workload Credential Schema", err.Error())
		return
	}

	_, probeProps, err := utils.DescribeV20260501("probe")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Probe Schema", err.Error())
		return
	}

	_, hooksProps, err := utils.DescribeV20260501("endpointHooks")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Hooks Schema", err.Error())
		return
	}

	_, hookProps, err := utils.DescribeV20260501("endpointHook")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Hook Schema", err.Error())
		return
	}

	_, reloadInfoProps, err := utils.DescribeV20260501("endpointReloadInfo")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Reload Info Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: workload,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: workloadProps["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: workloadProps["name"],
				Computed:            true,
			},
			"workload_type": schema.StringAttribute{
				MarkdownDescription: workloadProps["workloadType"],
				Computed:            true,
			},
			"credentials": schema.ListNestedAttribute{
				MarkdownDescription: workloadProps["credentials"],
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"credential_id": schema.StringAttribute{
							MarkdownDescription: credentialProps["credentialId"],
							Computed:            true,
						},
						"probes": schema.ListNestedAttribute{
							MarkdownDescription: credentialProps["probes"],
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"target": schema.StringAttribute{
										MarkdownDescription: probeProps["target"],
										Computed:            true,
									},
									"protocol": schema.StringAttribute{
										MarkdownDescription: probeProps["protocol"],
										Computed:            true,
									},
									"crt_file": schema.StringAttribute{
										MarkdownDescription: probeProps["crtFile"],
										Computed:            true,
									},
									"key_file": schema.StringAttribute{
										MarkdownDescription: probeProps["keyFile"],
										Computed:            true,
									},
									"root_file": schema.StringAttribute{
										MarkdownDescription: probeProps["rootFile"],
										Computed:            true,
									},
									"server_name": schema.StringAttribute{
										MarkdownDescription: probeProps["serverName"],
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
			"hooks": schema.SingleNestedAttribute{
				MarkdownDescription: hooksProps["hooks"],
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"sign": schema.SingleNestedAttribute{
						MarkdownDescription: hooksProps["sign"],
						Computed:            true,
						Attributes: map[string]schema.Attribute{
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
							"shell": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: hookProps["shell"],
							},
						},
					},
					"renew": schema.SingleNestedAttribute{
						MarkdownDescription: hooksProps["renew"],
						Computed:            true,
						Attributes: map[string]schema.Attribute{
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
							"shell": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: hookProps["shell"],
							},
						},
					},
				},
			},
			"reload_info": schema.SingleNestedAttribute{
				MarkdownDescription: reloadInfoProps["reloadInfo"],
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["method"],
						Computed:            true,
					},
					"pid_file": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["pidFile"],
						Computed:            true,
					},
					"signal": schema.Int64Attribute{
						MarkdownDescription: reloadInfoProps["signal"],
						Computed:            true,
					},
					"unit_name": schema.StringAttribute{
						MarkdownDescription: reloadInfoProps["unitName"],
						Computed:            true,
					},
				},
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
	var config *Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workloadID := config.ID.ValueString()
	if workloadID == "" {
		resp.Diagnostics.AddError("Invalid Read Workload Request", "Workload ID is required")
		return
	}

	httpResp, err := ds.client.GetWorkload(ctx, workloadID, &v20260501.GetWorkloadParams{})
	if err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to read workload: %v", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError("Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading workload: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)))
		return
	}

	workload := &v20260501.Workload{}
	if err := json.NewDecoder(httpResp.Body).Decode(workload); err != nil {
		resp.Diagnostics.AddError("Smallstep API Client Error", fmt.Sprintf("Failed to unmarshal workload: %v", err))
		return
	}

	model, d := fromAPI(ctx, workload)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
