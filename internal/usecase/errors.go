package usecase

import "errors"

var (
	ErrInvalidAdminToken  = errors.New("invalid admin token")
	ErrInvalidLogin       = errors.New("login must contain at least 8 latin letters or digits")
	ErrInvalidPassword    = errors.New("password must contain at least 8 characters and include lower-case, upper-case, digit and symbol characters")
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrInvalidSession     = errors.New("invalid or expired session")
	ErrForbidden          = errors.New("access forbidden")
)
