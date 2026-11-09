---
page_title: "cidrsubtract_set Function"
---

# cidrsubtract_set Function

Given a list of base CIDR blocks and a list of CIDR blocks to exclude, returns the minimal set of CIDR blocks that cover the union of the bases minus the union of the excludes.

Unlike [`cidrsubtract`](cidrsubtract.md), the excludes do not need to be contained within a base. Bases may overlap one another, and an exclude may partially overlap a base, contain it entirely, or miss it altogether. This makes it suitable for subtracting one published range list from another, for example a cloud provider's customer ranges from its full address space.

Results are sorted in ascending order by network address, producing stable Terraform plans with no spurious diffs.

## Example Usage

```hcl
provider::cidrsubtract::cidrsubtract_set(
  ["10.0.0.0/24", "10.0.2.0/24", "192.168.0.0/24"],
  ["10.0.0.128/25", "10.0.2.0/23", "192.168.0.0/16"]
)
# result: ["10.0.0.0/25"]
```

## Signature

```text
cidrsubtract_set(bases list(string), excludes list(string)) list(string)
```

## Arguments

1. `bases` (list of string) - IPv4 CIDR blocks to subtract from. Each entry must be valid CIDR notation with no host bits set. Overlapping, nested and duplicate entries are handled gracefully.
2. `excludes` (list of string) - IPv4 CIDR blocks to exclude. Each entry must be valid CIDR notation with no host bits set. Entries may overlap the bases in any way.

## Return Value

A list of strings representing the minimal set of CIDR blocks covering the bases minus the excludes. The list is sorted by network address ascending. Returns an empty list if the excludes fully cover the bases or if `bases` is empty. Returns the collapsed bases if `excludes` is empty.

## Error Conditions

- Invalid CIDR notation in any entry
- Host bits set in any entry (e.g. `10.0.1.5/24` instead of `10.0.1.0/24`)
- An IPv6 entry (only IPv4 is supported)
