---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_device Resource - terraform-provider-smallstep"
subcategory: ""
description: |-
  A device represents a computer that can be issued x509 certificates for securely connecting to company resources.
---

# smallstep_device (Resource)

A device represents a computer that can be issued x509 certificates for securely connecting to company resources.

## Example Usage

```terraform
resource "smallstep_device" "laptop1" {
  permanent_identifier = "782BF520"
  display_id           = "99-TUR"
  display_name         = "Employee Laptop"
  serial               = "6789"
  tags                 = ["ubuntu"]
  metadata = {
    k1 = "v1"
  }
  os        = "Linux"
  ownership = "company"
  user = {
    email = "user@example.com"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `permanent_identifier` (String) The cryptographic identity of the device. High-assurance devices are only issued certificates when this identifier is attested by a trusted source. All devices with the same permanent identifier appear as a single device in the Smallstep API. For Windows and Linux devices this is the hash of the TPM endorsement key and for Apple devices it is the serial number.

### Optional

- `display_id` (String) An opaque identifier that may be used to link this device to an external inventory.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.
- `display_name` (String) A friendly name for the device.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.
- `metadata` (Map of String) A map of key-value pairs available as template data when a provisioner with a webhook is used to issue a certificate to a device.
- `os` (String) The device operating system.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.
 Allowed values: `Linux` `Windows` `macOS` `iOS` `tvOS` `watchOS` `visionOS`
- `ownership` (String) Whether the device is owned by the user or the company.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.
 Allowed values: `company` `user`
- `serial` (String) The serial number of the device.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.
- `tags` (Set of String) A set of tags that can be used to group devices.
- `user` (Attributes) The user that a device is assigned to. A device cannot be approved for high-assurance certificates until a user has been assigned to it. (see [below for nested schema](#nestedatt--user))

### Read-Only

- `approved_at` (String) Timestamp in RFC3339 format when the device was approved to connect to company resources. Read only.
- `connected` (Boolean) The device is currently connected to Smallstep. Read only.
- `enrolled_at` (String) Timestamp in RFC3339 format when the device first connected to Smallstep. Read only.
- `high_assurance` (Boolean) The device has been issued certificates using high assurance device attestation. Read only.
- `host_id` (String) The identifier for the smallstep agent on the device.
- `id` (String) A UUID identifying this device. Read only.
- `last_seen` (String) Timestamp in RFC3339 format when the device last connected to Smallstep. Read only.

<a id="nestedatt--user"></a>
### Nested Schema for `user`

Required:

- `email` (String) Email of the user the device is assigned to.
This field may be populated with a value derived from data synced from your team's MDMs.
Setting this value explicitly will mask any MDM-derived value.

Optional:

- `display_name` (String) Full name of the user the device is assigned to. Synced from team's identity provider. Read only.

## Import

Import is supported using the following syntax:

```shell
terraform import smallstep_device.laptop_12 b1161f78-d251-401e-b17c-fe38fc26ae7b
```
