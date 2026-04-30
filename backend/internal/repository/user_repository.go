package repository

import (
	"context"
	"database/sql"
	"konbi/internal/errors"
	"konbi/internal/models"

	"github.com/sirupsen/logrus"
)

// user repository handles user database operations
type UserRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// create new user repository
func NewUserRepository(db *sql.DB, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// create inserts new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		r.logger.WithError(err).WithField("email", user.Email).Error("failed to create user")
		if err.Error() == "UNIQUE constraint failed: users.email" || err.Error() == "duplicate key value violates unique constraint \"users_email_key\"" {
			return errors.NewConflictError("email already exists")
		}
		return errors.NewInternalError("failed to create user", err)
	}

	return nil
}

// get by id retrieves user by id
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		r.logger.WithError(err).WithField("user_id", id).Error("failed to get user by id")
		return nil, errors.NewInternalError("failed to get user", err)
	}

	return &user, nil
}

// get by email retrieves user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		r.logger.WithError(err).WithField("email", email).Error("failed to get user by email")
		return nil, errors.NewInternalError("failed to get user", err)
	}

	return &user, nil
}
