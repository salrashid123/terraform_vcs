terraform {
  required_version = "~> 0.12"
  backend "remote" {
    hostname      = "app.terraform.io"
    organization  = "esodemoapp2"

    workspaces {
      name = "default-workspace"
    }
  }
}
