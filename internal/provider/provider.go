package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure interfaces are satisfied at compile time.
var (
	_ provider.Provider              = (*CidrMathProvider)(nil)
	_ provider.ProviderWithFunctions = (*CidrMathProvider)(nil)
)

// CidrMathProvider implements a Terraform provider that exposes the
// cidrmath provider-defined function. It has no configuration, resources,
// or data sources.
type CidrMathProvider struct{
	version string
	date    string
	commit  string
}

func (p *CidrMathProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cidrmath"
	resp.Version = p.version
}

func (p *CidrMathProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *CidrMathProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *CidrMathProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *CidrMathProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *CidrMathProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewCidrSubtractFunction,
		NewCidrSubtractSetFunction,
	}
}

// New returns a new provider.Provider instance for use with providerserver.Serve.
func New() provider.Provider {
	return &CidrMathProvider{}
}
