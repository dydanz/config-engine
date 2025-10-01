package validation

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// Validator handles configuration validation against schemas
type Validator struct {
	schemas map[string]*gojsonschema.Schema
}

// NewValidator creates a new validator with predefined schemas
func NewValidator() (*Validator, error) {
	v := &Validator{
		schemas: make(map[string]*gojsonschema.Schema),
	}

	// Register payment_config schema
	paymentSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"max_limit": map[string]interface{}{
				"type": "integer",
			},
			"enabled": map[string]interface{}{
				"type": "boolean",
			},
		},
		"required":             []string{"max_limit", "enabled"},
		"additionalProperties": false,
	}

	if err := v.RegisterSchema("payment_config", paymentSchema); err != nil {
		return nil, fmt.Errorf("failed to register payment_config schema: %w", err)
	}

	return v, nil
}

// RegisterSchema registers a new schema for a configuration type
func (v *Validator) RegisterSchema(configType string, schema map[string]interface{}) error {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	compiledSchema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	v.schemas[configType] = compiledSchema
	return nil
}

// Validate validates configuration data against its type's schema
func (v *Validator) Validate(configType string, data map[string]interface{}) error {
	schema, exists := v.schemas[configType]
	if !exists {
		return fmt.Errorf("no schema found for config type: %s", configType)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(dataJSON)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		errors := ""
		for i, desc := range result.Errors() {
			if i > 0 {
				errors += "; "
			}
			errors += fmt.Sprintf("%s: %s", desc.Field(), desc.Description())
		}
		return fmt.Errorf("%s", errors)
	}

	return nil
}

// HasSchema checks if a schema exists for the given config type
func (v *Validator) HasSchema(configType string) bool {
	_, exists := v.schemas[configType]
	return exists
}