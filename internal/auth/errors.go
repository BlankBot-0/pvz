package auth

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalidToken = errors.New("invalid token")
var ErrForbidden = errors.New("forbidden")
