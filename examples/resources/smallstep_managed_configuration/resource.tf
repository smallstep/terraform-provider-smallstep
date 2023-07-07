
resource "smallstep_managed_configuration" "mc" {
	agent_configuration_id = smallstep_agent_configuration.agent1.id
	host_id = "9cdaf513-3296-4037-bd9b-d0634f51cd79"
	name = "DB Server"
	managed_endpoints = [
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.x509.id
			x509_certificate_data = {
				common_name = "db"
				sans = [
					"db",
					"db.default",
					"db.default.svc",
					"db.defaulst.svc.cluster.local",
				]
			}
		},
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.ssh.id
      ssh_certificate_data = {
        key_id = "abc"
        principals = ["ops", "eng"]
      }
		},
	]
}
