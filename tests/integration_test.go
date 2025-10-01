package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"config-engine/internal/handlers"
	"config-engine/internal/models"
	"config-engine/internal/repository"
	"config-engine/internal/service"
	"config-engine/internal/validation"
)

func setupTestServer(t *testing.T) (*httptest.Server, *repository.InMemoryRepository) {
	validator, err := validation.NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	repo := repository.NewInMemoryRepository()
	svc := service.NewConfigService(repo, validator)
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	handler := handlers.NewConfigHandler(svc, logger)
	router := handlers.SetupRouter(handler, logger)

	// Gin's Engine implements http.Handler, so it works with httptest
	server := httptest.NewServer(router)
	return server, repo
}

func TestCreateConfigEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	reqBody := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		server.URL+"/api/v1/configs",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var config models.Config
	json.NewDecoder(resp.Body).Decode(&config)

	if config.Name != "payment_config" {
		t.Errorf("Expected name 'payment_config', got '%s'", config.Name)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}
}

func TestCreateConfigValidationError(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	reqBody := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": "invalid", // Should be integer
			"enabled":   true,
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		server.URL+"/api/v1/configs",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateConfigDuplicate(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	reqBody := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}

	body, _ := json.Marshal(reqBody)

	// Create first time
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Try to create again
	body, _ = json.Marshal(reqBody)
	resp, _ := http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.StatusCode)
	}
}

func TestGetConfigEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Create config first
	createReq := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Get config
	resp, err := http.Get(server.URL + "/api/v1/configs/payment_config")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var config models.Config
	json.NewDecoder(resp.Body).Decode(&config)

	if config.Name != "payment_config" {
		t.Errorf("Expected name 'payment_config', got '%s'", config.Name)
	}
}

func TestGetConfigNotFound(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, _ := http.Get(server.URL + "/api/v1/configs/nonexistent")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestGetConfigWithVersion(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Create config
	createReq := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Update config
	updateReq := models.UpdateConfigRequest{
		Data: map[string]interface{}{
			"max_limit": 2000,
			"enabled":   false,
		},
	}
	body, _ = json.Marshal(updateReq)
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/configs/payment_config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)

	// Get version 1
	resp, _ := http.Get(server.URL + "/api/v1/configs/payment_config?version=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var config models.Config
	json.NewDecoder(resp.Body).Decode(&config)

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Data["max_limit"].(float64) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", config.Data["max_limit"])
	}
}

func TestUpdateConfigEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Create config first
	createReq := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Update config
	updateReq := models.UpdateConfigRequest{
		Data: map[string]interface{}{
			"max_limit": 2000,
			"enabled":   false,
		},
	}
	body, _ = json.Marshal(updateReq)

	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/configs/payment_config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var config models.Config
	json.NewDecoder(resp.Body).Decode(&config)

	if config.Version != 2 {
		t.Errorf("Expected version 2, got %d", config.Version)
	}

	if config.Data["max_limit"].(float64) != 2000 {
		t.Errorf("Expected max_limit 2000, got %v", config.Data["max_limit"])
	}
}

func TestListVersionsEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Create config
	createReq := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Update config
	updateReq := models.UpdateConfigRequest{
		Data: map[string]interface{}{
			"max_limit": 2000,
			"enabled":   false,
		},
	}
	body, _ = json.Marshal(updateReq)
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/configs/payment_config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)

	// List versions
	resp, err := http.Get(server.URL + "/api/v1/configs/payment_config/versions")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var versionsResp models.VersionsResponse
	json.NewDecoder(resp.Body).Decode(&versionsResp)

	if len(versionsResp.Versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versionsResp.Versions))
	}
}

func TestRollbackConfigEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Create config
	createReq := models.CreateConfigRequest{
		Name: "payment_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))

	// Update config
	updateReq := models.UpdateConfigRequest{
		Data: map[string]interface{}{
			"max_limit": 2000,
			"enabled":   false,
		},
	}
	body, _ = json.Marshal(updateReq)
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/configs/payment_config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)

	// Rollback to version 1
	rollbackReq := models.RollbackRequest{Version: 1}
	body, _ = json.Marshal(rollbackReq)
	resp, err := http.Post(
		server.URL+"/api/v1/configs/payment_config/rollback",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var config models.Config
	json.NewDecoder(resp.Body).Decode(&config)

	if config.Version != 3 {
		t.Errorf("Expected version 3, got %d", config.Version)
	}

	if config.Data["max_limit"].(float64) != 1000 {
		t.Errorf("Expected max_limit 1000, got %v", config.Data["max_limit"])
	}
}

func TestHealthCheckEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("running")) {
		t.Error("Expected response to contain 'running'")
	}
}

func TestFullWorkflow(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	client := &http.Client{}

	// 1. Create config
	createReq := models.CreateConfigRequest{
		Name: "workflow_config",
		Type: "payment_config",
		Data: map[string]interface{}{
			"max_limit": 1000,
			"enabled":   true,
		},
	}
	body, _ := json.Marshal(createReq)
	resp, _ := http.Post(server.URL+"/api/v1/configs", "application/json", bytes.NewBuffer(body))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create config: status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Update config multiple times
	for i := 2; i <= 5; i++ {
		updateReq := models.UpdateConfigRequest{
			Data: map[string]interface{}{
				"max_limit": i * 1000,
				"enabled":   i%2 == 0,
			},
		}
		body, _ = json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", server.URL+"/api/v1/configs/workflow_config", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ = client.Do(req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to update config: status %d", resp.StatusCode)
		}
		resp.Body.Close()
	}

	// 3. List versions
	resp, _ = http.Get(server.URL + "/api/v1/configs/workflow_config/versions")
	var versionsResp models.VersionsResponse
	json.NewDecoder(resp.Body).Decode(&versionsResp)
	resp.Body.Close()

	if len(versionsResp.Versions) != 5 {
		t.Errorf("Expected 5 versions, got %d", len(versionsResp.Versions))
	}

	// 4. Get specific version
	resp, _ = http.Get(server.URL + "/api/v1/configs/workflow_config?version=2")
	var v2Config models.Config
	json.NewDecoder(resp.Body).Decode(&v2Config)
	resp.Body.Close()

	if v2Config.Version != 2 {
		t.Errorf("Expected version 2, got %d", v2Config.Version)
	}

	// 5. Rollback to version 1
	rollbackReq := models.RollbackRequest{Version: 1}
	body, _ = json.Marshal(rollbackReq)
	resp, _ = http.Post(
		server.URL+"/api/v1/configs/workflow_config/rollback",
		"application/json",
		bytes.NewBuffer(body),
	)
	var rolledBackConfig models.Config
	json.NewDecoder(resp.Body).Decode(&rolledBackConfig)
	resp.Body.Close()

	if rolledBackConfig.Version != 6 {
		t.Errorf("Expected version 6 after rollback, got %d", rolledBackConfig.Version)
	}

	if rolledBackConfig.Data["max_limit"].(float64) != 1000 {
		t.Errorf("Expected rolled back max_limit 1000, got %v", rolledBackConfig.Data["max_limit"])
	}

	// 6. Get latest version
	resp, _ = http.Get(server.URL + "/api/v1/configs/workflow_config")
	var latestConfig models.Config
	json.NewDecoder(resp.Body).Decode(&latestConfig)
	resp.Body.Close()

	if latestConfig.Version != 6 {
		t.Errorf("Expected latest version 6, got %d", latestConfig.Version)
	}

	fmt.Println("Full workflow test completed successfully")
}
