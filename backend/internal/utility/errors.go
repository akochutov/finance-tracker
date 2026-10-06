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

	ErrAddressInactive          = errors.New("address is inactive")
	ErrAddressHasActiveAccounts = errors.New("address has active utility accounts")
	ErrAccountInactive          = errors.New("utility account is inactive")

	ErrInitialReadingRequired = errors.New("initial reading is required")
	ErrInitialReadingLocked   = errors.New("the initial reading cannot be deleted; edit it or delete the meter")
	ErrZonesNotSplit          = errors.New("this account has a single tariff zone")

	ErrInvalidInput = errors.New("invalid input")
)
