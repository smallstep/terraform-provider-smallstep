
data "smallstep_managed_radius" "my_radius" {
  id = "cd4452b0-809a-4fc1-aafe-1814042ce1fc"
}

output "radius_ip" {
  value = data.smallstep_managed_radius.my_radius.server_ip
}

output "radius_port" {
  value = data.smallstep_managed_radius.my_radius.server_port
}

output "radius_hostname" {
  value = data.smallstep_managed_radius.my_radius.server_hostname
}

output "radius_ca" {
  value = data.smallstep_managed_radius.my_radius.server_ca
}
