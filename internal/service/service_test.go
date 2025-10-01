package service

import (
	"config-engine/internal/models"
	"config-engine/internal/repository"
	"config-engine/internal/validation"
	"testing"
)

func setupService(t *testing.T) *ConfigService {
	validator, err := validation.NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	repo := repository.NewInMemoryRepository()
	return NewConfigService(repo, validator)
}

func TestCreateConfig(t *testing.T) {
	svc := setupService(t)

	req := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}

	config, err := svc.CreateConfig(req)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if config.Name != "test_config" {
		t.Errorf("Expected name 'test_config', got '%s'", config.Name)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}
}

func TestCreateConfigValidation(t *testing.T) {
	svc := setupService(t)

	tests := []struct {
		name        string
		req         *models.CreateConfigRequest
		expectError bool
	}{
		{
			name: "missing name",
			req: &models.CreateConfigRequest{
				Type: "payment_config",
				Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
			},
			expectError: true,
		},
		{
			name: "missing type",
			req: &models.CreateConfigRequest{
				Name: "test",
				Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
			},
			expectError: true,
		},
		{
			name: "missing data",
			req: &models.CreateConfigRequest{
				Name: "test",
				Type: "payment_config",
			},
			expectError: true,
		},
		{
			name: "invalid schema - missing required field",
			req: &models.CreateConfigRequest{
				Name: "test",
				Type: "payment_config",
				Data: map[string]interface{}{"max_limit": 1000},
			},
			expectError: true,
		},
		{
			name: "invalid schema - wrong type",
			req: &models.CreateConfigRequest{
				Name: "test",
				Type: "payment_config",
				Data: map[string]interface{}{
					"max_limit": "not_a_number",
					"enabled":   true,
				},
			},
			expectError: true,
		},
		{
			name: "unknown config type",
			req: &models.CreateConfigRequest{
				Name: "test",
				Type: "unknown_type",
				Data: map[string]interface{}{"some": "data"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateConfig(tt.req)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Get config
	config, err := svc.GetConfig("test_config", nil)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if config.Name != "test_config" {
		t.Errorf("Expected name 'test_config', got '%s'", config.Name)
	}

	if config.Data["max_limit"].(int) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", config.Data["max_limit"])
	}
}

func TestGetConfigNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.GetConfig("nonexistent", nil)
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestGetConfigSpecificVersion(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Update config
	updateReq := &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	}
	svc.UpdateConfig("test_config", updateReq)

	// Get version 1
	version := 1
	config, err := svc.GetConfig("test_config", &version)
	if err != nil {
		t.Fatalf("Failed to get version 1: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Data["max_limit"].(int) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", config.Data["max_limit"])
	}
}

func TestUpdateConfig(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Update config
	updateReq := &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	}

	config, err := svc.UpdateConfig("test_config", updateReq)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	if config.Version != 2 {
		t.Errorf("Expected version 2, got %d", config.Version)
	}

	if config.Data["max_limit"].(int) != 2000 {
		t.Errorf("Expected max_limit 2000, got %v", config.Data["max_limit"])
	}
}

func TestUpdateConfigValidation(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Try to update with invalid data
	updateReq := &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": "invalid"},
	}

	_, err := svc.UpdateConfig("test_config", updateReq)
	if _, ok := err.(*models.SchemaValidationError); !ok {
		t.Errorf("Expected SchemaValidationError, got %v", err)
	}
}

func TestUpdateConfigNotFound(t *testing.T) {
	svc := setupService(t)

	updateReq := &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	}

	_, err := svc.UpdateConfig("nonexistent", updateReq)
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestRollbackConfig(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Update config multiple times
	svc.UpdateConfig("test_config", &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	})

	svc.UpdateConfig("test_config", &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 3000, "enabled": true},
	})

	// Rollback to version 1
	rollbackReq := &models.RollbackRequest{Version: 1}
	config, err := svc.RollbackConfig("test_config", rollbackReq)
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Should create version 4 with data from version 1
	if config.Version != 4 {
		t.Errorf("Expected version 4, got %d", config.Version)
	}

	if config.Data["max_limit"].(int) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", config.Data["max_limit"])
	}

	if config.Data["enabled"].(bool) != true {
		t.Errorf("Expected enabled true, got %v", config.Data["enabled"])
	}
}

func TestRollbackConfigInvalidVersion(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Try to rollback to non-existent version
	rollbackReq := &models.RollbackRequest{Version: 10}
	_, err := svc.RollbackConfig("test_config", rollbackReq)

	if _, ok := err.(*models.VersionNotFoundError); !ok {
		t.Errorf("Expected VersionNotFoundError, got %v", err)
	}
}

func TestRollbackConfigValidation(t *testing.T) {
	svc := setupService(t)

	tests := []struct {
		name        string
		req         *models.RollbackRequest
		expectError bool
	}{
		{
			name:        "invalid version - zero",
			req:         &models.RollbackRequest{Version: 0},
			expectError: true,
		},
		{
			name:        "invalid version - negative",
			req:         &models.RollbackRequest{Version: -1},
			expectError: true,
		},
		{
			name:        "valid version",
			req:         &models.RollbackRequest{Version: 1},
			expectError: false,
		},
	}

	// Create a config first
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.RollbackConfig("test_config", tt.req)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestListVersions(t *testing.T) {
	svc := setupService(t)

	// Create config
	createReq := &models.CreateConfigRequest{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	svc.CreateConfig(createReq)

	// Update multiple times
	svc.UpdateConfig("test_config", &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	})

	svc.UpdateConfig("test_config", &models.UpdateConfigRequest{
		Data: map[string]interface{}{"max_limit": 3000, "enabled": true},
	})

	// List versions
	response, err := svc.ListVersions("test_config")
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}

	if response.Name != "test_config" {
		t.Errorf("Expected name 'test_config', got '%s'", response.Name)
	}

	if len(response.Versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(response.Versions))
	}

	// Verify version data
	if response.Versions[0].Data["max_limit"].(int) != 1000 {
		t.Error("Version 1 data mismatch")
	}
	if response.Versions[1].Data["max_limit"].(int) != 2000 {
		t.Error("Version 2 data mismatch")
	}
	if response.Versions[2].Data["max_limit"].(int) != 3000 {
		t.Error("Version 3 data mismatch")
	}
}

func TestListVersionsNotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.ListVersions("nonexistent")
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}