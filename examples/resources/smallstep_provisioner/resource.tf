
resource "smallstep_provisioner" "my_oidc" {
  authority_id = smallstep_authority.basic.id
  name         = "My OIDC example.com"
  type         = "OIDC"

  oidc = {
    client_id              = "abc123"
    client_secret          = "xyz789"
    configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
  }
}

resource "smallstep_provisioner" "my_jwk" {
  authority_id = smallstep_authority.basic.id
  name         = "jwk1"
  type         = "JWK"

  jwk = {
    key           = file("${path.module}/foo.pub")
    encrypted_key = file("${path.module}/foo.priv")
  }
}
