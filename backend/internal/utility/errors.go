package utility

import "errors"

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
	checkViolation      = "23514"
)

var (
	ErrAddressNotFound = errors.New("address not found")
	ErrAccountNotFound = errors.New("utility account not found")
	ErrMeterNotFound   = errors.New("meter not found")
	ErrReadingNotFound = errors.New("reading not found")

	ErrAddressTaken  = errors.New("address already exists")
	ErrAccountTaken  = errors.New("account with this number already exists for the address and service")
	ErrSerialTaken   = errors.New("meter with this serial already exists on the account")
	ErrReadingExists = errors.New("reading for this date already exists")

	ErrInvalidInput = errors.New("invalid input")
)
