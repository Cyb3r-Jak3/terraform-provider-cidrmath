package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func subtractConfig(expr string) string {
	return fmt.Sprintf(`
output "test" {
  value = %s
}
`, expr)
}

func stringList(vals ...string) knownvalue.Check {
	checks := make([]knownvalue.Check, len(vals))
	for i, v := range vals {
		checks[i] = knownvalue.StringExact(v)
	}
	return knownvalue.ListExact(checks)
}

func TestCidrSubtractFunction(t *testing.T) {
	t.Parallel()

	success := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "no excludes returns base",
			expr:     `provider::cidrmath::cidrsubtract("10.0.0.0/16", [])`,
			expected: []string{"10.0.0.0/16"},
		},
		{
			name:     "exclude lower half",
			expr:     `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["10.0.0.0/25"])`,
			expected: []string{"10.0.0.128/25"},
		},
		{
			name:     "exclude entire base",
			expr:     `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["10.0.0.0/24"])`,
			expected: []string{},
		},
		{
			name:     "exclude middle",
			expr:     `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["10.0.0.64/26"])`,
			expected: []string{"10.0.0.0/26", "10.0.0.128/25"},
		},
	}

	for _, tc := range success {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: subtractConfig(tc.expr),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownOutputValue("test", stringList(tc.expected...)),
					},
				}},
			})
		})
	}

	failures := []struct {
		name  string
		expr  string
		error string
	}{
		{
			name:  "invalid base",
			expr:  `provider::cidrmath::cidrsubtract("not-a-cidr", [])`,
			error: `invalid base CIDR`,
		},
		{
			name:  "base with host bits",
			expr:  `provider::cidrmath::cidrsubtract("10.0.0.1/24", [])`,
			error: `host bits set`,
		},
		{
			name:  "invalid exclude",
			expr:  `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["bogus"])`,
			error: `invalid exclude entry at index 0`,
		},
		{
			name:  "exclude with host bits",
			expr:  `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["10.0.0.1/25"])`,
			error: `host bits set`,
		},
		{
			name:  "exclude outside base",
			expr:  `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["192.168.0.0/24"])`,
			error: `not contained within base`,
		},
		{
			name:  "exclude larger than base",
			expr:  `provider::cidrmath::cidrsubtract("10.0.0.0/24", ["10.0.0.0/16"])`,
			error: `not contained within base`,
		},
	}

	for _, tc := range failures {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      subtractConfig(tc.expr),
					ExpectError: regexp.MustCompile(tc.error),
				}},
			})
		})
	}
}
