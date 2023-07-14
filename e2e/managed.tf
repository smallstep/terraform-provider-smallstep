
resource "smallstep_collection" "tpms" {
  slug = "tpms"
}

resource "smallstep_collection_instance" "server1" {
  id = "urn:ek:sha256:RAzbOveN1Y45fYubuTxu5jOXWtOK1HbfZ7yHjBuWlyE="
  data = "{}"
  collection_slug = smallstep_collection.tpms.slug
  depends_on = [smallstep_collection.tpms]
}

resource "smallstep_attestation_authority" "aa" {
  name = "foo"
  catalog = smallstep_collection.tpms.slug
  attestor_roots = file("${path.module}/ca.crt")
  depends_on = [smallstep_collection.tpms]
}

