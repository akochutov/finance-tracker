package exchangerate

import "errors"

var ErrRateNotFound = errors.New("exchangerate: no rate at or before the given time")
var ErrPivotRate = errors.New("exchangerate: cannot store a rate for the pivot currency")
var ErrUnknownProvider = errors.New("exchangerate: unknown rate provider")
