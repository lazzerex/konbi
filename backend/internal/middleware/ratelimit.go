package middleware

import (
	"konbi/internal/errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rate limiter middleware with per-IP buckets
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	logger   *logrus.Logger
}

// create new rate limiter
func NewRateLimiter(perSecond, burst int, logger *logrus.Logger) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Every(time.Second / time.Duration(perSecond)),
		burst:    burst,
		logger:   logger,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if il, ok := rl.limiters[ip]; ok {
		il.lastSeen = time.Now()
		return il.limiter
	}

	l := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

// cleanup evicts IP entries not seen in the last 10 minutes
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, il := range rl.limiters {
			if time.Since(il.lastSeen) > 10*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// middleware handler
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.getLimiter(ip).Allow() {
			rl.logger.WithField("ip", ip).Warn("rate limit exceeded")
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
