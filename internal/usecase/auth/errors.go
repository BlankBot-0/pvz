package auth

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrIncorrectPassword = errors.New("incorrect password")
var ErrIncorrectToken = errors.New("incorrect token")
var ErrInvalidRole = errors.New("invalid role")
