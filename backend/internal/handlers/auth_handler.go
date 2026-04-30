package handlers

import (
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// auth handler handles authentication endpoints
type AuthHandler struct {
	service *services.AuthService
	logger  *logrus.Logger
}

// create new auth handler
func NewAuthHandler(service *services.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid request body", nil))
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid request body", nil))
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// refresh generates new access token from refresh token
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid request body", nil))
		return
	}

	accessToken, err := h.service.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// me returns current user info
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, errors.NewUnauthorizedError("user not found in context"))
		return
	}

	user := &models.User{
		ID: userID.(string),
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// logout invalidates token (frontend should discard token)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// private helper to respond with error
func (h *AuthHandler) respondWithError(c *gin.Context, err error) {
	var status int
	var message string

	if appErr, ok := err.(*errors.AppError); ok {
		status = appErr.StatusCode
		message = appErr.Message
	} else {
		status = http.StatusInternalServerError
		message = "internal server error"
	}

	h.logger.WithFields(logrus.Fields{
		"status":  status,
		"message": message,
	}).Error("auth error")

	c.JSON(status, gin.H{
		"error": message,
	})
}
