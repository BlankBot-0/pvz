package pvz

import (
	"context"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
	"time"
)

type (
	PVZService interface {
		ListPVZPaginated(ctx context.Context, startDate time.Time, endDate time.Time, page uint32, limit uint32) ([]models.PVZInfo, error)
		ListPVZ(ctx context.Context) ([]models.PVZ, error)
		CreatePVZ(ctx context.Context, city string, id *string, registrationDate *time.Time) (*models.PVZ, error)

		CreateReception(ctx context.Context, pvzID string) (models.Reception, error)
		CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error)

		CreateProduct(ctx context.Context, productType string, pvzID string) (models.Product, error)
		DeleteLastProduct(ctx context.Context, pvzID string) error
	}
	AuthService interface {
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
	PVZ  PVZService
	Auth AuthService
}

type Service struct {
	pvzpb.UnimplementedPVZServiceServer
	Deps

	datetimeFormat string
}

func NewService(datetimeFormat string, deps Deps) *Service {
	return &Service{
		Deps:           deps,
		datetimeFormat: datetimeFormat,
	}
}

const (
	employeeRole  = "employee"
	moderatorRole = "moderator"
)
