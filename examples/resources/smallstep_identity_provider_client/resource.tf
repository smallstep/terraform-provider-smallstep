
resource "smallstep_identity_provider" "my_idp" {
  trust_roots = file("${path.module}/root.crt")
}

resource "smallstep_identity_provider_client" "my_idp_client" {
  redirect_uri = "https://example.com/callback"
  store_secret = true
  depends_on   = [smallstep_identity_provider.my_idp]
}

output "idp_client_id" {
  value = smallstep_identity_provider_client.my_idp_client.id
}

output "idp_client_secret" {
  value = smallstep_identity_provider_client.my_idp_client.secret
}
