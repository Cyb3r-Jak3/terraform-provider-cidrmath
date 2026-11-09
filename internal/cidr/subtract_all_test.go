package cidr

import (
	"encoding/binary"
	"math/rand"
	"net/netip"
	"slices"
	"testing"
)

func mustPrefixes(t *testing.T, ss ...string) []netip.Prefix {
	t.Helper()
	out := make([]netip.Prefix, len(ss))
	for i, s := range ss {
		out[i] = netip.MustParsePrefix(s)
	}
	return out
}

func prefixStrings(ps []netip.Prefix) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.String()
	}
	return out
}

func TestSubtractAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		bases    []string
		excludes []string
		want     []string
	}{
		{
			name:  "no excludes returns collapsed bases",
			bases: []string{"10.0.1.0/24", "10.0.0.0/24"},
			want:  []string{"10.0.0.0/23"},
		},
		{
			name:     "exclude partially overlaps base",
			bases:    []string{"10.0.0.0/24"},
			excludes: []string{"10.0.0.128/25", "10.0.1.0/24"},
			want:     []string{"10.0.0.0/25"},
		},
		{
			name:     "exclude larger than base removes it",
			bases:    []string{"10.0.0.0/24", "192.168.0.0/24"},
			excludes: []string{"10.0.0.0/8"},
			want:     []string{"192.168.0.0/24"},
		},
		{
			name:     "exclude spanning several bases",
			bases:    []string{"10.0.0.0/24", "10.0.2.0/24", "10.0.4.0/24"},
			excludes: []string{"10.0.1.0/24", "10.0.2.0/23"},
			want:     []string{"10.0.0.0/24", "10.0.4.0/24"},
		},
		{
			name:     "nested and duplicate bases",
			bases:    []string{"10.0.0.0/16", "10.0.1.0/24", "10.0.0.0/16"},
			excludes: []string{"10.0.0.0/17"},
			want:     []string{"10.0.128.0/17"},
		},
		{
			name:     "excludes cover everything",
			bases:    []string{"10.0.0.0/24"},
			excludes: []string{"10.0.0.0/24"},
			want:     []string{},
		},
		{
			name:     "no bases",
			excludes: []string{"10.0.0.0/24"},
			want:     []string{},
		},
		{
			name:     "single address hole",
			bases:    []string{"10.0.0.0/30"},
			excludes: []string{"10.0.0.1/32"},
			want:     []string{"10.0.0.0/32", "10.0.0.2/31"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prefixStrings(SubtractAll(mustPrefixes(t, tt.bases...), mustPrefixes(t, tt.excludes...)))
			if !slices.Equal(got, tt.want) {
				t.Errorf("SubtractAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

// universe is the address space the randomized test checks exhaustively.
var universe = netip.MustParsePrefix("10.0.0.0/22")

func randomPrefix(r *rand.Rand) netip.Prefix {
	bits := 22 + r.Intn(11) // /22 .. /32
	base := binary.BigEndian.Uint32(universe.Addr().AsSlice())
	offset := uint32(r.Intn(1 << 10))
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], base+offset)
	return netip.PrefixFrom(netip.AddrFrom4(b), bits).Masked()
}

func covered(addr netip.Addr, ps []netip.Prefix) bool {
	for _, p := range ps {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// TestSubtractAllRandomized checks SubtractAll against a brute-force
// address-by-address computation over a small universe.
func TestSubtractAllRandomized(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	start := binary.BigEndian.Uint32(universe.Addr().AsSlice())

	for iter := range 300 {
		bases := make([]netip.Prefix, r.Intn(6))
		for i := range bases {
			bases[i] = randomPrefix(r)
		}
		excludes := make([]netip.Prefix, r.Intn(10))
		for i := range excludes {
			excludes[i] = randomPrefix(r)
		}

		got := SubtractAll(bases, excludes)

		// Sorted, non-overlapping.
		for i := 1; i < len(got); i++ {
			if got[i-1].Overlaps(got[i]) {
				t.Fatalf("iter %d: result overlaps: %v and %v", iter, got[i-1], got[i])
			}
			if got[i-1].Addr().Compare(got[i].Addr()) >= 0 {
				t.Fatalf("iter %d: result not sorted: %v", iter, got)
			}
		}
		// Minimal: no sibling pair left unmerged.
		for i := 1; i < len(got); i++ {
			if areSiblings(got[i-1], got[i]) {
				t.Fatalf("iter %d: unmerged siblings %v and %v", iter, got[i-1], got[i])
			}
		}
		// Exact coverage over every address in the universe.
		for off := uint32(0); off < 1<<10; off++ {
			var b [4]byte
			binary.BigEndian.PutUint32(b[:], start+off)
			addr := netip.AddrFrom4(b)
			want := covered(addr, bases) && !covered(addr, excludes)
			if have := covered(addr, got); have != want {
				t.Fatalf("iter %d: %v covered=%v want %v\nbases=%v\nexcludes=%v\ngot=%v",
					iter, addr, have, want, bases, excludes, got)
			}
		}
	}
}
