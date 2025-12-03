package identity_provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewIdentityProviderResource,
		NewClientResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewIdentityProviderDataSource,
		NewClientDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

func init() {
	helper.AddTestSweepers("smallstep_identity_provider", &helper.Sweeper{
		Name: "smallstep_identity_provider",
		F: func(region string) error {
			return sweep()
		},
	})
}

func sweep() error {
	ctx := context.Background()

	client, err := utils.SmallstepAPIClientFromEnv()
	if err != nil {
		return err
	}

	resp, err := client.DeleteIdentityProvider(ctx, &v20250101.DeleteIdentityProviderParams{})
	if err != nil {
		return fmt.Errorf("delete identity provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete identity provider: %d: %s", resp.StatusCode, body)
	}

	return nil
}
