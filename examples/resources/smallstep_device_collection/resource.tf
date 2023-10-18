
resource "smallstep_device_collection" "gcp" {
  slug         = "gce"
  display_name = "GCE"
  device_type  = "gcp-vm"
  gcp_vm = {
    service_accounts = ["pki@prod-1234.iam.gserviceaccount.com"]
  }
  admin_emails = ["admin@example.com"]
}

data "google_compute_instance" "dbserver" {
  name = "dbserver"
  zone = "us-central1-b"
}

resource "smallstep_collection_instance" "dbserver" {
  depends_on      = [smallstep_device_collection.gcp]
  collection_slug = smallstep_device_collection.gcp.slug
  id              = data.google_compute_instance.dbserver.instance_id
  data = jsonencode({
    "hostname"   = data.google_compute_instance.dbserver.name
    "private_ip" = data.google_compute_instance.dbserver.network_interface.0.network_ip
    "public_ip"  = data.google_compute_instance.dbserver.network_interface.0.access_config[0].nat_ip
  })
}
