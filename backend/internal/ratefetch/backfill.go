package ratefetch

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/ratesource"
)

const BackfillRunning = "running"
const BackfillDone = "done"
const BackfillFailed = "failed"

const dateLayout = "2006-01-02"

const cryptoThrottle = 150 * time.Millisecond

type BackfillJob struct {
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	From            string     `json:"from"`
	To              string     `json:"to"`
	TotalDays       int        `json:"total_days"`
	DoneDays        int        `json:"done_days"`
	Stored          int        `json:"stored"`
	Errors          int        `json:"errors"`
	LastError       string     `json:"last_error,omitempty"`
	CurrentCurrency string     `json:"current_currency,omitempty"`
	CurrentDate     string     `json:"current_date,omitempty"`
	Error           string     `json:"error,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type Backfiller struct {
	currencies *currency.Service
	sources    *ratesource.Service
	registry   *exchangerate.Registry
	rates      *exchangerate.Service

	mu   sync.Mutex
	jobs map[string]*BackfillJob
}

func NewBackfiller(
	currencies *currency.Service,
	sources *ratesource.Service,
	registry *exchangerate.Registry,
	rates *exchangerate.Service,
) *Backfiller {
	return &Backfiller{
		currencies: currencies,
		sources:    sources,
		registry:   registry,
		rates:      rates,
		jobs:       map[string]*BackfillJob{},
	}
}

func (b *Backfiller) Start(kind string) (BackfillJob, error) {
	if kind != ratesource.KindFiat && kind != ratesource.KindCrypto {
		return BackfillJob{}, ratesource.ErrInvalidKind
	}

	cfg, err := b.sources.Get(context.Background(), kind)
	if err != nil {
		return BackfillJob{}, err
	}
	if cfg.Source == "" {
		return BackfillJob{}, fmt.Errorf("%w: %s", ErrNoProvider, kind)
	}
	if cfg.BackfillStart == nil {
		return BackfillJob{}, ErrNoBackfillStart
	}

	provider, err := b.registry.Build(cfg.Source, cfg.URLTemplate, cfg.RequestTimeoutDuration())
	if err != nil {
		return BackfillJob{}, err
	}

	from := cfg.BackfillStart.UTC()
	to := time.Now().UTC()

	b.mu.Lock()
	if j, ok := b.jobs[kind]; ok && j.Status == BackfillRunning {
		b.mu.Unlock()
		return BackfillJob{}, ErrBackfillRunning
	}
	b.jobs[kind] = &BackfillJob{
		Kind:      kind,
		Status:    BackfillRunning,
		Source:    cfg.Source,
		From:      from.Format(dateLayout),
		To:        to.Format(dateLayout),
		StartedAt: time.Now().UTC(),
	}
	b.mu.Unlock()

	go b.run(kind, provider, cfg.Source, from, to)

	return b.snapshot(kind)
}

func (b *Backfiller) Status(kind string) (BackfillJob, error) {
	return b.snapshot(kind)
}

func (b *Backfiller) run(kind string, provider exchangerate.RateProvider, source string, from, to time.Time) {
	ctx := context.Background()

	if from.After(to) {
		b.finish(kind, nil)
		return
	}

	all, err := b.currencies.List(ctx)
	if err != nil {
		b.finish(kind, fmt.Errorf("list currencies: %w", err))
		return
	}
	var targets []string
	for _, c := range all {
		if c.Kind == kind && c.IsActive && c.Code != exchangerate.Pivot {
			targets = append(targets, c.Code)
		}
	}

	skipWeekends := kind == ratesource.KindFiat
	throttle := time.Duration(0)
	if kind == ratesource.KindCrypto {
		throttle = cryptoThrottle
	}

	b.setTotal(kind, int(to.Sub(from).Hours()/24)+1)

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		weekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
		if !(skipWeekends && weekend) {
			for _, code := range targets {
				b.setCurrent(kind, code, d)

				fr, err := provider.FetchRate(ctx, code, d)
				if err != nil {
					b.recordErr(kind, fmt.Errorf("%s @ %s: %w", code, d.Format(dateLayout), err))
				} else if _, err := b.rates.Record(ctx, code, source, fr.Rate, fr.RateAt); err != nil {
					b.recordErr(kind, fmt.Errorf("store %s @ %s: %w", code, d.Format(dateLayout), err))
				} else {
					b.incStored(kind)
				}

				if throttle > 0 {
					select {
					case <-ctx.Done():
						b.finish(kind, ctx.Err())
						return
					case <-time.After(throttle):
					}
				}
			}
		}

		if err := b.sources.SetBackfillStart(ctx, kind, d.AddDate(0, 0, 1)); err != nil {
			b.recordErr(kind, fmt.Errorf("checkpoint %s: %w", d.Format(dateLayout), err))
		}
		b.tickDay(kind)
	}

	b.finish(kind, nil)
}

func (b *Backfiller) setTotal(kind string, n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.jobs[kind].TotalDays = n
}

func (b *Backfiller) tickDay(kind string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.jobs[kind].DoneDays++
}

func (b *Backfiller) incStored(kind string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.jobs[kind].Stored++
}

func (b *Backfiller) setCurrent(kind, code string, d time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	j := b.jobs[kind]
	j.CurrentCurrency = code
	j.CurrentDate = d.Format(dateLayout)
}

func (b *Backfiller) recordErr(kind string, err error) {
	log.Printf("backfill[%s]: %v", kind, err)
	b.mu.Lock()
	defer b.mu.Unlock()
	j := b.jobs[kind]
	j.Errors++
	j.LastError = err.Error()
}

func (b *Backfiller) finish(kind string, fatal error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	j := b.jobs[kind]
	now := time.Now().UTC()
	j.FinishedAt = &now
	j.CurrentCurrency = ""
	j.CurrentDate = ""
	if fatal != nil {
		j.Status = BackfillFailed
		j.Error = fatal.Error()
	} else {
		j.Status = BackfillDone
	}
}

func (b *Backfiller) snapshot(kind string) (BackfillJob, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	j, ok := b.jobs[kind]
	if !ok {
		return BackfillJob{}, ErrNoBackfillJob
	}
	return *j, nil
}
