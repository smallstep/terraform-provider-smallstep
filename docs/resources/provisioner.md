---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "smallstep_provisioner Resource - terraform-provider-smallstep"
subcategory: ""
description: |-
  Provisioners https://smallstep.com/docs/step-ca/provisioners/ are methods of using the CA to get certificates with different modes of authorization.
---

# smallstep_provisioner (Resource)

[Provisioners](https://smallstep.com/docs/step-ca/provisioners/) are methods of using the CA to get certificates with different modes of authorization.

## Example Usage

```terraform
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

resource "smallstep_provisioner" "my_acme" {
  authority_id = smallstep_authority.basic.id
  name         = "acme1"
  type         = "ACME"
  acme = {
    challenges  = ["http-01", "dns-01", "tls-alpn-01"]
    require_eab = true
    force_cn    = false
  }
}

resource "smallstep_provisioner" "my_x5c" {
  authority_id = smallstep_authority.basic.id
  name         = "x5c foo"
  type         = "X5C"
  x5c = {
    roots = ["-----BEGIN CERTIFICATE-----\n..."]
  }
}

resource "smallstep_provisioner" "my_attest" {
  authority_id = smallstep_authority.basic.id
  name         = "attest foo"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["apple", "step"]
  }
}

resource "smallstep_provisioner" "tpm_attest" {
  authority_id = smallstep_authority.basic.id
  name         = "attest foo"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["tpm"]
    attestation_roots   = ["-----BEGIN CERTIFICATE-----\n..."]
  }
}

resource "smallstep_provisioner" "my_aws" {
  authority_id = smallstep_authority.basic.id
  name         = "AWS foo"
  type         = "AWS"
  aws = {
    accounts                   = ["0123456789"]
    instance_age               = "1h"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}
resource "smallstep_provisioner" "my_gcp" {
  authority_id = smallstep_authority.basic.id
  name         = "GCP foo"
  type         = "GCP"
  gcp = {
    project_ids                = ["prod-1234"]
    service_accounts           = ["pki@prod-1234.iam.gserviceaccount.com"]
    instance_age               = "1h"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}

resource "smallstep_provisioner" "my_azure" {
  authority_id = smallstep_authority.basic.id
  name         = "Azure foo"
  type         = "AZURE"
  azure = {
    tenant_id                  = "948920a7-8fc1-431f-be94-030a80232e51"
    resource_groups            = ["prod-1234"]
    audience                   = "example.com"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `authority_id` (String) The UUID of the authority this provisioner is attached to
- `name` (String) The name of the provisioner.
- `type` (String) The type of provisioner. Allowed values: `OIDC` `JWK` `ACME` `ACME_ATTESTATION` `X5C` `AWS` `GCP` `AZURE` `SCEP`

### Optional

- `acme` (Attributes) A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#acme) that enables automation with the [ACME protocol](https://smallstep.com/docs/step-ca/acme-basics/#acme-challenges). This object is required when type is `ACME` and is otherwise ignored. (see [below for nested schema](#nestedatt--acme))
- `acme_attestation` (Attributes) A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#acme) that enables automation with the [device-attest-01 challenge of the ACME protocol](https://smallstep.com/blog/acme-managed-device-attestation-explained/). This object is required when type is `ACME_ATTESTATION` and is otherwise ignored. (see [below for nested schema](#nestedatt--acme_attestation))
- `aws` (Attributes) The [AWS provisioner](https://smallstep.com/docs/step-ca/provisioners/#aws) grants a certificate to an Amazon EC2 instance using the Instance Identity Document. This object is required when type is `AWS` and is otherwise ignored. (see [below for nested schema](#nestedatt--aws))
- `azure` (Attributes) The [Azure provisioner](https://smallstep.com/docs/step-ca/provisioners/#azure) grants certificates to Microsoft Azure instances using the managed identities tokens. This object is required when type is `AZURE` and is otherwise ignored. (see [below for nested schema](#nestedatt--azure))
- `claims` (Attributes) A set of constraints configuring how this provisioner can be used to issue certificates. (see [below for nested schema](#nestedatt--claims))
- `gcp` (Attributes) The [GCP provisioner](https://smallstep.com/docs/step-ca/provisioners/#gcp) grants a certificate to a Google Compute Engine instance using its identity token. At least one service account or project ID must be set. This object is required when type is `GCP` and is otherwise ignored. (see [below for nested schema](#nestedatt--gcp))
- `jwk` (Attributes) A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#jwk) that uses public-key cryptography to sign and validate a JSON Web Token (JWT). This object is required when type is `JWK` and is otherwise ignored. (see [below for nested schema](#nestedatt--jwk))
- `oidc` (Attributes) A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#oauthoidc-single-sign-on) that is configured to trust and accept an OAuth provider's ID tokens for authentication. By default, the issued certificate will use the subject (sub) claim from the identity token as its subject. The value of the token's email claim is also included as an email SAN in the certificate. This object is required when type is `OIDC` and is otherwise ignored. (see [below for nested schema](#nestedatt--oidc))
- `options` (Attributes) Options that apply when issuing certificates with this provisioner. (see [below for nested schema](#nestedatt--options))
- `x5c` (Attributes) A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#x5c---x509-certificate) that authenticates a certificate request with an existing x509 certificate. This object is required when type is `X5C` and is otherwise ignored. (see [below for nested schema](#nestedatt--x5c))

### Read-Only

- `created_at` (String) Timestamp of when the provisioner was created in RFC 3339 format. Generated server-side.
- `id` (String) A UUID identifying this provisioner. Generated server-side when the provisioner is created.

<a id="nestedatt--acme"></a>
### Nested Schema for `acme`

Required:

- `challenges` (Set of String) Which ACME challenge types are allowed. Allowed values: `http-01` `dns-01` `tls-alpn-01`
- `require_eab` (Boolean) Only ACME clients that have been preconfigured with valid EAB credentials will be able to create an account with this provisioner. Must be `true` for all new provisioners.

Optional:

- `force_cn` (Boolean) Force one of the SANs to become the Common Name, if a Common Name is not provided.


<a id="nestedatt--acme_attestation"></a>
### Nested Schema for `acme_attestation`

Required:

- `attestation_formats` (Set of String) The allowed attestation formats for the device-attest-01 challenge. Valid values are `apple`, `step`, and `tpm`. The apple format is for Apple devices, and adds trust for Apple's CAs. The step format is for non-TPM devices that can issue attestation certificates, such as YubiKey PIV. It adds trust for Yubico's root CA. The tpm format is for TPMs and does not trust any CAs by default. Allowed values: `apple` `step` `tpm`

Optional:

- `attestation_roots` (Set of String) A trust bundle of root certificates in PEM format that will be used to verify attestation certificates. The default value depends on the value of attestationFormats. If provided, this PEM bundle will override the CA trust established by setting attestationFormats to apple or step. At least one root certificate is required when using the tpm attestationFormat.
- `force_cn` (Boolean) Force one of the SANs to become the Common Name, if a Common Name is not provided.
- `require_eab` (Boolean) Only ACME clients that have been preconfigured with valid EAB credentials will be able to create an account with this provisioner.


<a id="nestedatt--aws"></a>
### Nested Schema for `aws`

Required:

- `accounts` (Set of String) The list of AWS account IDs that are allowed to use an AWS cloud provisioner.

Optional:

- `disable_custom_sans` (Boolean) By default custom SANs are valid, but if this option is set to `true` only the SANs available in the instance identity document will be valid. These are the private IP and the DNS ip-<private-ip>.<region>.compute.internal.
- `disable_trust_on_first_use` (Boolean) By default only one certificate will be granted per instance, but if the option is set to `true` this limit is not set and different tokens can be used to get different certificates.
- `instance_age` (String) The maximum age of an instance that should be allowed to obtain a certificate. Limits certificate issuance to new instances to mitigate the risk of credential-misuse from instances that don't need a certificate. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).


<a id="nestedatt--azure"></a>
### Nested Schema for `azure`

Required:

- `resource_groups` (Set of String) The list of resource group names that are allowed to use this provisioner.
- `tenant_id` (String) The Azure account tenant ID for this provisioner. This ID is the Directory ID available in the Azure Active Directory properties.

Optional:

- `audience` (String) Defaults to https://management.azure.com/ but it can be changed if necessary.
- `disable_custom_sans` (Boolean) By default custom SANs are valid, but if this option is set to `true` only the SANs available in the token will be valid, in Azure only the virtual machine name is available.
- `disable_trust_on_first_use` (Boolean) By default only one certificate will be granted per instance, but if the option is set to true this limit is not set and different tokens can be used to get different certificates.


<a id="nestedatt--claims"></a>
### Nested Schema for `claims`

Optional:

- `allow_renewal_after_expiry` (Boolean) Allow renewals for expired certificates generated by this provisioner.
- `default_host_ssh_cert_duration` (String) The default duration for an SSH host certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `default_tls_cert_duration` (String) The default duration for an x509 certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `default_user_ssh_cert_duration` (String) The default duration for an SSH user certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `disable_renewal` (Boolean) Disable renewal for all certificates generated by this provisioner.
- `enable_ssh_ca` (Boolean) Allow this provisioner to be used to generate SSH certificates.
- `max_host_ssh_cert_duration` (String) The maximum duration for an SSH host certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `max_tls_cert_duration` (String) The maximum duration for an x509 certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `max_user_ssh_cert_duration` (String) The maximum duration for an SSH user certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `min_host_ssh_cert_duration` (String) The minimum duration for an SSH host certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `min_tls_cert_duration` (String) The minimum duration for an x509 certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).
- `min_user_ssh_cert_duration` (String) The minimum duration for an SSH user certificate generated by this provisioner. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).


<a id="nestedatt--gcp"></a>
### Nested Schema for `gcp`

Required:

- `project_ids` (Set of String) The list of project identifiers that are allowed to use a GCP cloud provisioner.
- `service_accounts` (Set of String) The list of service accounts that are allowed to use a GCP cloud provisioner.

Optional:

- `disable_custom_sans` (Boolean) By default custom SANs are valid, but if this option is set to `true` only the SANs available in the instance identity document will be valid, these are the DNS `<instance-name>.c.<project-id>.internal` and `<instance-name>.<zone>.c.<project-id>.internal`.
- `disable_trust_on_first_use` (Boolean) By default only one certificate will be granted per instance, but if the option is set to `true` this limit is not set and different tokens can be used to get different certificates.
- `instance_age` (String) The maximum age of an instance that should be allowed to obtain a certificate. Limits certificate issuance to new instances to mitigate the risk of credential-misuse from instances that don't need a certificate. Parsed as a [Golang duration](https://pkg.go.dev/time#ParseDuration).


<a id="nestedatt--jwk"></a>
### Nested Schema for `jwk`

Required:

- `key` (String) The public JSON web key.

Optional:

- `encrypted_key` (String) The JWE encrypted private key.


<a id="nestedatt--oidc"></a>
### Nested Schema for `oidc`

Required:

- `client_id` (String) The id used to validate the audience in an OpenID Connect token.
- `configuration_endpoint` (String) OpenID Connect configuration URL.

Optional:

- `admins` (Set of String) The emails of admin users in an OpenID Connect provisioner. These users will not have restrictions in the certificates to sign.
- `client_secret` (String) The secret used to obtain the OpenID Connect tokens.
- `domains` (Set of String) The domains used to validate the email claim in an OpenID Connect provisioner.
- `groups` (Set of String) The group list used to validate the groups extension in an OpenID Connect token.
- `listen_address` (String) The callback address used in the OpenID Connect flow.
- `tenant_id` (String) The tenant-id used to replace the templatized tenantid value in the OpenID Configuration.


<a id="nestedatt--options"></a>
### Nested Schema for `options`

Optional:

- `ssh` (Attributes) Options that apply when issuing SSH certificates (see [below for nested schema](#nestedatt--options--ssh))
- `x509` (Attributes) Options that apply when issuing x509 certificates. (see [below for nested schema](#nestedatt--options--x509))

<a id="nestedatt--options--ssh"></a>
### Nested Schema for `options.ssh`

Optional:

- `template` (String) A JSON representation of the SSH certificate to be created. [More info](https://smallstep.com/docs/step-ca/templates/#ssh-templates).
- `template_data` (String) A map of data that can be used by the certificate template.


<a id="nestedatt--options--x509"></a>
### Nested Schema for `options.x509`

Optional:

- `template` (String) A JSON representation of the x509 certificate to be created. [More info](https://smallstep.com/docs/step-ca/templates/#x509-templates).
- `template_data` (String) A map of data that can be used by the certificate template.



<a id="nestedatt--x5c"></a>
### Nested Schema for `x5c`

Required:

- `roots` (Set of String) A list of pem-encoded x509 certificates. Any certificate bundle that chains up to any of these roots can be used in a certificate request.

## Import

Import is supported using the following syntax:

```shell
terraform import smallstep_provisioner.my_jwk_provisioner b1161f78-d251-401e-b17c-fe38fc26ae7b/my_jwk_provisioner
```
