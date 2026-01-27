package wifi

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
	helper.AddTestSweepers("smallstep_wifi", &helper.Sweeper{
		Name: "smallstep_wifi",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListWifi(ctx, &v20250101.ListWifiParams{})
			if err != nil {
				return fmt.Errorf("list wifi: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read list wifi response body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to list wifi: %d: %s", resp.StatusCode, body)
			}

			var list []*v20250101.Wifi
			if err := json.Unmarshal(body, &list); err != nil {
				return fmt.Errorf("failed to parse wifi list: %w", err)
			}

			for _, wifi := range list {
				if wifi.Name == nil || !strings.HasPrefix(*wifi.Name, "tfprovider") {
					continue
				}

				resp, err := client.DeleteWifi(ctx, *wifi.Id, &v20250101.DeleteWifiParams{})
				if err != nil {
					return fmt.Errorf("failed to delete wifi %q: %w", *wifi.Id, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete wifi %q: %d: %s", *wifi.Id, resp.StatusCode, body)
				}
				log.Printf("Successfully swept wifi %s\n", *wifi.Name)
			}

			return nil
		},
	})
}
