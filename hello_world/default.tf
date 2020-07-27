
provider "http" {}

variable "user_agent" {
  type = string
  default = "terraform"
}

data "http" ga {
  url = "https://httpbin.org/get"
  request_headers = {
    User-Agent = "${var.user_agent}"
  }
}

output "ga_resp" {
  value = data.http.ga.body
}
