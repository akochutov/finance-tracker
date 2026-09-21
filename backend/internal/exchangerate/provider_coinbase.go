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

var _ RateProvider = (*CoinbaseProvider)(nil)

type CoinbaseProvider struct {
	urlTemplate string
	timeout     time.Duration
	client      *http.Client
}

func NewCoinbaseProvider(urlTemplate string, timeout time.Duration) *CoinbaseProvider {
	return &CoinbaseProvider{
		urlTemplate: urlTemplate,
		timeout:     timeout,
		client:      &http.Client{},
	}
}

type coinbaseResponse struct {
	Data struct {
		Amount decimal.Decimal `json:"amount"`
	} `json:"data"`
}

func (p *CoinbaseProvider) FetchRate(ctx context.Context, currency string, at time.Time) (FetchedRate, error) {
	symbol := coinbaseSymbol(currency)

	url := p.urlTemplate
	url = strings.ReplaceAll(url, "{base}", symbol)
	url = strings.ReplaceAll(url, "{quote}", Pivot)
	url = strings.ReplaceAll(url, "{date}", at.Format("2006-01-02"))

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchedRate{}, fmt.Errorf("coinbase: build request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return FetchedRate{}, fmt.Errorf("coinbase: fetch %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FetchedRate{}, fmt.Errorf("coinbase: %s returned status %d", symbol, resp.StatusCode)
	}

	var body coinbaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return FetchedRate{}, fmt.Errorf("coinbase: decode %s: %w", symbol, err)
	}

	if !body.Data.Amount.IsPositive() {
		return FetchedRate{}, fmt.Errorf("coinbase: %s returned non-positive amount %s", symbol, body.Data.Amount)
	}

	return FetchedRate{Rate: body.Data.Amount, RateAt: at}, nil
}

func coinbaseSymbol(currency string) string {
	if i := strings.IndexByte(currency, ' '); i >= 0 {
		return currency[:i]
	}

	return currency
}
