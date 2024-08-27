package authority

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	resource.AddTestSweepers("smallstep_authority", &resource.Sweeper{
		Name: "smallstep_authority",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.GetAuthorities(ctx, &v20231101.GetAuthoritiesParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list authorities: %d: %s", resp.StatusCode, body)
			}
			var list []*v20231101.Authority
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, authority := range list {
				if !strings.HasPrefix(authority.Domain, "tfprovider-") {
					continue
				}
				// Don't delete authorities that may be used by running tests
				age := time.Minute
				if sweepAge := os.Getenv("SWEEP_AGE"); sweepAge != "" {
					d, err := time.ParseDuration(sweepAge)
					if err != nil {
						return err
					}
					age = d
				}
				if authority.CreatedAt.After(time.Now().Add(age * -1)) {
					continue
				}
				resp, err := client.DeleteAuthority(ctx, authority.Id, &v20231101.DeleteAuthorityParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete authority %q: %d: %s", authority.Domain, resp.StatusCode, body)
				}
				log.Printf("Successfully swept %s\n", authority.Domain)
			}

			return nil
		},
	})
}
