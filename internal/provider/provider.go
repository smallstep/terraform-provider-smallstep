package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
)

// Ensure SmallstepProvider satisfies various provider interfaces.
var _ provider.Provider = &SmallstepProvider{}

// SmallstepProvider defines the provider implementation.
type SmallstepProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SmallstepProviderModel describes the provider data model.
type SmallstepProviderModel struct {
	BearerToken types.String `tfsdk:"bearer_token"`
}

func (p *SmallstepProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "smallstep"
	resp.Version = p.version
}

func (p *SmallstepProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: `
Credential used to authenticate to the Smallstep API.
May also be provided via the SMALLSTEP_API_TOKEN environment variable.
`,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *SmallstepProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SmallstepProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.BearerToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("bearer_token"),
			"Unknown Smallstep bearer_token",
			"The provider cannot connect to the Smallstep API since the bearer_token is unknown",
		)
		return
	}

	token := os.Getenv("SMALLSTEP_API_TOKEN")
	if !data.BearerToken.IsNull() {
		token = data.BearerToken.ValueString()
	}

	server := os.Getenv("SMALLSTEP_API_URL")
	if server == "" {
		server = "https://smallstep.com"
	}
	client, err := v20230301.NewClient(server, v20230301.WithRequestEditorFn(v20230301.RequestEditorFn(func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	})))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Smallstep API client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SmallstepProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuthorityResource,
	}
}

func (p *SmallstepProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuthorityDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SmallstepProvider{
			version: version,
		}
	}
}
