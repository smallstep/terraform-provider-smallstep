# Terraform Provider Smallstep

## Usage

Documentation for using the Smallstep provider can be found [here](https://registry.terraform.io/providers/smallstep/smallstep)

## Developing

This repository is based on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

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

A sweeper is defined to clean up all authorities older than 1 day unless the authority domain begins with `keep-`.
