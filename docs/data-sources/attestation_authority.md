---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_attestation_authority Data Source - terraform-provider-smallstep"
subcategory: ""
description: |-
  An attestation authority used with the device-attest-01 ACME challenge to verify a device's hardware identity. This object is experimental and subject to change.
---

# smallstep_attestation_authority (Data Source)

An attestation authority used with the device-attest-01 ACME challenge to verify a device's hardware identity. This object is experimental and subject to change.

## Example Usage

```terraform
data "smallstep_attestation_authority" "aa" {
  id = "4958f125-8e2a-4c99-8c32-832b25e5569e"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) A UUID identifying this attestation authority. Read only.

### Read-Only

- `attestor_intermediates` (String) The pem-encoded list of intermediate certificates used to build a chain of trust to verify the attestation certificates submitted by devices.
- `attestor_roots` (String) The pem-encoded list of certificates used to verify the attestation certificates submitted by devices.
- `created_at` (String) Timestamp in RFC3339 format when the attestation authority was created.
- `name` (String) The name of the attestation authority.
- `root` (String) The pem-encoded root certificate of this attestation authority. This is generated server-side when the attestation authority is created. This certificate should be used in the `attestationRoots` field of an ACME_ATTESTATION provisioner with the `tpm` format.
- `slug` (String) A short name for this attestation authority. Read only.


