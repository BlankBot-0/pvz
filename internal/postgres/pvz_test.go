//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
)

const defaultCity = "Москва"

const defaultPvzID = "1"

const defaultProductType = "обувь"

// filled after reception creation
var receptionID string

func TestPVZRepozitory(t *testing.T) {
	t.Parallel()

	t.Run("AddPVZ", testAddPVZ)

	t.Run("AddReception", testAddReception)

	t.Run("AddProduct", testAddProduct)

	t.Run("DeleteProduct", testDeleteProduct)

	t.Run("CloseLastReception", testCloseLastReception)

	t.Run("AssertPVZ", testAssertPVZ)
}

func testAddPVZ(t *testing.T) {
	ctx := context.Background()

	repo := db.RWPvz()

	pvz, err := repo.AddPVZ(ctx, defaultPvzID, defaultCity)
	if err != nil {
		t.Fatalf("failed to create pvz: %s", err)
	}

	if pvz.ID != defaultPvzID {
		t.Fatalf("unexpected ID: %s", pvz.ID)
	}
	if pvz.City != defaultCity {
		t.Fatalf("unexpected city: %s", pvz.City)
	}

	_, err = repo.AddPVZ(ctx, defaultPvzID, defaultCity)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected unique violation, got: %s", err)
	}
}

func testAddReception(t *testing.T) {
	ctx := context.Background()

	repo := db.RWPvz()

	reception, err := repo.AddReception(ctx, defaultPvzID, "in_progress")
	if err != nil {
		t.Fatalf("failed to create reception: %s", err)
	}

	receptionID = reception.ID
}

func testAddProduct(t *testing.T) {
	ctx := context.Background()

	repo := db.RWPvz()

	_, err := repo.AddProductToReception(ctx, defaultProductType, receptionID)
	if err != nil {
		t.Fatalf("failed to add product: %s", err)
	}

	_, err = repo.AddProductToReception(ctx, "голубцы2", receptionID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expecter foreign key violation: %s", err)
	}
}

func testDeleteProduct(t *testing.T) {
	ctx := context.Background()

	repo := db.RWPvz()

	err := repo.DeleteLastProduct(ctx, defaultPvzID)
	if err != nil {
		t.Fatalf("failed to delete product: %s", err)
	}
}

func testCloseLastReception(t *testing.T) {
	ctx := context.Background()

	repo := db.RWPvz()

	lastReception, err := repo.GetLastReceptionByPVZ(ctx, defaultPvzID)
	if err != nil {
		t.Fatalf("got unexpecter error trying to get last reception: %s", err)
	}

	reception, err := repo.CloseLastReceptionByPVZ(ctx, defaultPvzID)
	if err != nil {
		t.Fatalf("failed to close last reception: %s", err)
	}

	if reception.ID != lastReception.ID {
		t.Fatalf("last reception was expected to be closed")
	} else if reception.ReceptionStatus != "closed" {
		t.Fatalf("reception is expected to be closed")
	}
}

func testAssertPVZ(t *testing.T) {
	ctx := context.Background()

	repo := db.ROPvz()

	pvz, err := repo.GetPVZ(ctx, defaultPvzID)
	if err != nil {
		t.Fatalf("failed to get pvz: %s", err)
	}

	if pvz.ID != defaultPvzID {
		t.Fatal("got unexpected pvz")
	}

	receptions, err := repo.ListReceptionsByPVZ(ctx, []string{pvz.ID})
	if err != nil {
		t.Fatal("failed to list receptions")
	}

	if len(receptions) != 1 {
		t.Fatal("expected only one reception")
	}
}
