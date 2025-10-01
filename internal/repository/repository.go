package repository

import (
	"sync"
	"time"

	"config-engine/internal/models"
)

// ConfigRepository defines the interface for configuration storage
type ConfigRepository interface {
	Create(config *models.Config) error
	Get(name string) (*models.Config, error)
	Update(config *models.Config) error
	GetVersion(name string, version int) (*models.ConfigVersion, error)
	ListVersions(name string) ([]models.ConfigVersion, error)
	Exists(name string) bool
}

// InMemoryRepository implements ConfigRepository using in-memory storage
type InMemoryRepository struct {
	mu       sync.RWMutex
	configs  map[string]*models.Config
	versions map[string][]models.ConfigVersion // key: config name, value: list of versions
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		configs:  make(map[string]*models.Config),
		versions: make(map[string][]models.ConfigVersion),
	}
}

// Create creates a new configuration
func (r *InMemoryRepository) Create(config *models.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[config.Name]; exists {
		return &models.ConfigExistsError{Name: config.Name}
	}

	// Set initial version and timestamps
	config.Version = 1
	config.CreatedAt = time.Now()
	config.UpdatedAt = config.CreatedAt

	// Store the config
	r.configs[config.Name] = config

	// Store the version
	version := models.ConfigVersion{
		Version:   config.Version,
		Data:      copyData(config.Data),
		CreatedAt: config.CreatedAt,
	}
	r.versions[config.Name] = []models.ConfigVersion{version}

	return nil
}

// Get retrieves the latest version of a configuration
func (r *InMemoryRepository) Get(name string) (*models.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.configs[name]
	if !exists {
		return nil, &models.ConfigNotFoundError{Name: name}
	}

	// Return a copy to prevent external modifications
	configCopy := *config
	configCopy.Data = copyData(config.Data)
	return &configCopy, nil
}

// Update updates an existing configuration
func (r *InMemoryRepository) Update(config *models.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.configs[config.Name]
	if !exists {
		return &models.ConfigNotFoundError{Name: config.Name}
	}

	// Increment version
	config.Version = existing.Version + 1
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()

	// Update the config
	r.configs[config.Name] = config

	// Store the new version
	version := models.ConfigVersion{
		Version:   config.Version,
		Data:      copyData(config.Data),
		CreatedAt: config.UpdatedAt,
	}
	r.versions[config.Name] = append(r.versions[config.Name], version)

	return nil
}

// GetVersion retrieves a specific version of a configuration
func (r *InMemoryRepository) GetVersion(name string, version int) (*models.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.versions[name]
	if !exists {
		return nil, &models.ConfigNotFoundError{Name: name}
	}

	if version < 1 || version > len(versions) {
		return nil, &models.VersionNotFoundError{Name: name, Version: version}
	}

	// Versions are 1-indexed, array is 0-indexed
	versionCopy := versions[version-1]
	versionCopy.Data = copyData(versionCopy.Data)
	return &versionCopy, nil
}

// ListVersions lists all versions of a configuration
func (r *InMemoryRepository) ListVersions(name string) ([]models.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.versions[name]
	if !exists {
		return nil, &models.ConfigNotFoundError{Name: name}
	}

	// Return a copy of the versions
	versionsCopy := make([]models.ConfigVersion, len(versions))
	for i, v := range versions {
		versionsCopy[i] = models.ConfigVersion{
			Version:   v.Version,
			Data:      copyData(v.Data),
			CreatedAt: v.CreatedAt,
		}
	}

	return versionsCopy, nil
}

// Exists checks if a configuration exists
func (r *InMemoryRepository) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.configs[name]
	return exists
}

// copyData creates a deep copy of the data map
func copyData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	copy := make(map[string]interface{}, len(data))
	for k, v := range data {
		// For nested maps, recursively copy
		if nested, ok := v.(map[string]interface{}); ok {
			copy[k] = copyData(nested)
		} else {
			copy[k] = v
		}
	}
	return copy
}

// Clear removes all configurations (useful for testing)
func (r *InMemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.configs = make(map[string]*models.Config)
	r.versions = make(map[string][]models.ConfigVersion)
}

// Stats returns statistics about the repository (useful for monitoring)
func (r *InMemoryRepository) Stats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalVersions := 0
	for _, versions := range r.versions {
		totalVersions += len(versions)
	}

	return map[string]interface{}{
		"total_configs":  len(r.configs),
		"total_versions": totalVersions,
	}
}

// Validate that InMemoryRepository implements ConfigRepository
var _ ConfigRepository = (*InMemoryRepository)(nil)
