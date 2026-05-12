package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"CallItCureIt/backend/internal/config"
	"CallItCureIt/backend/internal/db"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	repo Repository
	cfg  config.Config
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token string
	User  *db.User
}

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewService(repo Repository, cfg config.Config) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *Service) EnsureDevAdmin(ctx context.Context) error {
	if !s.cfg.DevSeedAdmin {
		return nil
	}

	email := strings.ToLower(strings.TrimSpace(s.cfg.DevAdminEmail))

	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(s.cfg.DevAdminPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user := &db.User{
		ID:           uuid.NewString(),
		Email:        email,
		Name:         s.cfg.DevAdminName,
		PasswordHash: string(passwordHash),
		Role:         "admin",
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.TrimSpace(input.Password)

	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(s.cfg.JWTSecret), nil
		},
		jwt.WithIssuer(s.cfg.JWTIssuer),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidCredentials
	}

	return claims, nil
}

func (s *Service) issueToken(user *db.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.cfg.JWTExpirationMinutes) * time.Minute)

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWTIssuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}