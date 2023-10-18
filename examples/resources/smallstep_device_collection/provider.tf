# This file is not shown in example docs but is used for testing
terraform {
  required_providers {
    smallstep = {
      source = "smallstep/smallstep"
    }
    google = {
      source  = "hashicorp/google"
      version = "5.2.0"
    }
  }
}

provider "smallstep" {}

provider "google" {
  project = "prod-1234"
  region  = "us-central1"
}
