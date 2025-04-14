package mocks

import (
	"context"
	"github.com/jackc/pgx/v5"
	"pvz/internal/postgres"
	"testing"
)

var _ postgres.DB = &DBMock{}

// DBMock is a type that implements DBInterface and aggregates repository mocks.
type DBMock struct {
	*PVZMock
	*UserMock
}

func NewDBMock(t *testing.T) DBMock {
	return DBMock{
		PVZMock:  NewPVZMock(t),
		UserMock: NewUserMock(t),
	}
}

func (d DBMock) ROPvz() postgres.ROPVZ {
	return d.PVZMock
}

func (d DBMock) RWPvz() postgres.RWPVZ {
	return d.PVZMock
}

func (d DBMock) ROUsers() postgres.ROUsers {
	return d.UserMock
}

func (d DBMock) RWUsers() postgres.RWUsers {
	return d.UserMock
}

func (d DBMock) RunInTx(_ context.Context, f func(tx postgres.RepositoryProvider) error, isoLevel pgx.TxIsoLevel) error {
	return f(d)
}
