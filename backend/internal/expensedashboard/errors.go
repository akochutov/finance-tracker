package expensedashboard

import "errors"

var ErrInvalidPeriod = errors.New("invalid period: to is before from")
