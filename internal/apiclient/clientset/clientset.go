package clientset

import (
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	v20260501 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20260501"
)

// Clients holds every versioned Smallstep API client. It is the value passed as
// the provider's ResourceData/DataSourceData.
type Clients struct {
	V20250101 *v20250101.Client
	V20260501 *v20260501.Client
}
