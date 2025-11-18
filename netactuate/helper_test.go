package netactuate

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Helper function to create a test ResourceData with a simple schema
func testResourceData(t *testing.T) *schema.ResourceData {
	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"string_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"int_field": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"bool_field": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"list_field": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}, map[string]any{})
}

func TestSetValue(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         any
		expectError   bool
		errorContains string
		desc          string
	}{
		{
			name:  "set string value successfully",
			key:   "string_field",
			value: "test-value",
			desc:  "should set string value without error",
		},
		{
			name:  "set int value successfully",
			key:   "int_field",
			value: 42,
			desc:  "should set int value without error",
		},
		{
			name:  "set bool value successfully",
			key:   "bool_field",
			value: true,
			desc:  "should set bool value without error",
		},
		{
			name:          "set nonexistent key",
			key:           "nonexistent_field",
			value:         "value",
			expectError:   true,
			errorContains: "nonexistent_field",
			desc:          "should add error diagnostic for nonexistent key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testResourceData(t)
			var diags diag.Diagnostics

			setValue(tt.key, tt.value, d, &diags)

			if tt.expectError {
				if !diags.HasError() {
					t.Errorf("expected error but got none (%s)", tt.desc)
				}
				return
			}
			if diags.HasError() {
				t.Errorf("unexpected error: %v (%s)", diags, tt.desc)
			}
			// Verify value was actually set
			actualValue := d.Get(tt.key)
			if actualValue != tt.value {
				t.Errorf("expected value %v, got %v (%s)", tt.value, actualValue, tt.desc)
			}
		})
	}
}

func TestSetValueList(t *testing.T) {
	d := testResourceData(t)
	var diags diag.Diagnostics

	testList := []any{"item1", "item2"}
	setValue("list_field", testList, d, &diags)

	if len(diags) > 0 {
		t.Errorf("unexpected error: %v", diags)
	}

	// Verify list was set
	actualValue := d.Get("list_field")
	actualList, ok := actualValue.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", actualValue)
	}

	if len(actualList) != len(testList) {
		t.Errorf("expected list length %d, got %d", len(testList), len(actualList))
	}

	for i, expected := range testList {
		if actualList[i] != expected {
			t.Errorf("list[%d]: expected %v, got %v", i, expected, actualList[i])
		}
	}
}

func TestUpdateValue(t *testing.T) {
	tests := []struct {
		name         string
		setupKey     string
		setupValue   any
		updateKey    string
		updateValue  any
		shouldUpdate bool
		expectError  bool
		desc         string
	}{
		{
			name:         "update existing string field",
			setupKey:     "string_field",
			setupValue:   "original",
			updateKey:    "string_field",
			updateValue:  "updated",
			shouldUpdate: true,
			desc:         "should update existing field",
		},
		{
			name:         "update existing int field",
			setupKey:     "int_field",
			setupValue:   10,
			updateKey:    "int_field",
			updateValue:  20,
			shouldUpdate: true,
			desc:         "should update existing int field",
		},
		{
			name:        "skip update for non-existent field",
			setupKey:    "string_field",
			setupValue:  "value",
			updateKey:   "int_field",
			updateValue: 42,
			desc:        "should skip update when field doesn't exist in state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testResourceData(t)
			var diags diag.Diagnostics

			// Setup: set initial value
			err := d.Set(tt.setupKey, tt.setupValue)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Execute: call updateValue
			updateValue(tt.updateKey, tt.updateValue, d, &diags)

			// Verify: check errors
			if tt.expectError {
				if !diags.HasError() {
					t.Errorf("expected error but got none (%s)", tt.desc)
				}
			} else {
				if diags.HasError() {
					t.Errorf("unexpected error: %v (%s)", diags, tt.desc)
				}
			}

			// Verify: check if value was updated or not
			actualValue := d.Get(tt.updateKey)
			if tt.shouldUpdate {
				if actualValue != tt.updateValue {
					t.Errorf("expected value %v, got %v (%s)", tt.updateValue, actualValue, tt.desc)
				}
			} else {
				// Value should not have been set (should be zero value)
				zeroValue := actualValue
				if tt.updateKey == "string_field" && zeroValue != "" {
					t.Errorf("expected value to not be updated, but got %v (%s)", zeroValue, tt.desc)
				} else if tt.updateKey == "int_field" && zeroValue != 0 {
					t.Errorf("expected value to not be updated, but got %v (%s)", zeroValue, tt.desc)
				} else if tt.updateKey == "bool_field" && zeroValue != false {
					t.Errorf("expected value to not be updated, but got %v (%s)", zeroValue, tt.desc)
				}
			}
		})
	}
}

func TestSetValueDiagnosticsAccumulation(t *testing.T) {
	d := testResourceData(t)
	var diags diag.Diagnostics

	// Call setValue multiple times with errors
	setValue("nonexistent1", "value1", d, &diags)
	setValue("nonexistent2", "value2", d, &diags)
	setValue("string_field", "valid", d, &diags) // This should succeed

	if !diags.HasError() {
		t.Errorf("expected errors but got none")
	}

	// Should have 2 errors accumulated (not 3)
	if len(diags) != 2 {
		t.Errorf("expected 2 diagnostics, got %d: %v", len(diags), diags)
	}

	// All should be errors
	if slices.ContainsFunc(diags, func(diagnostic diag.Diagnostic) bool {
		return diagnostic.Severity != diag.Error
	}) {
		t.Errorf("expected all diagnostics to be errors: %v", diags)
	}
}

func TestUpdateValueDiagnosticsAccumulation(t *testing.T) {
	d := testResourceData(t)
	var diags diag.Diagnostics

	// Setup an existing field
	err := d.Set("string_field", "initial")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Call updateValue - this should skip (no error)
	updateValue("int_field", 42, d, &diags)

	// Call setValue on the same field with invalid key through updateValue
	// First set it, then update to a different invalid field
	err = d.Set("int_field", 10)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// This should succeed since int_field exists now
	updateValue("int_field", 20, d, &diags)

	// Should have no errors
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %v", len(diags), diags)
	}
}
