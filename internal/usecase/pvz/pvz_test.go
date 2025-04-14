package pvz

import (
	"context"
	"github.com/go-test/deep"
	"github.com/samber/lo"
	"pvz/internal/models"
	"pvz/internal/postgres"
	pgmocks "pvz/internal/postgres/mocks"
	"testing"
	"time"
)

type dumbUUIDGenerator struct{}

func (dumbUUIDGenerator) GenerateUUID(_ context.Context) string {
	return lo.RandomString(10, lo.AlphanumericCharset)
}

type constUUIDGenerator struct {
	id string
}

func (c constUUIDGenerator) GenerateUUID(_ context.Context) string {
	return c.id
}

func TestPVZ_ListPVZPaginated_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	end := time.Now()
	start := end.Add(-time.Hour)

	dbMock.ListPVZPaginatedMock.Set(func(_ context.Context, _ time.Time, _ time.Time, _ uint32, _ uint32) (_ []models.PVZ, _ error) {
		return []models.PVZ{
			{
				ID:               "1",
				RegistrationDate: start,
				City:             "Москва",
			},
		}, nil
	})

	dbMock.ListReceptionsByPVZMock.Set(func(_ context.Context, _ []string) (_ []models.Reception, _ error) {
		return []models.Reception{
			{
				ID:              "10",
				DateTime:        start,
				PvzID:           "1",
				ReceptionStatus: "closed",
			},
		}, nil
	})

	dbMock.ListProductsByReceptionMock.Set(func(_ context.Context, _ []string) (_ []models.Product, _ error) {
		return []models.Product{
			{
				ID:          "100",
				DateTime:    start,
				Type:        "голубцы",
				ReceptionID: "10",
			},
		}, nil
	})

	got, err := uc.ListPVZPaginated(context.Background(), start, end, 1, 10)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	want := []models.PVZInfo{
		{
			PVZ: models.PVZ{
				ID:               "1",
				RegistrationDate: start,
				City:             "Москва",
			},
			Receptions: []models.ReceptionInfo{
				{
					Reception: models.Reception{
						ID:              "10",
						DateTime:        start,
						PvzID:           "1",
						ReceptionStatus: "closed",
					},
					Products: []models.Product{{
						ID:          "100",
						DateTime:    start,
						Type:        "голубцы",
						ReceptionID: "10"},
					},
				},
			},
		},
	}

	if diff := deep.Equal(got, want); len(diff) > 0 {
		t.Fatalf("got unexpected diff in list page: %s", diff)
	}
}

func TestPVZ_ListPVZ_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	dbMock.ListPVZMock.Set(func(_ context.Context) (_ []models.PVZ, _ error) {
		return []models.PVZ{
			{
				ID:               "1",
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
		}, nil
	})

	pvz, err := uc.ListPVZ(context.Background())
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if len(pvz) != 1 {
		t.Fatalf("expected only on PVZ")
	} else if pvz[0].ID != "1" {
		t.Fatalf("got unexpected PVZ")
	}
}

func TestPVZ_CreatePVZ_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: constUUIDGenerator{"1"},
	})

	now := time.Now()

	dbMock.AddPVZMock.Set(func(_ context.Context, _ string, _ string) (_ *models.PVZ, _ error) {
		return &models.PVZ{
			ID:               "1",
			RegistrationDate: now,
			City:             "Москва",
		}, nil
	})

	pvz, err := uc.CreatePVZ(context.Background(), "Москва", nil, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if pvz.ID != "1" {
		t.Fatalf("got unexpected pvz")
	}
}

func TestPVZ_CreatePVZ_Ok_WithCustomFields(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	now := time.Now()

	dbMock.AddPVZWIthDateMock.Set(func(_ context.Context, _ string, _ string, _ time.Time) (_ *models.PVZ, _ error) {
		return &models.PVZ{
			ID:               "1",
			RegistrationDate: now,
			City:             "Москва",
		}, nil
	})

	pvz, err := uc.CreatePVZ(context.Background(), "Москва", lo.ToPtr("1"), &now)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if pvz.ID != "1" {
		t.Fatalf("got unexpected pvz")
	}
}

func TestPVZ_CreateReception_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	now := time.Now()

	dbMock.GetPVZMock.Return(models.PVZ{
		ID:               "1",
		RegistrationDate: now,
		City:             "Москва",
	}, nil)

	dbMock.GetLastReceptionByPVZMock.Return(models.Reception{}, postgres.ErrNotFound)

	dbMock.AddReceptionMock.Return(models.Reception{
		ID:              "10",
		DateTime:        now,
		PvzID:           "1",
		ReceptionStatus: "in_progress",
	}, nil)

	reception, err := uc.CreateReception(context.Background(), "1")
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if reception.ID != "10" {
		t.Fatalf("got unexpected reception")
	}
}

func TestPVZ_CloseLastReception_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	now := time.Now()

	dbMock.GetPVZMock.Return(models.PVZ{
		ID:               "1",
		RegistrationDate: now,
		City:             "Москва",
	}, nil)

	dbMock.CloseLastReceptionByPVZMock.Return(models.Reception{
		ID:              "10",
		DateTime:        now,
		PvzID:           "1",
		ReceptionStatus: "closed",
	}, nil)

	reception, err := uc.CloseLastReception(context.Background(), "1")
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if reception.ID != "10" {
		t.Fatalf("got unexpected reception")
	}
}

func TestPVZ_CreateProduct_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	now := time.Now()

	dbMock.GetPVZMock.Return(models.PVZ{
		ID:               "1",
		RegistrationDate: now,
		City:             "Москва",
	}, nil)

	dbMock.GetLastReceptionByPVZMock.Return(models.Reception{
		ID:              "10",
		DateTime:        now,
		PvzID:           "1",
		ReceptionStatus: "in_progress",
	}, nil)

	dbMock.AddProductToReceptionMock.Return(models.Product{
		ID:          "100",
		DateTime:    now,
		Type:        "голубцы",
		ReceptionID: "10",
	}, nil)

	product, err := uc.CreateProduct(context.Background(), "голубцы", "1")
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if product.ID != "100" {
		t.Fatalf("got unexpected reception")
	}
}

func TestPVZ_DeleteLastProduct_Ok(t *testing.T) {
	dbMock := pgmocks.NewDBMock(t)

	uc := New(Deps{
		Repo:          dbMock,
		UUIDGenerator: dumbUUIDGenerator{},
	})

	now := time.Now()

	dbMock.GetPVZMock.Return(models.PVZ{
		ID:               "1",
		RegistrationDate: now,
		City:             "Москва",
	}, nil)

	dbMock.DeleteLastProductMock.Return(nil)

	err := uc.DeleteLastProduct(context.Background(), "1")
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}
}
