terraform {
  required_providers {
    smallstep = {
      source = "smallstep/smallstep"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "smallstep" {
  # bearer_token = "ey..."

  # client_certificate = {
  #   certificate = file("api.crt")
  #   private_key = file("api.key")
  #   team_id     = "94a7dd82-1360-4493-b1bf-b14a97c45786"
  # }
}
#
# Configure the AWS Provider
provider "aws" {
  region = "us-west-1"
}

