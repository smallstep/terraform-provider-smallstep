package agent_configuration

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
	resource.AddTestSweepers("smallstep_agent_configuration", &resource.Sweeper{
		Name: "smallstep_agent_configuration",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListAgentConfigurations(ctx, &v20230301.ListAgentConfigurationsParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list agents: %d: %s", resp.StatusCode, body)
			}
			var list []*v20230301.AgentConfiguration
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, ec := range list {
				if !strings.HasPrefix(ec.Name, "tfprovider") {
					continue
				}
				resp, err := client.DeleteAgentConfiguration(ctx, *ec.Id, &v20230301.DeleteAgentConfigurationParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete agent configuration %q: %d: %s", ec.Name, resp.StatusCode, body)
				}
				log.Printf("Successfully swept %s\n", ec.Name)
			}

			return nil
		},
	})
}
