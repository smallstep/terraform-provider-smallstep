
resource "smallstep_device_collection" "ec2_west" {
  slug         = "ec2west"
  display_name = "EC2 West"
  device_type  = "aws-vm"
  aws_vm = {
    accounts = ["0123456789"]
  }
  admin_emails = ["admin@example.com"]
}

resource "smallstep_workload" "generic" {
  depends_on             = [smallstep_device_collection.ec2_west]
  workload_type          = "generic"
  device_collection_slug = resource.smallstep_device_collection.ec2_west.slug
  slug                   = "ec2generic"
  display_name           = "Generic Workload"
  admin_emails           = ["admin@example.com"]

  certificate_info = {
    type = "X509"
  }

  key_info = {
    format = "DEFAULT"
    type   = "ECDSA_P256"
  }
}
