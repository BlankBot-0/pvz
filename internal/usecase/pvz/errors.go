package pvz

import "errors"

var (
	ErrPVZNotFound                = errors.New("pvz not found")
	ErrAnotherReceptionInProgress = errors.New("another reception is in progress")
	ErrNoReceptionsInProgress     = errors.New("no receptions in progress")
	ErrNoProducts                 = errors.New("no products found")

	ErrCityNotFound        = errors.New("city is not found")
	ErrProductTypeNotFound = errors.New("product type is not found")
)
