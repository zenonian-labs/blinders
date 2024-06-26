provider "aws" {
  region  = var.region
  profile = var.profile
}

terraform {
  backend "s3" {
    key = "blinders-prod-state"
  }
}

module "core" {
  source = "../core"

  project = {
    name           = "blinders"
    environment    = "prod"
    default_region = "ap-south-1"
  }

  domains = {
    http      = "api.peakee.co"
    websocket = "ws.peakee.co"
  }

  env_filename = "../../.env.prod"
}

