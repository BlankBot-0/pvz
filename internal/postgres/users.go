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

func (ro *roUsers) UserByID(ctx context.Context, userId string) (models.User, error) {
	const queryName = "UsersRepository/UserCoins"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
        select users.id, users.email, users.password_hash, roles.name
        from users
        join roles on roles.id = users.role_id
        where users.id = $1`

	var user models.User
	if err := pgxscan.Get(ctx, ro.query, &user, q, userId); errIsNoRows(err) {
		return user, handleError(queryName, ErrNotFound)
	} else if err != nil {
		return user, handleError(queryName, err)
	}

	return user, nil
}

func (ro *roUsers) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const queryName = "UsersRepository/UserByEmail"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
        select users.id, users.email, users.password_hash, roles.name as role
        from users
        join roles on roles.id = users.role_id
        where email = $1`

	var user models.User
	if err := pgxscan.Get(ctx, ro.query, &user, q, email); errIsNoRows(err) {
		return user, handleError(queryName, ErrNotFound)
	} else if err != nil {
		return user, handleError(queryName, err)
	}

	return user, nil
}

func (ro *roUsers) ValidateRole(ctx context.Context, role string) error {
	const queryName = "UsersRepository/ValidateRole"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select exists(select 1 from roles where name = $1)`

	var exists bool
	err := pgxscan.Get(ctx, ro.query, &exists, q, role)
	if err != nil {
		return handleError(queryName, err)
	} else if !exists {
		return ErrNotFound
	}
	return nil
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
		insert into users(id, email, password_hash, role_id)
		values (gen_random_uuid(), $1, $2, coalesce((select id from roles where name = $3), -1))
		returning id`

	var id string
	if err := pgxscan.Get(ctx, rw.exec, &id, q, email, passwordHash, role); isUniqueViolated(err) {
		return "", handleError(queryName, ErrAlreadyExists)
	} else if isForeignKeyViolated(err) {
		return "", handleError(queryName, ErrInvalidReference)
	} else if err != nil {
		return "", handleError(queryName, err)
	}

	return id, nil
}
