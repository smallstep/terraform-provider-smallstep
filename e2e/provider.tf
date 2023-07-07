terraform {
  required_providers {
    smallstep = {
      source = "smallstep/smallstep"
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
