package fake

import (
	"context"
	"github.com/google/uuid"
	"pvz/internal/models"
	"pvz/internal/postgres"
	"time"
)

var _ postgres.RWPVZ = &PVZRepoFake{}

type PVZRepoFake struct {
	Pvzs                  []models.PVZ
	PvzsById              map[string]models.PVZ
	ReceptionsByPVZId     map[string]map[string]models.Reception
	ReceptionsById        map[string]models.Reception
	ProductsByReceptionId map[string]map[string]models.Product
	ProductsById          map[string]models.Product

	LastReceptionIdByPVZId     map[string]string
	LastProductIdByReceptionId map[string]string

	Users map[string]models.User

	AddPVZErr                    error
	AddReceptionErr              error
	UpdateReceptionErr           error
	AddProductToReceptionErr     error
	DeleteProductErr             error
	ListPVZErr                   error
	GetPVZErr                    error
	ListReceptionByPVZIdErr      error
	GetLastReceptionByPVZErr     error
	ListProductsByReceptionIdErr error
	GetLastProductByReceptionErr error
}

func (p PVZRepoFake) GetPVZ(ctx context.Context, pvzId string) (models.PVZ, error) {
	if p.AddPVZErr != nil {
		return models.PVZ{}, p.AddPVZErr
	}

	return p.PvzsById[pvzId], nil
}

func (p PVZRepoFake) AddPVZ(_ context.Context, city string) (models.PVZ, error) {
	if p.AddPVZErr != nil {
		return models.PVZ{}, p.AddPVZErr
	}

	pvz := models.PVZ{
		ID:               uuid.New().String(),
		RegistrationDate: time.Now(),
		City:             city,
	}

	p.Pvzs = append(p.Pvzs, pvz)
	p.PvzsById[pvz.ID] = p.Pvzs[len(p.Pvzs)-1]
	return pvz, nil
}

func (p PVZRepoFake) AddReception(_ context.Context, pvzId string, status string) (models.Reception, error) {
	if p.AddReceptionErr != nil {
		return models.Reception{}, p.AddReceptionErr
	}

	reception := models.Reception{
		ID:              uuid.New().String(),
		DateTime:        time.Now(),
		PvzID:           pvzId,
		ReceptionStatus: status,
	}

	p.ReceptionsByPVZId[reception.PvzID][reception.ID] = reception
	p.ReceptionsById[reception.ID] = reception
	p.LastReceptionIdByPVZId[reception.PvzID] = reception.ID
	return reception, nil
}

func (p PVZRepoFake) UpdateReception(_ context.Context, receptionId string, status string) (models.Reception, error) {
	if p.UpdateReceptionErr != nil {
		return models.Reception{}, p.UpdateReceptionErr
	}

	r, ok := p.ReceptionsById[receptionId]
	if !ok {
		return models.Reception{}, postgres.ErrNotFound
	}

	r.ReceptionStatus = status
	p.ReceptionsById[receptionId] = r
	p.ReceptionsByPVZId[r.PvzID][r.ID] = r

	return r, nil
}

func (p PVZRepoFake) AddProductToReception(_ context.Context, productType string, receptionId string) (models.Product, error) {
	if p.AddProductToReceptionErr != nil {
		return models.Product{}, p.AddProductToReceptionErr
	}

	product := models.Product{
		ID:          uuid.New().String(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: receptionId,
	}

	p.ProductsByReceptionId[product.ReceptionID][product.ID] = product
	p.ProductsById[product.ID] = product
	p.LastProductIdByReceptionId[product.ReceptionID] = product.ID
	return product, nil
}

func (p PVZRepoFake) DeleteProduct(_ context.Context, productId string) error {
	if p.DeleteProductErr != nil {
		return p.DeleteProductErr
	}

	receptionId := p.ProductsById[productId].ReceptionID

	delete(p.ProductsByReceptionId[receptionId], productId)
	delete(p.ProductsById, productId)
	return nil
}

func (p PVZRepoFake) ListPVZ(_ context.Context, startDate time.Time, endDate time.Time, offset uint32, limit uint32) ([]models.PVZ, error) {
	if p.ListPVZErr != nil {
		return nil, p.ListPVZErr
	}

	return p.Pvzs, nil
}

func (p PVZRepoFake) ListReceptionsByPVZId(_ context.Context, pvzIds []string) ([]models.Reception, error) {
	if p.ListReceptionByPVZIdErr != nil {
		return nil, p.ListReceptionByPVZIdErr
	}

	receptionsByPVZId := make([]models.Reception, 0)
	for _, pvzId := range pvzIds {
		for _, reception := range p.ReceptionsByPVZId[pvzId] {
			receptionsByPVZId = append(receptionsByPVZId, reception)
		}
	}
	return receptionsByPVZId, nil
}

func (p PVZRepoFake) GetLastReceptionByPVZ(_ context.Context, pvzId string) (models.Reception, error) {
	if p.GetLastReceptionByPVZErr != nil {
		return models.Reception{}, p.GetLastReceptionByPVZErr
	}

	recId := p.LastReceptionIdByPVZId[pvzId]
	return p.ReceptionsById[recId], nil
}

func (p PVZRepoFake) ListProductsByReceptionId(_ context.Context, receptionIds []string) ([]models.Product, error) {
	if p.ListProductsByReceptionIdErr != nil {
		return nil, p.ListProductsByReceptionIdErr
	}

	products := make([]models.Product, 0)
	for _, receptionId := range receptionIds {
		for _, product := range p.ProductsByReceptionId[receptionId] {
			products = append(products, product)
		}
	}
	return products, nil
}

func (p PVZRepoFake) GetLastProductByReception(_ context.Context, receptionId string) (models.Product, error) {
	if p.GetLastProductByReceptionErr != nil {
		return models.Product{}, p.GetLastProductByReceptionErr
	}

	pId := p.LastProductIdByReceptionId[receptionId]
	return p.ProductsById[pId], nil
}
