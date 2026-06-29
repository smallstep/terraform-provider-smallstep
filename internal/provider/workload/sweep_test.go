package workload

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
	helper.AddTestSweepers("smallstep_workload", &helper.Sweeper{
		Name: "smallstep_workload",
		F: func(region string) error {
			ctx := context.Background()

			client, err := utils.SmallstepAPIClientV20260501FromEnv()
			if err != nil {
				return err
			}

			httpResp, err := client.ListWorkloads(ctx, &v20260501.ListWorkloadsParams{})
			if err != nil {
				return err
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(httpResp.Body)
				return fmt.Errorf("failed to list workloads: %d: %s", httpResp.StatusCode, body)
			}

			var list []*v20260501.Workload
			if err := json.NewDecoder(httpResp.Body).Decode(&list); err != nil {
				return err
			}

			for _, workload := range list {
				if !strings.HasPrefix(*workload.Name, "tfprovider") {
					continue
				}

				resp, err := client.DeleteWorkload(ctx, *workload.Id, &v20260501.DeleteWorkloadParams{})
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("failed to delete workload %q: %d: %s", *workload.Name, resp.StatusCode, body)
					continue
				}
				log.Printf("Successfully swept workload %s\n", *workload.Name)
			}

			return nil
		},
	})
}

func TestMain(m *testing.M) {
	helper.TestMain(m)
}
