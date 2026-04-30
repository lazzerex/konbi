package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"konbi/internal/config"
	"konbi/internal/errors"
	"konbi/internal/models"
	"konbi/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// auth service handles authentication operations
type AuthService struct {
	userRepo *repository.UserRepository
	config   *config.Config
	logger   *logrus.Logger
}

// create new auth service
func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config, logger *logrus.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   cfg,
		logger:   logger,
	}
}

// register creates new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		s.logger.WithField("email", req.Email).Warn("registration attempted with existing email")
		return nil, errors.NewConflictError("email already registered")
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.WithError(err).Error("failed to hash password")
		return nil, errors.NewInternalError("password hashing failed", err)
	}

	// generate user id
	id := generateID()

	// create user
	user := &models.User{
		ID:           id,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithError(err).Error("failed to create user")
		return nil, errors.NewInternalError("registration failed", err)
	}

	s.logger.WithField("user_id", id).Info("user registered successfully")

	// generate tokens
	return s.generateAuthResponse(user)
}

// login authenticates user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.WithError(err).WithField("email", req.Email).Error("database error fetching user")
		return nil, errors.NewInternalError("login failed", err)
	}

	if user == nil {
		s.logger.WithField("email", req.Email).Warn("login attempted with non-existent email")
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.WithField("user_id", user.ID).Warn("login failed with incorrect password")
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	s.logger.WithField("user_id", user.ID).Info("user logged in successfully")

	// generate tokens
	return s.generateAuthResponse(user)
}

// verify access token and extract claims
func (s *AuthService) VerifyAccessToken(tokenString string) (*models.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Server.JWTSecret), nil
	})

	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.NewUnauthorizedError("invalid token claims")
	}

	userID, ok := (*claims)["user_id"].(string)
	if !ok {
		return nil, errors.NewUnauthorizedError("invalid user id in token")
	}

	email, ok := (*claims)["email"].(string)
	if !ok {
		return nil, errors.NewUnauthorizedError("invalid email in token")
	}

	return &models.TokenClaims{
		UserID: userID,
		Email:  email,
	}, nil
}

// verify refresh token and return new access token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// parse refresh token (same verification, different purpose)
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Server.JWTRefreshSecret), nil
	})

	if err != nil {
		return "", errors.NewUnauthorizedError("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.NewUnauthorizedError("invalid refresh token claims")
	}

	userID, ok := (*claims)["user_id"].(string)
	if !ok {
		return "", errors.NewUnauthorizedError("invalid user id in refresh token")
	}

	// get user to ensure still exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("failed to get user for refresh")
		return "", errors.NewInternalError("token refresh failed", err)
	}

	if user == nil {
		return "", errors.NewUnauthorizedError("user not found")
	}

	// generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// private helper to generate tokens
func (s *AuthService) generateAuthResponse(user *models.User) (*models.AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.Server.JWTExpiry.Seconds()),
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// generate access token
func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().UTC().Add(s.config.Server.JWTExpiry).Unix(),
		"iat":     time.Now().UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Server.JWTSecret))
	if err != nil {
		s.logger.WithError(err).Error("failed to sign access token")
		return "", errors.NewInternalError("token generation failed", err)
	}

	return tokenString, nil
}

// generate refresh token
func (s *AuthService) generateRefreshToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().UTC().Add(s.config.Server.JWTRefreshExpiry).Unix(),
		"iat":     time.Now().UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Server.JWTRefreshSecret))
	if err != nil {
		s.logger.WithError(err).Error("failed to sign refresh token")
		return "", errors.NewInternalError("token generation failed", err)
	}

	return tokenString, nil
}

// generate unique id
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
