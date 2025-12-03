package identity_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*IdentityProviderDataSource)(nil)

func NewIdentityProviderDataSource() datasource.DataSource {
	return &IdentityProviderDataSource{}
}

type IdentityProviderDataSource struct {
	client *v20250101.Client
}

func (a *IdentityProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = idp_name
}

// Configure adds the Smallstep API client to the data source.
func (ds *IdentityProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	idp, props, err := utils.Describe("identityProvider")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Identity Provider Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: idp,

		Attributes: map[string]schema.Attribute{
			"trust_roots": schema.StringAttribute{
				MarkdownDescription: props["trustRoots"],
				Computed:            true,
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: props["issuer"],
				Computed:            true,
			},
			"authorize_endpoint": schema.StringAttribute{
				MarkdownDescription: props["authorizeEndpoint"],
				Computed:            true,
			},
			"token_endpoint": schema.StringAttribute{
				MarkdownDescription: props["tokenEndpoint"],
				Computed:            true,
			},
			"jwks_endpoint": schema.StringAttribute{
				MarkdownDescription: props["jwksEndpoint"],
				Computed:            true,
			},
		},
	}
}

func (ds *IdentityProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	httpResp, err := ds.client.GetIdentityProvider(ctx, &v20250101.GetIdentityProviderParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read identity provider: %v", err),
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
			fmt.Sprintf("Request %q received status %d reading identity provider: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	idp := &v20250101.IdentityProvider{}
	if err := json.NewDecoder(httpResp.Body).Decode(idp); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal identity provider: %v", err),
		)
		return
	}

	remote := idpFromAPI(idp)

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
