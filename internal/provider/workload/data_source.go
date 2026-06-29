package workload

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
	workload, props, err := utils.DescribeV20260501("workload")
	if err != nil {
		resp.Diagnostics.AddError("Parse Smallstep OpenAPI Workload Schema", err.Error())
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: workload,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"workload_type": schema.StringAttribute{
				MarkdownDescription: props["workloadType"],
				Computed:            true,
			},
			"credentials": schema.ListAttribute{
				MarkdownDescription: props["credentials"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"hooks": schema.SingleNestedAttribute{
				MarkdownDescription: "Hooks configuration for the workload",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"sign": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"before": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"after": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"on_error": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"shell": schema.StringAttribute{
								Computed: true,
							},
						},
					},
					"renew": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"before": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"after": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"on_error": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"shell": schema.StringAttribute{
								Computed: true,
							},
						},
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

	model, _ := fromAPI(ctx, workload)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
