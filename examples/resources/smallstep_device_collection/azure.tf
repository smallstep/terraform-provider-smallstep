
resource "smallstep_device_collection" "azure" {
  slug         = "azure"
  display_name = "Azure VMs"
  admin_emails = ["admin@example.com"]
  device_type  = "azure-vm"
  azure_vm = {
    tenant_id           = "76543210"
    resource_groups     = ["0123456789"]
    disable_custom_sans = false
    audience            = ""
  }
}
