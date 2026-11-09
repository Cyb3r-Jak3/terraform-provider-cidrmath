// Package cidr provides pure CIDR arithmetic functions for computing the
// complement of a base prefix minus a set of excluded prefixes.
package cidr

import (
	"encoding/binary"
	"net/netip"
	"slices"
)

// prefixContains reports whether outer fully contains inner.
// Both prefixes must be valid. outer contains inner when inner's network
// address and last address both fall within outer's range, and inner's
// prefix length is >= outer's prefix length.
func prefixContains(outer, inner netip.Prefix) bool {
	if inner.Bits() < outer.Bits() {
		return false
	}
	// inner's first address must be in outer.
	if !outer.Contains(inner.Addr()) {
		return false
	}
	// inner's last address must also be in outer.
	last := prefixLastAddr(inner)
	return outer.Contains(last)
}

// prefixLastAddr returns the last (highest) IP address in a prefix.
func prefixLastAddr(p netip.Prefix) netip.Addr {
	addr := p.Addr()
	b := addr.As4()
	ip := binary.BigEndian.Uint32(b[:])
	bits := p.Bits()
	if bits == 32 {
		return addr
	}
	// Set all host bits to 1.
	hostMask := uint32(1)<<(32-bits) - 1
	ip |= hostMask
	binary.BigEndian.PutUint32(b[:], ip)
	return netip.AddrFrom4(b)
}

// prefixOverlaps reports whether two prefixes share any IP address.
func prefixOverlaps(a, b netip.Prefix) bool {
	// If one contains the other's network address, they overlap.
	// We check both directions to handle the case where one is
	// a superset of the other.
	return a.Overlaps(b)
}

// splitPrefix splits a prefix into its two child halves by incrementing
// the prefix length by 1. Given a /N prefix, it returns two /(N+1) prefixes:
// lo has the same network address, hi has the bit at position N+1 set.
// Panics if called on a /32 prefix (cannot split further).
func splitPrefix(p netip.Prefix) (lo, hi netip.Prefix) {
	bits := p.Bits()
	if bits >= 32 {
		panic("cannot split a /32 prefix")
	}
	newBits := bits + 1

	// lo half: same network address, one more bit of prefix.
	lo = netip.PrefixFrom(p.Addr(), newBits)

	// hi half: set the bit at position (bits) in the address.
	b := p.Addr().As4()
	ip := binary.BigEndian.Uint32(b[:])
	ip |= 1 << (31 - bits) // set the bit at position `bits` (0-indexed from MSB)
	binary.BigEndian.PutUint32(b[:], ip)
	hi = netip.PrefixFrom(netip.AddrFrom4(b), newBits)

	return lo, hi
}

// sortPrefixes sorts a slice of prefixes by network address ascending,
// then by prefix length ascending (smaller length = larger block first)
// for prefixes sharing the same network address.
func sortPrefixes(prefixes []netip.Prefix) {
	slices.SortFunc(prefixes, func(a, b netip.Prefix) int {
		ac := a.Addr().Compare(b.Addr())
		if ac != 0 {
			return ac
		}
		return a.Bits() - b.Bits()
	})
}

// mergeAdjacent merges sibling CIDR pairs into their parent prefix,
// repeating until no more merges are possible. Two prefixes are siblings
// if they have the same prefix length N and combining them produces a
// valid /(N-1) prefix.
func mergeAdjacent(prefixes []netip.Prefix) []netip.Prefix {
	for {
		sortPrefixes(prefixes)
		merged := false
		var result []netip.Prefix
		i := 0
		for i < len(prefixes) {
			if i+1 < len(prefixes) && areSiblings(prefixes[i], prefixes[i+1]) {
				// Merge into parent: use the lower address with one less bit.
				parent := netip.PrefixFrom(prefixes[i].Addr(), prefixes[i].Bits()-1)
				result = append(result, parent)
				i += 2
				merged = true
			} else {
				result = append(result, prefixes[i])
				i++
			}
		}
		prefixes = result
		if !merged {
			break
		}
	}
	return prefixes
}

// areSiblings reports whether two prefixes are siblings — same prefix length,
// and they combine to form a valid parent prefix at one bit shorter.
func areSiblings(a, b netip.Prefix) bool {
	if a.Bits() != b.Bits() {
		return false
	}
	bits := a.Bits()
	if bits == 0 {
		return false // /0 has no parent
	}
	// The parent of both should be the same /(bits-1) prefix.
	parentA := netip.PrefixFrom(a.Addr(), bits-1).Masked()
	parentB := netip.PrefixFrom(b.Addr(), bits-1).Masked()
	return parentA == parentB
}

