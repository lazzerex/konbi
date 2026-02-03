package handlers

import (
	"konbi/internal/config"
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// url handler handles url shortening endpoints
type URLHandler struct {
	service *services.URLService
	cfg     *config.Config
	logger  *logrus.Logger
}

// create new url handler
func NewURLHandler(service *services.URLService, cfg *config.Config, logger *logrus.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		cfg:     cfg,
		logger:  logger,
	}
}

// shorten creates a shortened url
// POST /api/shorten
func (h *URLHandler) Shorten(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid request body", err))
		return
	}

	response, err := h.service.ShortenURL(c.Request.Context(), &req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// redirect redirects to original url
// GET /s/:code
func (h *URLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		h.respondWithError(c, errors.NewBadRequestError("short code required", nil))
		return
	}

	// get client info for analytics
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")

	originalURL, err := h.service.Redirect(c.Request.Context(), shortCode, ipAddress, userAgent, referrer)
	if err != nil {
		// for not found, show a nice error page instead of json
		if appErr, ok := err.(*errors.AppError); ok && appErr.StatusCode == http.StatusNotFound {
			c.HTML(http.StatusNotFound, "", gin.H{
				"message": "Short URL not found or expired",
			})
			return
		}
		h.respondWithError(c, err)
		return
	}

	// redirect to original url
	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// get stats retrieves url analytics
// GET /api/shorten/:code/stats
func (h *URLHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		h.respondWithError(c, errors.NewBadRequestError("short code required", nil))
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), shortCode)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// delete url soft deletes a shortened url
// DELETE /api/shorten/:code
func (h *URLHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		h.respondWithError(c, errors.NewBadRequestError("short code required", nil))
		return
	}

	err := h.service.DeleteURL(c.Request.Context(), shortCode)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "short url deleted successfully"})
}

// respond with error helper
func (h *URLHandler) respondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		h.logger.WithFields(logrus.Fields{
			"code":    appErr.Code,
			"message": appErr.Message,
			"error":   appErr.Err,
		}).Error("request error")

		c.JSON(appErr.StatusCode, gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
		return
	}

	// fallback for unknown errors
	h.logger.WithError(err).Error("unknown error")
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
		"code":  "INTERNAL_ERROR",
	})
}

