package managed_configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	resource.AddTestSweepers("smallstep_managed_configuration", &resource.Sweeper{
		Name: "smallstep_managed_configuration",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListManagedConfigurations(ctx, &v20230301.ListManagedConfigurationsParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list configurations: %d: %s", resp.StatusCode, body)
			}
			var list []*v20230301.ManagedConfiguration
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, mc := range list {
				if !strings.HasPrefix(mc.Name, "tfprovider") {
					continue
				}
				resp, err := client.DeleteManagedConfiguration(ctx, *mc.Id, &v20230301.DeleteManagedConfigurationParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete managed configuration %q: %d: %s", mc.Name, resp.StatusCode, body)
				}
				log.Printf("Successfully swept %s\n", mc.Name)
			}

			return nil
		},
	})
}
