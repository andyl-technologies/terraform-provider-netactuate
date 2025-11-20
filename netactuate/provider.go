package netactuate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netactuate/gona/gona"
)

const (
	ProviderVersion = "0.4.0-dev"
)

// Provider returns the SDK v2 provider (legacy)
// This is kept for backward compatibility during migration
func Provider() *schema.Provider {
	return NewSDKProvider(ProviderVersion)
}

// NewSDKProvider creates a new SDK v2 provider instance
// Used during mux-based migration
func NewSDKProvider(version string) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETACTUATE_API_KEY", nil),
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"netactuate_server":       resourceServer(),
			"netactuate_sshkey":       resourceSshKey(),
			"netactuate_bgp_sessions": resourceBGPSessions(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"netactuate_server":       dataSourceServer(),
			"netactuate_sshkey":       dataSourceSshKey(),
			"netactuate_bgp_sessions": dataSourceBGPSessions(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiKey := d.Get("api_key").(string)
	apiUrl := d.Get("api_url").(string)

	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create NetActuate API client",
			Detail: `Unable to find NetActuate API key. It can be set with either NETACTUATE_API_KEY environment
variable or 'api_key' property`,
		})
		return nil, diags
	}

	if apiUrl == "" {
		return gona.NewClient(apiKey), nil
	} else {
		return gona.NewClientCustom(apiKey, apiUrl), nil
	}

}
