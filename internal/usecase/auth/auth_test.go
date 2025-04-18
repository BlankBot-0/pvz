package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	tokenmocks "pvz/internal/auth/mocks"
	"pvz/internal/models"
	"pvz/internal/postgres"
	pgmocks "pvz/internal/postgres/mocks"
	"testing"
)

func TestAuthService_DummyLogin_Ok(t *testing.T) {
	t.Parallel()

	dbMock := pgmocks.NewDBMock(t)
	dbMock.ValidateRoleMock.Return(nil)

	issuerMock := tokenmocks.NewIssuerMock(t)
	issuerMock.IssueMock.Set(func(userRole string) (token string, err error) {
		return userRole, nil
	})

	auth := AuthService{Deps{
		Issuer: issuerMock,
		Repo:   dbMock,
	}}
	role := "employee"
	token, err := auth.DummyLogin(context.Background(), role)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if token != "employee" {
		t.Fatalf("got unexpected token: %s", token)
	}

	role = "moderator"
	token, err = auth.DummyLogin(context.Background(), role)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if token != "moderator" {
		t.Fatalf("got unexpected token: %s", token)
	}
}

func TestAuthService_DummyLogin_InvalidRole(t *testing.T) {
	t.Parallel()

	dbMock := pgmocks.NewDBMock(t)
	dbMock.UserMock.ValidateRoleMock.Set(func(_ context.Context, role string) error {
		if role != "employee" && role != "moderator" {
			return postgres.ErrNotFound
		}
		return nil
	})

	issuerMock := tokenmocks.NewIssuerMock(t)
	issuerMock.IssueMock.Set(func(userRole string) (token string, err error) {
		return userRole, nil
	})

	auth := AuthService{Deps{
		Issuer: issuerMock,
		Repo:   dbMock,
	}}
	role := "employee"
	token, err := auth.DummyLogin(context.Background(), role)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if token != "employee" {
		t.Fatalf("got unexpected token: %s", token)
	}

	role = "unexpectedRole"
	token, err = auth.DummyLogin(context.Background(), role)
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got: %s", err)
	}
}

func TestAuthService_Register_Ok(t *testing.T) {
	t.Parallel()

	userID := "userID"
	dbMock := pgmocks.NewDBMock(t)
	dbMock.UserMock.CreateUserMock.Return(userID, nil)

	auth := AuthService{Deps{
		Repo: dbMock,
	}}

	email := "user@example.com"
	password := "password"
	role := "employee"

	user, err := auth.Register(context.Background(), email, password, role)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if user.Email != email {
		t.Errorf("got unexpected email: %s", user.Email)
	}
	if user.Role != role {
		t.Errorf("got unexpected role: %s", user.Role)
	}
	if user.ID != userID {
		t.Errorf("got unexpected user ID: %s", userID)
	}
}

func TestAuthService_Register_ErrorHandling(t *testing.T) {
	t.Parallel()

	userID := "userID"
	password := "password"
	role := "employee"

	unexpectedErr := fmt.Errorf("unexpected error")
	tcs := []struct {
		name        string
		innerErr    error
		expectedErr error
	}{
		{"user already exists", postgres.ErrAlreadyExists, ErrUserAlreadyExists},
		{"invalid user role", postgres.ErrInvalidReference, ErrInvalidRole},
		{"unexpected error", unexpectedErr, unexpectedErr},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dbMock := pgmocks.NewDBMock(t)
			dbMock.UserMock.CreateUserMock.Return("", tc.innerErr)

			auth := AuthService{Deps{
				Repo: dbMock,
			}}

			_, err := auth.Register(context.Background(), userID, password, role)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected %s, got: %s", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_UserToken_Ok(t *testing.T) {
	t.Parallel()

	userPassword := "password"
	passwordHashRaw, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("unexpected error while hashing password: %s", err)
	}

	dbMock := pgmocks.NewDBMock(t)
	dbMock.UserMock.UserByEmailMock.Return(models.User{
		Email:        "employee@example.com",
		PasswordHash: string(passwordHashRaw),
		Role:         "employee",
	}, nil)

	issuerMock := tokenmocks.NewIssuerMock(t)
	issuerMock.IssueMock.Set(func(userRole string) (token string, err error) {
		return userRole, nil
	})

	auth := AuthService{Deps{
		Issuer: issuerMock,
		Repo:   dbMock,
	}}

	token, err := auth.UserToken(context.Background(), "", userPassword)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}
	if token != "employee" {
		t.Fatalf("got unexpected token: %s", token)
	}
}

func TestAuthService_UserToken_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("user not registered", func(t *testing.T) {
		t.Parallel()
		dbMock := pgmocks.NewDBMock(t)
		dbMock.UserMock.UserByEmailMock.Return(models.User{}, postgres.ErrNotFound)
		auth := AuthService{Deps{
			Repo: dbMock,
		}}

		_, err := auth.UserToken(context.Background(), "", "")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %s", err)
		}
	})

	t.Run("unexpected repository error", func(t *testing.T) {
		t.Parallel()
		dbMock := pgmocks.NewDBMock(t)
		dbMock.UserMock.UserByEmailMock.Return(models.User{}, fmt.Errorf("some error"))
		auth := AuthService{Deps{
			Repo: dbMock,
		}}

		_, err := auth.UserToken(context.Background(), "", "")
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("issuer error", func(t *testing.T) {
		t.Parallel()
		userPassword := "password"
		passwordHashRaw, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("unexpected error while hashing password: %s", err)
		}

		dbMock := pgmocks.NewDBMock(t)
		dbMock.UserMock.UserByEmailMock.Return(models.User{
			Email:        "employee@example.com",
			PasswordHash: string(passwordHashRaw),
			Role:         "employee",
		}, nil)

		issuerMock := tokenmocks.NewIssuerMock(t)
		issuerMock.IssueMock.Return("", fmt.Errorf("some error"))

		auth := AuthService{Deps{
			Issuer: issuerMock,
			Repo:   dbMock,
		}}

		if _, err = auth.UserToken(context.Background(), "", userPassword); err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		t.Parallel()
		dbMock := pgmocks.NewDBMock(t)

		expectedPasswordHashedRaw, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("unexpected error while hashing password: %s", err)
		}

		dbMock.UserMock.UserByEmailMock.Return(models.User{
			PasswordHash: string(expectedPasswordHashedRaw),
		}, nil)

		auth := AuthService{Deps{
			Repo: dbMock,
		}}

		if _, err = auth.UserToken(context.Background(), "", "wrong-password"); !errors.Is(err, ErrIncorrectPassword) {
			t.Fatalf("expected ErrIncorrectPassword, got: %s", err)
		}
	})
}
