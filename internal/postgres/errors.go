package postgres

import (
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotChanged       = errors.New("entity is not changed")
	ErrNotFound         = errors.New("entity not found")
	ErrAlreadyExists    = errors.New("entity already exists")
	ErrInvalidReference = errors.New("entity has invalid reference")
)

func handleError(queryName string, err error) error {
	if err == nil {
		return nil
	} else if errIsNoRows(err) {
		return ErrNotFound
	} else if isUniqueViolated(err) {
		return ErrAlreadyExists
	} else if isForeignKeyViolated(err) {
		return ErrInvalidReference
	}
	return fmt.Errorf("executing %s: %w", queryName, err)
}

func ensureRowIsAffected(tag pgconn.CommandTag) error {
	if tag.RowsAffected() == 1 {
		return nil
	}
	return ErrNotChanged
}

func errIsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isUniqueViolated(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func isForeignKeyViolated(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation
}
