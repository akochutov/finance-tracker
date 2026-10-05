package tariff

import "errors"

const (
	uniqueViolation = "23505"
)

var (
	ErrNotFound      = errors.New("tariff not found")
	ErrAlreadyExists = errors.New("tariff for this service and zone already starts on this date")
	ErrInvalidInput  = errors.New("invalid input")
)
