package validation

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	if validator == nil {
		t.Error("Validator should not be nil")
	}
}

func TestValidatePaymentConfig(t *testing.T) {
	validator, _ := NewValidator()

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config",
			data: map[string]interface{}{
				"max_limit": 1000,
				"enabled":   true,
			},
			expectError: false,
		},
		{
			name: "missing max_limit",
			data: map[string]interface{}{
				"enabled": true,
			},
			expectError: true,
		},
		{
			name: "missing enabled",
			data: map[string]interface{}{
				"max_limit": 1000,
			},
			expectError: true,
		},
		{
			name: "wrong type for max_limit",
			data: map[string]interface{}{
				"max_limit": "not_a_number",
				"enabled":   true,
			},
			expectError: true,
		},
		{
			name: "wrong type for enabled",
			data: map[string]interface{}{
				"max_limit": 1000,
				"enabled":   "not_a_boolean",
			},
			expectError: true,
		},
		{
			name: "additional properties not allowed",
			data: map[string]interface{}{
				"max_limit": 1000,
				"enabled":   true,
				"extra":     "field",
			},
			expectError: true,
		},
		{
			name: "max_limit as float",
			data: map[string]interface{}{
				"max_limit": 1000.5,
				"enabled":   true,
			},
			expectError: true, // JSON schema expects integer, not number
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("payment_config", tt.data)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateUnknownType(t *testing.T) {
	validator, _ := NewValidator()

	data := map[string]interface{}{
		"some": "data",
	}

	err := validator.Validate("unknown_type", data)
	if err == nil {
		t.Error("Expected error for unknown type")
	}
}

func TestHasSchema(t *testing.T) {
	validator, _ := NewValidator()

	if !validator.HasSchema("payment_config") {
		t.Error("Should have payment_config schema")
	}

	if validator.HasSchema("unknown_type") {
		t.Error("Should not have unknown_type schema")
	}
}

func TestRegisterSchema(t *testing.T) {
	validator, _ := NewValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"name"},
	}

	err := validator.RegisterSchema("custom_config", schema)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	if !validator.HasSchema("custom_config") {
		t.Error("Should have custom_config schema")
	}

	// Test validation with the new schema
	validData := map[string]interface{}{
		"name": "test",
	}

	err = validator.Validate("custom_config", validData)
	if err != nil {
		t.Errorf("Validation should succeed: %v", err)
	}

	// Test with invalid data
	invalidData := map[string]interface{}{
		"name": 123,
	}

	err = validator.Validate("custom_config", invalidData)
	if err == nil {
		t.Error("Expected validation error")
	}
}