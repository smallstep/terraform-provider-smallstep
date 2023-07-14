
resource "smallstep_collection" "tpms" {
  slug = "tpms"
}

resource "smallstep_collection_instance" "server1" {
  id              = "urn:ek:sha256:RAzbOveN1Y45fYubuTxu5jOXWtOK1HbfZ7yHjBuWlyE="
  data            = "{}"
  collection_slug = smallstep_collection.tpms.slug
  depends_on      = [smallstep_collection.tpms]
}

resource "smallstep_attestation_authority" "aa" {
  name           = "foo"
  catalog        = smallstep_collection.tpms.slug
  attestor_roots = file("${path.module}/ca.crt")
  depends_on     = [smallstep_collection.tpms]
}

resource "smallstep_authority" "agents" {
  subdomain    = "agents"
  name         = "Agents Authority"
  type         = "devops"
  admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
  authority_id = smallstep_authority.agents.id
  name         = "Agents"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["tpm"]
    attestation_roots   = [smallstep_attestation_authority.aa.root]
  }
}

resource "smallstep_authority" "endpoints" {
  subdomain    = "endpoints"
  name         = "Endpoints Authority"
  type         = "devops"
  admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "endpoints" {
  authority_id = smallstep_authority.endpoints.id
  name         = "Endpoints"
  type         = "X5C"
  x5c = {
    # TODO need a data source to get the authority root certificate
    # roots = [smallstep_authority.agents.root]
    roots = ["-----BEGIN CERTIFICATE-----\nMIIBazCCARGgAwIBAgIQIbPd9RVuu0l/eTKVVjDL9zAKBggqhkjOPQQDAjAUMRIw\nEAYDVQQDEwlyb290LWNhLTEwHhcNMjIxMjE0MjIzMzE0WhcNMzIxMjExMjIzMzE0\nWjAUMRIwEAYDVQQDEwlyb290LWNhLTEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\nAASIOTDFWdDWOIyzASpf9EKEokDkc1ckVt50pjgkalI9wQ8WyH88xutdUNDjIVDK\nVryJaH2DMUdWwFh07kmEyu8co0UwQzAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/\nBAgwBgEB/wIBATAdBgNVHQ4EFgQUq5LjtW5K1PrPYV8mJjcCTvuKvyYwCgYIKoZI\nzj0EAwIDSAAwRQIhAJRTP0IUbOVmW2ufj8VINXnBhYCCauSa1/fHNNM/5o1JAiAx\n+Pkrk5eAijv8H5xDEW2ik3bUDnai7rQLzdbrEb5RPw==\n-----END CERTIFICATE-----"]
  }
}

resource "smallstep_provisioner_webhook" "devices" {
  authority_id    = smallstep_authority.agents.id
  provisioner_id  = smallstep_provisioner.agents.id
  name            = "devices"
  kind            = "ENRICHING"
  cert_type       = "X509"
  server_type     = "HOSTED_ATTESTATION"
  collection_slug = smallstep_collection.tpms.slug
  depends_on      = [smallstep_collection.tpms]
}
