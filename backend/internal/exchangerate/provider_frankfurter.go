package exchangerate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var _ RateProvider = (*FrankfurterProvider)(nil)

type FrankfurterProvider struct {
	urlTemplate string
	timeout     time.Duration
	client      *http.Client
}

func NewFrankfurterProvider(urlTemplate string, timeout time.Duration) *FrankfurterProvider {
	return &FrankfurterProvider{
		urlTemplate: urlTemplate,
		timeout:     timeout,
		client:      &http.Client{},
	}
}

type frankfurterResponse struct {
	Date string          `json:"date"`
	Rate decimal.Decimal `json:"rate"`
}

func (p *FrankfurterProvider) FetchRate(ctx context.Context, currency string, at time.Time) (FetchedRate, error) {
	url := p.urlTemplate
	url = strings.ReplaceAll(url, "{base}", currency)
	url = strings.ReplaceAll(url, "{quote}", Pivot)
	url = strings.ReplaceAll(url, "{date}", at.Format("2006-01-02"))

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchedRate{}, fmt.Errorf("frankfurter: build request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return FetchedRate{}, fmt.Errorf("frankfurter: fetch %s: %w", currency, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FetchedRate{}, fmt.Errorf("frankfurter: %s returned status %d", currency, resp.StatusCode)
	}

	var body frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return FetchedRate{}, fmt.Errorf("frankfurter: decode %s: %w", currency, err)
	}

	if !body.Rate.IsPositive() {
		return FetchedRate{}, fmt.Errorf("frankfurter: %s returned non-positive rate %s", currency, body.Rate)
	}

	rateAt, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		return FetchedRate{}, fmt.Errorf("frankfurter: parse date %q: %w", body.Date, err)
	}

	return FetchedRate{Rate: body.Rate, RateAt: rateAt}, nil
}
