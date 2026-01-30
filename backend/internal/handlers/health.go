package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// health check handler
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// root handler
func Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "konbi api",
		"version": "1.0",
		"endpoints": gin.H{
			"health":  "/health",
			"upload":  "POST /api/upload",
			"note":    "POST /api/note",
			"content": "GET /api/content/:id",
		},
	})
}
