
data "smallstep_managed_radius_secret" "my_radius" {
  id = "cd4452b0-809a-4fc1-aafe-1814042ce1fc"
}

output "radius_secret" {
  sensitive = true
  value     = data.smallstep_managed_radius_ssecret.my_radius.secret
}
