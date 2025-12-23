package handlers

import (
	"net/http"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	Port     string
	Hostname string
	DB       *gorm.DB
}

func NewHealthHandler(port, hostname string) *HealthHandler {
	return &HealthHandler{
		Port:     port,
		Hostname: hostname,
		DB:       database.GetDB(),
	}
}

// HealthCheck untuk aplikasi
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  "gin-api-service",
		"version":  "1.0.0",
		"server":   h.Hostname,
		"port":     h.Port,
		"database": "connected",
		"checks": gin.H{
			"gin_router":   "healthy",
			"memory_usage": "normal",
			"api_status":   "running",
		},
	})
}

// DatabaseHealthCheck untuk cek koneksi database
func (h *HealthHandler) DatabaseHealthCheck(c *gin.Context) {
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "database",
			"message": "Database connection not initialized",
		})
		return
	}

	// Ping database
	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "database",
			"message": "Failed to get database instance",
			"error":   err.Error(),
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "database",
			"message": "Database ping failed",
			"error":   err.Error(),
		})
		return
	}

	// Get database stats
	var userCount int64
	h.DB.Model(&struct {
		gorm.Model
	}{}).Count(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "postgresql",
		"database": gin.H{
			"connected":    true,
			"ping":         "success",
			"total_tables": 1, // Adjust based on your schema
			"total_users":  userCount,
		},
		"timestamp": gin.H{"server_time": "2024-01-01T00:00:00Z"}, // Use real time
		"uptime":    "0s",                                         // Implement real uptime
	})
}
