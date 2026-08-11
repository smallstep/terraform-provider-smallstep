package sso_integration

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
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	helper.AddTestSweepers("smallstep_sso_integration", &helper.Sweeper{
		Name: "smallstep_sso_integration",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientV20260501FromEnv()
			if err != nil {
				return err
			}

			httpResp, err := client.ListSsoIntegrations(ctx, &v20260501.ListSsoIntegrationsParams{})
			if err != nil {
				return err
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(httpResp.Body)
				return fmt.Errorf("failed to list sso integrations: %d: %s", httpResp.StatusCode, body)
			}

			var list []*v20260501.SsoIntegration
			if err := json.NewDecoder(httpResp.Body).Decode(&list); err != nil {
				return err
			}

			for _, integration := range list {
				if !strings.HasPrefix(integration.RedirectURI, "https://tfprovider.example.com") {
					continue
				}

				resp, err := client.DeleteSsoIntegration(ctx, *integration.Id, &v20260501.DeleteSsoIntegrationParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("failed to delete sso integration %q: %d: %s", *integration.Id, resp.StatusCode, body)
					continue
				}
				log.Printf("Successfully swept sso integration %s\n", *integration.Id)
			}

			return nil
		},
	})
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
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
