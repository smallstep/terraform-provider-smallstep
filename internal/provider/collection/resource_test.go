package collection

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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func init() {
	helper.AddTestSweepers("smallstep_collection", &helper.Sweeper{
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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_collection.employees", "display_name", "Current Employees"),
					helper.TestCheckResourceAttr("smallstep_collection.employees", "schema_uri", "https://schema.infra.smallstep.com/manage-workloads/device/aws-vm"),
					helper.TestCheckResourceAttr("smallstep_collection.employees", "instance_count", "0"),
					helper.TestMatchResourceAttr("smallstep_collection.employees", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestMatchResourceAttr("smallstep_collection.employees", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				Config: updatedConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection.employees", "display_name", "Employees"),
					helper.TestCheckNoResourceAttr("smallstep_collection.employees", "schema_uri"),
				),
				ConfigPlanChecks: helper.ConfigPlanChecks{
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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug2),
					helper.TestCheckResourceAttr("smallstep_collection.employees", "display_name", ""),
					helper.TestCheckResourceAttr("smallstep_collection.employees", "schema_uri", ""),
				),
			},
		},
	})

	slug3 := utils.Slug(t)
	nullConfig := fmt.Sprintf(`
resource "smallstep_collection" "employees" {
	slug = %q
}`, slug3)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: nullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection.employees", "slug", slug3),
					helper.TestCheckNoResourceAttr("smallstep_collection.employees", "display_name"),
					helper.TestCheckNoResourceAttr("smallstep_collection.employees", "schema_uri"),
				),
			},
		},
	})
}
