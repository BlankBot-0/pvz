package fake

import (
	"context"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"pvz/internal/models"
	"pvz/internal/postgres"
)

var _ postgres.RWUsers = &UsersRepoFake{}

type UsersRepoFake struct {
	lastId string
	users  []models.User

	CreateUserErr  error
	UserByIdErr    error
	UserByLoginErr error
}

func (u *UsersRepoFake) CreateUser(_ context.Context, email, passwordHash, role string) (string, error) {
	if u.CreateUserErr != nil {
		return "", u.CreateUserErr
	}
	u.lastId = uuid.New().String()
	u.users = append(u.users, models.User{
		ID:           u.lastId,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	})
	return u.lastId, nil
}

func (u *UsersRepoFake) UserByID(_ context.Context, userId string) (models.User, error) {
	if u.UserByIdErr != nil {
		return models.User{}, u.UserByIdErr
	}

	user, ok := lo.Find(u.users, func(item models.User) bool {
		return item.ID == userId
	})
	if !ok {
		return user, postgres.ErrNotFound
	}

	return user, nil
}

func (u *UsersRepoFake) UserByEmail(_ context.Context, email string) (models.User, error) {
	if u.UserByLoginErr != nil {
		return models.User{}, u.UserByLoginErr
	}

	user, ok := lo.Find(u.users, func(item models.User) bool {
		return item.Email == email
	})
	if !ok {
		return user, postgres.ErrNotFound
	}

	return user, nil
}
