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
var _ function.Function = (*CidrSubtractSetFunction)(nil)

// CidrSubtractSetFunction implements the cidrsubtract_set provider-defined
// function: the union of several base ranges minus the union of several
// excluded ranges, with no containment requirement between the two.
type CidrSubtractSetFunction struct{}

// NewCidrSubtractSetFunction returns a new function.Function for cidrsubtract_set.
func NewCidrSubtractSetFunction() function.Function {
	return &CidrSubtractSetFunction{}
}

func (f *CidrSubtractSetFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "cidrsubtract_set"
}

func (f *CidrSubtractSetFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Compute CIDR set difference",
		Description: "Returns the minimal set of IPv4 CIDRs covering the union of bases minus the union of excludes",
		Parameters: []function.Parameter{
			function.ListParameter{
				Name:        "bases",
				Description: "IPv4 CIDR blocks to subtract from. May overlap one another.",
				ElementType: types.StringType,
			},
			function.ListParameter{
				Name:        "excludes",
				Description: "IPv4 CIDR blocks to exclude. May overlap the bases in any way.",
				ElementType: types.StringType,
			},
		},
		Return: function.ListReturn{
			ElementType: types.StringType,
		},
	}
}

func (f *CidrSubtractSetFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var bases, excludes []string

	resp.Error = function.ConcatFuncErrors(
		req.Arguments.Get(ctx, &bases, &excludes),
	)
	if resp.Error != nil {
		return
	}

	basePfxs, baseErr := parseIPv4Prefixes(bases, 0, "base")
	excludePfxs, excludeErr := parseIPv4Prefixes(excludes, 1, "exclude")
	if funcErrors := function.ConcatFuncErrors(baseErr, excludeErr); funcErrors != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErrors)
		return
	}

	result := cidr.SubtractAll(basePfxs, excludePfxs)

	resultStrs := make([]string, len(result))
	for i, p := range result {
		resultStrs[i] = p.String()
	}

	resp.Error = function.ConcatFuncErrors(resp.Error,
		resp.Result.Set(ctx, resultStrs),
	)
}

// parseIPv4Prefixes parses entries as IPv4 CIDRs with no host bits set,
// reporting every invalid entry against the given argument position.
func parseIPv4Prefixes(entries []string, arg int, kind string) ([]netip.Prefix, *function.FuncError) {
	var funcErrors *function.FuncError
	pfxs := make([]netip.Prefix, 0, len(entries))

	for i, entry := range entries {
		pfx, err := netip.ParsePrefix(entry)
		switch {
		case err != nil:
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(int64(arg),
					fmt.Sprintf("invalid %s entry at index %d: %s", kind, i, entry)))
		case !pfx.Addr().Is4():
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(int64(arg),
					fmt.Sprintf("%s entry at index %d: %s is not an IPv4 CIDR", kind, i, entry)))
		case pfx != pfx.Masked():
			funcErrors = function.ConcatFuncErrors(funcErrors,
				function.NewArgumentFuncError(int64(arg),
					fmt.Sprintf("%s entry at index %d: %s has host bits set; did you mean %s?", kind, i, entry, pfx.Masked())))
		default:
			pfxs = append(pfxs, pfx)
		}
	}

	return pfxs, funcErrors
}
