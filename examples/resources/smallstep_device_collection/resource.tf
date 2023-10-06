
resource "smallstep_device_collection" "aws" {
  slug         = "ec2west"
  display_name = "EC2 West"
  admin_emails = ["admin@example.com"]
  device_type  = "aws-vm"
  aws_vm = {
    accounts            = ["0123456789"]
    disable_custom_sans = false
  }
}

resource "smallstep_device_collection" "gcp" {
  slug         = "gce"
  display_name = "GCE"
  admin_emails = ["admin@example.com"]
  device_type  = "gcp-vm"
  gcp_vm = {
    service_accounts    = ["pki@prod-1234.iam.gserviceaccount.com"]
    project_ids         = ["prod-1234"]
    disable_custom_sans = false
  }
}

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

resource "smallstep_device_collection" "tpm" {
  slug         = "tmpservers"
  display_name = "TPM Servers"
  admin_emails = ["admin@example.com"]
  device_type  = "tpm"
  tpm = {
    attestor_roots         = "-----BEGIN..."
    attestor_intermediates = "-----BEGIN..."
    force_cn               = false
    require_eab            = false
  }
}
