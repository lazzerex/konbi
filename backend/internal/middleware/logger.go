package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// logger middleware logs all http requests
type LoggerMiddleware struct {
	logger *logrus.Logger
}

// create new logger middleware
func NewLoggerMiddleware(logger *logrus.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// middleware handler
func (lm *LoggerMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// generate request id
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// start timer
		start := time.Now()

		// process request
		c.Next()

		// calculate latency
		latency := time.Since(start)

		// log request details
		lm.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": latency.Milliseconds(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("request completed")
	}
}
