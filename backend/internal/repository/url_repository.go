package repository

import (
	"context"
	"database/sql"
	"fmt"
	"konbi/internal/errors"
	"konbi/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

// url repository handles database operations for shortened urls
type URLRepository struct {
	db         *sql.DB
	logger     *logrus.Logger
	isPostgres bool
}

// create new url repository
func NewURLRepository(db *sql.DB, logger *logrus.Logger, isPostgres bool) *URLRepository {
	return &URLRepository{
		db:         db,
		logger:     logger,
		isPostgres: isPostgres,
	}
}

// helper to get current timestamp function
func (r *URLRepository) nowFunc() string {
	if r.isPostgres {
		return "NOW()"
	}
	return "datetime('now')"
}

// helper to convert ? to $1, $2 for postgres
func (r *URLRepository) convertQuery(query string) string {
	if !r.isPostgres {
		return query
	}
	
	result := ""
	paramCount := 1
	for _, char := range query {
		if char == '?' {
			result += fmt.Sprintf("$%d", paramCount)
			paramCount++
		} else {
			result += string(char)
		}
	}
	return result
}

// create inserts new shortened url
func (r *URLRepository) Create(ctx context.Context, url *models.ShortenedURL) error {
	query := r.convertQuery(`
		INSERT INTO shortened_urls (short_code, original_url, custom_alias, expires_at, click_count)
		VALUES (?, ?, ?, ?, 0)
		RETURNING id, created_at
	`)

	if r.isPostgres {
		err := r.db.QueryRowContext(ctx, query,
			url.ShortCode,
			url.OriginalURL,
			url.CustomAlias,
			url.ExpiresAt,
		).Scan(&url.ID, &url.CreatedAt)

		if err != nil {
			r.logger.WithError(err).Error("failed to create shortened url")
			return errors.NewInternalError("failed to create short url", err)
		}
	} else {
		// sqlite doesn't support returning, so do insert then query
		query = r.convertQuery(`
			INSERT INTO shortened_urls (short_code, original_url, custom_alias, expires_at, click_count)
			VALUES (?, ?, ?, ?, 0)
		`)
		result, err := r.db.ExecContext(ctx, query,
			url.ShortCode,
			url.OriginalURL,
			url.CustomAlias,
			url.ExpiresAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("failed to create shortened url")
			return errors.NewInternalError("failed to create short url", err)
		}

		id, _ := result.LastInsertId()
		url.ID = id
		url.CreatedAt = time.Now().UTC()
	}

	r.logger.WithField("short_code", url.ShortCode).Info("shortened url created")
	return nil
}

// find by short code retrieves url by short code
func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*models.ShortenedURL, error) {
	query := r.convertQuery(fmt.Sprintf(`
		SELECT id, short_code, original_url, custom_alias, created_at, expires_at, click_count, deleted_at
		FROM shortened_urls
		WHERE short_code = ? AND (deleted_at IS NULL OR deleted_at > %s)
	`, r.nowFunc()))

	url := &models.ShortenedURL{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CustomAlias,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.ClickCount,
		&url.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("short url not found")
	}
	if err != nil {
		r.logger.WithError(err).WithField("short_code", shortCode).Error("failed to find short url")
		return nil, errors.NewInternalError("database error", err)
	}

	return url, nil
}

// find active by short code retrieves non-expired url
func (r *URLRepository) FindActiveByShortCode(ctx context.Context, shortCode string) (*models.ShortenedURL, error) {
	query := r.convertQuery(fmt.Sprintf(`
		SELECT id, short_code, original_url, custom_alias, created_at, expires_at, click_count, deleted_at
		FROM shortened_urls
		WHERE short_code = ? 
		AND (expires_at IS NULL OR expires_at > %s)
		AND (deleted_at IS NULL OR deleted_at > %s)
	`, r.nowFunc(), r.nowFunc()))

	url := &models.ShortenedURL{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CustomAlias,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.ClickCount,
		&url.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("short url not found or expired")
	}
	if err != nil {
		r.logger.WithError(err).WithField("short_code", shortCode).Error("failed to find active short url")
		return nil, errors.NewInternalError("database error", err)
	}

	return url, nil
}

// short code exists checks if short code already exists
func (r *URLRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var exists bool
	query := r.convertQuery("SELECT EXISTS(SELECT 1 FROM shortened_urls WHERE short_code = ?)")
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).WithField("short_code", shortCode).Error("failed to check short code existence")
		return false, errors.NewInternalError("database error", err)
	}
	return exists, nil
}

// increment click count increases click counter
func (r *URLRepository) IncrementClickCount(ctx context.Context, id int64) error {
	query := r.convertQuery("UPDATE shortened_urls SET click_count = click_count + 1 WHERE id = ?")
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("url_id", id).Error("failed to increment click count")
		return errors.NewInternalError("failed to update click count", err)
	}
	return nil
}

// record click saves click analytics
func (r *URLRepository) RecordClick(ctx context.Context, click *models.URLClick) error {
	query := r.convertQuery(`
		INSERT INTO url_clicks (url_id, ip_address, user_agent, referrer)
		VALUES (?, ?, ?, ?)
	`)

	_, err := r.db.ExecContext(ctx, query,
		click.URLID,
		click.IPAddress,
		click.UserAgent,
		click.Referrer,
	)

	if err != nil {
		r.logger.WithError(err).Error("failed to record click")
		return errors.NewInternalError("failed to record click", err)
	}

	return nil
}

// get recent clicks retrieves recent click analytics
func (r *URLRepository) GetRecentClicks(ctx context.Context, urlID int64, limit int) ([]*models.URLClick, error) {
	query := r.convertQuery(`
		SELECT id, url_id, clicked_at, ip_address, user_agent, referrer
		FROM url_clicks
		WHERE url_id = ?
		ORDER BY clicked_at DESC
		LIMIT ?
	`)

	rows, err := r.db.QueryContext(ctx, query, urlID, limit)
	if err != nil {
		r.logger.WithError(err).Error("failed to get recent clicks")
		return nil, errors.NewInternalError("database error", err)
	}
	defer rows.Close()

	var clicks []*models.URLClick
	for rows.Next() {
		click := &models.URLClick{}
		err := rows.Scan(
			&click.ID,
			&click.URLID,
			&click.ClickedAt,
			&click.IPAddress,
			&click.UserAgent,
			&click.Referrer,
		)
		if err != nil {
			r.logger.WithError(err).Error("failed to scan click row")
			continue
		}
		clicks = append(clicks, click)
	}

	return clicks, nil
}

// soft delete marks url as deleted
func (r *URLRepository) SoftDelete(ctx context.Context, id int64) error {
	query := r.convertQuery(fmt.Sprintf("UPDATE shortened_urls SET deleted_at = %s WHERE id = ?", r.nowFunc()))
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("url_id", id).Error("failed to soft delete url")
		return errors.NewInternalError("failed to delete url", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NewNotFoundError("url not found")
	}

	r.logger.WithField("url_id", id).Info("url soft deleted")
	return nil
}
