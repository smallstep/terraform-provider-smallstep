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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	resource.AddTestSweepers("smallstep_collection", &resource.Sweeper{
		Name: "smallstep_collection",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientFromEnv()
			if err != nil {
				return err
			}

			resp, err := client.ListCollections(ctx, &v20230301.ListCollectionsParams{})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list collections: %d: %s", resp.StatusCode, body)
			}
			var list []*v20230301.Collection
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			for _, collection := range list {
				if !strings.HasPrefix(collection.Slug, "tfprovider") {
					continue
				}
				// Don't delete collections that may be used by running tests
				age := time.Minute
				if sweepAge := os.Getenv("SWEEP_AGE"); sweepAge != "" {
					d, err := time.ParseDuration(sweepAge)
					if err != nil {
						return err
					}
					age = d
				}
				if collection.CreatedAt.After(time.Now().Add(age * -1)) {
					continue
				}
				resp, err := client.DeleteCollection(ctx, collection.Slug, &v20230301.DeleteCollectionParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to delete collection %q: %d: %s", collection.Slug, resp.StatusCode, body)
				}
				log.Printf("Successfully swept collection %s\n", collection.Slug)
			}

			return nil
		},
	})
}

func TestAccCollectionResource(t *testing.T) {
	t.Parallel()

	slug := utils.Slug(t)
	config := fmt.Sprintf(`
resource "smallstep_collection" "employees" {
	slug = %q
	display_name = "Current Employees"
	schema_uri = "https://schema.infra.smallstep.com/manage-workloads/device/aws-vm"
}`, slug)

	updatedConfig := fmt.Sprintf(`
resource "smallstep_collection" "employees" {
	slug = %q
	display_name = "Employees"
}`, slug)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug),
					resource.TestCheckResourceAttr("smallstep_collection.employees", "display_name", "Current Employees"),
					resource.TestCheckResourceAttr("smallstep_collection.employees", "schema_uri", "https://schema.infra.smallstep.com/manage-workloads/device/aws-vm"),
					resource.TestCheckResourceAttr("smallstep_collection.employees", "instance_count", "0"),
					resource.TestMatchResourceAttr("smallstep_collection.employees", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("smallstep_collection.employees", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection.employees", "display_name", "Employees"),
					resource.TestCheckNoResourceAttr("smallstep_collection.employees", "schema_uri"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_collection.employees", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      "smallstep_collection.employees",
				ImportState:       true,
				ImportStateId:     slug,
				ImportStateVerify: true,
			},
		},
	})

	slug2 := utils.Slug(t)
	emptyConfig := fmt.Sprintf(`
resource "smallstep_collection" "employees" {
	slug = %q
	display_name = ""
	schema_uri = ""
}`, slug2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: emptyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug2),
					resource.TestCheckResourceAttr("smallstep_collection.employees", "display_name", ""),
					resource.TestCheckResourceAttr("smallstep_collection.employees", "schema_uri", ""),
				),
			},
		},
	})

	slug3 := utils.Slug(t)
	nullConfig := fmt.Sprintf(`
resource "smallstep_collection" "employees" {
	slug = %q
}`, slug3)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: nullConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug3),
					resource.TestCheckNoResourceAttr("smallstep_collection.employees", "display_name"),
					resource.TestCheckNoResourceAttr("smallstep_collection.employees", "schema_uri"),
				),
			},
		},
	})
}
