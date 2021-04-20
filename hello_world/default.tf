terraform {
  required_providers {
    http = {
      source = "hashicorp/http"
      version = "2.1.0"
    }
  }
}

provider "http" {
  # Configuration options
}

variable "user_agent" {
  type = string
  default = "terraform"
}

data "http" ga {
  url = "https://httpbin.org/get"
  request_headers = {
    User-Agent = "var.user_agent"
  }
}

output "ga_resp" {
  value = data.http.ga.body
}
