package merch_store

import (
	"context"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
	"time"
)

type (
	pvzService interface {
		ListPVZPaginated(ctx context.Context, startDate time.Time, endDate time.Time, page uint32, limit uint32) ([]models.PVZInfo, error)
		ListPVZ(ctx context.Context) ([]models.PVZ, error)
		CreatePVZ(ctx context.Context, city string, id *string, registrationDate *time.Time) (*models.PVZ, error)

		CreateReception(ctx context.Context, pvzID string) (models.Reception, error)
		CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error)

		CreateProduct(ctx context.Context, productType string, pvzID string) (models.Product, error)
		DeleteLastProduct(ctx context.Context, pvzID string) error
	}
	authService interface {
		DummyLogin(ctx context.Context, role string) (string, error)
		Register(ctx context.Context, email, password, role string) (models.User, error)
		UserToken(ctx context.Context, email, password string) (string, error)
	}
)

type ListPvzParams struct {
	StartDate int64
	EndDate   int64
	Page      uint32
	Limit     uint32
}

type Deps struct {
	PVZ  pvzService
	Auth authService
}

type Service struct {
	pvzpb.PVZServiceServer
	Deps
}

func NewService(deps Deps) *Service {
	return &Service{
		Deps: deps,
	}
}
