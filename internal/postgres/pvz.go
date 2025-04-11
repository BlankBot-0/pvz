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

func (ro *roPVZ) ListPVZ(ctx context.Context, startDate, endDate time.Time, offset, limit uint32) ([]models.PVZ, error) {
	const queryName = "PVZRepository/ListPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, registration_date, city
		from pvzs
		where registration_date >= $1 and registration_date <= $2
		limit $3 offset $4`

	var pvzs []models.PVZ
	if err := pgxscan.Select(ctx, ro.query, pvzs, q, startDate, endDate, limit, offset); err != nil {
		return pvzs, formatError(queryName, err)
	}

	return pvzs, nil
}

func (ro *roPVZ) ListReceptionsByPVZId(ctx context.Context, pvzIds []string) ([]models.Reception, error) {
	const queryName = "PVZRepository/ListReceptionsByPVZId"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, date, status, pvz_id
		from receptions
		where pvz_id in ($1)`

	var receptions []models.Reception
	if err := pgxscan.Select(ctx, ro.query, receptions, q, pvzIds); err != nil {
		return nil, formatError(queryName, err)
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
	if err := pgxscan.Get(ctx, ro.query, &reception, q, pvzId); errIsNoRows(err) {
		return reception, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return reception, formatError(queryName, err)
	}

	return reception, nil
}

func (ro *roPVZ) ListProductsByReceptionId(ctx context.Context, receptionIds []string) ([]models.Product, error) {
	const queryName = "PVZRepository/ListProductsByReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, date, type, reception_id
		from products
		where reception_id in ($1)`

	var products []models.Product
	if err := pgxscan.Select(ctx, ro.query, products, q, receptionIds); err != nil {
		return nil, formatError(queryName, err)
	}

	return products, nil
}

func (ro *roPVZ) GetLastProductByReception(ctx context.Context, receptionId string) (models.Product, error) {
	const queryName = "PVZRepository/GetLastProductByReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, date, type, reception_id
		from products
		where reception_id = $1
		order by date desc
		limit 1`

	var product models.Product
	if err := pgxscan.Get(ctx, ro.query, &product, q, receptionId); errIsNoRows(err) {
		return product, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return product, formatError(queryName, err)
	}

	return product, nil
}

func (ro *roPVZ) GetPVZ(ctx context.Context, pvzId string) (models.PVZ, error) {
	const queryName = "PVZRepository/GetPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		select id, registration_date, city
		from pvzs
		where id = $1`

	var pvz models.PVZ
	if err := pgxscan.Get(ctx, ro.query, q, pvzId); errIsNoRows(err) {
		return pvz, formatError(queryName, ErrNotFound)
	} else if err != nil {
		return pvz, formatError(queryName, err)
	}

	return pvz, nil
}

type rwPVZ struct {
	ROPVZ
	exec executor
}

func (rw *rwPVZ) AddPVZ(ctx context.Context, city string) (models.PVZ, error) {
	const queryName = "PVZRepository/AddPVZ"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into pvzs (id, city)
		values (gen_random_uuid(), $1)
		returning id, registration_date, city`

	var pvz models.PVZ
	if err := pgxscan.Get(ctx, rw.exec, pvz, q, city); err != nil {
		return pvz, formatError(queryName, err)
	}

	return pvz, nil
}

func (rw *rwPVZ) AddReception(ctx context.Context, pvzId, status string) (models.Reception, error) {
	const queryName = "PVZRepository/AddReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into receptions (id, pvz_id, status)
		values (gen_random_uuid(), $1, $2)
		returning id, date, status, pvz_id`

	var reception models.Reception
	if err := pgxscan.Get(ctx, rw.exec, &reception, q, pvzId, status); err != nil {
		return reception, formatError(queryName, err)
	}

	return reception, nil
}

func (rw *rwPVZ) UpdateReception(ctx context.Context, receptionId, status string) (models.Reception, error) {
	const queryName = "PVZRepository/UpdateReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		update receptions set status = $2
		where id = $1
		returning id, date, status, pvz_id`

	var reception models.Reception
	if err := pgxscan.Get(ctx, rw.exec, &reception, q, receptionId, status); err != nil {
		return reception, formatError(queryName, err)
	}

	return reception, nil
}

func (rw *rwPVZ) AddProductToReception(ctx context.Context, productType, receptionId string) (models.Product, error) {
	const queryName = "PVZRepository/AddProductToReception"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		insert into products (id, type, reception_id)
		values (gen_random_uuid(), $1, $2)
		returning id, date, type, reception_id`

	var product models.Product
	if err := pgxscan.Get(ctx, rw.exec, &product, q, productType, receptionId); err != nil {
		return product, formatError(queryName, err)
	}

	return product, nil
}

func (rw *rwPVZ) DeleteProduct(ctx context.Context, productId string) error {
	const queryName = "PVZRepository/DeleteProduct"
	span, ctx := opentracing.StartSpanFromContext(ctx, queryName)
	defer span.Finish()

	const q = `
		delete from products where id = $1`

	if _, err := rw.exec.Exec(ctx, q, productId); err != nil {
		return formatError(queryName, err)
	}
	return nil
}
