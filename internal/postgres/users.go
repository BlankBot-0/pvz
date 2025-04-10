package postgres

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/opentracing/opentracing-go"
	"pvz/internal/models"
)

type roUsers struct {
	query querier
}

func (ro *roUsers) UserById(ctx context.Context, userId string) (models.User, error) {
	const queryName = "UsersRepository/UserCoins"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
        select id, email, password_hash, role
        from users
        where id = $1`

	var user models.User
	if err := pgxscan.Get(ctx, ro.query, &user, q, userId); errIsNoRows(err) {
		return user, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return user, formatError(queryName, err)
	}

	return user, nil
}

func (ro *roUsers) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const queryName = "UsersRepository/UserByEmail"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
        select id, email, password_hash, role
        from users
        where email = $1`

	var user models.User
	if err := pgxscan.Get(ctx, ro.query, &user, q, email); errIsNoRows(err) {
		return user, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return user, formatError(queryName, err)
	}

	return user, nil
}

type rwUsers struct {
	ROUsers
	exec executor
}

func (rw *rwUsers) CreateUser(ctx context.Context, email, passwordHash, role string) (string, error) {
	const queryName = "UsersRepository/createUser"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into users(email, password_hash, role)
		values ($1, $2, $3)
		returning id`

	var id string
	if err := pgxscan.Get(ctx, rw.exec, &id, q, email, passwordHash, role); isUniqueViolated(err) {
		return "", formatError(queryName, ErrAlreadyExists)
	} else if err != nil {
		return "", formatError(queryName, err)
	}

	return id, nil
}
