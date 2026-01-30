package handlers

import (
	"fmt"
	"io"
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/services"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// content handler handles http requests for content operations
type ContentHandler struct {
	service *services.ContentService
	logger  *logrus.Logger
}

// create new content handler
func NewContentHandler(service *services.ContentService, logger *logrus.Logger) *ContentHandler {
	return &ContentHandler{
		service: service,
		logger:  logger,
	}
}

// upload handles file upload requests
func (h *ContentHandler) Upload(c *gin.Context) {
	ctx := c.Request.Context()

	// parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		appErr := errors.NewBadRequestError("no file provided", err)
		h.respondWithError(c, appErr)
		return
	}
	defer file.Close()

	// read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		appErr := errors.NewInternalError("failed to read file", err)
		h.respondWithError(c, appErr)
		return
	}

	// prepare request
	req := &models.UploadRequest{
		File:     fileBytes,
		Filename: header.Filename,
		Size:     header.Size,
	}

	// upload file
	content, err := h.service.UploadFile(ctx, req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// respond
	c.JSON(http.StatusOK, gin.H{
		"id":        content.ID,
		"filename":  *content.Filename,
		"size":      *content.Filesize,
		"expiresAt": content.ExpiresAt.Format(time.RFC3339),
	})
}

// note handles note creation requests
func (h *ContentHandler) Note(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewBadRequestError("invalid request", err)
		h.respondWithError(c, appErr)
		return
	}

	// create note
	content, err := h.service.CreateNote(ctx, &req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// respond
	response := gin.H{
		"id":        content.ID,
		"expiresAt": content.ExpiresAt.Format(time.RFC3339),
	}
	if content.Title != nil {
		response["title"] = *content.Title
	}

	c.JSON(http.StatusOK, response)
}

// get content retrieves content by id
func (h *ContentHandler) GetContent(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		appErr := errors.NewBadRequestError("id required", nil)
		h.respondWithError(c, appErr)
		return
	}

	// get content
	content, err := h.service.GetContent(ctx, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// prepare response based on content type
	if content.Type == models.ContentTypeNote {
		response := gin.H{
			"type": "note",
		}
		if content.Title != nil {
			response["title"] = *content.Title
		}
		if content.Content != nil {
			response["content"] = *content.Content
		}
		c.JSON(http.StatusOK, response)
	} else if content.Type == models.ContentTypeFile {
		// check if file exists
		if content.Filepath == nil {
			appErr := errors.NewNotFoundError("file not found")
			h.respondWithError(c, appErr)
			return
		}

		if _, err := os.Stat(*content.Filepath); os.IsNotExist(err) {
			appErr := errors.NewNotFoundError("file not found")
			h.respondWithError(c, appErr)
			return
		}

		response := gin.H{
			"type":        "file",
			"downloadUrl": fmt.Sprintf("/api/content/%s/download", id),
		}
		if content.Filename != nil {
			response["filename"] = *content.Filename
		}
		if content.Filesize != nil {
			response["size"] = *content.Filesize
		}
		c.JSON(http.StatusOK, response)
	}
}

// download handles file download requests
func (h *ContentHandler) Download(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		appErr := errors.NewBadRequestError("id required", nil)
		h.respondWithError(c, appErr)
		return
	}

	// get content
	content, err := h.service.GetContent(ctx, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// validate content type
	if content.Type != models.ContentTypeFile {
		appErr := errors.NewBadRequestError("content is not a file", nil)
		h.respondWithError(c, appErr)
		return
	}

	// validate file path
	if content.Filepath == nil {
		appErr := errors.NewNotFoundError("file not found")
		h.respondWithError(c, appErr)
		return
	}

	// check if file exists
	if _, err := os.Stat(*content.Filepath); os.IsNotExist(err) {
		appErr := errors.NewNotFoundError("file not found")
		h.respondWithError(c, appErr)
		return
	}

	// serve file
	filename := "download"
	if content.Filename != nil {
		filename = *content.Filename
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(*content.Filepath)
}

// get stats retrieves content statistics
func (h *ContentHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		appErr := errors.NewBadRequestError("id required", nil)
		h.respondWithError(c, appErr)
		return
	}

	// get stats
	content, err := h.service.GetStats(ctx, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"viewCount": content.ViewCount,
		"createdAt": content.CreatedAt.Format(time.RFC3339),
		"expiresAt": content.ExpiresAt.Format(time.RFC3339),
	})
}

// list admin retrieves all content for admin
func (h *ContentHandler) ListAdmin(c *gin.Context) {
	ctx := c.Request.Context()

	// list all content
	contents, err := h.service.ListAll(ctx)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// prepare response
	var response []gin.H
	for _, content := range contents {
		item := gin.H{
			"id":         content.ID,
			"type":       content.Type,
			"created_at": content.CreatedAt.Format(time.RFC3339),
			"expires_at": content.ExpiresAt.Format(time.RFC3339),
			"view_count": content.ViewCount,
		}

		if content.Title != nil {
			item["title"] = *content.Title
		}
		if content.Filename != nil {
			item["filename"] = *content.Filename
		}
		if content.Filesize != nil {
			item["filesize"] = *content.Filesize
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(response),
		"contents": response,
	})
}

// respond with error handles error responses
func (h *ContentHandler) respondWithError(c *gin.Context, err error) {
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
