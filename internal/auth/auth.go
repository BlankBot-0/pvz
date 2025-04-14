package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"pvz/internal/config"
	"time"
)

type Claims interface {
	UserRole() string
}

type Auth struct {
	PrivateKey     string
	ExpirationTime time.Duration
}

func New(cfg config.Auth) *Auth {
	return &Auth{
		PrivateKey:     cfg.PrivateKey,
		ExpirationTime: cfg.ExpirationTime,
	}
}

func (a *Auth) Issue(userRole string) (string, error) {
	expirationTime := time.Now().Add(a.ExpirationTime)
	claims := &privateClaims{
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.PrivateKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *Auth) Verify(token string) (Claims, error) {
	claims := &privateClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(a.PrivateKey), nil
	})
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !tkn.Valid {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

type privateClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (p *privateClaims) UserRole() string {
	return p.Role
}

type userRoleKey struct{}

func GetUserRoleFromCtx(ctx context.Context) string {
	userRole, ok := ctx.Value(userRoleKey{}).(string)
	if !ok {
		return ""
	}
	return userRole
}

func SetUserRoleToCtx(ctx context.Context, userRole string) context.Context {
	return context.WithValue(ctx, userRoleKey{}, userRole)
}
