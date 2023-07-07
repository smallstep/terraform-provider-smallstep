---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_agent_configuration Data Source - terraform-provider-smallstep"
subcategory: ""
description: |-
  The agent configuration describes the attestation authority used by the agent to grant workload certificates.
---

# smallstep_agent_configuration (Data Source)

The agent configuration describes the attestation authority used by the agent to grant workload certificates.

## Example Usage

```terraform
data "smallstep_agent_configuration" "agent1" {
  id = "0496154c-ea90-4642-a2b9-96e76e69d219"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) A UUID identifying this agent configuration. Generated server-side on creation.

### Read-Only

- `attestation_slug` (String) The slug of the attestation authority the agent connects to to get a certificate.
- `authority_id` (String) UUID identifying the authority the agent uses to generate endpoint certificates.
- `name` (String) The name of this agent configuration.
- `provisioner_name` (String) The name of the provisioner on the authority the agent uses to generate endpoint certificates.

