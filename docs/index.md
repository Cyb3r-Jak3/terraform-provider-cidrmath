---
page_title: "cidrsubtract Provider"
---

# cidrsubtract Provider

The `cidrsubtract` provider exposes provider-defined functions, `cidrsubtract` and `cidrsubtract_set`, that compute the minimal set of CIDR blocks covering a base range minus a list of excluded ranges. `cidrsubtract_set` accepts lists of bases and excludes that may overlap in any way.

This is useful when you need to create security group rules, ALB listener conditions, or network ACLs that allow traffic from a network range while excluding specific subnets.

The provider has no configuration, manages no state, and makes no API calls. It requires Terraform >= 1.8 for provider-defined function support.

## Example Usage

```hcl
terraform {
  required_version = ">= 1.8.0"

  required_providers {
    cidrsubtract = {
      source = "Cyb3r-Jak3/cidrmath"
    }
  }
}

provider "cidrsubtract" {}

locals {
  # Allow traffic from the VPC except two restricted subnets
  allowed_cidrs = provider::cidrsubtract::cidrsubtract(
    "10.0.0.0/16",
    [
      "10.0.10.0/24",
      "10.0.11.0/24",
    ]
  )
}
```
