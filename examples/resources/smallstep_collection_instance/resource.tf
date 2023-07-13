
resource "smallstep_collection_instance" "server1" {
  id              = "urn:ek:sha256:RAzbOveN1Y45fYubuTxu5jOXWtOK1HbfZ7yHjBuWlyE="
  data            = "{}"
  collection_slug = smallstep_collection.tpms.slug
  depends_on      = [smallstep_collection.tpms]
}

