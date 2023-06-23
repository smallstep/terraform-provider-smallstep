package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
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
				if !strings.HasPrefix(authority.Domain, "tfprovider-") {
					continue
				}
				// Don't delete authorities that may be used by running tests
				age := time.Minute
				if sweepAge := os.Getenv("SWEEP_AGE"); sweepAge != "" {
					d, err : time.ParseDuration(sweepAge)
					if err != nil {
						return err
					}
					age = d
				if authority.CreatedAt.After(time.Now().Add(age) {
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

func TestAccAuthorityResource(t *testing.T) {
	t.Parallel()

	devopsSlug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	devopsConfig := fmt.Sprintf(`
resource "smallstep_authority" "devops" {
	subdomain = "%s"
	name = "%s Authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}
`, devopsSlug, devopsSlug)

	caDomain := os.Getenv("SMALLSTEP_CA_DOMAIN")
	if caDomain == "" {
		caDomain = ".step-e2e.ca.smallstep.com"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: devopsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_authority.devops", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_authority.devops", "domain", devopsSlug+caDomain),
					resource.TestMatchResourceAttr("smallstep_authority.devops", "fingerprint", regexp.MustCompile(`^[0-9a-z]{64}$`)),
					resource.TestMatchResourceAttr("smallstep_authority.devops", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				ResourceName:      "smallstep_authority.devops",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	advancedSlug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	advancedConfig := fmt.Sprintf(`
resource "smallstep_authority" "advanced" {
	subdomain = "%s"
	name = "%s Authority"
	type = "advanced"
	admin_emails = ["andrew@smallstep.com"]
	active_revocation = true
	intermediate_issuer = {
		name = "%s Intermediate"
		key_version = "RSA_SIGN_PKCS1_2048_SHA256"
		duration = "100h"
		max_path_length = 0
		name_constraints = {
			critical = true
			permitted_ip_ranges = ["10.32.0.0/12"]
			permitted_dns_domains = [".cluster.local"]
			permitted_email_addresses = ["eng@smallstep.com"]
			permitted_uri_domains = ["uri.cluster.local"]
		}
		subject = {
			common_name = "Issuer"
			country = "US"
			email_address = "test@smallstep.com"
			locality = "San Francisco"
			organization = "Engineering"
			organizational_unit = "Core"
			postal_code = "94108"
			province = "CA"
			serial_number = "1"
			street_address = "26 O'Farrell St"
		}
	}
	root_issuer = {
		name = "Root %s"
		key_version = "RSA_SIGN_PKCS1_2048_SHA256"
		duration = "1000h"
		max_path_length = "1"
		name_constraints = {
			critical = false
			excluded_ip_ranges = ["10.96.0.0/12"]
			excluded_dns_domains = ["example.com"]
			excluded_email_addresses = ["eng@example.com"]
			excluded_uri_domains = ["uri:example.com"]
		}
	}
}
`, advancedSlug, advancedSlug, advancedSlug, advancedSlug)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: advancedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_authority.advanced", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_authority.advanced", "domain", advancedSlug+caDomain),
					resource.TestMatchResourceAttr("smallstep_authority.advanced", "fingerprint", regexp.MustCompile(`^[0-9a-z]{64}$`)),
					resource.TestMatchResourceAttr("smallstep_authority.advanced", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
