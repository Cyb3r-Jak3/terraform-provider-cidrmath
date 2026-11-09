package provider

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Cyb3r-Jak3/terraform-provider-cidrmath/internal/cidr"
)

// Ensure interface is satisfied at compile time.
var _ function.Function = (*CidrSubtractFunction)(nil)

// CidrSubtractFunction implements the cidrsubtract provider-defined function.
type CidrSubtractFunction struct{}

// NewCidrSubtractFunction returns a new function.Function for cidrsubtract.
func NewCidrSubtractFunction() function.Function {
	return &CidrSubtractFunction{}
}

func (f *CidrSubtractFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "cidrsubtract"
}

func (f *CidrSubtractFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Compute CIDR complement",
		Description: "Returns the minimal set of CIDRs covering base minus excludes",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "base",
				Description: "Base CIDR block (e.g. 10.0.0.0/16)",
			},
			function.ListParameter{
				Name:        "excludes",
				Description: "CIDR blocks to exclude from the base",
				ElementType: types.StringType,
			},
		},
		Return: function.ListReturn{
			ElementType: types.StringType,
		},
	}
}

func (f *CidrSubtractFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var base string
	var excludes []string

	resp.Error = function.ConcatFuncErrors(
		req.Arguments.Get(ctx, &base, &excludes),
	)
	if resp.Error != nil {
		return
	}

	// Validate base CIDR.
	var funcErrors *function.FuncError
	basePfx, err := netip.ParsePrefix(base)
	if err != nil {
		funcErrors = function.ConcatFuncErrors(funcErrors,
			function.NewArgumentFuncError(0,
				fmt.Sprintf("invalid base CIDR: %s — expected format like 10.0.0.0/16", base),
			),
		)
	} else if basePfx != basePfx.Masked() {
		funcErrors = function.ConcatFuncErrors(funcErrors,
			function.NewArgumentFuncError(0,
				fmt.Sprintf("%s has host bits set; did you mean %s?", base, basePfx.Masked()),
			),
		)
	}

	// Validate each exclude entry.
	var excludePfxs []netip.Prefix
	baseValid := funcErrors == nil // base is valid only if no errors so far

	for i, ex := range excludes {
		pfx, err := netip.ParsePrefix(ex)
		if err != nil {
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(1,
					fmt.Sprintf("invalid exclude entry at index %d: %s", i, ex),
				),
			)
			continue
		}
		if pfx != pfx.Masked() {
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(1,
					fmt.Sprintf("exclude entry at index %d: %s has host bits set; did you mean %s?", i, ex, pfx.Masked()),
				),
			)
			continue
		}
		// Only check containment if base was valid.
		if baseValid && !prefixContains(basePfx, pfx) {
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(1,
					fmt.Sprintf("exclude entry at index %d: %s is not contained within base %s", i, ex, base),
				),
			)
			continue
		}
		excludePfxs = append(excludePfxs, pfx)
	}

	if funcErrors != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErrors)
		return
	}

	// Compute the CIDR complement.
	result := cidr.Subtract(basePfx, excludePfxs)

	// Format result as string list.
	resultStrs := make([]string, len(result))
	for i, p := range result {
		resultStrs[i] = p.String()
	}

	resp.Error = function.ConcatFuncErrors(resp.Error,
		resp.Result.Set(ctx, resultStrs),
	)
}

// prefixContains reports whether outer fully contains inner.
func prefixContains(outer, inner netip.Prefix) bool {
	if inner.Bits() < outer.Bits() {
		return false
	}
	if !outer.Contains(inner.Addr()) {
		return false
	}
	// Check last address of inner is within outer.
	// For a prefix, the last address has all host bits set to 1.
	addr := inner.Addr()
	b := addr.As4()
	ip := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	bits := inner.Bits()
	if bits < 32 {
		hostMask := uint32(1)<<(32-bits) - 1
		ip |= hostMask
	}
	lastAddr := netip.AddrFrom4([4]byte{byte(ip >> 24), byte(ip >> 16), byte(ip >> 8), byte(ip)})
	return outer.Contains(lastAddr)
}
