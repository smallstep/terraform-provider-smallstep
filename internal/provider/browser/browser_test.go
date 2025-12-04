package browser

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
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

func init() {
	helper.AddTestSweepers("smallstep_browser", &helper.Sweeper{
		Name: "smallstep_browser",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListBrowser(ctx, &v20250101.ListBrowserParams{})
			if err != nil {
				return fmt.Errorf("list browsers: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read list browsers response body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to list browsers: %d: %s", resp.StatusCode, body)
			}

			var list []*v20250101.Browser
			if err := json.Unmarshal(body, &list); err != nil {
				return fmt.Errorf("failed to parse browsers list: %w", err)
			}

			for _, browser := range list {
				if browser.Name == nil || !strings.HasPrefix(*browser.Name, "tfprovider") {
					continue
				}

				resp, err := client.DeleteBrowser(ctx, *browser.Id, &v20250101.DeleteBrowserParams{})
				if err != nil {
					return fmt.Errorf("failed to delete browser %q: %w", *browser.Id, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete browser %q: %d: %s", *browser.Id, resp.StatusCode, body)
				}
				log.Printf("Successfully swept browser %s\n", *browser.Name)
			}

			return nil
		},
	})
}
