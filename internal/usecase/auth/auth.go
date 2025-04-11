package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"pvz/internal/auth"
	"pvz/internal/models"
	"pvz/internal/postgres"
)

type Issuer interface {
	Issue(userRole string) (string, error)
}

type Deps struct {
	Issuer Issuer
	Repo   postgres.DB
}
type AuthService struct {
	Deps
}

func NewAuthService(deps Deps) *AuthService {
	return &AuthService{
		Deps: deps,
	}
}

// UserToken checks input credentials and returns a token.
func (a *AuthService) UserToken(ctx context.Context, email, password string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Auth/UserToken")
	defer span.Finish()

	tokenInfo, err := a.Repo.ROUsers().UserByEmail(ctx, email)

	switch {
	case errors.Is(err, postgres.ErrNotFound):
		return "", fmt.Errorf("authorization: %w", err)
	case err != nil:
		return "", fmt.Errorf("could not get user while attempting to authorize: %w", err)
	default:
		err := bcrypt.CompareHashAndPassword([]byte(tokenInfo.PasswordHash), []byte(password))
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", fmt.Errorf("incorrect password: %w", auth.ErrUnauthorized)
		} else if err != nil {
			return "", fmt.Errorf("failed to verify password: %w", err)
		}
	}

	token, err := a.Deps.Issuer.Issue(tokenInfo.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Register creates user and returns their id.
func (a *AuthService) Register(ctx context.Context, email, password, role string) (models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Auth/createUser")
	defer span.Finish()

	passwordHashRaw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return models.User{}, fmt.Errorf("could not hash password: %w", err)
	}

	id, err := a.Deps.Repo.RWUsers().CreateUser(ctx, email, string(passwordHashRaw), role)
	if errors.Is(err, postgres.ErrAlreadyExists) {
		return models.User{}, ErrUserAlreadyExists
	} else if err != nil {
		return models.User{}, err
	}

	return models.User{
		Id:    id,
		Email: email,
		Role:  role,
	}, nil
}

// DummyLogin returns valid gwt-token reflecting input role.
func (a *AuthService) DummyLogin(ctx context.Context, role string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Auth/DummyLogin")
	defer span.Finish()

	return a.Deps.Issuer.Issue(role)
}
