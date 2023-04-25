This directory is for testing against the staging testacc account during development.

First ensure your $HOME/.terraformrc has a dev_override for this provider:
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

Then you can run `go install` from the root of this repo and use `terraform apply` in this directory to test changes.
