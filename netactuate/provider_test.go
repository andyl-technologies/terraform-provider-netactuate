package netactuate

import (
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	p := Provider()

	if err := p.InternalValidate(); err != nil {
		t.Fatalf("InternalValidate() error: %s", err)
	}

	// Test schema fields
	t.Run("schema_api_key", func(t *testing.T) {
		apiKeySchema, ok := p.Schema["api_key"]
		if !ok {
			t.Fatal("api_key not found in provider schema")
		}
		if apiKeySchema.Type != schema.TypeString {
			t.Errorf("api_key type: expected TypeString, got %v", apiKeySchema.Type)
		}
		if !apiKeySchema.Optional {
			t.Error("api_key should be optional")
		}
	})

	t.Run("schema_api_url", func(t *testing.T) {
		apiUrlSchema, ok := p.Schema["api_url"]
		if !ok {
			t.Fatal("api_url not found in provider schema")
		}
		if apiUrlSchema.Type != schema.TypeString {
			t.Errorf("api_url type: expected TypeString, got %v", apiUrlSchema.Type)
		}
		if !apiUrlSchema.Optional {
			t.Error("api_url should be optional")
		}
	})

	// Test resources are registered
	t.Run("resources_registered", func(t *testing.T) {
		expectedResources := map[string]struct{}{
			"netactuate_server":       {},
			"netactuate_sshkey":       {},
			"netactuate_bgp_sessions": {},
		}

		for resourceName := range expectedResources {
			if _, ok := p.ResourcesMap[resourceName]; !ok {
				t.Errorf("resource %q not registered", resourceName)
			}
		}

		// Report unexpected resources
		for resourceName := range p.ResourcesMap {
			if _, ok := expectedResources[resourceName]; !ok {
				t.Errorf("unexpected resource %q registered", resourceName)
			}
		}
	})

	// Test data sources are registered
	t.Run("data_sources_registered", func(t *testing.T) {
		expectedDataSources := map[string]struct{}{
			"netactuate_server":       {},
			"netactuate_sshkey":       {},
			"netactuate_bgp_sessions": {},
		}

		for dataSourceName := range expectedDataSources {
			if _, ok := p.DataSourcesMap[dataSourceName]; !ok {
				t.Errorf("data source %q not registered", dataSourceName)
			}
		}

		// Report unexpected data sources
		for dataSourceName := range p.DataSourcesMap {
			if _, ok := expectedDataSources[dataSourceName]; !ok {
				t.Errorf("unexpected data source %q registered", dataSourceName)
			}
		}
	})

	// Test ConfigureContextFunc is set
	t.Run("configure_func_set", func(t *testing.T) {
		if p.ConfigureContextFunc == nil {
			t.Error("ConfigureContextFunc should be set")
		}
	})
}

func TestProviderConfigure(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		apiUrl        string
		useEnvVar     bool
		envVarValue   string
		expectError   bool
		errorContains string
		desc          string
	}{
		{
			name:   "valid api_key from config",
			apiKey: "test-api-key-123",
			desc:   "should configure with api_key from config",
		},
		{
			name:   "valid api_key with custom url",
			apiKey: "test-api-key-456",
			apiUrl: "https://custom.api.example.com",
			desc:   "should configure with custom api_url",
		},
		{
			name:   "valid api_key with empty url uses default",
			apiKey: "test-api-key-789",
			desc:   "should use default API URL when api_url is empty",
		},
		{
			name:          "missing api_key returns error",
			expectError:   true,
			errorContains: "Unable to create NetActuate API client",
			desc:          "should return error when api_key is missing",
		},
		{
			name:        "api_key from environment variable",
			useEnvVar:   true,
			envVarValue: "env-api-key-999",
			desc:        "should use NETACTUATE_API_KEY environment variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variable if needed
			if tt.useEnvVar {
				oldValue := os.Getenv("NETACTUATE_API_KEY")
				os.Setenv("NETACTUATE_API_KEY", tt.envVarValue)
				defer os.Setenv("NETACTUATE_API_KEY", oldValue)
			}

			// Create test ResourceData
			p := Provider()
			raw := map[string]any{}
			if tt.apiKey != "" {
				raw["api_key"] = tt.apiKey
			}
			if tt.apiUrl != "" {
				raw["api_url"] = tt.apiUrl
			}

			d := schema.TestResourceDataRaw(t, p.Schema, raw)

			// Call providerConfigure
			ctx := context.Background()
			client, diags := providerConfigure(ctx, d)

			// Check for errors
			if !tt.expectError {
				if diags.HasError() {
					t.Errorf("unexpected error: %v (%s)", diags, tt.desc)
				}
				if client == nil {
					t.Errorf("expected non-nil client, got nil (%s)", tt.desc)
				}
				return
			}
			if client != nil {
				t.Errorf("expected nil client on error, got %T (%s)", client, tt.desc)
			}
			if !diags.HasError() {
				t.Errorf("expected error but got none (%s)", tt.desc)
			} else if tt.errorContains != "" {
				// Check error message contains expected text
				if !slices.ContainsFunc(diags, func(diag diag.Diagnostic) bool {
					return strings.Contains(diag.Summary, tt.errorContains)
				}) {
					t.Errorf("expected error containing %q but got: %v (%s)",
						tt.errorContains, diags, tt.desc)
				}
			}
		})
	}
}

func TestProviderConfigureEnvironmentVariable(t *testing.T) {
	// Test that environment variable is properly used as default
	oldValue := os.Getenv("NETACTUATE_API_KEY")
	testKey := "env-test-key-from-environment"
	os.Setenv("NETACTUATE_API_KEY", testKey)
	defer os.Setenv("NETACTUATE_API_KEY", oldValue)

	p := Provider()

	// Create ResourceData without api_key in config
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{})

	ctx := context.Background()
	client, diags := providerConfigure(ctx, d)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags)
	}

	if client == nil {
		t.Error("expected non-nil client when using environment variable")
	}
}

func TestProviderConfigureCustomURL(t *testing.T) {
	p := Provider()
	customURL := "https://custom.netactuate.example.com/api"

	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"api_key": "test-key",
		"api_url": customURL,
	})

	ctx := context.Background()
	client, diags := providerConfigure(ctx, d)

	if len(diags) > 0 {
		t.Errorf("unexpected error: %v", diags)
	}

	if client == nil {
		t.Error("expected non-nil client with custom URL")
	}
}

func TestProviderConfigureMissingAPIKey(t *testing.T) {
	// Ensure no environment variable is set
	oldValue := os.Getenv("NETACTUATE_API_KEY")
	os.Unsetenv("NETACTUATE_API_KEY")
	defer os.Setenv("NETACTUATE_API_KEY", oldValue)

	p := Provider()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{})

	ctx := context.Background()
	client, diags := providerConfigure(ctx, d)

	if !diags.HasError() {
		t.Error("expected error for missing api_key")
	}

	if client != nil {
		t.Errorf("expected nil client on error, got %T", client)
	}

	// Verify error message is helpful
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	diag := diags[0]
	if !strings.Contains(diag.Summary, "Unable to create NetActuate API client") {
		t.Errorf("unexpected error summary: %s", diag.Summary)
	}
	if !strings.Contains(diag.Detail, "NETACTUATE_API_KEY") {
		t.Errorf("error detail should mention NETACTUATE_API_KEY, got: %s", diag.Detail)
	}
}
