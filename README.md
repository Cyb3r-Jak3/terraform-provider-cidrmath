# terraform-provider-cidrsubtract

## Forked from [danifr/terraform-provider-cidrsubtract](https://github.com/danifr/terraform-provider-cidrsubtract)

A Terraform provider that exposes provider-defined functions for CIDR subtraction: `cidrsubtract` computes the minimal set of CIDR blocks covering a base range minus a list of excluded ranges, and `cidrsubtract_set` does the same for lists of bases and excludes that may overlap in any way.

Useful for security group rules, ALB listener conditions, or network ACLs where you need to allow traffic from a network range while excluding specific subnets.

- No configuration, no state, no API calls — just a pure function
- Deterministic, sorted output for stable Terraform plans
- Handles overlapping and duplicate excludes gracefully
- Requires Terraform >= 1.8 (provider-defined function support)

## Usage

```hcl
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
  allowed_cidrs = provider::cidrsubtract::cidrsubtract(
    "10.0.0.0/16",
    [
      "10.0.10.0/24",
      "10.0.11.0/24",
    ]
  )
}
```

## Function Signature

```text
cidrsubtract(base string, excludes list(string)) list(string)
```

**Arguments:**

- `base` — Base CIDR block (e.g. `"10.0.0.0/16"`). Must be valid CIDR notation with no host bits set.
- `excludes` — List of CIDR blocks to exclude. Each must be valid CIDR notation, no host bits set, and fully contained within the base.

**Returns:** Minimal list of CIDR blocks covering the base minus the excludes, sorted by network address.

### cidrsubtract_set

```text
cidrsubtract_set(bases list(string), excludes list(string)) list(string)
```

Subtracts the union of `excludes` from the union of `bases`. Excludes need not be contained within a base, and bases may overlap one another, so two published range lists can be subtracted directly. IPv4 only. See [docs/functions/cidrsubtract_set.md](docs/functions/cidrsubtract_set.md).

```hcl
provider::cidrsubtract::cidrsubtract_set(
  ["10.0.0.0/24", "10.0.2.0/24"],
  ["10.0.0.128/25", "10.0.2.0/23"]
)
# result: ["10.0.0.0/25"]
```

## Building from source

```bash
go build -o terraform-provider-cidrsubtract
```

## Local development

Add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "danifr/cidrsubtract" = "/path/to/terraform-provider-cidrsubtract"
  }
  direct {}
}
```

Then run `terraform plan` in any configuration that uses the provider — no `terraform init` needed.

## License

MPL-2.0
