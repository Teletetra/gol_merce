package service

import (
	"context"
	"strings"
	"time"

	"ecommerce_go/internal/domain"
)

type AuthService struct {
	users       domain.UserRepository
	tokenSecret string
	tokenTTL    time.Duration
}

func NewAuthService(users domain.UserRepository, tokenSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{users: users, tokenSecret: tokenSecret, tokenTTL: tokenTTL}
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput, role domain.Role) (domain.User, string, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" || len(input.Password) < 8 {
		return domain.User{}, "", ErrValidation
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, found, err := s.users.FindByEmail(ctx, email); err != nil {
		return domain.User{}, "", err
	} else if found {
		return domain.User{}, "", ErrConflict
	}

	user := domain.User{
		ID:           newID("usr"),
		Name:         strings.TrimSpace(input.Name),
		Email:        email,
		PasswordHash: hashPassword(input.Password),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	created, err := s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, "", err
	}

	token, err := signToken(s.tokenSecret, tokenClaims{
		UserID: created.ID,
		Role:   created.Role,
		Exp:    time.Now().Add(s.tokenTTL).Unix(),
	})
	if err != nil {
		return domain.User{}, "", err
	}

	return created, token, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (domain.User, string, error) {
	user, found, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		return domain.User{}, "", err
	}
	if !found || !verifyPassword(input.Password, user.PasswordHash) {
		return domain.User{}, "", ErrInvalidCredentials
	}

	token, err := signToken(s.tokenSecret, tokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Exp:    time.Now().Add(s.tokenTTL).Unix(),
	})
	if err != nil {
		return domain.User{}, "", err
	}

	return user, token, nil
}

func (s *AuthService) ValidateToken(token string) (tokenClaims, error) {
	return parseToken(s.tokenSecret, token)
}

func (s *AuthService) Me(ctx context.Context, userID string) (domain.User, error) {
	user, found, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if !found {
		return domain.User{}, ErrUnauthorized
	}
	return user, nil
}
