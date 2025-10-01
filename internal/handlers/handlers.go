package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"config-engine/internal/models"
	"config-engine/internal/service"

	"github.com/gin-gonic/gin"
)

// ConfigHandler handles HTTP requests for configuration management
type ConfigHandler struct {
	service *service.ConfigService
	logger  *log.Logger
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(service *service.ConfigService, logger *log.Logger) *ConfigHandler {
	return &ConfigHandler{
		service: service,
		logger:  logger,
	}
}

// CreateConfig handles POST /api/v1/configs
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req models.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Printf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	config, err := h.service.CreateConfig(&req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetConfig handles GET /api/v1/configs/{name}
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	name := c.Param("name")

	// Check for version query parameter
	var version *int
	if versionStr := c.Query("version"); versionStr != "" {
		v, err := strconv.Atoi(versionStr)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid version parameter",
				Details: "version must be a positive integer",
			})
			return
		}
		version = &v
	}

	config, err := h.service.GetConfig(name, version)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig handles PUT /api/v1/configs/{name}
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	name := c.Param("name")

	var req models.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Printf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	config, err := h.service.UpdateConfig(name, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, config)
}

// RollbackConfig handles POST /api/v1/configs/{name}/rollback
func (h *ConfigHandler) RollbackConfig(c *gin.Context) {
	name := c.Param("name")

	var req models.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Printf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	config, err := h.service.RollbackConfig(name, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListVersions handles GET /api/v1/configs/{name}/versions
func (h *ConfigHandler) ListVersions(c *gin.Context) {
	name := c.Param("name")

	versions, err := h.service.ListVersions(name)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, versions)
}

// HealthCheck handles GET /health
func (h *ConfigHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "running"})
}

// handleServiceError maps service errors to appropriate HTTP responses
func (h *ConfigHandler) handleServiceError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *models.ValidationError:
		h.logger.Printf("Validation error: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   err.Error(),
			Details: "",
		})
	case *models.ConfigNotFoundError:
		h.logger.Printf("Config not found: %v", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   err.Error(),
			Details: "",
		})
	case *models.ConfigExistsError:
		h.logger.Printf("Config already exists: %v", err)
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   err.Error(),
			Details: "",
		})
	case *models.VersionNotFoundError:
		h.logger.Printf("Version not found: %v", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   err.Error(),
			Details: "",
		})
	case *models.SchemaValidationError:
		h.logger.Printf("Schema validation error: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Schema validation failed",
			Details: e.Details,
		})
	default:
		// TODO: Ideally not exposing internal error details to the client side
		h.logger.Printf("Internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Printf("%s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "Internal server error",
					Details: fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// SetupRouter configures and returns the HTTP router
func SetupRouter(handler *ConfigHandler, logger *log.Logger) *gin.Engine {
	r := gin.New()

	// Apply middleware
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))

	// Health check
	r.GET("/health", handler.HealthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		api.POST("/configs", handler.CreateConfig)
		api.GET("/configs/:name", handler.GetConfig)
		api.PUT("/configs/:name", handler.UpdateConfig)
		api.GET("/configs/:name/versions", handler.ListVersions)
		api.POST("/configs/:name/rollback", handler.RollbackConfig)
	}

	return r
}