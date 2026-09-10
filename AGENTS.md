# AGENTS.md

Guidance for AI coding agents working in this repository.

## Overview

`terraform-provider-smallstep` is the Terraform provider for the Smallstep SaaS platform, published to the Terraform Registry as `smallstep/smallstep`. It is built on the HashiCorp Terraform Plugin Framework (protocol v6) and talks to the Smallstep customer REST API through Go clients generated from Smallstep's OpenAPI spec. Module path: `github.com/smallstep/terraform-provider-smallstep`. It depends on `go.step.sm/crypto` (test helpers only) and has no private dependencies; `go build` works with no `GOPRIVATE` or extra setup.

## Commands

```bash
go build ./...                 # build (what CI runs; produces no binary)
go install                     # install the provider binary into $GOPATH/bin for dev_overrides
go vet ./...
go generate ./...              # terraform fmt on examples/ + regenerate docs/ with tfplugindocs
make generate-docs             # same as go generate
make testacc                   # full acceptance suite: TF_ACC=1 go test ./... -v -timeout 20m
make sweep                     # delete leftover test resources (tfprovider-* prefixed)
go test -v -run TestAccProxyResource ./internal/provider/proxy/   # single test (still needs env, see below)
```

`go generate` requires the `terraform` binary on `PATH`. CI runs it and fails the build if `docs/` or `examples/` change, so regenerate and commit whenever you touch a schema, description, or example.

There are no pure unit tests. Every `*_test.go` is an acceptance test that creates real resources against a live Smallstep team, and the test helpers call `t.Fatal` when `SMALLSTEP_API_TOKEN` is unset, so a bare `go test ./...` fails rather than skipping. Required environment:

| Variable | Purpose |
|----------|---------|
| `TF_ACC=1` | Enables `terraform-plugin-testing` acceptance tests |
| `SMALLSTEP_API_TOKEN` | API token for the team the tests run against |
| `SMALLSTEP_API_URL` | API base URL, e.g. `https://gateway.smallstep.com/api` |
| `SMALLSTEP_CA_DOMAIN` | Optional; CA domain suffix asserted by the authority tests |
| `RELAY_HOSTNAME`, `RELAY_HOSTNAME_2` | Optional; hostnames used by the relay tests |
| `SWEEP_AGE` | Optional; min age of `tfprovider-` authorities the sweeper deletes (default `1m`) |
| `TF_ACC_LOG=INFO` | Optional; Terraform log level during tests |

CI (`.github/workflows/test.yml`) runs `go build`, the `go generate` diff check, the acceptance suite against a single Terraform version, then `make sweep`. There is a `concurrency` group on the workflow because a team may have only one attestation authority, so acceptance runs are serialized. `.golangci.yml` exists but no CI job runs golangci-lint.

## Generated code - do not edit

| Pattern | Generator |
|---------|-----------|
| `internal/apiclient/v*/api.gen.go` | oapi-codegen v2, from Smallstep's OpenAPI spec. Not generated in this repo: the file is produced from the spec and copied in as a whole. Each package embeds its spec (`GetSwagger()`). |
| `docs/**/*.md` | `tfplugindocs` via `go generate` (from schemas + `examples/`) |
| `examples/**/*.tf` formatting | `terraform fmt -recursive ./examples/` via `go generate` |

To pick up a new API version, add a new `internal/apiclient/vYYYYMMDD/` package, add a client field to `clientset.Clients`, wire it in `provider.go` (`Configure`) and `testprovider/provider.go`, and add a `DescribeVYYYYMMDD` helper in `internal/provider/utils/utils.go`.

## Architecture

