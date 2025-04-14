package pvz

import "errors"

var (
	ErrIDIsOccupied               = errors.New("id is already occupied")
	ErrPVZNotFound                = errors.New("pvz not found")
	ErrAnotherReceptionInProgress = errors.New("another reception is in progress")
	ErrNoReceptionsInProgress     = errors.New("no receptions in progress")
	ErrNoProducts                 = errors.New("no products found")

	ErrCityNotFound        = errors.New("city is not found")
	ErrProductTypeNotFound = errors.New("product type is not found")
)
