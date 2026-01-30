package middleware

import (
	"konbi/internal/config"
	"konbi/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// admin auth middleware
type AdminAuth struct {
	config *config.Config
	logger *logrus.Logger
}

// create new admin auth middleware
func NewAdminAuth(cfg *config.Config, logger *logrus.Logger) *AdminAuth {
	return &AdminAuth{
		config: cfg,
		logger: logger,
	}
}

// middleware handler
func (a *AdminAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if admin endpoint is enabled
		if a.config.Security.AdminSecret == "" {
			err := errors.NewForbiddenError("admin endpoint disabled")
			c.JSON(err.StatusCode, gin.H{
				"error": err.Message,
				"code":  err.Code,
			})
			c.Abort()
			return
		}

		// validate admin secret
		providedSecret := c.GetHeader("X-Admin-Secret")
		if providedSecret != a.config.Security.AdminSecret {
			a.logger.WithField("ip", c.ClientIP()).Warn("unauthorized admin access attempt")
			err := errors.NewUnauthorizedError("unauthorized")
			c.JSON(err.StatusCode, gin.H{
				"error": err.Message,
				"code":  err.Code,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
