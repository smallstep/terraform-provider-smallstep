package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccAuthorityDataSource(t *testing.T) {
	t.Parallel()
	authority := utils.NewAuthority(t)
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
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "active_revocation", fmt.Sprintf("%t", utils.Deref(authority.ActiveRevocation))),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "domain", authority.Domain),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "created_at", authority.CreatedAt.Format(time.RFC3339)),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "fingerprint", *authority.Fingerprint),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "root", *authority.Root),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "admin_emails.0", (*authority.AdminEmails)[0]),
				),
			},
		},
	})
}
