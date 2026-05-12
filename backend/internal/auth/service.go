package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"CallItCureIt/backend/internal/db"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserDisabled = errors.New("user disabled")
var ErrForbidden = errors.New("forbidden")

type Service struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) *Service {
	if strings.TrimSpace(jwtSecret) == "" {
		jwtSecret = "dev-only-change-me"
	}

	return &Service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	User  *db.User
	Token string
}

type CreateUserInput struct {
	Email    string
	Password string
	FullName string
	Role     string
}

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) Login(
	ctx context.Context,
	input LoginInput,
) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != "active" {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:  user,
		Token: token,
	}, nil
}

func (s *Service) AuthenticateToken(
	ctx context.Context,
	tokenString string,
) (*db.User, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, ErrInvalidCredentials
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidCredentials
			}

			return s.jwtSecret, nil
		},
	)

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != "active" {
		return nil, ErrUserDisabled
	}

	return user, nil
}

func (s *Service) CreateUser(
	ctx context.Context,
	input CreateUserInput,
) (*db.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.Password == "" || strings.TrimSpace(input.FullName) == "" {
		return nil, ErrInvalidCredentials
	}

	role := input.Role
	if role == "" {
		role = "trainee"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &db.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Role:         role,
		Status:       "active",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) issueToken(user *db.User) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    "call-it-cure-it",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(12 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecret)
}