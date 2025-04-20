//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
)

const defaultEmail = "employee@example.com"

const defaultPasswordHash = "password"

const employeeRole = "employee"
const moderatorRole = "moderator"
const invalidRole = "invalidRole"

var userID string

func TestUserRepository_Ok(t *testing.T) {
	t.Parallel()

	t.Run("ValidateRole", testValidateRole)

	t.Run("CreateUser", testCreateUser)

	t.Run("UserNyEmail", testUserByEmail)
}

func testValidateRole(t *testing.T) {
	ctx := context.Background()

	repo := db.ROUsers()

	err := repo.ValidateRole(ctx, employeeRole)
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("expected role to be found")
	} else if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	err = repo.ValidateRole(ctx, moderatorRole)
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("expected role to be found")
	} else if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	err = repo.ValidateRole(ctx, invalidRole)
	if err == nil {
		t.Fatalf("expected error on invalid role")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on invalid role, got: %v", err)
	}
}

func testCreateUser(t *testing.T) {
	ctx := context.Background()

	repo := db.RWUsers()

	_, err := repo.CreateUser(ctx, defaultEmail, defaultPasswordHash, invalidRole)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got: %v", err)
	}

	userID, err = repo.CreateUser(ctx, defaultEmail, defaultPasswordHash, employeeRole)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = repo.CreateUser(ctx, defaultEmail, "newPasswordHash", employeeRole)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected unique violation, got: %v", err)
	}
}

func testUserByEmail(t *testing.T) {
	ctx := context.Background()

	repo := db.ROUsers()

	user, err := repo.UserByEmail(ctx, "nonexistent@email.com")
	if err == nil {
		t.Fatalf("expected error on nonexistent email")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on nonexistent email, got: %v", err)
	}

	user, err = repo.UserByEmail(ctx, defaultEmail)
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("expected user with default email to be found")
	} else if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	} else if user.Email != defaultEmail {
		t.Fatalf("got unexpected user email: %s", user.Email)
	} else if user.PasswordHash != defaultPasswordHash {
		t.Fatalf("got unexpected user password: %s", user.PasswordHash)
	}
}