// Subtract returns the minimal set of CIDR prefixes that cover the base
// prefix minus all excluded prefixes. The result is sorted by network
// address ascending.
//
// Algorithm:
//  1. Normalize excludes — clip each to the base range, discard non-overlapping,
//     merge overlapping/adjacent entries.
//  2. Recursive binary splitting — if no excludes overlap the current prefix,
//     return it; if excludes fully cover it, return nothing; otherwise split
//     into two halves and recurse.
//  3. Merge pass — combine adjacent sibling CIDRs for minimality.
func Subtract(base netip.Prefix, excludes []netip.Prefix) []netip.Prefix {
	base = base.Masked()

	// Normalize: clip excludes to base, discard non-overlapping.
	var clipped []netip.Prefix
	for _, ex := range excludes {
		ex = ex.Masked()
		if !prefixOverlaps(base, ex) {
			continue
		}
		// Clip: if exclude is larger than base, narrow it to base.
		if prefixContains(ex, base) {
			clipped = append(clipped, base)
		} else if prefixContains(base, ex) {
			clipped = append(clipped, ex)
		} else {
			// Partial overlap — the exclude extends beyond base on one side.
			// We keep only the portion inside base by splitting recursively.
			clipped = append(clipped, clipToBase(base, ex)...)
		}
	}

	if len(clipped) == 0 {
		return []netip.Prefix{base}
	}

	// Merge overlapping/adjacent excludes into a minimal sorted set.
	clipped = mergeOverlapping(clipped)

	// Recursive subtract.
	result := subtractRecursive(base, clipped)

	// Final merge pass for minimality and sort.
	result = mergeAdjacent(result)
	sortPrefixes(result)
	return result
}

// clipToBase returns the portions of ex that fall within base.
// This handles the case where ex partially overlaps base.
func clipToBase(base, ex netip.Prefix) []netip.Prefix {
	if !prefixOverlaps(base, ex) {
		return nil
	}
	if prefixContains(base, ex) {
		return []netip.Prefix{ex}
	}
	if prefixContains(ex, base) {
		return []netip.Prefix{base}
	}
	// Partial overlap: split ex and recurse.
	if ex.Bits() >= 32 {
		// /32 that overlaps but isn't contained — shouldn't happen after above checks.
		return nil
	}
	lo, hi := splitPrefix(ex)
	var result []netip.Prefix
	result = append(result, clipToBase(base, lo)...)
	result = append(result, clipToBase(base, hi)...)
	return result
}

// mergeOverlapping merges a list of prefixes into a minimal non-overlapping
// sorted set. It first sorts, then greedily merges overlapping/contained
// prefixes, and finally merges adjacent siblings.
func mergeOverlapping(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) <= 1 {
		return prefixes
	}
	sortPrefixes(prefixes)

	var merged []netip.Prefix
	current := prefixes[0]
	for i := 1; i < len(prefixes); i++ {
		if prefixContains(current, prefixes[i]) {
			// prefixes[i] is fully within current, skip it.
			continue
		}
		// Not contained — emit current and move on.
		merged = append(merged, current)
		current = prefixes[i]
	}
	merged = append(merged, current)

	// Also merge adjacent siblings for a truly minimal set.
	return mergeAdjacent(merged)
}

// subtractRecursive performs the core recursive binary-splitting subtraction.
func subtractRecursive(prefix netip.Prefix, excludes []netip.Prefix) []netip.Prefix {
	// Filter excludes to those overlapping this prefix.
	var relevant []netip.Prefix
	for _, ex := range excludes {
		if prefixOverlaps(prefix, ex) {
			relevant = append(relevant, ex)
		}
	}

	// No excludes overlap — this entire prefix is in the result.
	if len(relevant) == 0 {
		return []netip.Prefix{prefix}
	}

	// Check if any single exclude fully covers this prefix.
	for _, ex := range relevant {
		if prefixContains(ex, prefix) {
			return nil
		}
	}

	// Cannot split /32 further.
	if prefix.Bits() >= 32 {
		// If we got here, the /32 is not fully covered but has overlapping
		// excludes — which means the /32 IS one of the excludes. Return nil.
		return nil
	}

	// Split and recurse.
	lo, hi := splitPrefix(prefix)
	loResult := subtractRecursive(lo, relevant)
	hiResult := subtractRecursive(hi, relevant)

	var result []netip.Prefix
	result = append(result, loResult...)
	result = append(result, hiResult...)
	return result
}
