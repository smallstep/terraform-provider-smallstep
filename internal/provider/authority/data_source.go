package authority

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource implements data.smallstep_authority
type DataSource struct {
	client *v20231101.Client
}

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = authorityTypeName
}

// Configure adds the Smallstep API client to the data source.
func (a *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	a.client = client
}

func (a *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" {
		id = data.Domain.ValueString()
	}

	httpResp, err := a.client.GetAuthority(ctx, id, &v20231101.GetAuthorityParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read authority %s: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading authority %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	authority := &v20231101.Authority{}
	if err := json.NewDecoder(httpResp.Body).Decode(authority); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal authority %s: %v", data.ID.String(), err),
		)
		return
	}

	data.ID = types.StringValue(authority.Id)
	data.Name = types.StringValue(authority.Name)
	data.Type = types.StringValue(string(authority.Type))
	data.Domain = types.StringValue(authority.Domain)
	data.Fingerprint = types.StringValue(utils.Deref(authority.Fingerprint))
	data.Root = types.StringValue(utils.Deref(authority.Root))
	data.CreatedAt = types.StringValue(authority.CreatedAt.Format(time.RFC3339))
	data.ActiveRevocation = types.BoolValue(utils.Deref(authority.ActiveRevocation))
	var adminEmails []attr.Value
	if authority.AdminEmails != nil {
		for _, email := range *authority.AdminEmails {
			adminEmails = append(adminEmails, types.StringValue(email))
		}
	}
	adminEmailsSet, diags := types.SetValue(types.StringType, adminEmails)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.AdminEmails = adminEmailsSet

	tflog.Trace(ctx, fmt.Sprintf("read authority %q data source", data.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, properties, err := utils.Describe("authority")
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
				MarkdownDescription: properties["id"],
				Optional:            true,
				Computed:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: properties["domain"],
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: properties["name"],
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: properties["type"],
				Computed:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: properties["fingerprint"],
				Computed:            true,
			},
			"root": schema.StringAttribute{
				MarkdownDescription: properties["root"],
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: properties["createdAt"],
				Computed:            true,
			},
			"active_revocation": schema.BoolAttribute{
				MarkdownDescription: properties["activeRevocation"],
				Computed:            true,
			},
			"admin_emails": schema.SetAttribute{
				MarkdownDescription: properties["adminEmails"],
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}
