package vpn

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
	helper.AddTestSweepers("smallstep_vpn", &helper.Sweeper{
		Name: "smallstep_vpn",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListVpn(ctx, &v20250101.ListVpnParams{})
			if err != nil {
				return fmt.Errorf("list vpn: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read list vpn response body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to list vpn: %d: %s", resp.StatusCode, body)
			}

			var list []*v20250101.Vpn
			if err := json.Unmarshal(body, &list); err != nil {
				return fmt.Errorf("failed to parse vpn list: %w", err)
			}

			for _, vpn := range list {
				if vpn.Name == nil || !strings.HasPrefix(*vpn.Name, "tfprovider") {
					continue
				}

				resp, err := client.DeleteVpn(ctx, *vpn.Id, &v20250101.DeleteVpnParams{})
				if err != nil {
					return fmt.Errorf("failed to delete vpn %q: %w", *vpn.Id, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete vpn %q: %d: %s", *vpn.Id, resp.StatusCode, body)
				}
				log.Printf("Successfully swept vpn %s\n", *vpn.Name)
			}

			return nil
		},
	})
}
