package middleware

import (
	"konbi/internal/errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// rate limiter middleware
type RateLimiter struct {
	limiter *rate.Limiter
	logger  *logrus.Logger
}

// create new rate limiter
func NewRateLimiter(perSecond, burst int, logger *logrus.Logger) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(perSecond)), burst),
		logger:  logger,
	}
}

// middleware handler
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiter.Allow() {
			rl.logger.WithField("ip", c.ClientIP()).Warn("rate limit exceeded")
			err := errors.NewRateLimitError()
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
