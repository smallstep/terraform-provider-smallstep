
resource "smallstep_identity_provider" "my_idp" {
  trust_roots = file("${path.module}/root.crt")
}

resource "smallstep_sso_integration" "my_sso" {
  redirect_uri          = "https://example.com/callback"
  lifecycle_failure_uri = "https://example.com/device-not-active"

  # The secret is only returned when the integration is created and cannot be
  # recovered later. Set store_secret to keep it in terraform state, or set
  # write_secret_file to write it to disk instead.
  store_secret = true

  depends_on = [smallstep_identity_provider.my_idp]
}

output "sso_client_id" {
  value = smallstep_sso_integration.my_sso.id
}

output "sso_client_secret" {
  value     = smallstep_sso_integration.my_sso.secret
  sensitive = true
}
