package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAuthorityDataSource(t *testing.T) {
	t.Parallel()
	authority := newAuthority(t)
	config := fmt.Sprintf(`
data "smallstep_authority" "test" {
	id = "%s"
}
`, authority.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "id", authority.Id),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "name", authority.Name),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "type", string(authority.Type)),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "active_revocation", fmt.Sprintf("%t", deref(authority.ActiveRevocation))),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "domain", authority.Domain),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "created_at", authority.CreatedAt.Format(time.RFC3339)),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "fingerprint", *authority.Fingerprint),
				),
			},
		},
	})
}
