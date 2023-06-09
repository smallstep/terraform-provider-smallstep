---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep Provider"
subcategory: ""
description: |-
  
---

# smallstep Provider



## Example Usage

```terraform
provider "smallstep" {
  bearer_token = "ey..." // ignored if client_certificate is provided

  client_certificate = {
    certificate = file("api.crt")
    private_key = file("api.key")
    team_id     = "94a7dd82-1360-4493-b1bf-b14a97c45786"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `bearer_token` (String, Sensitive) Credential used to authenticate to the Smallstep API. May also be provided via the SMALLSTEP_API_TOKEN environment variable. Use the Smallstep dashboard to manage API tokens. Ignored if a client certificate is set.
- `client_certificate` (Attributes) Get an API token with a client certificate key pair signed by your trusted root. Use the Smallstep dashboard to manage trusted roots. (see [below for nested schema](#nestedatt--client_certificate))

<a id="nestedatt--client_certificate"></a>
### Nested Schema for `client_certificate`

Required:

- `certificate` (String) The PEM encoded certificate signed by your trusted root.
- `private_key` (String) The PEM encoded private key
- `team_id` (String) Your team's UUID
