---
page_title: "cidrsubtract Function"
---

# cidrsubtract Function

Given a base CIDR block and a list of CIDR blocks to exclude, returns the minimal set of CIDR blocks that cover the base range minus the excluded ranges.

Results are sorted in ascending order by network address, producing stable Terraform plans with no spurious diffs.

## Example Usage

```hcl
provider::cidrsubtract::cidrsubtract("10.0.0.0/16", ["10.0.31.0/25", "10.0.31.128/25"])
# result: ["10.0.0.0/17", "10.0.28.0/22", "10.0.30.0/24", "10.0.31.0/25", ...]
```

## Signature

```text
cidrsubtract(base string, excludes list(string)) list(string)
```

## Arguments

1. `base` (string) - The base CIDR block to subtract from (e.g. `"10.0.0.0/16"`). Must be valid CIDR notation with no host bits set.
2. `excludes` (list of string) - CIDR blocks to exclude from the base. Each entry must be valid CIDR notation with no host bits set, and must be fully contained within the base CIDR. Overlapping and duplicate entries are handled gracefully.

## Return Value

A list of strings representing the minimal set of CIDR blocks covering the base range minus the excluded ranges. The list is sorted by network address ascending. Returns an empty list if the excludes fully cover the base. Returns the base unchanged if the exclude list is empty.

## Error Conditions

- Invalid CIDR notation in base or any exclude entry
- Host bits set in base or any exclude entry (e.g. `10.0.1.5/24` instead of `10.0.1.0/24`)
- Exclude entry not contained within the base CIDR
