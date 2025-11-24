package netactuate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/netactuate/gona/gona"
	"github.com/stretchr/testify/assert"
)

var _ provider.Provider = (*FrameworkProvider)(nil)

func TestNewFrameworkProvider(t *testing.T) {
	const version = "1.2.3"
	p := NewFrameworkProvider(version)

	assert.NotNil(t, p, "NewFrameworkProvider should not return nil")

	fp, ok := p.(*FrameworkProvider)
	assert.True(t, ok, "NewFrameworkProvider should return *FrameworkProvider")
	assert.Equal(t, version, fp.version, "version should match")
	assert.NotEmpty(t, fp.gonaVersion, "gonaVersion should be set from gona.Version")
}

func TestFrameworkProvider_Metadata(t *testing.T) {
	const version = "1.2.3"
	p := &FrameworkProvider{version: version}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	assert.Equal(t, "netactuate", resp.TypeName, "TypeName should be 'netactuate'")
	assert.Equal(t, version, resp.Version, "Version should match")
}

func TestFrameworkProvider_Schema(t *testing.T) {
	p := &FrameworkProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	assert.NotNil(t, resp.Schema.Attributes, "Schema.Attributes should not be nil")

	// Check that api_key attribute exists and is configured correctly
	apiKeyAttr, ok := resp.Schema.Attributes["api_key"]
	assert.True(t, ok, "api_key attribute should be in schema")
	assert.NotNil(t, apiKeyAttr, "api_key attribute should not be nil")

	// Check that api_url attribute exists
	apiUrlAttr, ok := resp.Schema.Attributes["api_url"]
	assert.True(t, ok, "api_url attribute should be in schema")
	assert.NotNil(t, apiUrlAttr, "api_url attribute should not be nil")
}

func TestFrameworkProvider_Configure_Success(t *testing.T) {
	p := &FrameworkProvider{version: "test"}

	// Create a config with api_key set
	configValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key": tftypes.String,
				"api_url": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"api_key": tftypes.NewValue(tftypes.String, "test-api-key"),
			"api_url": tftypes.NewValue(tftypes.String, nil),
		},
	)

	// Get the provider schema to create config
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "should not have errors")
	assert.NotNil(t, resp.ResourceData, "ResourceData should be set")
	assert.NotNil(t, resp.DataSourceData, "DataSourceData should be set")

	// Verify it's a gona client
	_, ok := resp.ResourceData.(*gona.Client)
	assert.True(t, ok, "ResourceData should be *gona.Client")

	_, ok = resp.DataSourceData.(*gona.Client)
	assert.True(t, ok, "DataSourceData should be *gona.Client")
}

func TestFrameworkProvider_Configure_WithCustomURL(t *testing.T) {
	p := &FrameworkProvider{version: "test"}

	// Create a config with both api_key and api_url set
	configValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key": tftypes.String,
				"api_url": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"api_key": tftypes.NewValue(tftypes.String, "test-api-key"),
			"api_url": tftypes.NewValue(tftypes.String, "https://custom.api.example.com"),
		},
	)

	// Get the provider schema to create config
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "should not have errors")
	assert.NotNil(t, resp.ResourceData, "ResourceData should be set")
	assert.NotNil(t, resp.DataSourceData, "DataSourceData should be set")
}

func TestFrameworkProvider_Configure_MissingAPIKey(t *testing.T) {
	p := &FrameworkProvider{version: "test"}

	// Create a config with empty api_key
	configValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key": tftypes.String,
				"api_url": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"api_key": tftypes.NewValue(tftypes.String, ""),
			"api_url": tftypes.NewValue(tftypes.String, nil),
		},
	)

	// Get the provider schema to create config
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "should have error when api_key is missing")
	assert.NotEmpty(t, resp.Diagnostics.Errors(), "should have at least one error diagnostic")

	errorSummary := resp.Diagnostics.Errors()[0].Summary()
	assert.Equal(t, "Unable to create NetActuate API client", errorSummary, "error summary should match")
}

func TestFrameworkProvider_Resources(t *testing.T) {
	p := &FrameworkProvider{}

	resources := p.Resources(context.Background())

	assert.NotNil(t, resources, "Resources should not return nil")
	// Currently should be empty during mux phase
	assert.Empty(t, resources, "should have 0 resources during mux phase")
}

func TestFrameworkProvider_DataSources(t *testing.T) {
	p := &FrameworkProvider{}

	dataSources := p.DataSources(context.Background())

	assert.NotNil(t, dataSources, "DataSources should not return nil")
	// Currently should be empty during mux phase
	assert.Empty(t, dataSources, "should have 0 data sources during mux phase")
}

// TestFrameworkProvider_ProviderServer verifies the provider can be served
func TestFrameworkProvider_ProviderServer(t *testing.T) {
	p := NewFrameworkProvider("test")

	// This should not panic
	server := providerserver.NewProtocol6(p)
	assert.NotNil(t, server, "providerserver.NewProtocol6 should not return nil")
}
