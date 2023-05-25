data "smallstep_provisioner" "prov" {
  authority_id = "c7f77ceb-0dbe-4754-af1e-aa8487fe3118"
  id = "ab71ec41-2a74-48ae-a219-53fa08b4cca4"
}

output "provisioner_name" {
  value = data.smallstep_provisioner.prov.name
}

resource "smallstep_provisioner" "my_oidc" {
  authority_id = smallstep_authority.basic.id
  name = "My OIDC example.com"
  type = "OIDC"

  oidc = {
    client_id = "abc123"
    client_secret = "xyz789"
    configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
  }
}

resource "smallstep_provisioner" "jwk1" {
  authority_id = smallstep_authority.basic.id
  name = "jwk1"
  type = "JWK"

  jwk = {
    key = file("${path.module}/foo.pub")
    encrypted_key = file("${path.module}/foo.priv")
  }
}
