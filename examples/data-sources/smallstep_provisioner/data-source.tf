
data "smallstep_provisioner" "by_name" {
  authority_id = "34bd2a7f-68e5-4f5e-81a5-531a4c3b5d99"
  name         = "my_jwk"
}

data "smallstep_provisioner" "by_id" {
  authority_id = "34bd2a7f-68e5-4f5e-81a5-531a4c3b5d99"
  id           = "e58e57ce-c88f-4acc-97fa-bed59aae611e"
}
