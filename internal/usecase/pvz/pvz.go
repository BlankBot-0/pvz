package pvz

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"pvz/internal/models"
	"pvz/internal/postgres"
	"time"
)

type uuidGenerator interface {
	GenerateUuid(ctx context.Context) string
}

type Deps struct {
	Repo       postgres.DB
	UuidIssuer uuidGenerator
}

type PVZ struct {
	Deps Deps
}

func New(deps Deps) *PVZ {
	return &PVZ{
		Deps: deps,
	}
}

func (p *PVZ) ListPVZ(ctx context.Context, startDate, endDate time.Time, page, limit uint32) ([]models.PVZInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/ListPVZPaginated")
	defer span.Finish()

	offset := (page - 1) * limit
	pvzs, err := p.Deps.Repo.ROPvz().ListPVZPaginated(ctx, startDate, endDate, offset, limit)
	if errors.Is(err, postgres.ErrNotFound) {
		return nil, fmt.Errorf("list pvz: no items with such parameters: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("list pvz: %w", err)
	}

	pvzInfos := make([]models.PVZInfo, len(pvzs))
	var pvzInfosMap map[string]*models.PVZInfo
	pvzIds := make([]string, len(pvzs))
	for i, pvz := range pvzs {
		pvzInfos[i] = models.PVZInfo{
			PVZ:        pvz,
			Receptions: nil,
		}
		pvzInfosMap[pvz.Id] = &pvzInfos[i]
		pvzIds[i] = pvz.Id
	}

	receptions, err := p.Deps.Repo.ROPvz().ListReceptionsByPVZId(ctx, pvzIds)
	if err != nil {
		return nil, fmt.Errorf("list pvz receptions: %w", err)
	}

	receptionIds := make([]string, len(receptions))
	receptionInfos := make([]models.ReceptionInfo, len(receptions))
	var receptionInfosMap map[string]*models.ReceptionInfo
	for i, reception := range receptions {
		receptionInfo := models.ReceptionInfo{
			Reception: reception,
			Products:  nil,
		}
		receptionInfos[i] = receptionInfo
		receptionInfosMap[reception.Id] = &receptionInfos[len(receptionInfos)-1]
		receptionIds[i] = reception.Id
	}

	products, err := p.Deps.Repo.ROPvz().ListProductsByReceptionId(ctx, receptionIds)
	if err != nil {
		return nil, fmt.Errorf("list pvz products: %w", err)
	}

	// building output
	for _, product := range products {
		receptionInfosMap[product.ReceptionId].Products = append(
			receptionInfosMap[product.ReceptionId].Products,
			product,
		)
	}
	for _, receptionInfo := range receptionInfos {
		pvzInfosMap[receptionInfo.Reception.PvzId].Receptions = append(
			pvzInfosMap[receptionInfo.Reception.PvzId].Receptions,
			receptionInfo,
		)
	}

	return pvzInfos, nil
}
func (p *PVZ) GetPVZList(ctx context.Context) ([]models.PVZ, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/ListPVZ")
	defer span.Finish()

	pvzs, err := p.Deps.Repo.ROPvz().ListPVZ(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pvz list: %w", err)
	}

	return pvzs, nil
}
func (p *PVZ) CreatePVZ(ctx context.Context, id, city string, registrationDate *time.Time) (models.PVZ, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreatePVZ")
	defer span.Finish()

	validCity, err := p.Deps.Repo.ROPvz().CheckExistsCity(ctx, city)
	if err != nil {
		return models.PVZ{}, fmt.Errorf("create pvz: %w", err)
	} else if !validCity {
		return models.PVZ{}, ErrCityNotFound
	}

	if len(id) == 0 {
		id = p.Deps.UuidIssuer.Uuid(ctx)
	}
	if registrationDate == nil {
		t := time.Now()
		registrationDate = &t
	}
	pvz, err := p.Deps.Repo.RWPvz().AddPVZ(ctx, id, city, *registrationDate)
	if err != nil {
		return pvz, fmt.Errorf("create pvz: %w", err)
	}

	return pvz, nil
}

func (p *PVZ) CreateReception(ctx context.Context, pvzId string) (models.Reception, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreateReception")
	defer span.Finish()

	var reception models.Reception
	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("create reception: %w", err)
		}

		lastReception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzId)
		if err != nil && !errors.Is(err, postgres.ErrNotFound) {
			return fmt.Errorf("create reception: %w", err)
		}
		if lastReception.ReceptionStatus == models.ReceptionStatusInProgress {
			return ErrAnotherReceptionInProgress
		}

		reception, err = tx.RWPvz().AddReception(ctx, pvzId, models.ReceptionStatusInProgress)
		if err != nil {
			return fmt.Errorf("create reception: %w", err)
		}
		return nil
	}, pgx.Serializable)

	return reception, err
}
func (p *PVZ) CloseLastReception(ctx context.Context, pvzId string) (models.Reception, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CloseLastReception")
	defer span.Finish()

	var reception models.Reception
	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("close last reception: %w", err)
		}

		lastReception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrNoReceptionsInProgress
		} else if err != nil {
			return fmt.Errorf("close last reception: %w", err)
		}
		if lastReception.ReceptionStatus == models.ReceptionStatusClosed {
			return ErrNoReceptionsInProgress
		}

		reception, err = tx.RWPvz().UpdateReception(ctx, lastReception.Id, models.ReceptionStatusClosed)
		if err != nil {
			return fmt.Errorf("close last reception: %w", err)
		}
		return nil
	}, pgx.Serializable)

	return reception, err
}

func (p *PVZ) CreateProduct(ctx context.Context, productType, pvzId string) (models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/CreateProduct")
	defer span.Finish()

	existsProductType, err := p.Deps.Repo.ROPvz().CheckExistsProductType(ctx, productType)
	if err != nil {
		return models.Product{}, fmt.Errorf("create product: %w", err)
	} else if !existsProductType {
		return models.Product{}, ErrProductTypeNotFound
	}

	var product models.Product
	err = p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("create product: %w", err)
		}

		reception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrNoReceptionsInProgress
		} else if err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		if reception.ReceptionStatus == models.ReceptionStatusClosed {
			return ErrNoReceptionsInProgress
		}

		product, err = tx.RWPvz().AddProductToReception(ctx, productType, pvzId)
		if err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		return nil
	}, pgx.Serializable)

	return product, err
}
func (p *PVZ) DeleteLastProduct(ctx context.Context, pvzId string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PVZ/DeleteLastProduct")
	defer span.Finish()

	err := p.Deps.Repo.RunInTx(ctx, func(tx postgres.RepositoryProvider) error {
		_, err := tx.ROPvz().GetPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrPVZNotFound
		} else if err != nil {
			return fmt.Errorf("delete last product: %w", err)
		}

		reception, err := tx.ROPvz().GetLastReceptionByPVZ(ctx, pvzId)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrNoReceptionsInProgress
		} else if err != nil {
			return fmt.Errorf("delete last product: %w", err)
		}
		if reception.ReceptionStatus == models.ReceptionStatusClosed {
			return ErrNoReceptionsInProgress
		}

		product, err := tx.ROPvz().GetLastProductByReception(ctx, reception.Id)
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrNoProducts
		} else if err != nil {
			return fmt.Errorf("delete last product: %w", err)
		}

		err = tx.RWPvz().DeleteProduct(ctx, product.Id)
		if err != nil {
			return fmt.Errorf("delete last product: %w", err)
		}
		return nil
	}, pgx.Serializable)

	if err != nil {
		return "last product not deleted", err
	}
	return "last product deleted", err
}

var (
	minimumDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	maximumDate = time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC)
)
