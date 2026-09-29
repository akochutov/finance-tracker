package expensedashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/settings"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	dateLayout   = "2006-01-02"
	monthLayout  = "2006-01"
	trendMonths  = 12
	largestCount = 6
	moneyPlaces  = 2
)

type Service struct {
	repo     *Repository
	rates    *exchangerate.Service
	settings *settings.Service
	incomes  *income.Service
}

func NewService(repo *Repository, rates *exchangerate.Service, settings *settings.Service, incomes *income.Service) *Service {
	return &Service{repo: repo, rates: rates, settings: settings, incomes: incomes}
}

type categoryAcc struct {
	name  string
	total decimal.Decimal
}

type groupAcc struct {
	name        string
	total, prev decimal.Decimal
	categories  map[uuid.UUID]*categoryAcc
}

type monthAcc struct {
	total, fixed, variable, income decimal.Decimal
	groups                         map[uuid.UUID]decimal.Decimal
}

func (s *Service) Get(ctx context.Context, from, to time.Time) (Dashboard, error) {
	from, to = dateOnly(from), dateOnly(to)
	if to.Before(from) {
		return Dashboard{}, ErrInvalidPeriod
	}

	set, err := s.settings.Get(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("expense dashboard: read settings: %w", err)
	}
	base := set.ExpenseBaseCurrency

	prevFrom, prevTo := previousPeriod(from, to)
	trendFrom := firstOfMonth(to).AddDate(0, -(trendMonths - 1), 0)
	trendTo := firstOfMonth(to).AddDate(0, 1, -1)

	loadFrom := trendFrom
	if prevFrom.Before(loadFrom) {
		loadFrom = prevFrom
	}

	months := make(map[string]*monthAcc, trendMonths)
	monthKeys := make([]string, 0, trendMonths)
	for i := 0; i < trendMonths; i++ {
		key := trendFrom.AddDate(0, i, 0).Format(monthLayout)
		months[key] = &monthAcc{groups: map[uuid.UUID]decimal.Decimal{}}
		monthKeys = append(monthKeys, key)
	}

	conv := newConverter(s.rates, base)
	groups := map[uuid.UUID]*groupAcc{}
	unconvertedDays := map[string]bool{}
	var total, prevTotal, fixed, variable decimal.Decimal

	rows, err := s.repo.Amounts(ctx, loadFrom, trendTo)
	if err != nil {
		return Dashboard{}, err
	}
	for _, r := range rows {
		f, ok, err := conv.factor(ctx, r.Currency, r.Day)
		if err != nil {
			return Dashboard{}, err
		}
		if !ok {
			unconvertedDays[r.Day.Format(dateLayout)] = true
			continue
		}
		amount := r.Amount.Mul(f)

		inPeriod := within(r.Day, from, to)
		inPrev := within(r.Day, prevFrom, prevTo)

		if inPeriod || inPrev {
			g := groups[r.GroupID]
			if g == nil {
				g = &groupAcc{name: r.GroupName, categories: map[uuid.UUID]*categoryAcc{}}
				groups[r.GroupID] = g
			}
			if inPeriod {
				g.total = g.total.Add(amount)
				c := g.categories[r.CategoryID]
				if c == nil {
					c = &categoryAcc{name: r.CategoryName}
					g.categories[r.CategoryID] = c
				}
				c.total = c.total.Add(amount)

				total = total.Add(amount)
				if r.Fixed {
					fixed = fixed.Add(amount)
				} else {
					variable = variable.Add(amount)
				}
			}
			if inPrev {
				g.prev = g.prev.Add(amount)
				prevTotal = prevTotal.Add(amount)
			}
		}

		if m := months[r.Day.Format(monthLayout)]; m != nil {
			m.total = m.total.Add(amount)
			if r.Fixed {
				m.fixed = m.fixed.Add(amount)
			} else {
				m.variable = m.variable.Add(amount)
			}
			m.groups[r.GroupID] = m.groups[r.GroupID].Add(amount)
		}
	}

	incomes, err := s.incomes.List(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("expense dashboard: list incomes: %w", err)
	}
	var periodIncome, prevIncome decimal.Decimal
	unconvertedIncomes := 0
	for _, in := range incomes {
		if !in.IsActive {
			continue
		}
		day := dateOnly(in.OccurredAt)
		inPeriod := within(day, from, to)
		inPrev := within(day, prevFrom, prevTo)
		m := months[day.Format(monthLayout)]
		if !inPeriod && !inPrev && m == nil {
			continue
		}

		amount, err := s.rates.Convert(ctx, in.Amount, in.Currency, base, in.OccurredAt)
		if err != nil {
			if errors.Is(err, exchangerate.ErrRateNotFound) {
				unconvertedIncomes++
				continue
			}
			return Dashboard{}, fmt.Errorf("expense dashboard: convert income %s: %w", in.ID, err)
		}

		if inPeriod {
			periodIncome = periodIncome.Add(amount)
		}
		if inPrev {
			prevIncome = prevIncome.Add(amount)
		}
		if m != nil {
			m.income = m.income.Add(amount)
		}
	}

	receipts, err := s.repo.Receipts(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	largest, err := s.largest(ctx, conv, receipts)
	if err != nil {
		return Dashboard{}, err
	}

	days := daysBetween(from, to)
	prevDays := daysBetween(prevFrom, prevTo)
	elapsed := elapsedDays(from, to, dateOnly(time.Now()))

	return Dashboard{
		Currency: base,
		Period:   Period{From: from.Format(dateLayout), To: to.Format(dateLayout), Days: days},
		Previous: Period{From: prevFrom.Format(dateLayout), To: prevTo.Format(dateLayout), Days: prevDays},
		Summary: Summary{
			Total:                total.Round(moneyPlaces),
			PreviousTotal:        prevTotal.Round(moneyPlaces),
			DailyAverage:         total.DivRound(decimal.NewFromInt(int64(elapsed)), moneyPlaces),
			PreviousDailyAverage: prevTotal.DivRound(decimal.NewFromInt(int64(prevDays)), moneyPlaces),
			Receipts:             len(receipts),
			Fixed:                fixed.Round(moneyPlaces),
			Variable:             variable.Round(moneyPlaces),
			Income:               periodIncome.Round(moneyPlaces),
			SavingsRate:          savingsRate(periodIncome, total),
			PreviousSavingsRate:  savingsRate(prevIncome, prevTotal),
		},
		Groups:  buildGroups(groups),
		Trend:   buildTrend(monthKeys, months),
		Largest: largest,
		Coverage: Coverage{
			UnconvertedDays:    len(unconvertedDays),
			UnconvertedIncomes: unconvertedIncomes,
		},
	}, nil
}

func (s *Service) largest(ctx context.Context, conv *converter, receipts []ReceiptRow) ([]Receipt, error) {
	type converted struct {
		row   ReceiptRow
		total decimal.Decimal
	}

	list := make([]converted, 0, len(receipts))
	for _, rc := range receipts {
		f, ok, err := conv.factor(ctx, rc.Currency, rc.Day)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		list = append(list, converted{row: rc, total: rc.Total.Mul(f)})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].total.GreaterThan(list[j].total) })
	if len(list) > largestCount {
		list = list[:largestCount]
	}

	out := make([]Receipt, 0, len(list))
	for _, c := range list {
		summary := c.row.Descriptions
		if c.row.Note != nil && *c.row.Note != "" {
			summary = *c.row.Note
		}
		out = append(out, Receipt{
			ExpenseID:  c.row.ExpenseID.String(),
			OccurredOn: c.row.Day.Format(dateLayout),
			Total:      c.total.Round(moneyPlaces),
			Group:      c.row.Group,
			Summary:    summary,
			Lines:      c.row.Lines,
		})
	}
	return out, nil
}

