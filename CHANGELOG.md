## 0.3.0

FEATURES:
* Add schema_uri field to smallstep_collection data source and resource.
* Changing the data field on a smallstep_collection_instance resources updates the instance in place. Previously changing the instance data required replacing the instance.

BUG FIXES:

* Changing smallstep_collection_instance.id forces replace of the instance. Previously a new instance would be created and the instance with the old id would remain in the collection.

## 0.2.0

FEATURES:

* Add smallstep_provisioner data source and resource with import. The supported provisioner types are OIDC, JWK, ACME, ACME_ATTESTATION, X5C, AWS, GCP and AZURE.
* Add smallstep_collection data source and resource with import.
* Add smallstep_collection_instance data source and resource with import.
* Add smallstep_provisioner_webhook data source and resource with import.
* Add smallstep_attestation_authority data source and resource with import.
* Add smallstep_agent_configuration data source and resource with import.
* Add smallstep_endpoint_configuration data source and resource with import.
* Add smallstep_managed_configuration data source and resource with import.

BUG FIXES:

* smallstep_authority.admin_emails type changed from list to set

## 0.1.0

FEATURES:

* Add smallstep_authority resource and data source
