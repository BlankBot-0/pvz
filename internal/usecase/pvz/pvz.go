package pvz

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/samber/lo"
	"pvz/internal/models"
	"pvz/internal/postgres"
	"time"
)

type uuidGenerator interface {
	GenerateUUID(ctx context.Context) string
}

type Deps struct {
	Repo          postgres.DB
	UUIDGenerator uuidGenerator
}

type PVZ struct {
	Deps Deps
}

func New(deps Deps) *PVZ {
	return &PVZ{
		Deps: deps,
	}
}

func (p *PVZ) ListPVZPaginated(ctx context.Context, startDate, endDate time.Time, page, limit uint32) ([]models.PVZInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/ListPVZPaginated")
	defer span.Finish()

	offset := (page - 1) * limit
	pvzs, err := p.Deps.Repo.ROPvz().ListPVZPaginated(ctx, startDate, endDate, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list pvz: %w", err)
	}

	pvzInfos := make([]models.PVZInfo, len(pvzs))
	pvzIDs := make([]string, len(pvzs))
	for i, pvz := range pvzs {
		pvzInfos[i] = models.PVZInfo{PVZ: pvz}
		pvzIDs[i] = pvz.ID
	}

	receptions, err := p.Deps.Repo.ROPvz().ListReceptionsByPVZ(ctx, pvzIDs)
	if err != nil {
		return nil, fmt.Errorf("list pvz receptions: %w", err)
	}

	receptionIDs := make([]string, len(receptions))
	receptionByPvz := make(map[string][]models.ReceptionInfo, len(receptions))
	for i, reception := range receptions {
		receptionByPvz[reception.PvzID] = append(receptionByPvz[reception.PvzID], models.ReceptionInfo{
			Reception: reception,
		})
		receptionIDs[i] = reception.ID
	}

	products, err := p.Deps.Repo.ROPvz().ListProductsByReception(ctx, receptionIDs)
	if err != nil {
		return nil, fmt.Errorf("list pvz products: %w", err)
	}

	productsByReception := lo.GroupBy(products, func(product models.Product) string {
		return product.ReceptionID
	})

	for i := range pvzInfos {
		pvzID := pvzInfos[i].PVZ.ID
		pvzReceptions := receptionByPvz[pvzID]
		for _, reception := range pvzReceptions {
			reception.Products = productsByReception[reception.Reception.ID]
		}

		pvzInfos[i].Receptions = pvzReceptions
	}

	return pvzInfos, nil
}
func (p *PVZ) ListPVZ(ctx context.Context) ([]models.PVZ, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/ListPVZ")
	defer span.Finish()

	pvzs, err := p.Deps.Repo.ROPvz().ListPVZ(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list pvz: %w", err)
	}

	return pvzs, nil
}

func (p *PVZ) CreatePVZ(ctx context.Context, city string, id *string, registrationDate *time.Time) (*models.PVZ, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreatePVZ")
	defer span.Finish()

	var pvzID string
	if id != nil {
		pvzID = *id
	} else {
		pvzID = p.Deps.UUIDGenerator.GenerateUUID(ctx)
	}

	var pvz *models.PVZ
	var err error
	if registrationDate != nil {
		pvz, err = p.Deps.Repo.RWPvz().AddPVZWIthDate(ctx, pvzID, city, *registrationDate)
	} else {
		pvz, err = p.Deps.Repo.RWPvz().AddPVZ(ctx, pvzID, city)
	}

	if errors.Is(err, postgres.ErrAlreadyExists) {
		return nil, ErrIDIsOccupied
	} else if errors.Is(err, postgres.ErrInvalidReference) {
		return nil, ErrCityNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to create pvz: %w", err)
	}

	return pvz, nil
}

func (p *PVZ) CreateReception(ctx context.Context, pvzID string) (models.Reception, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreateReception")
	defer span.Finish()

	var reception models.Reception
	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzID)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("failed to resolve pvz: %w", err)
		}

		lastReception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzID)
		if err != nil && !errors.Is(err, postgres.ErrNotFound) {
			return fmt.Errorf("failed to pick last reception: %w", err)
		}
		if lastReception.ReceptionStatus == models.ReceptionStatusInProgress {
			return ErrAnotherReceptionInProgress
		}

		reception, err = tx.RWPvz().AddReception(ctx, pvzID, models.ReceptionStatusInProgress)
		if err != nil {
			return fmt.Errorf("failed to create reception: %w", err)
		}
		return nil
	}, pgx.Serializable)

	return reception, err
}
func (p *PVZ) CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CloseLastReception")
	defer span.Finish()

	var reception models.Reception
	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzID)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("failed to resolve pvz: %w", err)
		}

		if err := tx.RWPvz().CloseLastReceptionByPVZ(ctx, pvzID); errors.Is(err, postgres.ErrNotChanged) {
			return ErrNoReceptionsInProgress
		} else if err != nil {
			return fmt.Errorf("failed to close last reception: %w", err)
		}

		return nil
	}, pgx.Serializable)

	return reception, err
}

func (p *PVZ) CreateProduct(ctx context.Context, productType, pvzID string) (models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreateProduct")
	defer span.Finish()

	var product models.Product
	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		if _, err := tx.ROPvz().GetPVZ(ctx, pvzID); errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("failed to resolve pvz: %w", err)
		}

		reception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzID)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrNoReceptionsInProgress
		} else if err != nil {
			return fmt.Errorf("failed to get last reception: %w", err)
		} else if reception.ReceptionStatus == models.ReceptionStatusClosed {
			return ErrNoReceptionsInProgress
		}

		product, err = tx.RWPvz().AddProductToReception(ctx, productType, pvzID)
		if errors.Is(err, postgres.ErrInvalidReference) {
			return ErrProductTypeNotFound
		} else if err != nil {
			return fmt.Errorf("failed to create product in reception: %w", err)
		}
		return nil
	}, pgx.Serializable)

	return product, err
}
func (p *PVZ) DeleteLastProduct(ctx context.Context, pvzID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/DeleteLastProduct")
	defer span.Finish()

	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		if _, err := tx.ROPvz().GetPVZ(ctx, pvzID); errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("failed to resolve pvz: %w", err)
		}

		if err := tx.RWPvz().DeleteLastProduct(ctx, pvzID); errors.Is(err, postgres.ErrNotChanged) {
			return ErrNoProducts
		} else if err != nil {
			return fmt.Errorf("failed to delete last product: %w", err)
		}

		return nil
	}, pgx.Serializable)

	return err
}

var (
	minimumDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	maximumDate = time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC)
)
