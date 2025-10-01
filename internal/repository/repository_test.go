package repository

import (
	"config-engine/internal/models"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}

	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if config.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestCreateDuplicate(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}

	repo.Create(config)
	err := repo.Create(config)

	if _, ok := err.(*models.ConfigExistsError); !ok {
		t.Errorf("Expected ConfigExistsError, got %v", err)
	}
}

func TestGet(t *testing.T) {
	repo := NewInMemoryRepository()

	original := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}

	repo.Create(original)

	retrieved, err := repo.Get("test_config")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, retrieved.Name)
	}

	if retrieved.Version != 1 {
		t.Errorf("Expected version 1, got %d", retrieved.Version)
	}
}

func TestGetNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	_, err := repo.Get("nonexistent")
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	repo := NewInMemoryRepository()

	// Create initial config
	original := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(original)

	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	// Update config
	updated := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 2000, "enabled": false},
	}

	err := repo.Update(updated)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	if updated.Version != 2 {
		t.Errorf("Expected version 2, got %d", updated.Version)
	}

	if updated.UpdatedAt.Before(original.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}

	// Verify the update is stored
	retrieved, _ := repo.Get("test_config")
	if retrieved.Version != 2 {
		t.Errorf("Expected stored version 2, got %d", retrieved.Version)
	}

	if retrieved.Data["max_limit"].(int) != 2000 {
		t.Errorf("Expected max_limit 2000, got %v", retrieved.Data["max_limit"])
	}
}

func TestUpdateNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "nonexistent",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}

	err := repo.Update(config)
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	repo := NewInMemoryRepository()

	// Create and update config multiple times
	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	config.Data = map[string]interface{}{"max_limit": 2000, "enabled": false}
	repo.Update(config)

	config.Data = map[string]interface{}{"max_limit": 3000, "enabled": true}
	repo.Update(config)

	// Get version 1
	v1, err := repo.GetVersion("test_config", 1)
	if err != nil {
		t.Fatalf("Failed to get version 1: %v", err)
	}

	if v1.Version != 1 {
		t.Errorf("Expected version 1, got %d", v1.Version)
	}

	if v1.Data["max_limit"].(int) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", v1.Data["max_limit"])
	}

	// Get version 2
	v2, err := repo.GetVersion("test_config", 2)
	if err != nil {
		t.Fatalf("Failed to get version 2: %v", err)
	}

	if v2.Data["max_limit"].(int) != 2000 {
		t.Errorf("Expected max_limit 2000, got %v", v2.Data["max_limit"])
	}
}

func TestGetVersionNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	// Try to get non-existent version
	_, err := repo.GetVersion("test_config", 5)
	if _, ok := err.(*models.VersionNotFoundError); !ok {
		t.Errorf("Expected VersionNotFoundError, got %v", err)
	}

	// Try to get version of non-existent config
	_, err = repo.GetVersion("nonexistent", 1)
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestListVersions(t *testing.T) {
	repo := NewInMemoryRepository()

	// Create and update config
	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	config.Data = map[string]interface{}{"max_limit": 2000, "enabled": false}
	repo.Update(config)

	versions, err := repo.ListVersions("test_config")
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	if versions[0].Version != 1 {
		t.Errorf("Expected first version to be 1, got %d", versions[0].Version)
	}

	if versions[1].Version != 2 {
		t.Errorf("Expected second version to be 2, got %d", versions[1].Version)
	}
}

func TestListVersionsNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	_, err := repo.ListVersions("nonexistent")
	if _, ok := err.(*models.ConfigNotFoundError); !ok {
		t.Errorf("Expected ConfigNotFoundError, got %v", err)
	}
}

func TestExists(t *testing.T) {
	repo := NewInMemoryRepository()

	if repo.Exists("test_config") {
		t.Error("Config should not exist yet")
	}

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	if !repo.Exists("test_config") {
		t.Error("Config should exist")
	}
}

func TestConcurrency(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	// Run concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				repo.Get("test_config")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				updated := &models.Config{
					Name: "test_config",
					Type: "payment_config",
					Data: map[string]interface{}{
						"max_limit": 1000 + id*100 + j,
						"enabled":   true,
					},
				}
				repo.Update(updated)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify final state is consistent
	final, err := repo.Get("test_config")
	if err != nil {
		t.Fatalf("Failed to get final config: %v", err)
	}

	if final.Version < 1 {
		t.Errorf("Expected version >= 1, got %d", final.Version)
	}
}

func TestDataIsolation(t *testing.T) {
	repo := NewInMemoryRepository()

	config := &models.Config{
		Name: "test_config",
		Type: "payment_config",
		Data: map[string]interface{}{"max_limit": 1000, "enabled": true},
	}
	repo.Create(config)

	// Get config and modify the returned data
	retrieved, _ := repo.Get("test_config")
	retrieved.Data["max_limit"] = 9999

	// Get config again and verify it wasn't affected
	retrieved2, _ := repo.Get("test_config")
	if retrieved2.Data["max_limit"].(int) != 1000 {
		t.Error("Data modification should not affect stored config")
	}
}