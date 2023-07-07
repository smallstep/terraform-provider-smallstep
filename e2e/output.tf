output "agent_configuration_id" {
  value = smallstep_agent_configuration.agent1.id
}

output "ep_x509_id" {
  value = smallstep_endpoint_configuration.ep_x509.id
}

output "ep_ssh_id" {
  value = smallstep_endpoint_configuration.ep_ssh.id
}

output "managed_configuration_id" {
  value = smallstep_managed_configuration.mc.id
}
