package cidr

import "net/netip"

// SubtractAll returns the minimal set of CIDR prefixes that cover the union
// of bases minus the union of excludes. The result is sorted by network
// address ascending.
//
// Unlike Subtract, bases may overlap one another and excludes may overlap
// bases in any way: partially, entirely, or not at all.
func SubtractAll(bases, excludes []netip.Prefix) []netip.Prefix {
	masked := make([]netip.Prefix, len(bases))
	for i, b := range bases {
		masked[i] = b.Masked()
	}
	// Collapse duplicate and nested bases so each address is considered once.
	masked = mergeOverlapping(masked)

	var result []netip.Prefix
	for _, base := range masked {
		result = append(result, Subtract(base, excludes)...)
	}

	if len(result) == 0 {
		return []netip.Prefix{}
	}
	result = mergeAdjacent(result)
	sortPrefixes(result)
	return result
}
