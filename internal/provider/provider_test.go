package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("SMALLSTEP_API_URL") == "" {
		t.Fatal("SMALLSTEP_API_URL environment variable is required")
	}
	if os.Getenv("SMALLSTEP_API_TOKEN") == "" {
		t.Fatal("SMALLSTEP_API_TOKEN environment variable is required")
	}
}
