package ratefetch

import "errors"

var ErrNoProvider = errors.New("ratefetch: no provider configured for this class")
var ErrBackfillRunning = errors.New("ratefetch: backfill already running for this class")
var ErrNoBackfillStart = errors.New("ratefetch: backfill start date not configured")
var ErrNoBackfillJob = errors.New("ratefetch: no backfill job for this class")
