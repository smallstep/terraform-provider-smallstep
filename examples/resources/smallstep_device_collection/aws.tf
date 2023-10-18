
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