```
main.go                          # providerserver.Serve; go:generate directives for docs
internal/
  apiclient/
    clientset/clientset.go       # Clients{V20250101, V20260501}: the ProviderData passed to every resource
    v20250101/api.gen.go         # generated client + embedded spec, API version 2025-01-01
    v20260501/api.gen.go         # generated client + embedded spec, API version 2026-05-01
  provider/
    provider.go                  # provider schema (bearer_token, client_certificate), Configure, resource/data-source registry
    cert_client.go               # exchanges a client certificate for an API token (client_certificate auth)
    utils/                       # shared helpers: Describe*/DescribeV20260501, Deref, APIErrorMsg, plan modifiers, test helpers
    <resource>/                  # one package per Terraform type, e.g. authority, credential, proxy, wifi
      model.go                   #   typeName, Model struct (tfsdk tags), fromAPI / toAPI converters
      resource.go                #   resource.Resource (+ ImportState) CRUD
      data_source.go             #   datasource.DataSource
      resource_test.go           #   acceptance tests
      data_source_test.go
      sweep_test.go              #   TestMain; most types also AddTestSweepers("smallstep_<type>")
  testprovider/provider.go       # minimal provider for tests; builds clients from env vars
docs/                            # generated registry docs (index.md, resources/, data-sources/)
examples/                        # provider.tf, per-resource resource.tf + import.sh; feed tfplugindocs
tools/tools.go                   # pins tfplugindocs in go.mod
```

Resource packages: `authority`, `browser`, `credential`, `device`, `ethernet`, `identity_provider` (two resources: identity provider and client), `managed_radius` (plus a `_secret` data source), `provisioner`, `proxy`, `relay`, `sso_integration`, `vpn`, `webhook` (`smallstep_provisioner_webhook`), `wifi`, `workload`. Every resource and data source must be registered in `Resources()` / `DataSources()` in `internal/provider/provider.go`.

Older resources use the `V20250101` client; resources added in 0.8.0+ (`proxy`, `relay`, `sso_integration`, `workload`) and `credential` use `V20260501`. Pick the client in `Configure` from `clientset.Clients` and keep a resource on one version.

## Conventions

- Attribute descriptions come from the OpenAPI spec, not hand-written strings: call `utils.Describe("componentName")` (2025-01-01) or `utils.DescribeV20260501(...)` in `Schema()` and use the returned `props["jsonField"]` map. This keeps registry docs in sync with the API.
- API calls use the generated `client.PostX / GetX / PutX / DeleteX` methods that return `*http.Response`; check the status code explicitly, and on failure report `X-Request-Id` and `utils.APIErrorMsg(body)` in the diagnostic. Decode success bodies with `json.NewDecoder`.
- Errors surface as `resp.Diagnostics.AddError(summary, detail)`; there is no error-wrapping convention beyond `fmt.Errorf` in helpers. Logging is `tflog` from `terraform-plugin-log`.
- Generated client structs use pointers for optional fields; use `utils.Deref`, `utils.ToStringPointer`, `utils.ToIntPointer`, and Go's `new(value)` when converting to and from `Model`.
- Computed IDs use `stringplanmodifier.UseStateForUnknown()`; `utils.MaybeUseStateForUnknown` handles server-defaulted nested objects (see its doc comment).
- Tests use `terraform-plugin-testing` (`helper.Test` with `ProtoV6ProviderFactories` built from `testprovider.SmallstepTestProvider`), `testify` for setup assertions, and helpers in `internal/provider/utils/testutils.go` (`NewAuthority`, `NewCredential`, `NewDevice`, `CACerts`, `Slug`, ...) that create prerequisites over the API and register `t.Cleanup` deletes. Name test resources with the `tfprovider-` prefix so the sweepers can find them, and add an `AddTestSweepers` entry in `sweep_test.go` for any new resource type.
- Each test package defines its own `provider` and `providerFactories` vars listing only the factories it exercises.
- Every resource ships an `examples/resources/smallstep_<type>/resource.tf` and `import.sh`, and each data source an `examples/data-sources/smallstep_<type>/data-source.tf`; tfplugindocs renders them into `docs/`.
- Record user-facing changes in `CHANGELOG.md` under a version heading. Releases are cut by tagging `v*`, which runs goreleaser (`.github/workflows/release.yml`) with GPG-signed checksums.

## Local development against Terraform

Add a `dev_overrides` block for `smallstep/smallstep` pointing at `$GOPATH/bin` in `~/.terraformrc` (see README.md), run `go install`, then `terraform apply` in a directory using the provider. The provider reads `SMALLSTEP_API_TOKEN` and `SMALLSTEP_API_URL` at runtime when `bearer_token` is not set in configuration.
