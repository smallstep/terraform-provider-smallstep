
resource "smallstep_relay" "corp_relay" {
  name     = "CorpRelay"
  hostname = "relay.example.com"

  ca_chain = file("${path.module}/relay_ca_chain.crt")

  allowed_targets = ["ssh.internal.example.com", "db.internal.example.com"]

  issuing_authority_id = smallstep_authority.my_authority.id
}
