package dashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/settings"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const monthLayout = "2006-01"
const avgPrecision = 2

type Service struct {
	incomes   *income.Service
	companies *company.Service
	rates     *exchangerate.Service
	settings  *settings.Service
}

func NewService(
	incomes *income.Service,
	companies *company.Service,
	rates *exchangerate.Service,
	settings *settings.Service,
) *Service {
	return &Service{
		incomes:   incomes,
		companies: companies,
		rates:     rates,
		settings:  settings,
	}
}

func (s *Service) Get(ctx context.Context) (Dashboard, error) {
	set, err := s.settings.Get(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("dashboard: read settings: %w", err)
	}
	base := set.BaseCurrency

	incomes, err := s.incomes.List(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("dashboard: list incomes: %w", err)
	}

	companies, err := s.companies.List(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("dashboard: list companies: %w", err)
	}
	names := make(map[uuid.UUID]string, len(companies))
	for _, c := range companies {
		names[c.ID] = c.Name
	}

	now := time.Now().UTC()
	curYear := now.Year()
	curMonthKey := fmt.Sprintf("%04d-%02d", now.Year(), int(now.Month()))
	lm := now.AddDate(0, 0, -now.Day()+1).AddDate(0, -1, 0)
	lastMonthKey := fmt.Sprintf("%04d-%02d", lm.Year(), int(lm.Month()))

	total := decimal.Zero
	byMonth := map[string]decimal.Decimal{}
	byCurrency := map[string]decimal.Decimal{}
	byCompany := map[uuid.UUID]decimal.Decimal{}
	var thisYear, lastMonth, thisMonth decimal.Decimal
	var recTotal, converted, unconverted int
	minMonth, maxMonth := "", ""

	for _, in := range incomes {
		if !in.IsActive {
			continue
		}
		recTotal++

		monthKey := in.OccurredAt.UTC().Format(monthLayout)
		if minMonth == "" || monthKey < minMonth {
			minMonth = monthKey
		}
		if maxMonth == "" || monthKey > maxMonth {
			maxMonth = monthKey
		}

		conv, err := s.rates.Convert(ctx, in.Amount, in.Currency, base, in.OccurredAt)
		if err != nil {
			if errors.Is(err, exchangerate.ErrRateNotFound) {
				unconverted++
				continue
			}
			return Dashboard{}, fmt.Errorf("dashboard: convert income %s: %w", in.ID, err)
		}
		converted++

		total = total.Add(conv)
		byMonth[monthKey] = byMonth[monthKey].Add(conv)
		byCurrency[in.Currency] = byCurrency[in.Currency].Add(conv)
		byCompany[in.PayerID] = byCompany[in.PayerID].Add(conv)

		if in.OccurredAt.UTC().Year() == curYear {
			thisYear = thisYear.Add(conv)
		}
		if monthKey == curMonthKey {
			thisMonth = thisMonth.Add(conv)
		}
		if monthKey == lastMonthKey {
			lastMonth = lastMonth.Add(conv)
		}
	}

	monthsSeries := buildMonthSeries(minMonth, maxMonth, byMonth)

	avgAllTime := decimal.Zero
	if len(monthsSeries) > 0 {
		avgAllTime = total.DivRound(decimal.NewFromInt(int64(len(monthsSeries))), avgPrecision)
	}
	avgThisYear := thisYear.DivRound(decimal.NewFromInt(int64(now.Month())), avgPrecision)

	return Dashboard{
		Currency:    base,
		TotalIncome: total,
		Records:     RecordStats{Total: recTotal, Converted: converted, Unconverted: unconverted},
		Period:      Period{From: minMonth, To: maxMonth},
		ByMonth:     monthsSeries,
		ByCurrency:  sortedCurrencies(byCurrency),
		ByCompany:   sortedCompanies(byCompany, names),
		Averages:    Averages{MonthlyAllTime: avgAllTime, MonthlyThisYear: avgThisYear},
		Periods:     PeriodTotals{ThisYear: thisYear, LastMonth: lastMonth, ThisMonth: thisMonth},
	}, nil
}

func buildMonthSeries(from, to string, byMonth map[string]decimal.Decimal) []MonthBucket {
	out := make([]MonthBucket, 0)
	if from == "" {
		return out
	}
	start, _ := time.Parse(monthLayout, from)
	end, _ := time.Parse(monthLayout, to)
	for m := start; !m.After(end); m = m.AddDate(0, 1, 0) {
		k := m.Format(monthLayout)
		out = append(out, MonthBucket{Month: k, Total: byMonth[k]})
	}
	return out
}

func sortedCurrencies(m map[string]decimal.Decimal) []CurrencyBucket {
	out := make([]CurrencyBucket, 0, len(m))
	for cur, tot := range m {
		out = append(out, CurrencyBucket{Currency: cur, Total: tot})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Total.GreaterThan(out[j].Total) })
	return out
}

func sortedCompanies(m map[uuid.UUID]decimal.Decimal, names map[uuid.UUID]string) []CompanyBucket {
	out := make([]CompanyBucket, 0, len(m))
	for id, tot := range m {
		out = append(out, CompanyBucket{CompanyID: id.String(), Name: names[id], Total: tot})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Total.GreaterThan(out[j].Total) })
	return out
}
