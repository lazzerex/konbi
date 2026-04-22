package handlers

import (
	"archive/zip"
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
		Passcode: c.PostForm("passcode"),
	}

	// upload file
	content, err := h.service.UploadFile(ctx, req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

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

	response := gin.H{
		"id":        content.ID,
		"expiresAt": content.ExpiresAt.Format(time.RFC3339),
	}
	if content.Title != nil {
		response["title"] = *content.Title
	}
	c.JSON(http.StatusOK, response)
}

// bundle handles multi-file bundle upload
func (h *ContentHandler) Bundle(c *gin.Context) {
	ctx := c.Request.Context()

	form, err := c.MultipartForm()
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid multipart form", err))
		return
	}

	fileHeaders := form.File["files"]
	if len(fileHeaders) == 0 {
		h.respondWithError(c, errors.NewBadRequestError("no files provided", nil))
		return
	}

	var requests []*models.UploadRequest
	for _, fh := range fileHeaders {
		f, err := fh.Open()
		if err != nil {
			h.respondWithError(c, errors.NewInternalError("failed to read file", err))
			return
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			h.respondWithError(c, errors.NewInternalError("failed to read file", err))
			return
		}
		requests = append(requests, &models.UploadRequest{
			File:     data,
			Filename: fh.Filename,
			Size:     fh.Size,
		})
	}

	bundle, err := h.service.CreateBundle(ctx, requests)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        bundle.ID,
		"fileCount": len(requests),
		"expiresAt": bundle.ExpiresAt.Format(time.RFC3339),
	})
}

// bundle zip streams all files in a bundle as a zip archive
func (h *ContentHandler) BundleZip(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		h.respondWithError(c, errors.NewBadRequestError("id required", nil))
		return
	}

	bundle, err := h.service.GetContent(ctx, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	if bundle.Type != models.ContentTypeBundle {
		h.respondWithError(c, errors.NewBadRequestError("content is not a bundle", nil))
		return
	}

	if bundle.PasscodeHash != nil {
		passcode := c.GetHeader("X-Passcode")
		if passcode == "" {
			h.respondWithError(c, errors.NewUnauthorizedError("passcode required"))
			return
		}
		if err := h.service.VerifyPasscode(bundle, passcode); err != nil {
			h.respondWithError(c, err)
			return
		}
	}

	files, err := h.service.GetBundleFiles(ctx, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	if len(files) == 0 {
		h.respondWithError(c, errors.NewNotFoundError("no files found in bundle"))
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="bundle-%s.zip"`, id))
	c.Status(http.StatusOK)

	zw := zip.NewWriter(c.Writer)
	for _, f := range files {
		if f.Filepath == nil || f.Filename == nil {
			continue
		}
		w, err := zw.Create(*f.Filename)
		if err != nil {
			h.logger.WithError(err).Error("failed to create zip entry")
			continue
		}
		src, err := os.Open(*f.Filepath)
		if err != nil {
			h.logger.WithError(err).WithField("filepath", *f.Filepath).Error("failed to open file for zip")
			continue
		}
		io.Copy(w, src)
		src.Close()
	}
	zw.Close()
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

	// if passcode-protected, return metadata only — no content or download URL
	if content.PasscodeHash != nil {
		response := gin.H{
			"type":         content.Type,
			"id":           content.ID,
			"has_passcode": true,
			"expiresAt":    content.ExpiresAt.Format(time.RFC3339),
		}
		if content.Type == models.ContentTypeFile {
			if content.Filename != nil {
				response["filename"] = *content.Filename
			}
			if content.Filesize != nil {
				response["size"] = *content.Filesize
			}
		} else if content.Type == models.ContentTypeNote {
			if content.Title != nil {
				response["title"] = *content.Title
			}
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// prepare response based on content type
	if content.Type == models.ContentTypeNote {
		response := gin.H{
			"type": "note",
			"id":   content.ID,
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
			"id":          content.ID,
			"downloadUrl": fmt.Sprintf("/api/content/%s/download", id),
		}
		if content.Filename != nil {
			response["filename"] = *content.Filename
		}
		if content.Filesize != nil {
			response["size"] = *content.Filesize
		}
		c.JSON(http.StatusOK, response)
	} else if content.Type == models.ContentTypeBundle {
		files, err := h.service.GetBundleFiles(ctx, id)
		if err != nil {
			h.respondWithError(c, err)
			return
		}
		var fileList []gin.H
		for _, f := range files {
			item := gin.H{"id": f.ID}
			if f.Filename != nil {
				item["filename"] = *f.Filename
			}
			if f.Filesize != nil {
				item["size"] = *f.Filesize
			}
			fileList = append(fileList, item)
		}
		c.JSON(http.StatusOK, gin.H{
			"type":        "bundle",
			"id":          content.ID,
			"fileCount":   len(files),
			"files":       fileList,
			"downloadUrl": fmt.Sprintf("/api/content/%s/zip", id),
		})
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

	// verify passcode if required
	if content.PasscodeHash != nil {
		passcode := c.GetHeader("X-Passcode")
		if passcode == "" {
			h.respondWithError(c, errors.NewUnauthorizedError("passcode required"))
			return
		}
		if err := h.service.VerifyPasscode(content, passcode); err != nil {
			h.respondWithError(c, err)
			return
		}
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

// unlock verifies a passcode and returns full content
func (h *ContentHandler) Unlock(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		h.respondWithError(c, errors.NewBadRequestError("id required", nil))
		return
	}

	var req models.UnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.NewBadRequestError("passcode required", err))
		return
	}

	content, err := h.service.UnlockContent(ctx, id, req.Passcode)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	if content.Type == models.ContentTypeNote {
		response := gin.H{
			"type": "note",
			"id":   content.ID,
		}
		if content.Title != nil {
			response["title"] = *content.Title
		}
		if content.Content != nil {
			response["content"] = *content.Content
		}
		c.JSON(http.StatusOK, response)
	} else if content.Type == models.ContentTypeFile {
		if content.Filepath == nil {
			h.respondWithError(c, errors.NewNotFoundError("file not found"))
			return
		}
		if _, err := os.Stat(*content.Filepath); os.IsNotExist(err) {
			h.respondWithError(c, errors.NewNotFoundError("file not found"))
			return
		}
		response := gin.H{
			"type":        "file",
			"id":          content.ID,
			"downloadUrl": fmt.Sprintf("/api/content/%s/download", content.ID),
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
			"id":           content.ID,
			"type":         content.Type,
			"has_passcode": content.PasscodeHash != nil,
			"created_at":   content.CreatedAt.Format(time.RFC3339),
			"expires_at":   content.ExpiresAt.Format(time.RFC3339),
			"view_count":   content.ViewCount,
		}
		if content.Code != nil {
			item["code"] = *content.Code
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