func buildGroups(groups map[uuid.UUID]*groupAcc) []GroupTotal {
	out := make([]GroupTotal, 0, len(groups))
	for id, g := range groups {
		cats := make([]CategoryTotal, 0, len(g.categories))
		for cid, c := range g.categories {
			cats = append(cats, CategoryTotal{CategoryID: cid.String(), Name: c.name, Total: c.total.Round(moneyPlaces)})
		}
		sort.Slice(cats, func(i, j int) bool { return cats[i].Total.GreaterThan(cats[j].Total) })

		out = append(out, GroupTotal{
			GroupID:    id.String(),
			Name:       g.name,
			Total:      g.total.Round(moneyPlaces),
			Previous:   g.prev.Round(moneyPlaces),
			Categories: cats,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Total.Equal(out[j].Total) {
			return out[i].Total.GreaterThan(out[j].Total)
		}
		return out[i].Previous.GreaterThan(out[j].Previous)
	})
	return out
}

func buildTrend(keys []string, months map[string]*monthAcc) []MonthTrend {
	out := make([]MonthTrend, 0, len(keys))
	for _, k := range keys {
		m := months[k]
		byGroup := make([]GroupAmount, 0, len(m.groups))
		for id, v := range m.groups {
			byGroup = append(byGroup, GroupAmount{GroupID: id.String(), Total: v.Round(moneyPlaces)})
		}
		sort.Slice(byGroup, func(i, j int) bool { return byGroup[i].Total.GreaterThan(byGroup[j].Total) })

		out = append(out, MonthTrend{
			Month:    k,
			Total:    m.total.Round(moneyPlaces),
			Fixed:    m.fixed.Round(moneyPlaces),
			Variable: m.variable.Round(moneyPlaces),
			Income:   m.income.Round(moneyPlaces),
			ByGroup:  byGroup,
		})
	}
	return out
}

func savingsRate(income, spent decimal.Decimal) *decimal.Decimal {
	if !income.IsPositive() || spent.IsZero() {
		return nil
	}
	r := income.Sub(spent).Div(income).Mul(decimal.NewFromInt(100)).Round(1)
	return &r
}

func previousPeriod(from, to time.Time) (time.Time, time.Time) {
	prevTo := from.AddDate(0, 0, -1)
	if from.Day() == 1 && to.AddDate(0, 0, 1).Day() == 1 {
		months := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month()) + 1
		return from.AddDate(0, -months, 0), prevTo
	}
	return prevTo.AddDate(0, 0, -(daysBetween(from, to) - 1)), prevTo
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours()/24) + 1
}

