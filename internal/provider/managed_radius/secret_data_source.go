package managed_radius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*SecretDataSource)(nil)

func NewSecretDataSource() datasource.DataSource {
	return &SecretDataSource{}
}

// By putting the secret in a separate data source users must opt-in to storing
// the secret in state.
type SecretDataSource struct {
	client *v20250101.Client
}

func (a *SecretDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "smallstep_managed_radius_secret"
}

// Configure adds the Smallstep API client to the data source.
func (ds *SecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	ds.client = client
}

func (d *SecretDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read the secret required to configure a network access server to connect to a managed RADIUS server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of a managed RADIUS resource.",
				Required:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The secret a network access server needs to authenticate to a managed RADIUS server.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (ds *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id string
	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := ds.client.GetManagedRadius(ctx, id, &v20250101.GetManagedRadiusParams{Secret: utils.Ref(true)})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read managed radius secret %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading managed radius %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	managedRadius := &v20250101.ManagedRadius{}
	if err := json.NewDecoder(httpResp.Body).Decode(managedRadius); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal managed radius %s: %v", id, err),
		)
		return
	}

	diags = resp.State.SetAttribute(ctx, path.Root("secret"), utils.Deref(managedRadius.Secret))
	resp.Diagnostics.Append(diags...)
}
