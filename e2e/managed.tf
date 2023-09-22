
resource "smallstep_device_collection" "ec2_west" {
  slug = "ec2west"
  display_name = "EC2 West"
  device_type = "aws-vm"
  aws_vm = {
    accounts = ["807492473263"]
  }
  admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_collection_instance" "my_ec2" {
  id = "i-01234"
  collection_slug = smallstep_device_collection.ec2_west.slug
  data = "{}"
}

output "outdata" {
  value = resource.smallstep_collection_instance.my_ec2.out_data
}