func elapsedDays(from, to, today time.Time) int {
	end := to
	if today.Before(end) {
		end = today
	}
	if end.Before(from) {
		return 1
	}
	return daysBetween(from, end)
}

func within(day, from, to time.Time) bool {
	return !day.Before(from) && !day.After(to)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

type converter struct {
	rates *exchangerate.Service
	base  string
	cache map[string]*decimal.Decimal
}

func newConverter(rates *exchangerate.Service, base string) *converter {
	return &converter{rates: rates, base: base, cache: map[string]*decimal.Decimal{}}
}

func (c *converter) factor(ctx context.Context, currency string, day time.Time) (decimal.Decimal, bool, error) {
	if currency == c.base {
		return decimal.NewFromInt(1), true, nil
	}

	key := currency + "|" + day.Format(dateLayout)
	if f, seen := c.cache[key]; seen {
		if f == nil {
			return decimal.Zero, false, nil
		}
		return *f, true, nil
	}

	at := day.AddDate(0, 0, 1).Add(-time.Nanosecond)
	f, err := c.rates.Convert(ctx, decimal.NewFromInt(1), currency, c.base, at)
	if err != nil {
		if errors.Is(err, exchangerate.ErrRateNotFound) {
			c.cache[key] = nil
			return decimal.Zero, false, nil
		}
		return decimal.Zero, false, fmt.Errorf("expense dashboard: convert %s on %s: %w", currency, day.Format(dateLayout), err)
	}
	c.cache[key] = &f
	return f, true, nil
}
