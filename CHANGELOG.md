## 0.4.0

FEATURES:
* Add smallstep_device_collection resource.
* Add smallstep_workload resource.

BUG FIXES:
* Hosted provisioner webhooks will have a secret of type null rather than type unknown after creation.

CHANGES:
* attestation_authority resource and data source no longer has a `catalog` attribute.
* Remove smallstep_managed_configuration resource and data source.
* Remove smallstep_endpoint_configuration resource and data source.
* Remove smallstep_agent_configuration resource and data source.

## 0.3.0

FEATURES:
* Add `schema_uri` attribute to smallstep_collection data source and resource.
* Changing the `data` attribute on a smallstep_collection_instance resource updates the instance in place. Previously changing the `data` attribute required replacing the instance.
* Changing the `display_name` attribute on a smallstep_collection resource updates the collection in place. Previously changing the `display_name` attribute required replacing the collection.
* All user-supplied attributes on smallstep_agent_configuration, smallstep_endpoint_configuration and smallstep_managed_configuration resources will update the resource in place. Previously changing any attribute required replacing the resource.

BUG FIXES:

* Changing smallstep_collection_instance.id forces replace of the instance. Previously a new instance would be created and the instance with the old id would remain in the collection.
* A smallstep_collection resource without the optional `display_name` attribute set would fail to apply.

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
