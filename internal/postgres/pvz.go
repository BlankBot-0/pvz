package postgres

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/opentracing/opentracing-go"
	"pvz/internal/models"
	"time"
)

type roPVZ struct {
	query querier
}

func (ro *roPVZ) ListPVZPaginated(ctx context.Context, startDate, endDate time.Time, offset, limit uint32) ([]models.PVZ, error) {
	const queryName = "PVZRepository/ListPVZPaginated"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select p.id, p.registration_date, c.name
		from pvzs p
		join cities c on p.city_id = c.id
		join receptions r on r.pvz_id = p.id
		where r.date >= $1 and r.date <= $2
		limit $3 offset $4`

	var pvzs []models.PVZ
	if err := pgxscan.Select(ctx, ro.query, pvzs, q, startDate, endDate, limit, offset); err != nil {
		return pvzs, handleError(queryName, err)
	}
	return pvzs, nil
}

func (ro *roPVZ) ListPVZ(ctx context.Context) ([]models.PVZ, error) {
	const queryName = "PVZRepository/ListPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select p.id, p.registration_date, c.name
		from pvzs p
		join cities c on p.city_id = c.id
		`

	var pvzs []models.PVZ
	if err := pgxscan.Select(ctx, ro.query, pvzs, q); err != nil {
		return pvzs, handleError(queryName, err)
	}
	return pvzs, nil
}

func (ro *roPVZ) ListReceptionsByPVZ(ctx context.Context, pvzIds []string) ([]models.Reception, error) {
	const queryName = "PVZRepository/ListReceptionsByPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, date, status, pvz_id
		from receptions
		where pvz_id in ($1)`

	var receptions []models.Reception
	if err := pgxscan.Select(ctx, ro.query, receptions, q, pvzIds); err != nil {
		return nil, handleError(queryName, err)
	}
	return receptions, nil
}

func (ro *roPVZ) GetLastReceptionByPVZ(ctx context.Context, pvzId string) (models.Reception, error) {
	const queryName = "PVZRepository/GetLastReceptionByPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, date, status, pvz_id
		from receptions
		where pvz_id = $1
		order by date desc
		limit 1`

	var reception models.Reception
	if err := pgxscan.Get(ctx, ro.query, &reception, q, pvzId); err != nil {
		return reception, handleError(queryName, err)
	}
	return reception, nil
}

func (ro *roPVZ) ListProductsByReception(ctx context.Context, receptionIds []string) ([]models.Product, error) {
	const queryName = "PVZRepository/ListProductsByReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select products.id, products.date, product_types.name, products.reception_id
		from products
		join product_types on products.type_id = product_types.id
		where reception_id in ($1)`

	var products []models.Product
	if err := pgxscan.Select(ctx, ro.query, products, q, receptionIds); err != nil {
		return nil, handleError(queryName, err)
	}
	return products, nil
}

func (ro *roPVZ) GetPVZ(ctx context.Context, pvzId string) (models.PVZ, error) {
	const queryName = "PVZRepository/GetPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select pvzs.id, pvzs.registration_date, cities.name
		from pvzs
		join cities on pvzs.city_id = cities.id
		where pvzs.id = $1`

	var pvz models.PVZ
	if err := pgxscan.Get(ctx, ro.query, q, pvzId); err != nil {
		return pvz, handleError(queryName, err)
	}
	return pvz, nil
}

type rwPVZ struct {
	ROPVZ
	exec executor
}

func (rw *rwPVZ) AddPVZ(ctx context.Context, id string, city string) (*models.PVZ, error) {
	const queryName = "PVZRepository/AddPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into pvzs (id, city_id)
		values ($1, coalesce((select id from cities where name = $2), -1))
		returning registration_date`

	var registrationDate time.Time
	if err := pgxscan.Get(ctx, rw.exec, &registrationDate, q, id, city); err != nil {
		return nil, handleError(queryName, err)
	}

	return &models.PVZ{
		ID:               id,
		RegistrationDate: registrationDate,
		City:             city,
	}, nil
}

func (rw *rwPVZ) AddPVZWIthDate(ctx context.Context, id string, city string, registrationDate time.Time) (*models.PVZ, error) {
	const queryName = "PVZRepository/AddPVZWIthDate"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into pvzs (id, city_id, registration_date)
		values ($1, coalesce((select id from cities where name = $2), -1), $3)`

	if _, err := rw.exec.Exec(ctx, q, id, city, registrationDate); err != nil {
		return nil, handleError(queryName, err)
	}

	return &models.PVZ{
		ID:               id,
		RegistrationDate: registrationDate,
		City:             city,
	}, nil
}

func (rw *rwPVZ) AddReception(ctx context.Context, pvzId, status string) (models.Reception, error) {
	const queryName = "PVZRepository/AddReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into receptions (id, pvz_id, status)
		values (gen_random_uuid(), $1, $2)
		returning id, date`

	reception := models.Reception{PvzID: pvzId, ReceptionStatus: status}
	if err := pgxscan.Get(ctx, rw.exec, &reception, q, pvzId, status); err != nil {
		return reception, handleError(queryName, err)
	}
	return reception, nil
}

func (rw *rwPVZ) CloseLastReceptionByPVZ(ctx context.Context, pvzId string) error {
	const queryName = "PVZRepository/CloseLastReceptionByPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
        update receptions set status = 'closed'
        where pvz_id = $1 and status == 'in_progress'`

	if tag, err := rw.exec.Exec(ctx, q, pvzId); err != nil {
		return handleError(queryName, err)
	} else if err = ensureRowIsAffected(tag); err != nil {
		return handleError(queryName, err)
	}
	return nil
}

func (rw *rwPVZ) AddProductToReception(ctx context.Context, productType, receptionId string) (models.Product, error) {
	const queryName = "PVZRepository/AddProductToReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into products (id, type_id, reception_id)
		values (gen_random_uuid(), coalesce((select id from cities where name = $1), -1), $2)
		returning id, date`

	product := models.Product{
		Type:        productType,
		ReceptionID: receptionId,
	}
	if err := pgxscan.Get(ctx, rw.exec, &product, q, productType, receptionId); err != nil {
		return product, handleError(queryName, err)
	}
	return product, nil
}

func (rw *rwPVZ) DeleteLastProduct(ctx context.Context, pvzId string) error {
	const queryName = "PVZRepository/DeleteLastProduct"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		delete from products
		where id = (
			select p.id from products p
			join receptions on receptions.id = products.reception_id
		    where pvz_id = $1 and status = 'in_progress'
		    order by p.date desc
		    limit 1)`

	if tag, err := rw.exec.Exec(ctx, q, pvzId); err != nil {
		return handleError(queryName, err)
	} else if err = ensureRowIsAffected(tag); err != nil {
		return err
	}
	return nil
}
