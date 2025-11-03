package managed_radius

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
		NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
		NewSecretDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

func init() {
	helper.AddTestSweepers("smallstep_managed_radius", &helper.Sweeper{
		Name: "smallstep_managed_radius",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListManagedRadius(ctx, &v20250101.ListManagedRadiusParams{})
			if err != nil {
				return fmt.Errorf("list managed radius: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read list managed radius response body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to list managed radius: %d: %s", resp.StatusCode, body)
			}

			var list []*v20250101.ManagedRadius
			if err := json.Unmarshal(body, &list); err != nil {
				return fmt.Errorf("failed to parse managed radius list: %w", err)
			}

			for _, radius := range list {
				if !strings.HasPrefix(radius.Name, "tfprovider") {
					continue
				}

				resp, err := client.DeleteManagedRadius(ctx, *radius.Id, &v20250101.DeleteManagedRadiusParams{})
				if err != nil {
					return fmt.Errorf("failed to delete managed radius server %q: %w", *radius.Id, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete managed radius server %q: %d: %s", *radius.Id, resp.StatusCode, body)
				}
				log.Printf("Successfully swept managed radius %s\n", radius.Name)
			}

			return nil
		},
	})
}
