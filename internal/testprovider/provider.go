package testprovider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ provider.Provider = (*SmallstepTestProvider)(nil)

type SmallstepTestProvider struct {
	ResourceFactories   []func() resource.Resource
	DataSourceFactories []func() datasource.DataSource
}

func (p *SmallstepTestProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "smallstep"
	resp.Version = "test"
}

func (p *SmallstepTestProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *SmallstepTestProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	server := os.Getenv("SMALLSTEP_API_URL")
	if server == "" {
		server = "https://gateway.smallstep.com/api"
	}

	client, err := utils.SmallstepAPIClientFromEnv()
	if err != nil {
		resp.Diagnostics.AddError(
			"Get API client configured from environment",
			err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SmallstepTestProvider) Resources(ctx context.Context) []func() resource.Resource {
	return p.ResourceFactories
}

func (p *SmallstepTestProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.DataSourceFactories
}

func (p *SmallstepTestProvider) New() provider.Provider {
	return p
}
