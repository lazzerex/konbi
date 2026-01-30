package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"konbi/internal/config"
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/repository"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// content service handles business logic for content operations
type ContentService struct {
	repo   *repository.ContentRepository
	config *config.Config
	logger *logrus.Logger
}

// allowed file extensions
var allowedExtensions = map[string]bool{
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".zip": true, ".tar": true, ".gz": true,
	".mp4": true, ".mp3": true, ".wav": true,
	".csv": true, ".xlsx": true, ".xls": true,
	".json": true, ".xml": true, ".md": true,
}

// create new content service
func NewContentService(repo *repository.ContentRepository, cfg *config.Config, logger *logrus.Logger) *ContentService {
	return &ContentService{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

// upload file handles file upload logic
func (s *ContentService) UploadFile(ctx context.Context, req *models.UploadRequest) (*models.Content, error) {
	// validate file size
	if req.Size > s.config.Storage.MaxFileSize {
		s.logger.WithFields(logrus.Fields{
			"file_size": req.Size,
			"max_size":  s.config.Storage.MaxFileSize,
		}).Warn("file size exceeds limit")
		return nil, errors.NewFileTooLargeError(s.config.Storage.MaxFileSize)
	}

	// validate file extension
	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext != "" && !allowedExtensions[ext] {
		s.logger.WithField("extension", ext).Warn("file type not allowed")
		return nil, errors.NewFileTypeNotAllowedError()
	}

	// generate unique id
	id, err := s.generateUniqueID(ctx)
	if err != nil {
		return nil, err
	}

	// prepare file path
	filename := id + ext
	filePath := filepath.Join(s.config.Storage.UploadDir, filename)

	// save file to disk
	file, err := os.Create(filePath)
	if err != nil {
		s.logger.WithError(err).WithField("filepath", filePath).Error("failed to create file")
		return nil, errors.NewInternalError("failed to save file", err)
	}
	defer file.Close()

	if _, err := file.Write(req.File); err != nil {
		os.Remove(filePath)
		s.logger.WithError(err).WithField("filepath", filePath).Error("failed to write file")
		return nil, errors.NewInternalError("failed to save file", err)
	}

	// prepare content model
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.Storage.ExpirationDays) * 24 * time.Hour)
	content := &models.Content{
		ID:        id,
		Type:      models.ContentTypeFile,
		Filename:  &req.Filename,
		Filepath:  &filePath,
		Filesize:  &req.Size,
		ExpiresAt: expiresAt,
	}

	// save to database
	if err := s.repo.Create(ctx, content); err != nil {
		os.Remove(filePath)
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"content_id": id,
		"filename":   req.Filename,
		"size":       req.Size,
	}).Info("file uploaded successfully")

	return content, nil
}

// create note handles note creation logic
func (s *ContentService) CreateNote(ctx context.Context, req *models.NoteRequest) (*models.Content, error) {
	// validate content length (1mb limit)
	if len(req.Content) > 1024*1024 {
		s.logger.Warn("note content too large")
		return nil, errors.NewContentTooLargeError()
	}

	// generate unique id
	id, err := s.generateUniqueID(ctx)
	if err != nil {
		return nil, err
	}

	// prepare content model
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.Storage.ExpirationDays) * 24 * time.Hour)
	var title *string
	if req.Title != "" {
		title = &req.Title
	}

	content := &models.Content{
		ID:        id,
		Type:      models.ContentTypeNote,
		Title:     title,
		Content:   &req.Content,
		ExpiresAt: expiresAt,
	}

	// save to database
	if err := s.repo.Create(ctx, content); err != nil {
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"content_id": id,
		"title":      req.Title,
	}).Info("note created successfully")

	return content, nil
}

// get content retrieves content by id and increments view count
func (s *ContentService) GetContent(ctx context.Context, id string) (*models.Content, error) {
	content, err := s.repo.FindActiveByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// increment view count asynchronously
	go func() {
		if err := s.repo.IncrementViewCount(context.Background(), id); err != nil {
			s.logger.WithError(err).WithField("content_id", id).Error("failed to increment view count")
		}
	}()

	return content, nil
}

// get stats retrieves content statistics
func (s *ContentService) GetStats(ctx context.Context, id string) (*models.Content, error) {
	return s.repo.FindByID(ctx, id)
}

// list all content for admin
func (s *ContentService) ListAll(ctx context.Context) ([]*models.Content, error) {
	return s.repo.ListAll(ctx)
}

// cleanup expired content removes expired files and database records
func (s *ContentService) CleanupExpired(ctx context.Context) (int, error) {
	s.logger.Info("starting cleanup of expired content")

	// find expired file content
	expiredContent, err := s.repo.FindExpiredContent(ctx)
	if err != nil {
		return 0, err
	}

	// delete files from disk
	deletedFiles := 0
	for _, content := range expiredContent {
		if content.Filepath == nil {
			continue
		}

		if err := os.Remove(*content.Filepath); err != nil && !os.IsNotExist(err) {
			s.logger.WithError(err).WithField("filepath", *content.Filepath).Error("failed to delete file")
		} else {
			deletedFiles++
		}
	}

	// delete expired records from database
	deletedRecords, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		return deletedFiles, err
	}

	s.logger.WithFields(logrus.Fields{
		"deleted_files":   deletedFiles,
		"deleted_records": deletedRecords,
	}).Info("cleanup completed")

	return int(deletedRecords), nil
}

// generate unique id creates a unique content id
func (s *ContentService) generateUniqueID(ctx context.Context) (string, error) {
	const maxRetries = 5
	const idLength = 8

	for i := 0; i < maxRetries; i++ {
		id, err := generateRandomID(idLength)
		if err != nil {
			s.logger.WithError(err).Error("failed to generate random id")
			return "", errors.NewInternalError("failed to generate id", err)
		}

		// check if id exists
		exists, err := s.repo.IDExists(ctx, id)
		if err != nil {
			return "", err
		}

		if !exists {
			return id, nil
		}

		s.logger.WithField("id", id).Debug("id collision detected, retrying")
	}

	return "", errors.NewInternalError("failed to generate unique id after retries", nil)
}

// helper to generate random id
func generateRandomID(length int) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// base64 encode and make url-safe
	id := base64.RawURLEncoding.EncodeToString(bytes)

	// use only alphanumeric
	id = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, id)

	// ensure minimum length
	if len(id) < length {
		return generateRandomID(length)
	}

	return id[:length], nil
}
