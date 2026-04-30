package middleware

import (
	"konbi/internal/errors"
	"konbi/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// jwt auth middleware
type JWTAuth struct {
	authService *services.AuthService
	logger      *logrus.Logger
}

// create new jwt auth middleware
func NewJWTAuth(authService *services.AuthService, logger *logrus.Logger) *JWTAuth {
	return &JWTAuth{
		authService: authService,
		logger:      logger,
	}
}

// middleware validates jwt token and attaches user to context
func (j *JWTAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			j.logger.WithField("ip", c.ClientIP()).Warn("missing authorization header")
			err := errors.NewUnauthorizedError("missing authorization header")
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			c.Abort()
			return
		}

		// extract bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			j.logger.WithField("ip", c.ClientIP()).Warn("invalid authorization header format")
			err := errors.NewUnauthorizedError("invalid authorization header format")
			c.JSON(err.StatusCode, gin.H{"error": err.Message})
			c.Abort()
			return
		}

		token := parts[1]

		// verify token
		claims, err := j.authService.VerifyAccessToken(token)
		if err != nil {
			j.logger.WithField("ip", c.ClientIP()).Warn("invalid token")
			appErr := err.(*errors.AppError)
			c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
			c.Abort()
			return
		}

		// attach user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
