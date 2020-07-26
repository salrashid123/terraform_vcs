
provider "http" {}

resource "random_id" "id" {
  byte_length = 4
  prefix      = "ts-"
}

variable "ga_id" {
  type = string
  default = "UA-155117088-1"
}

variable "s_id" {
  type = string
  default = "https://cloud.google.com/solutions/gaming/minecraft-server"
}

variable "user_agent" {
  type = string
  default = "terraform"
}

data "http" ga {
  # url = "https://www.google-analytics.com/collect?v=1&tid=${var.ga_id}&cid=${random_id.id.b64_url}&dp=${var.s_id}"
  url = "https://httpbin.org/get?v=1&tid=${var.ga_id}&cid=${random_id.id.b64_url}&dp=${var.s_id}"
  request_headers = {
    User-Agent = "${var.user_agent}"
  }
}

output "ga_resp" {
  value = data.http.ga.body
}
