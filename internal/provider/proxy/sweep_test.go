package proxy

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
	helper.AddTestSweepers("smallstep_proxy", &helper.Sweeper{
		Name: "smallstep_proxy",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientV20260501FromEnv()
			if err != nil {
				return err
			}

			httpResp, err := client.ListProxy(ctx, &v20260501.ListProxyParams{})
			if err != nil {
				return err
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(httpResp.Body)
				return fmt.Errorf("failed to list proxies: %d: %s", httpResp.StatusCode, body)
			}

			var list []*v20260501.Proxy
			if err := json.NewDecoder(httpResp.Body).Decode(&list); err != nil {
				return err
			}

			for _, proxy := range list {
				if !strings.HasPrefix(*proxy.Name, "tfprovider-") {
					continue
				}

				resp, err := client.DeleteProxy(ctx, *proxy.Id, &v20260501.DeleteProxyParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("failed to delete proxy %q: %d: %s", *proxy.Name, resp.StatusCode, body)
					continue
				}
				log.Printf("Successfully swept proxy %s\n", *proxy.Name)
			}

			return nil
		},
	})
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}
