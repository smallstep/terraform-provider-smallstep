package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
)

func init() {
	resource.AddTestSweepers("smallstep_authority", &resource.Sweeper{
		Name: "smallstep_authority",
		F: func(region string) error {
			ctx := context.Background()

			client, err := newClient()
			if err != nil {
				return err
			}

			resp, err := client.GetAuthorities(ctx, &v20230301.GetAuthoritiesParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list authorities: %d: %s", resp.StatusCode, body)
			}
			var list []*v20230301.Authority
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, authority := range list {
				if strings.HasPrefix(authority.Domain, "keep-") {
					continue
				}
				// Keep authorities for 24 hours for debugging
				if authority.CreatedAt.After(time.Now().Add(time.Hour * -24)) {
					continue
				}
				resp, err := client.DeleteAuthority(ctx, authority.Id, &v20230301.DeleteAuthorityParams{})
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

func newAuthorityResourceConfig(slug string) string {
	return fmt.Sprintf(`
resource "smallstep_authority" "test" {
	subdomain = "%s"
	name = "%s Authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
	active_revocation = false
}
`, slug, slug)
}

func TestAccAuthorityResource(t *testing.T) {
	t.Parallel()

	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: newAuthorityResourceConfig(slug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_authority.test", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_authority.test", "domain", slug+".testacc.ca.pki.pub"),
					resource.TestMatchResourceAttr("smallstep_authority.test", "fingerprint", regexp.MustCompile(`^[0-9a-z]{64}$`)),
					resource.TestMatchResourceAttr("smallstep_authority.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
