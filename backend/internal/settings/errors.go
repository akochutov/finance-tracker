package settings

import "errors"

var ErrNotFound = errors.New("settings: not found")
var ErrNotFiat = errors.New("settings: base currency must be a fiat currency")
