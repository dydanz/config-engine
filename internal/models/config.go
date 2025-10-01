package models

import (
	"encoding/json"
	"time"
)

// Config represents a configuration with versioning support
type Config struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ConfigVersion represents a specific version of a configuration
type ConfigVersion struct {
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// CreateConfigRequest represents the request to create a new configuration
type CreateConfigRequest struct {
	Name string                 `json:"name"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// UpdateConfigRequest represents the request to update a configuration
type UpdateConfigRequest struct {
	Data map[string]interface{} `json:"data"`
}

// RollbackRequest represents the request to rollback to a specific version
type RollbackRequest struct {
	Version int `json:"version"`
}

// VersionsResponse represents the response containing all versions
type VersionsResponse struct {
	Name     string          `json:"name"`
	Versions []ConfigVersion `json:"versions"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Validate validates the CreateConfigRequest
func (r *CreateConfigRequest) Validate() error {
	if r.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if r.Type == "" {
		return &ValidationError{Field: "type", Message: "type is required"}
	}
	if r.Data == nil {
		return &ValidationError{Field: "data", Message: "data is required"}
	}
	return nil
}

// Validate validates the UpdateConfigRequest
func (r *UpdateConfigRequest) Validate() error {
	if r.Data == nil {
		return &ValidationError{Field: "data", Message: "data is required"}
	}
	return nil
}

// Validate validates the RollbackRequest
func (r *RollbackRequest) Validate() error {
	if r.Version < 1 {
		return &ValidationError{Field: "version", Message: "version must be >= 1"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ConfigNotFoundError represents a configuration not found error
type ConfigNotFoundError struct {
	Name string
}

func (e *ConfigNotFoundError) Error() string {
	return "configuration not found: " + e.Name
}

// ConfigExistsError represents a configuration already exists error
type ConfigExistsError struct {
	Name string
}

func (e *ConfigExistsError) Error() string {
	return "configuration already exists: " + e.Name
}

// VersionNotFoundError represents a version not found error
type VersionNotFoundError struct {
	Name    string
	Version int
}

func (e *VersionNotFoundError) Error() string {
	return "version not found"
}

// SchemaValidationError represents a schema validation error
type SchemaValidationError struct {
	Details string
}

func (e *SchemaValidationError) Error() string {
	return "schema validation failed: " + e.Details
}

// UnmarshalCreateConfigRequest unmarshals JSON into CreateConfigRequest
func UnmarshalCreateConfigRequest(data []byte) (*CreateConfigRequest, error) {
	var req CreateConfigRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// UnmarshalUpdateConfigRequest unmarshals JSON into UpdateConfigRequest
func UnmarshalUpdateConfigRequest(data []byte) (*UpdateConfigRequest, error) {
	var req UpdateConfigRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// UnmarshalRollbackRequest unmarshals JSON into RollbackRequest
func UnmarshalRollbackRequest(data []byte) (*RollbackRequest, error) {
	var req RollbackRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}