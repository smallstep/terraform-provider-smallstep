# Terraform Provider Smallstep

## Usage

The Terraform Registry has [Documentation for using the Smallstep provider](https://registry.terraform.io/providers/smallstep/smallstep/latest/docs).

## Developing

This repository is based on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

It uses an API client generated from Smallstep's OpenAPI spec.
The generated client and server can be copied from the API repo to internal/apiclient and internal/apiserver.
The generated server is only used to get documentation for fields since it embedds OpenAPI spec.

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20

### Building The Provider

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

### Documentation

To generate or update documentation, run `go generate`.

### Testing
In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.
```shell
make testacc
```

Use these environment variables to configure the account the tests run against:
* SMALLSTEP_API_TOKEN
* SMALLSTEP_API_URL
* SMALLSTEP_CA_DOMAIN

A sweeper is defined to clean up all authorities older than 1 day unless the authority domain begins with `keep-`.

```shell
make sweep
```


To test the provider with terraform locally, first ensure your $HOME/.terraformrc has a dev_override for this provider:
```
provider_installation {

  dev_overrides {
      "smallstep/smallstep" = "/home/<USER>/go/bin/"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Then you can run `go install` from the root of this repo and use `terraform apply`.
