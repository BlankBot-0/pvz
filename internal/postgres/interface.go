package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"pvz/internal/models"
	"time"
)

var _ DB = (*Database)(nil)

type (
	// ROPVZ is a read-only repository
	ROPVZ interface {
		ListPVZPaginated(ctx context.Context, startDate time.Time, endDate time.Time, offset uint32, limit uint32) ([]models.PVZ, error)
		ListPVZ(ctx context.Context) ([]models.PVZ, error)
		ListReceptionsByPVZ(ctx context.Context, pvzIds []string) ([]models.Reception, error)
		GetLastReceptionByPVZ(ctx context.Context, pvzId string) (models.Reception, error)
		ListProductsByReception(ctx context.Context, receptionIds []string) ([]models.Product, error)
		GetPVZ(ctx context.Context, pvzId string) (models.PVZ, error)
	}

	// RWPVZ is a read-write repository
	RWPVZ interface {
		AddPVZ(ctx context.Context, id string, city string) (*models.PVZ, error)
		AddPVZWIthDate(ctx context.Context, id string, city string, registrationDate time.Time) (*models.PVZ, error)
		AddReception(ctx context.Context, pvzId string, status string) (models.Reception, error)
		CloseLastReceptionByPVZ(ctx context.Context, pvzId string) (models.Reception, error)
		AddProductToReception(ctx context.Context, productType string, receptionId string) (models.Product, error)
		DeleteLastProduct(ctx context.Context, pvzId string) error
		ROPVZ
	}

	ROUsers interface {
		UserByID(ctx context.Context, userId string) (models.User, error)
		UserByEmail(ctx context.Context, email string) (models.User, error)
		ValidateRole(ctx context.Context, role string) error
	}

	RWUsers interface {
		CreateUser(ctx context.Context, email, passwordHash, role string) (string, error)
		ROUsers
	}

	RepositoryProvider interface {
		ROPvz() ROPVZ
		RWPvz() RWPVZ
		ROUsers() ROUsers
		RWUsers() RWUsers
	}

	DB interface {
		RepositoryProvider
		RunInTx(ctx context.Context, f func(tx RepositoryProvider) error, isoLevel pgx.TxIsoLevel) error
	}
)
