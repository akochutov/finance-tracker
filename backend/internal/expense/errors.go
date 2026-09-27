package expense

import "errors"

const (
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
	checkViolation      = "23514"
)

var (
	ErrNotFound     = errors.New("expense not found")
	ErrInvalidInput = errors.New("invalid input")
)
