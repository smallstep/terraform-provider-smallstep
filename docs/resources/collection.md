---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_collection Resource - terraform-provider-smallstep"
subcategory: ""
description: |-
  A collection of instances.
---

# smallstep_collection (Resource)

A collection of instances.

## Example Usage

```terraform
resource "smallstep_collection" "tpms" {
  slug = "tpms"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `slug` (String) A lowercase name identifying the collection.

### Optional

- `display_name` (String) A user-friendly name for the collection.
- `schema_uri` (String) Reference to a schema that all instances in the collection must conform to.

### Read-Only

- `created_at` (String) Timestamp in RFC3339 format when the collections was created
- `instance_count` (Number) The number of instances in the collection.
- `updated_at` (String) Timestamp in RFC3339 format when the collections was last updated

## Import

Import is supported using the following syntax:

```shell
terraform import smallstep_collection.devices devices
```
