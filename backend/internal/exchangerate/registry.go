package exchangerate

import (
	"fmt"
	"sort"
	"time"
)

type providerFactory func(urlTemplate string, timeout time.Duration) RateProvider

type Registry struct {
	factories map[string]providerFactory
}

func NewRegistry() *Registry {
	return &Registry{
		factories: map[string]providerFactory{
			"frankfurter": func(urlTemplate string, timeout time.Duration) RateProvider {
				return NewFrankfurterProvider(urlTemplate, timeout)
			},
			"coinbase": func(urlTemplate string, timeout time.Duration) RateProvider {
				return NewCoinbaseProvider(urlTemplate, timeout)
			},
		},
	}
}

func (r *Registry) Build(source, urlTemplate string, timeout time.Duration) (RateProvider, error) {
	factory, ok := r.factories[source]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, source)
	}

	return factory(urlTemplate, timeout), nil
}

func (r *Registry) Has(source string) bool {
	_, ok := r.factories[source]
	return ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
