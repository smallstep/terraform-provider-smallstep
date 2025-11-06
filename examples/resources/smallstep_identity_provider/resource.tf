
resource "smallstep_identity_provider" "my_idp" {
  trust_roots = file("${path.module}/root.crt")
}

output "idp_authorize_url" {
  value = smallstep_identity_provider.my_idp.authorize_endpoint
}

output "idp_token_url" {
  value = smallstep_identity_provider.my_idp.token_endpoint
}

output "idp_jwks_url" {
  value = smallstep_identity_provider.my_idp.jwks_endpoint
}
