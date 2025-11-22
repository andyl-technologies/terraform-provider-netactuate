package netactuate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/netactuate/gona/gona"
)

var _ provider.Provider = (*FrameworkProvider)(nil)

func TestNewFrameworkProvider(t *testing.T) {
	const version = "1.2.3"
	p := NewFrameworkProvider(version)

	if p == nil {
		t.Fatal("NewFrameworkProvider returned nil")
	}

	fp, ok := p.(*FrameworkProvider)
	if !ok {
		t.Fatal("NewFrameworkProvider did not return *FrameworkProvider")
	}

	if fp.version != version {
		t.Errorf("expected version %q, got %q", version, fp.version)
	}

	if fp.gonaVersion == "" {
		t.Error("gonaVersion should be set from gona.Version")
	}
}

func TestFrameworkProvider_Metadata(t *testing.T) {
	const version = "1.2.3"
	p := &FrameworkProvider{version: version}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "netactuate" {
		t.Errorf("expected TypeName %q, got %q", "netactuate", resp.TypeName)
	}

	if resp.Version != version {
		t.Errorf("expected Version %q, got %q", version, resp.Version)
	}
}

func TestFrameworkProvider_Schema(t *testing.T) {
	p := &FrameworkProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Schema.Attributes should not be nil")
	}

	// Check that api_key attribute exists and is configured correctly
	apiKeyAttr, ok := resp.Schema.Attributes["api_key"]
	if !ok {
		t.Fatal("api_key attribute not found in schema")
	}

	// Verify it's a StringAttribute (we can't directly type assert, but we can check it exists)
	if apiKeyAttr == nil {
		t.Error("api_key attribute should not be nil")
	}

	// Check that api_url attribute exists
	apiUrlAttr, ok := resp.Schema.Attributes["api_url"]
	if !ok {
		t.Fatal("api_url attribute not found in schema")
	}

	if apiUrlAttr == nil {
		t.Error("api_url attribute should not be nil")
	}
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

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}

	if resp.ResourceData == nil {
		t.Error("ResourceData should be set")
	}

	if resp.DataSourceData == nil {
		t.Error("DataSourceData should be set")
	}

	// Verify it's a gona client
	if _, ok := resp.ResourceData.(*gona.Client); !ok {
		t.Error("ResourceData should be *gona.Client")
	}

	if _, ok := resp.DataSourceData.(*gona.Client); !ok {
		t.Error("DataSourceData should be *gona.Client")
	}
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

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}

	if resp.ResourceData == nil {
		t.Error("ResourceData should be set")
	}

	if resp.DataSourceData == nil {
		t.Error("DataSourceData should be set")
	}
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

	if !resp.Diagnostics.HasError() {
		t.Error("expected error when api_key is missing")
	}

	// Verify the error message is appropriate
	if len(resp.Diagnostics.Errors()) == 0 {
		t.Fatal("expected at least one error diagnostic")
	}

	errorSummary := resp.Diagnostics.Errors()[0].Summary()
	if errorSummary != "Unable to create NetActuate API client" {
		t.Errorf("unexpected error summary: %q", errorSummary)
	}
}

func TestFrameworkProvider_Resources(t *testing.T) {
	p := &FrameworkProvider{}

	resources := p.Resources(context.Background())

	if resources == nil {
		t.Error("Resources should not return nil")
	}

	// Currently should be empty during mux phase
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestFrameworkProvider_DataSources(t *testing.T) {
	p := &FrameworkProvider{}

	dataSources := p.DataSources(context.Background())

	if dataSources == nil {
		t.Error("DataSources should not return nil")
	}

	// Currently should be empty during mux phase
	if len(dataSources) != 0 {
		t.Errorf("expected 0 data sources, got %d", len(dataSources))
	}
}

// TestFrameworkProvider_ProviderServer verifies the provider can be served
func TestFrameworkProvider_ProviderServer(t *testing.T) {
	p := NewFrameworkProvider("test")

	// This should not panic
	if providerserver.NewProtocol6(p) == nil {
		t.Error("providerserver.NewProtocol6 returned nil")
	}
}
