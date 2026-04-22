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
	"golang.org/x/crypto/bcrypt"
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

	// hash passcode if provided
	var passcodeHash *string
	if req.Passcode != "" {
		if err := validatePasscode(req.Passcode); err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Passcode), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.NewInternalError("failed to hash passcode", err)
		}
		h := string(hash)
		passcodeHash = &h
	}

	// prepare content model
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.Storage.ExpirationDays) * 24 * time.Hour)
	content := &models.Content{
		ID:           id,
		Type:         models.ContentTypeFile,
		Filename:     &req.Filename,
		Filepath:     &filePath,
		Filesize:     &req.Size,
		PasscodeHash: passcodeHash,
		ExpiresAt:    expiresAt,
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

	// hash passcode if provided
	var passcodeHash *string
	if req.Passcode != "" {
		if err := validatePasscode(req.Passcode); err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Passcode), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.NewInternalError("failed to hash passcode", err)
		}
		h := string(hash)
		passcodeHash = &h
	}

	// prepare content model
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.Storage.ExpirationDays) * 24 * time.Hour)
	var title *string
	if req.Title != "" {
		title = &req.Title
	}

	content := &models.Content{
		ID:           id,
		Type:         models.ContentTypeNote,
		Title:        title,
		Content:      &req.Content,
		PasscodeHash: passcodeHash,
		ExpiresAt:    expiresAt,
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

// create bundle uploads multiple files under a single shared ID/code
func (s *ContentService) CreateBundle(ctx context.Context, files []*models.UploadRequest) (*models.Content, error) {
	if len(files) == 0 {
		return nil, errors.NewBadRequestError("no files provided", nil)
	}

	// validate all files up front before writing anything
	for _, req := range files {
		if req.Size > s.config.Storage.MaxFileSize {
			return nil, errors.NewFileTooLargeError(s.config.Storage.MaxFileSize)
		}
		ext := strings.ToLower(filepath.Ext(req.Filename))
		if ext != "" && !allowedExtensions[ext] {
			return nil, errors.NewFileTypeNotAllowedError()
		}
	}

	bundleID, err := s.generateUniqueID(ctx)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(s.config.Storage.ExpirationDays) * 24 * time.Hour)
	bundle := &models.Content{
		ID:        bundleID,
		Type:      models.ContentTypeBundle,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, bundle); err != nil {
		return nil, err
	}

	var uploadedPaths []string
	for _, req := range files {
		id, err := s.generateUniqueID(ctx)
		if err != nil {
			s.rollbackBundle(bundleID, uploadedPaths)
			return nil, err
		}

		ext := strings.ToLower(filepath.Ext(req.Filename))
		diskName := id + ext
		filePath := filepath.Join(s.config.Storage.UploadDir, diskName)

		f, err := os.Create(filePath)
		if err != nil {
			s.rollbackBundle(bundleID, uploadedPaths)
			return nil, errors.NewInternalError("failed to save file", err)
		}
		if _, err := f.Write(req.File); err != nil {
			f.Close()
			os.Remove(filePath)
			s.rollbackBundle(bundleID, uploadedPaths)
			return nil, errors.NewInternalError("failed to save file", err)
		}
		f.Close()
		uploadedPaths = append(uploadedPaths, filePath)

		fileContent := &models.Content{
			ID:        id,
			BundleID:  &bundleID,
			Type:      models.ContentTypeFile,
			Filename:  &req.Filename,
			Filepath:  &filePath,
			Filesize:  &req.Size,
			ExpiresAt: expiresAt,
		}
		if err := s.repo.Create(ctx, fileContent); err != nil {
			s.rollbackBundle(bundleID, uploadedPaths)
			return nil, err
		}
	}

	s.logger.WithFields(logrus.Fields{
		"bundle_id":  bundleID,
		"file_count": len(files),
	}).Info("bundle created successfully")

	return bundle, nil
}

// rollback bundle removes uploaded files and soft-deletes the bundle record on failure
func (s *ContentService) rollbackBundle(bundleID string, paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
	s.repo.SoftDelete(context.Background(), bundleID)
}

// get bundle files retrieves all files belonging to a bundle
func (s *ContentService) GetBundleFiles(ctx context.Context, bundleID string) ([]*models.Content, error) {
	return s.repo.FindBundleFiles(ctx, bundleID)
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

// unlock content verifies passcode and returns full content (increments view count on success)
func (s *ContentService) UnlockContent(ctx context.Context, id, passcode string) (*models.Content, error) {
	content, err := s.repo.FindActiveByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if content.PasscodeHash == nil {
		return content, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*content.PasscodeHash), []byte(passcode)); err != nil {
		s.logger.WithField("content_id", id).Warn("incorrect passcode attempt")
		return nil, errors.NewForbiddenError("incorrect passcode")
	}

	go func() {
		if err := s.repo.IncrementViewCount(context.Background(), id); err != nil {
			s.logger.WithError(err).WithField("content_id", id).Error("failed to increment view count")
		}
	}()

	return content, nil
}

// verify passcode checks a passcode against a content record without fetching from DB
func (s *ContentService) VerifyPasscode(content *models.Content, passcode string) error {
	if content.PasscodeHash == nil {
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*content.PasscodeHash), []byte(passcode)); err != nil {
		return errors.NewForbiddenError("incorrect passcode")
	}
	return nil
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

// validate passcode enforces length bounds (4–64 characters)
func validatePasscode(passcode string) error {
	if len(passcode) < 4 {
		return errors.NewBadRequestError("passcode must be at least 4 characters", nil)
	}
	if len(passcode) > 64 {
		return errors.NewBadRequestError("passcode must be at most 64 characters", nil)
	}
	return nil
}
