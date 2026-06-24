package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// health check handler pings the database to verify connectivity
func HealthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error":  "database unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
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
