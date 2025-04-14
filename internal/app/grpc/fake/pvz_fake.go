package fake

import (
	"context"
	pvz "pvz/internal/app/grpc"
	"pvz/internal/models"
	"time"
)

var _ pvz.PVZService = &PVZFake{}

type PVZFake struct {
}

func (P PVZFake) ListPVZPaginated(ctx context.Context, startDate time.Time, endDate time.Time, page uint32, limit uint32) ([]models.PVZInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) ListPVZ(ctx context.Context) ([]models.PVZ, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) CreatePVZ(ctx context.Context, city string, id *string, registrationDate *time.Time) (*models.PVZ, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) CreateReception(ctx context.Context, pvzID string) (models.Reception, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) CreateProduct(ctx context.Context, productType string, pvzID string) (models.Product, error) {
	//TODO implement me
	panic("implement me")
}

func (P PVZFake) DeleteLastProduct(ctx context.Context, pvzID string) error {
	//TODO implement me
	panic("implement me")
}
