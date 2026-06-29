package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	helper.AddTestSweepers("smallstep_relay", &helper.Sweeper{
		Name: "smallstep_relay",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientV20260501FromEnv()
			if err != nil {
				return err
			}

			httpResp, err := client.ListRelays(ctx, &v20260501.ListRelaysParams{})
			if err != nil {
				return err
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(httpResp.Body)
				return fmt.Errorf("failed to list relays: %d: %s", httpResp.StatusCode, body)
			}

			var list []*v20260501.Relay
			if err := json.NewDecoder(httpResp.Body).Decode(&list); err != nil {
				return err
			}

			for _, relay := range list {
				if !strings.HasPrefix(relay.Name, "tfprovider-") {
					continue
				}

				resp, err := client.DeleteRelay(ctx, *relay.Id, &v20260501.DeleteRelayParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("failed to delete relay %q: %d: %s", relay.Name, resp.StatusCode, body)
					continue
				}
				log.Printf("Successfully swept relay %s\n", relay.Name)
			}

			return nil
		},
	})
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}
