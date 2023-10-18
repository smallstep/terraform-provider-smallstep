
resource "smallstep_device_collection" "tpm" {
  slug         = "tmpservers"
  display_name = "TPM Servers"
  admin_emails = ["admin@example.com"]
  device_type  = "tpm"
  tpm = {
    attestor_roots = file("${path.module}/root.crt")
  }
}
