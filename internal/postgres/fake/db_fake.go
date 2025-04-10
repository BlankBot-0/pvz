package fake

import (
	"context"
	"github.com/jackc/pgx/v5"
	"pvz/internal/postgres"
)

var _ postgres.DB = &DBMock{}

// DBMock is a type that implements DBInterface and aggregates repository mocks.
type DBMock struct {
	*UsersRepoFake
	*PVZRepoFake
}

func (d DBMock) ROPvz() postgres.ROPVZ {
	return d.PVZRepoFake
}

func (d DBMock) RWPvz() postgres.RWPVZ {
	return d.PVZRepoFake
}

func (d DBMock) ROUsers() postgres.ROUsers {
	return d.UsersRepoFake
}

func (d DBMock) RWUsers() postgres.RWUsers {
	return d.UsersRepoFake
}

func (d DBMock) RunInTx(_ context.Context, f func(tx postgres.RepositoryProvider) error, isoLevel pgx.TxIsoLevel) error {
	return f(d)
}
