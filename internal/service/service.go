package service

import (
	"fmt"

	"config-engine/internal/models"
	"config-engine/internal/repository"
	"config-engine/internal/validation"
)

// ConfigService handles business logic for configuration management
type ConfigService struct {
	repo      repository.ConfigRepository
	validator *validation.Validator
}

// NewConfigService creates a new configuration service
func NewConfigService(repo repository.ConfigRepository, validator *validation.Validator) *ConfigService {
	return &ConfigService{
		repo:      repo,
		validator: validator,
	}
}

// CreateConfig creates a new configuration
func (s *ConfigService) CreateConfig(req *models.CreateConfigRequest) (*models.Config, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if schema exists for this config type
	if !s.validator.HasSchema(req.Type) {
		return nil, &models.ValidationError{
			Field:   "type",
			Message: fmt.Sprintf("unknown config type: %s", req.Type),
		}
	}

	// Validate data against schema
	if err := s.validator.Validate(req.Type, req.Data); err != nil {
		return nil, &models.SchemaValidationError{Details: err.Error()}
	}

	// Create config
	config := &models.Config{
		Name: req.Name,
		Type: req.Type,
		Data: req.Data,
	}

	if err := s.repo.Create(config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetConfig retrieves a configuration by name
func (s *ConfigService) GetConfig(name string, version *int) (*models.Config, error) {
	if name == "" {
		return nil, &models.ValidationError{Field: "name", Message: "name is required"}
	}

	// If specific version requested
	if version != nil {
		configVersion, err := s.repo.GetVersion(name, *version)
		if err != nil {
			return nil, err
		}

		// Get the config to retrieve type info
		config, err := s.repo.Get(name)
		if err != nil {
			return nil, err
		}

		// Return a config with the requested version's data
		return &models.Config{
			Name:      name,
			Type:      config.Type,
			Version:   configVersion.Version,
			Data:      configVersion.Data,
			CreatedAt: config.CreatedAt,
			UpdatedAt: configVersion.CreatedAt,
		}, nil
	}

	// Return latest version
	return s.repo.Get(name)
}

// UpdateConfig updates an existing configuration
func (s *ConfigService) UpdateConfig(name string, req *models.UpdateConfigRequest) (*models.Config, error) {
	if name == "" {
		return nil, &models.ValidationError{Field: "name", Message: "name is required"}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing config to retrieve type
	existing, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	}

	// Validate data against schema
	if err := s.validator.Validate(existing.Type, req.Data); err != nil {
		return nil, &models.SchemaValidationError{Details: err.Error()}
	}

	// Update config
	config := &models.Config{
		Name: name,
		Type: existing.Type,
		Data: req.Data,
	}

	if err := s.repo.Update(config); err != nil {
		return nil, err
	}

	return config, nil
}

// RollbackConfig rolls back a configuration to a previous version
func (s *ConfigService) RollbackConfig(name string, req *models.RollbackRequest) (*models.Config, error) {
	if name == "" {
		return nil, &models.ValidationError{Field: "name", Message: "name is required"}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get the target version
	targetVersion, err := s.repo.GetVersion(name, req.Version)
	if err != nil {
		return nil, err
	}

	// Get current config to retrieve type
	current, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	}

	// Validate the historical data against current schema
	// (in case schema has changed since that version)
	if err := s.validator.Validate(current.Type, targetVersion.Data); err != nil {
		return nil, &models.SchemaValidationError{
			Details: fmt.Sprintf("target version data is incompatible with current schema: %s", err.Error()),
		}
	}

	// Create a new version with the historical data
	config := &models.Config{
		Name: name,
		Type: current.Type,
		Data: targetVersion.Data,
	}

	if err := s.repo.Update(config); err != nil {
		return nil, err
	}

	return config, nil
}

// ListVersions lists all versions of a configuration
func (s *ConfigService) ListVersions(name string) (*models.VersionsResponse, error) {
	if name == "" {
		return nil, &models.ValidationError{Field: "name", Message: "name is required"}
	}

	versions, err := s.repo.ListVersions(name)
	if err != nil {
		return nil, err
	}

	return &models.VersionsResponse{
		Name:     name,
		Versions: versions,
	}, nil
}