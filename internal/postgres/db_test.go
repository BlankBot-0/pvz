//go:build integration

package postgres

import (
	"context"
	"os"
	"pvz/internal/logger"
	"testing"
)

var db *Database

func TestMain(m *testing.M) {
	const truncateQuery = `
    truncate
		pvzs,
		receptions,
		products,
		users
    cascade
  `

	ctx := context.Background()

	var err error
	db, err = Connect(ctx, os.Getenv("DSN"))
	if err != nil {
		logger.Fatalf("Connect() failed with err=%v", err)
	}

	exitCode := m.Run()

	_, err = db.pool.Exec(ctx, truncateQuery)
	if err != nil {
		logger.Fatalf("Truncation after tests failed with err=%v", err)
	}

	os.Exit(exitCode)
}
