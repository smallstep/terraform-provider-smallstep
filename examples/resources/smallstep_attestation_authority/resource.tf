
resource "smallstep_collection" "tpms" {
  slug = "tpms"
}

resource "smallstep_attestation_authority" "aa" {
  name                   = "Foo Attest"
  catalog                = smallstep_collection.tpms.slug
  attestor_roots         = "-----BEGIN CERTIFICATE-----\n..."
  attestor_intermediates = "----- BEGIN CERTIFICATE-----\n..."
  depends_on             = [smallstep_collection.tpms]
}
