---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_provisioner_webhook Resource - terraform-provider-smallstep"
subcategory: ""
description: |-
  A webhook https://smallstep.com/docs/step-ca/webhooks/ to call when a certificate request is being processed.
---

# smallstep_provisioner_webhook (Resource)

A [webhook](https://smallstep.com/docs/step-ca/webhooks/) to call when a certificate request is being processed.

## Example Usage

```terraform
resource "smallstep_provisioner_webhook" "external" {
  authority_id   = smallstep_authority.foo.id
  provisioner_id = smallstep_provisioner.bar.id
  name           = "devices"
  kind           = "ENRICHING"
  cert_type      = "X509"
  server_type    = "EXTERNAL"
  url            = "https://example.com/hook"
  bearer_token   = "secret123"
}

resource "smallstep_provisioner_webhook" "devices" {
  authority_id    = smallstep_authority.foo.id
  provisioner_id  = smallstep_provisioner.bar.id
  name            = "devices"
  kind            = "ENRICHING"
  cert_type       = "X509"
  server_type     = "HOSTED_ATTESTATION"
  collection_slug = smallstep_collection.tpms.slug
  depends_on      = [smallstep_collection.tpms]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `authority_id` (String)
- `cert_type` (String) Allowed values: `ALL` `X509` `SSH`
- `kind` (String) The webhook kind indicates how and when it is called.

ENRICHING webhooks are called before rendering the certificate template. They have two functions. First, they must allow the certificate request or it will be aborted. Second, they can return additional data to be referenced in the certificate template. The payload sent to the webhook server varies based on whether an X509 or SSH certificate is to be signed and based on the type of provisioner.
 Allowed values: `ENRICHING`
- `name` (String) The name of the webhook. For `ENRICHING` webhooks, the returned data can be referenced in the certificate under the path `.Webhooks.<name>`. Must be unique to the provisioner.
- `provisioner_id` (String)
- `server_type` (String) An EXTERNAL webhook server is not operated by Smallstep. The caller must use the returned ID and secret to configure the server.

A HOSTED_ATTESTATION webhook server is hosted by Smallstep and must be used with an `ENRICHING` webhook type and an ACME Attestation provisioner. The webhook server will verify the attested permanent identifier exists as the ID of an instance in the configured collection. The data of the instance in the collection will be added to the template data.
 Allowed values: `EXTERNAL` `HOSTED_ATTESTATION`

### Optional

- `basic_auth` (Attributes) Configures provisioner webhook requests to include an Authorization header with these credentials. Optional for `EXTERNAL` webhook servers; not allowed with hosted webhook servers. At most one of `bearerToken` and `basicAuth` may be set. (see [below for nested schema](#nestedatt--basic_auth))
- `bearer_token` (String, Sensitive) Webhook requests will include an Authorization header with the token. Optional for `EXTERNAL` webhook servers; not allowed with hosted webhook servers. At most one of `bearerToken` and `basicAuth` may be set.
- `collection_slug` (String) For HOSTED_ATTESTATION webhooks, the collectionSlug is a reference to the collection that holds the devices that may be issued certificates. This collection must already exist. Required for `HOSTED_ATTESTATION` webhook servers; not allowed for `EXTERNAL`.
- `disable_tls_client_auth` (Boolean) The CA will not send a client certificate when requested by the webhook server. Optional for `EXTERNAL` webhook servers; not allowed with hosted webhook servers.
- `url` (String) The URL of the webhook server. Required for `EXTERNAL` webhook servers; read-only for hosted webhook servers.

### Read-Only

- `id` (String) UUID identifying this webhook. Generated server-side when the webhook is created. Will be sent to the webhook server in every request in the `X-Smallstep-Webhook-ID` header.
- `secret` (String, Sensitive) The shared secret used to authenticate the payload sent to the webhook server. Generated server-side. This is returned only for `EXTERNAL` webhook servers and only once, at the time of creation.

<a id="nestedatt--basic_auth"></a>
### Nested Schema for `basic_auth`

Required:

- `password` (String, Sensitive)
- `username` (String, Sensitive)

## Import

Import is supported using the following syntax:

```shell
# <authority_id>/<provisioner_id>/<name>
AUTHORITY_ID=ed2e4f38-fd2d-4eb0-9280-52b697636873
PROVISIONER_ID=57b8ade4-5873-4a15-911c-a4fff5999600

terraform import smallstep_provisioner_webhook.devices ${AUTHORITY_ID}/${PROVISIONER_ID}/devices
```
