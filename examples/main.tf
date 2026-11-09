terraform {
  required_version = ">= 1.8.0"

  required_providers {
    cidrsubtract = {
      source = "danifr/cidrsubtract"
    }
  }
}

provider "cidrsubtract" {}

locals {
  # Compute the CIDR complement: VPC range minus excluded subnets
  allowed_cidrs = provider::cidrsubtract::cidrsubtract(
    "10.0.0.0/16",
    [
      "10.0.10.0/24", # will be excluded
      "10.0.11.0/24", # will be excluded
    ]
  )
}

output "allowed_cidrs" {
  description = "CIDR blocks covering the VPC minus excluded subnets"
  value       = local.allowed_cidrs
}
