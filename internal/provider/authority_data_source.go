package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
)

var _ datasource.DataSourceWithConfigure = (*AuthorityDataSource)(nil)

func NewAuthorityDataSource() datasource.DataSource {
	return &AuthorityDataSource{}
}

// AuthorityDataSource implements data.smallstep_authority
type AuthorityDataSource struct {
	client *v20230301.Client
}

func (a *AuthorityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = authorityTypeName
}

// Configure adds the Smallstep API client to the data source.
func (a *AuthorityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (a *AuthorityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AuthorityDataModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.GetAuthority(ctx, data.ID.ValueString(), &v20230301.GetAuthorityParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read authority %s: %v", data.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading authority %s: %s", httpResp.StatusCode, data.ID.String(), apiErrorMsg(httpResp.Body)),
		)
		return
	}

	authority := &v20230301.Authority{}
	if err := json.NewDecoder(httpResp.Body).Decode(authority); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal authority %s: %v", data.ID.String(), err),
		)
		return
	}

	data.Name = types.StringValue(authority.Name)
	data.Type = types.StringValue(string(authority.Type))
	data.Domain = types.StringValue(authority.Domain)
	data.Fingerprint = types.StringValue(deref(authority.Fingerprint))
	data.CreatedAt = types.StringValue(authority.CreatedAt.Format(time.RFC3339))
	data.ActiveRevocation = types.BoolValue(deref(authority.ActiveRevocation))

	tflog.Trace(ctx, fmt.Sprintf("read authority %q resource", data.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *AuthorityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, properties, err := describe("authority")
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
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: properties["name"],
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: properties["type"],
				Computed:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: properties["domain"],
				Computed:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: properties["fingerprint"],
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
		},
	}
}
