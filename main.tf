terraform {
  required_providers {
    http = {
      source = "hashicorp/http"
      version = "2.1.0"
    }
  }
  required_version = "~> 0.12"
  backend "remote" {
    hostname      = "app.terraform.io"
    organization  = "esodemoapp2"

    workspaces {
      name = "default-workspace"
    }
  }
}
