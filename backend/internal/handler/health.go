package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Health exposes a simple readiness probe including database ping.
type Health struct {
	DB *gorm.DB
}

func (h *Health) Register(r gin.IRoutes) {
	r.GET("/health", h.Get)
}

func (h *Health) Get(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "database": "unreachable"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "database": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Healthy"})
}
