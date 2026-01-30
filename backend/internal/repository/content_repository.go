package repository

import (
	"context"
	"database/sql"
	"fmt"
	"konbi/internal/errors"
	"konbi/internal/models"
	"os"

	"github.com/sirupsen/logrus"
)

// content repository handles database operations for content
type ContentRepository struct {
	db         *sql.DB
	logger     *logrus.Logger
	isPostgres bool
}

// create new repository
func NewContentRepository(db *sql.DB, logger *logrus.Logger) *ContentRepository {
	return &ContentRepository{
		db:         db,
		logger:     logger,
		isPostgres: os.Getenv("DATABASE_URL") != "",
	}
}

// helper to get current timestamp function based on database type
func (r *ContentRepository) nowFunc() string {
	if r.isPostgres {
		return "NOW()"
	}
	return "datetime('now')"
}

// helper to convert ? placeholders to postgresql $1, $2, etc
func (r *ContentRepository) convertQuery(query string) string {
	if !r.isPostgres {
		return query
	}
	
	// convert ? to $1, $2, $3...
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

// create inserts new content record
func (r *ContentRepository) Create(ctx context.Context, content *models.Content) error {
	query := r.convertQuery(`
		INSERT INTO content (id, type, title, filename, filepath, filesize, content, expires_at, view_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
	`)

	_, err := r.db.ExecContext(ctx, query,
		content.ID,
		content.Type,
		content.Title,
		content.Filename,
		content.Filepath,
		content.Filesize,
		content.Content,
		content.ExpiresAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("content_id", content.ID).Error("failed to create content")
		return errors.NewInternalError("failed to save content", err)
	}

	r.logger.WithFields(logrus.Fields{
		"content_id":   content.ID,
		"content_type": content.Type,
	}).Info("content created successfully")

	return nil
}

// find by id retrieves content by id
func (r *ContentRepository) FindByID(ctx context.Context, id string) (*models.Content, error) {
	query := r.convertQuery(fmt.Sprintf(`
		SELECT id, type, title, filename, filepath, filesize, content, created_at, expires_at, view_count, deleted_at
		FROM content
		WHERE id = ? AND (deleted_at IS NULL OR deleted_at > %s)
	`, r.nowFunc()))

	content := &models.Content{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&content.ID,
		&content.Type,
		&content.Title,
		&content.Filename,
		&content.Filepath,
		&content.Filesize,
		&content.Content,
		&content.CreatedAt,
		&content.ExpiresAt,
		&content.ViewCount,
		&content.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("content not found")
	}
	if err != nil {
		r.logger.WithError(err).WithField("content_id", id).Error("failed to find content")
		return nil, errors.NewInternalError("database error", err)
	}

	return content, nil
}

// find active by id retrieves non-expired content
func (r *ContentRepository) FindActiveByID(ctx context.Context, id string) (*models.Content, error) {
	query := r.convertQuery(fmt.Sprintf(`
		SELECT id, type, title, filename, filepath, filesize, content, created_at, expires_at, view_count, deleted_at
		FROM content
		WHERE id = ? AND expires_at > %s AND (deleted_at IS NULL OR deleted_at > %s)
	`, r.nowFunc(), r.nowFunc()))

	content := &models.Content{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&content.ID,
		&content.Type,
		&content.Title,
		&content.Filename,
		&content.Filepath,
		&content.Filesize,
		&content.Content,
		&content.CreatedAt,
		&content.ExpiresAt,
		&content.ViewCount,
		&content.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("content not found or expired")
	}
	if err != nil {
		r.logger.WithError(err).WithField("content_id", id).Error("failed to find active content")
		return nil, errors.NewInternalError("database error", err)
	}

	return content, nil
}

// id exists checks if content id already exists
func (r *ContentRepository) IDExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := r.convertQuery("SELECT EXISTS(SELECT 1 FROM content WHERE id = ?)")
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).WithField("content_id", id).Error("failed to check id existence")
		return false, errors.NewInternalError("database error", err)
	}
	return exists, nil
}

// increment view count increases view counter
func (r *ContentRepository) IncrementViewCount(ctx context.Context, id string) error {
	query := r.convertQuery("UPDATE content SET view_count = view_count + 1 WHERE id = ?")
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("content_id", id).Error("failed to increment view count")
		return errors.NewInternalError("failed to update view count", err)
	}
	return nil
}

// list all retrieves all content (for admin)
func (r *ContentRepository) ListAll(ctx context.Context) ([]*models.Content, error) {
	query := fmt.Sprintf(`
		SELECT id, type, title, filename, filesize, created_at, expires_at, view_count, deleted_at
		FROM content
		WHERE deleted_at IS NULL OR deleted_at > %s
		ORDER BY created_at DESC
	`, r.nowFunc())

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("failed to list content")
		return nil, errors.NewInternalError("database error", err)
	}
	defer rows.Close()

	var contents []*models.Content
	for rows.Next() {
		content := &models.Content{}
		err := rows.Scan(
			&content.ID,
			&content.Type,
			&content.Title,
			&content.Filename,
			&content.Filesize,
			&content.CreatedAt,
			&content.ExpiresAt,
			&content.ViewCount,
			&content.DeletedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("failed to scan content row")
			continue
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// find expired content retrieves expired file content
func (r *ContentRepository) FindExpiredContent(ctx context.Context) ([]*models.Content, error) {
	query := r.convertQuery(fmt.Sprintf(`
		SELECT id, filepath
		FROM content
		WHERE expires_at < %s AND type = ? AND (deleted_at IS NULL OR deleted_at > %s)
	`, r.nowFunc(), r.nowFunc()))

	rows, err := r.db.QueryContext(ctx, query, models.ContentTypeFile)
	if err != nil {
		r.logger.WithError(err).Error("failed to find expired content")
		return nil, errors.NewInternalError("database error", err)
	}
	defer rows.Close()

	var contents []*models.Content
	for rows.Next() {
		content := &models.Content{}
		err := rows.Scan(&content.ID, &content.Filepath)
		if err != nil {
			r.logger.WithError(err).Error("failed to scan expired content")
			continue
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// soft delete marks content as deleted
func (r *ContentRepository) SoftDelete(ctx context.Context, id string) error {
	query := r.convertQuery(fmt.Sprintf("UPDATE content SET deleted_at = %s WHERE id = ?", r.nowFunc()))
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("content_id", id).Error("failed to soft delete content")
		return errors.NewInternalError("failed to delete content", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NewNotFoundError("content not found")
	}

	r.logger.WithField("content_id", id).Info("content soft deleted")
	return nil
}

// delete expired permanently removes expired records
func (r *ContentRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("DELETE FROM content WHERE expires_at < %s", r.nowFunc())
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("failed to delete expired content")
		return 0, errors.NewInternalError("failed to delete expired content", err)
	}

	count, _ := result.RowsAffected()
	r.logger.WithField("deleted_count", count).Info("expired content deleted")
	return count, nil
}

// transaction wrapper for complex operations
func (r *ContentRepository) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithError(err).Error("failed to begin transaction")
		return errors.NewInternalError("failed to start transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			r.logger.WithField("panic", p).Error("panic in transaction, rolling back")
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.WithError(rbErr).Error("failed to rollback transaction")
			return errors.NewInternalError(fmt.Sprintf("transaction error: %v, rollback error: %v", err, rbErr), err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		r.logger.WithError(err).Error("failed to commit transaction")
		return errors.NewInternalError("failed to commit transaction", err)
	}

	return nil
}