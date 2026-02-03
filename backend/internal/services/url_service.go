package services

import (
	"context"
	"fmt"
	"konbi/internal/config"
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/repository"
	"math/rand"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	shortCodeLength = 6
	maxRetries = 5
)

// url service handles business logic for url shortening
type URLService struct {
	repo   *repository.URLRepository
	cfg    *config.Config
	logger *logrus.Logger
	rng    *rand.Rand
}

// create new url service
func NewURLService(repo *repository.URLRepository, cfg *config.Config, logger *logrus.Logger) *URLService {
	return &URLService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// shorten url creates a shortened url
func (s *URLService) ShortenURL(ctx context.Context, req *models.ShortenRequest) (*models.ShortenResponse, error) {
	// validate url
	if !isValidURL(req.URL) {
		return nil, errors.NewBadRequestError("invalid url format", nil)
	}

	// generate or use custom short code
	var shortCode string
	var err error

	if req.CustomAlias != nil && *req.CustomAlias != "" {
		shortCode = *req.CustomAlias
		// validate custom alias
		if len(shortCode) < 3 || len(shortCode) > 20 {
			return nil, errors.NewBadRequestError("custom alias must be 3-20 characters", nil)
		}
		if !isAlphanumeric(shortCode) {
			return nil, errors.NewBadRequestError("custom alias must be alphanumeric", nil)
		}

		// check if already exists
		exists, err := s.repo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.NewConflictError("custom alias already taken")
		}
	} else {
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	// calculate expiration
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expires := time.Now().UTC().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &expires
	}

	// create url record
	urlRecord := &models.ShortenedURL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CustomAlias: req.CustomAlias,
		ExpiresAt:   expiresAt,
	}

	err = s.repo.Create(ctx, urlRecord)
	if err != nil {
		return nil, err
	}

	// build short url
	shortURL := fmt.Sprintf("%s/s/%s", s.cfg.Server.BaseURL, shortCode)

	response := &models.ShortenResponse{
		ID:          urlRecord.ID,
		ShortCode:   shortCode,
		ShortURL:    shortURL,
		OriginalURL: req.URL,
		CreatedAt:   urlRecord.CreatedAt,
		ExpiresAt:   expiresAt,
	}

	s.logger.WithFields(logrus.Fields{
		"short_code": shortCode,
		"url_id":     urlRecord.ID,
	}).Info("url shortened successfully")

	return response, nil
}

// redirect gets original url and records analytics
func (s *URLService) Redirect(ctx context.Context, shortCode string, ipAddress string, userAgent string, referrer string) (string, error) {
	// find active url
	urlRecord, err := s.repo.FindActiveByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// record click analytics asynchronously
	go func() {
		click := &models.URLClick{
			URLID:     urlRecord.ID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Referrer:  referrer,
		}
		
		// use background context for async operation
		bgCtx := context.Background()
		
		if err := s.repo.RecordClick(bgCtx, click); err != nil {
			s.logger.WithError(err).Error("failed to record click asynchronously")
		}
		
		if err := s.repo.IncrementClickCount(bgCtx, urlRecord.ID); err != nil {
			s.logger.WithError(err).Error("failed to increment click count asynchronously")
		}
	}()

	return urlRecord.OriginalURL, nil
}

// get stats retrieves url statistics
func (s *URLService) GetStats(ctx context.Context, shortCode string) (*models.URLStatsResponse, error) {
	urlRecord, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// get recent clicks (last 100)
	clicks, err := s.repo.GetRecentClicks(ctx, urlRecord.ID, 100)
	if err != nil {
		s.logger.WithError(err).Error("failed to get recent clicks")
		// don't fail entire request if clicks fail
		clicks = []*models.URLClick{}
	}

	// convert to response format
	recentClicks := make([]models.URLClick, len(clicks))
	for i, click := range clicks {
		recentClicks[i] = *click
	}

	stats := &models.URLStatsResponse{
		ShortCode:    urlRecord.ShortCode,
		OriginalURL:  urlRecord.OriginalURL,
		ClickCount:   urlRecord.ClickCount,
		CreatedAt:    urlRecord.CreatedAt,
		ExpiresAt:    urlRecord.ExpiresAt,
		RecentClicks: recentClicks,
	}

	return stats, nil
}

// delete url soft deletes a shortened url
func (s *URLService) DeleteURL(ctx context.Context, shortCode string) error {
	urlRecord, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, urlRecord.ID)
}

// generate unique short code with collision handling
func (s *URLService) generateUniqueShortCode(ctx context.Context) (string, error) {
	for i := 0; i < maxRetries; i++ {
		code := s.generateBase62Code(shortCodeLength)
		
		exists, err := s.repo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return code, nil
		}
		
		s.logger.WithField("attempt", i+1).Warn("short code collision detected, retrying")
	}
	
	return "", errors.NewInternalError("failed to generate unique short code after max retries", nil)
}

// generate base62 encoded short code
func (s *URLService) generateBase62Code(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = base62Chars[s.rng.Intn(len(base62Chars))]
	}
	return string(code)
}

// validate url format
func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// check if string is alphanumeric
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return false
		}
	}
	return true
}
