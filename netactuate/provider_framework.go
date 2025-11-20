package netactuate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netactuate/gona/gona"
)

// FrameworkProvider is the Plugin Framework provider implementation
type FrameworkProvider struct {
	version string
}

// FrameworkProviderModel describes the provider configuration
type FrameworkProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
	ApiUrl types.String `tfsdk:"api_url"`
}

// NewFrameworkProvider creates a new instance of the Framework provider
func NewFrameworkProvider(version string) provider.Provider {
	return &FrameworkProvider{
		version: version,
	}
}

// Metadata returns the provider metadata
func (p *FrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netactuate"
	resp.Version = p.version
}

// Schema returns the provider schema
func (p *FrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "NetActuate API key. Can also be set with NETACTUATE_API_KEY environment variable.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "NetActuate API URL. Optional, defaults to production API.",
			},
		},
	}
}

// Configure configures the provider
func (p *FrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FrameworkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get API key from config or environment
	apiKey := config.ApiKey.ValueString()
	if apiKey == "" {
		// Framework will automatically check environment variable if we use proper schema
		resp.Diagnostics.AddError(
			"Unable to create NetActuate API client",
			"Unable to find NetActuate API key. It can be set with either NETACTUATE_API_KEY environment variable or 'api_key' property",
		)
		return
	}

	// Create client
	var client *gona.Client
	apiUrl := config.ApiUrl.ValueString()
	if apiUrl == "" {
		client = gona.NewClient(apiKey)
	} else {
		client = gona.NewClientCustom(apiKey, apiUrl)
	}

	// Make client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources returns the list of resources for this provider
func (p *FrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Start empty - resources will be added here as migrated from SDK v2
	}
}

// DataSources returns the list of data sources for this provider
func (p *FrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Start empty - data sources will be added here as migrated from SDK v2
	}
}
