package endpoint_configuration

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
	resource.AddTestSweepers("smallstep_endpoint_configuration", &resource.Sweeper{
		Name: "smallstep_endpoint_configuration",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListEndpointConfigurations(ctx, &v20230301.ListEndpointConfigurationsParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list endpoints: %d: %s", resp.StatusCode, body)
			}
			var list []*v20230301.EndpointConfiguration
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, ec := range list {
				if !strings.HasPrefix(ec.Name, "tfprovider") {
					continue
				}
				resp, err := client.DeleteEndpointConfiguration(ctx, *ec.Id, &v20230301.DeleteEndpointConfigurationParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete endpoint configuration %q: %d: %s", ec.Name, resp.StatusCode, body)
				}
				log.Printf("Successfully swept %s\n", ec.Name)
			}

			return nil
		},
	})
}
