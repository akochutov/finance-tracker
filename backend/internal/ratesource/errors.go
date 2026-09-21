package ratesource

import "errors"

var ErrNotFound = errors.New("ratesource: not found")
var ErrInvalidKind = errors.New("ratesource: kind must be 'fiat' or 'crypto'")
var ErrInvalidConfig = errors.New("ratesource: invalid configuration")
