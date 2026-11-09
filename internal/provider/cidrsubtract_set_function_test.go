package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestCidrSubtractSetFunction(t *testing.T) {
	t.Parallel()

	success := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "no excludes",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], [])`,
			expected: []string{"10.0.0.0/24"},
		},
		{
			name:     "no bases",
			expr:     `provider::cidrmath::cidrsubtract_set([], ["10.0.0.0/24"])`,
			expected: []string{},
		},
		{
			name:     "exclude lower half",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["10.0.0.0/25"])`,
			expected: []string{"10.0.0.128/25"},
		},
		{
			name:     "exclude outside base is ignored",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["192.168.0.0/24"])`,
			expected: []string{"10.0.0.0/24"},
		},
		{
			name:     "exclude larger than base removes it",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["10.0.0.0/16"])`,
			expected: []string{},
		},
		{
			name:     "multiple bases",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24", "10.0.1.0/24"], ["10.0.0.0/25", "10.0.1.128/25"])`,
			expected: []string{"10.0.0.128/25", "10.0.1.0/25"},
		},
		{
			name:     "overlapping bases are merged",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24", "10.0.0.0/25"], ["10.0.0.0/26"])`,
			expected: []string{"10.0.0.64/26", "10.0.0.128/25"},
		},
		{
			name:     "adjacent bases are aggregated",
			expr:     `provider::cidrmath::cidrsubtract_set(["10.0.0.0/25", "10.0.0.128/25"], [])`,
			expected: []string{"10.0.0.0/24"},
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
			expr:  `provider::cidrmath::cidrsubtract_set(["nope"], [])`,
			error: `invalid base entry at index 0`,
		},
		{
			name:  "invalid exclude",
			expr:  `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["nope"])`,
			error: `invalid exclude entry at index 0`,
		},
		{
			name:  "base with host bits",
			expr:  `provider::cidrmath::cidrsubtract_set(["10.0.0.1/24"], [])`,
			error: `host bits set`,
		},
		{
			name:  "exclude with host bits",
			expr:  `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["10.0.0.1/25"])`,
			error: `host bits set`,
		},
		{
			name:  "IPv6 base rejected",
			expr:  `provider::cidrmath::cidrsubtract_set(["2001:db8::/32"], [])`,
			error: `not an IPv4 CIDR`,
		},
		{
			name:  "IPv6 exclude rejected",
			expr:  `provider::cidrmath::cidrsubtract_set(["10.0.0.0/24"], ["2001:db8::/32"])`,
			error: `not an IPv4 CIDR`,
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
