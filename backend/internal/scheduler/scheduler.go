package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
)

const retryInterval = time.Minute

type Scheduler struct {
	fetcher *ratefetch.Service
	sources *ratesource.Service
	wg      sync.WaitGroup
}

func New(fetcher *ratefetch.Service, sources *ratesource.Service) *Scheduler {
	return &Scheduler{fetcher: fetcher, sources: sources}
}

func (s *Scheduler) Start(ctx context.Context) {
	for _, kind := range []string{ratesource.KindFiat, ratesource.KindCrypto} {
		s.wg.Add(1)
		go s.runClass(ctx, kind)
	}
}

func (s *Scheduler) Wait() {
	s.wg.Wait()
}

func (s *Scheduler) runClass(ctx context.Context, kind string) {
	defer s.wg.Done()

	for {
		if ctx.Err() != nil {
			return
		}

		cfg, err := s.sources.Get(ctx, kind)
		if err != nil {
			log.Printf("scheduler[%s]: read config: %v", kind, err)
			if !sleep(ctx, retryInterval) {
				return
			}
			continue
		}

		if cfg.Source == "" {
			if !sleep(ctx, cfg.PollIntervalDuration()) {
				return
			}
			continue
		}

		result, err := s.fetcher.FetchClass(ctx, kind, time.Now())
		if err != nil {
			log.Printf("scheduler[%s] via %s: %v", kind, cfg.Source, err)
		} else {
			log.Printf("scheduler[%s] via %s: fetched=%d skipped=%d",
				kind, cfg.Source, result.Fetched, result.Skipped)
		}

		if !sleep(ctx, cfg.PollIntervalDuration()) {
			return
		}
	}
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
