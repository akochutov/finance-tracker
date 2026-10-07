package utilitydashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/settings"
	"github.com/akochutov/finance-tracker/internal/tariff"
	"github.com/akochutov/finance-tracker/internal/utility"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrInvalidInput = errors.New("invalid input")

const (
	monthLayout = "2006-01"
	dayLayout   = "2006-01-02"
	chartMonths = 12
	maxMonths   = 36
)

type Service struct {
	repo     *Repository
	tariffs  *tariff.Service
	rates    *exchangerate.Service
	settings *settings.Service
}

func NewService(repo *Repository, tariffs *tariff.Service, rates *exchangerate.Service, settings *settings.Service) *Service {
	return &Service{repo: repo, tariffs: tariffs, rates: rates, settings: settings}
}

type registerKey struct {
	meter uuid.UUID
	zone  string
}

type seriesData struct {
	key, label, kind, zone string
	consumption            []*decimal.Decimal
	cost                   []*decimal.Decimal
}

func (s *Service) Get(ctx context.Context, from, to time.Time) (Dashboard, error) {
	from, to = firstOfMonth(from), firstOfMonth(to)
	if to.Before(from) {
		return Dashboard{}, fmt.Errorf("%w: the period ends before it starts", ErrInvalidInput)
	}
	length := monthsBetween(from, to) + 1
	if length > maxMonths {
		return Dashboard{}, fmt.Errorf("%w: a period is at most %d months", ErrInvalidInput, maxMonths)
	}

	prevFrom := from.AddDate(0, -length, 0)
	prevTo := from.AddDate(0, -1, 0)
	chartFrom := to.AddDate(0, -(chartMonths - 1), 0)

	start := prevFrom
	if chartFrom.Before(start) {
		start = chartFrom
	}
	months := monthRange(start, to)
	index := make(map[string]int, len(months))
	for i, m := range months {
		index[m.Format(monthLayout)] = i
	}
	within := func(i int, a, b time.Time) bool { return !months[i].Before(a) && !months[i].After(b) }
	inPeriod := func(i int) bool { return within(i, from, to) }

	set, err := s.settings.Get(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("utility dashboard: read settings: %w", err)
	}
	base := set.ExpenseBaseCurrency

	services, err := s.repo.Services(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	accounts, err := s.repo.Accounts(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	meters, err := s.repo.Meters(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	readings, err := s.repo.Readings(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	var categories []uuid.UUID
	seenCategory := map[uuid.UUID]bool{}
	for _, a := range accounts {
		if a.ExpenseCategoryID != nil && !seenCategory[*a.ExpenseCategoryID] {
			seenCategory[*a.ExpenseCategoryID] = true
			categories = append(categories, *a.ExpenseCategoryID)
		}
	}
	payments, err := s.repo.Payments(ctx, categories, start, to.AddDate(0, 1, 0))
	if err != nil {
		return Dashboard{}, err
	}

	cov := newCoverage()
	conv := newConverter(s.rates, base, cov)
	prices := newPriceBook(s.tariffs)

	perRegister := map[registerKey][]*decimal.Decimal{}
	var prev *ReadingRow
	for i := range readings {
		rd := &readings[i]
		if prev != nil && prev.MeterID == rd.MeterID && prev.Zone == rd.Zone {
			month := firstOfMonth(rd.TakenOn).AddDate(0, -1, 0)
			if k, ok := index[month.Format(monthLayout)]; ok {
				key := registerKey{rd.MeterID, rd.Zone}
				if perRegister[key] == nil {
					perRegister[key] = make([]*decimal.Decimal, len(months))
				}
				addTo(perRegister[key], k, rd.Value.Sub(prev.Value))
			}
		}
		prev = rd
	}

	metersOf := map[uuid.UUID][]MeterRow{}
	for _, m := range meters {
		metersOf[m.AccountID] = append(metersOf[m.AccountID], m)
	}
	accountVolume := map[uuid.UUID][]*decimal.Decimal{}
	for _, a := range accounts {
		vol := make([]*decimal.Decimal, len(months))
		for _, m := range metersOf[a.ID] {
			for _, z := range utility.ZonesOf(m.Registers) {
				for k, v := range perRegister[registerKey{m.ID, z}] {
					if v != nil {
						addTo(vol, k, *v)
					}
				}
			}
		}
		accountVolume[a.ID] = vol
	}

	paidBy := map[uuid.UUID][]decimal.Decimal{}
	for _, p := range payments {
		k, ok := index[p.Month.Format(monthLayout)]
		if !ok {
			continue
		}
		amount, err := conv.convert(ctx, p.Amount, p.Currency, lastDay(p.Month))
		if err != nil {
			return Dashboard{}, err
		}
		if amount == nil {
			continue
		}
		if paidBy[p.CategoryID] == nil {
			paidBy[p.CategoryID] = make([]decimal.Decimal, len(months))
		}
		paidBy[p.CategoryID][k] = paidBy[p.CategoryID][k].Add(*amount)
	}

	blocks := make([]ServiceBlock, 0)
	for _, svc := range services {
		var accs []AccountRow
		meterCount := 0
		for _, a := range accounts {
			if a.Service == svc.Code {
				accs = append(accs, a)
				meterCount += len(metersOf[a.ID])
			}
		}
		if len(accs) == 0 {
			continue
		}

		var series []seriesData
		dayNight := false
		for _, a := range accs {
			for _, m := range metersOf[a.ID] {
				for _, z := range utility.ZonesOf(m.Registers) {
					if z != utility.ZoneSingle {
						dayNight = true
					}
					sr, err := s.registerSeries(ctx, svc, m, z, meterCount > 1, months,
						perRegister[registerKey{m.ID, z}], accountVolume[a.ID], prices, conv, cov, inPeriod)
					if err != nil {
						return Dashboard{}, err
					}
					series = append(series, sr)
				}
			}
		}

		estimated := make([]*decimal.Decimal, len(months))
		for k := range months {
			sum, known, any := decimal.Zero, true, false
			for _, sr := range series {
				if sr.consumption[k] == nil {
					continue
				}
				any = true
				if sr.cost[k] == nil {
					known = false
					break
				}
				sum = sum.Add(*sr.cost[k])
			}
			if any && known {
				v := sum
				estimated[k] = &v
			}
		}

		paid := make([]decimal.Decimal, len(months))
		tracks := false
		counted := map[uuid.UUID]bool{}
		for _, a := range accs {
			if a.ExpenseCategoryID == nil || counted[*a.ExpenseCategoryID] {
				continue
			}
			counted[*a.ExpenseCategoryID] = true
			tracks = true
			for k, v := range paidBy[*a.ExpenseCategoryID] {
				paid[k] = paid[k].Add(v)
			}
		}

		summarize := func(a, b time.Time) Summary {
			sum := Summary{TracksPayments: tracks, Zones: []ZoneAmount{}}
			zones := map[string]decimal.Decimal{}
			for k := range months {
				if !within(k, a, b) {
					continue
				}
				hasConsumption := false
				for _, sr := range series {
					if sr.consumption[k] == nil {
						continue
					}
					hasConsumption = true
					sum.Consumption = sum.Consumption.Add(*sr.consumption[k])
					if sr.kind == KindZone {
						zones[sr.zone] = zones[sr.zone].Add(*sr.consumption[k])
					}
				}
				if estimated[k] != nil {
					sum.Estimated = sum.Estimated.Add(*estimated[k])
				} else if hasConsumption {
					sum.MonthsWithoutTariff++
				}
				sum.Paid = sum.Paid.Add(paid[k])
			}
			if dayNight {
				for _, z := range []string{utility.ZoneDay, utility.ZoneNight} {
					sum.Zones = append(sum.Zones, ZoneAmount{Zone: z, Quantity: zones[z]})
				}
			}
			return sum
		}

		first := len(months) - chartMonths
		out := ServiceBlock{
			Service:   svc.Code,
			Name:      svc.Name,
			Unit:      svc.Unit,
			Summary:   summarize(from, to),
			Previous:  summarize(prevFrom, prevTo),
			Series:    make([]Series, 0, len(series)),
			Estimated: estimated[first:],
			Paid:      paid[first:],
		}
		for _, sr := range series {
			out.Series = append(out.Series, Series{
				Key:         sr.key,
				Label:       sr.label,
				Kind:        sr.kind,
				Zone:        sr.zone,
				Consumption: sr.consumption[first:],
				Cost:        sr.cost[first:],
			})
		}
		blocks = append(blocks, out)
	}

	chart := make([]string, 0, chartMonths)
	for _, m := range months[len(months)-chartMonths:] {
		chart = append(chart, m.Format(monthLayout))
	}

	return Dashboard{
		Currency: base,
		Period:   Period{From: from.Format(monthLayout), To: to.Format(monthLayout)},
		Previous: Period{From: prevFrom.Format(monthLayout), To: prevTo.Format(monthLayout)},
		Months:   chart,
		Services: blocks,
		Coverage: cov.result(),
	}, nil
}

func (s *Service) registerSeries(ctx context.Context, svc ServiceRow, m MeterRow, zone string, withSerial bool, months []time.Time,
	consumption, accountVolume []*decimal.Decimal, prices *priceBook, conv *converter, cov *coverage, inPeriod func(int) bool) (seriesData, error) {

	sr := seriesData{
		key:         m.ID.String() + ":" + zone,
		label:       registerLabel(m.Serial, zone, withSerial),
		kind:        KindMeter,
		zone:        zone,
		consumption: make([]*decimal.Decimal, len(months)),
		cost:        make([]*decimal.Decimal, len(months)),
	}
	if zone != utility.ZoneSingle {
		sr.kind = KindZone
	}

	for k, month := range months {
		if consumption == nil || consumption[k] == nil {
			continue
		}
		sr.consumption[k] = consumption[k]
		cost, err := s.price(ctx, prices, conv, svc.Code, zone, month, *accountVolume[k], *consumption[k])
		if err != nil {
			return seriesData{}, err
		}
		sr.cost[k] = cost
		if cost == nil && inPeriod(k) {
			what := svc.Code
			if zone != utility.ZoneSingle {
				what += " " + zone
			}
			cov.missingTariff(month, what)
		}
	}
	return sr, nil
}

func registerLabel(serial, zone string, withSerial bool) string {
	switch zone {
	case utility.ZoneDay, utility.ZoneNight:
		label := "Day"
		if zone == utility.ZoneNight {
			label = "Night"
		}
		if withSerial {
			label += " · " + serial
		}
		return label
	default:
		return serial
	}
}

func (s *Service) price(ctx context.Context, prices *priceBook, conv *converter, service, zone string, month time.Time, volume, priced decimal.Decimal) (*decimal.Decimal, error) {
	t, ok, err := prices.get(ctx, service, zone, lastDay(month))
	if err != nil || !ok {
		return nil, err
	}
	cost := tariff.Cost(t, volume, priced)
	return conv.convert(ctx, cost, t.Currency, lastDay(month))
}

type priceBook struct {
	tariffs *tariff.Service
	cache   map[string]*tariff.Tariff
}

func newPriceBook(t *tariff.Service) *priceBook {
	return &priceBook{tariffs: t, cache: map[string]*tariff.Tariff{}}
}

func (p *priceBook) get(ctx context.Context, service, zone string, day time.Time) (tariff.Tariff, bool, error) {
	key := service + "|" + zone + "|" + day.Format(dayLayout)
	if t, seen := p.cache[key]; seen {
		if t == nil {
			return tariff.Tariff{}, false, nil
		}
		return *t, true, nil
	}
	t, err := p.tariffs.ActiveAt(ctx, service, zone, day)
	if errors.Is(err, tariff.ErrNotFound) {
		p.cache[key] = nil
		return tariff.Tariff{}, false, nil
	}
	if err != nil {
		return tariff.Tariff{}, false, fmt.Errorf("utility dashboard: tariff: %w", err)
	}
	p.cache[key] = &t
	return t, true, nil
}

type converter struct {
	rates *exchangerate.Service
	base  string
	cov   *coverage
	cache map[string]*decimal.Decimal
}

func newConverter(r *exchangerate.Service, base string, cov *coverage) *converter {
	return &converter{rates: r, base: base, cov: cov, cache: map[string]*decimal.Decimal{}}
}

func (c *converter) convert(ctx context.Context, amount decimal.Decimal, currency string, day time.Time) (*decimal.Decimal, error) {
	if currency == c.base {
		return &amount, nil
	}
	key := currency + "|" + day.Format(dayLayout)
	factor, seen := c.cache[key]
	if !seen {
		endOfDay := day.Add(24*time.Hour - time.Nanosecond)
		f, err := c.rates.Convert(ctx, decimal.NewFromInt(1), currency, c.base, endOfDay)
		switch {
		case errors.Is(err, exchangerate.ErrRateNotFound):
			c.cov.missingRate(day, currency)
		case err != nil:
			return nil, fmt.Errorf("utility dashboard: convert %s: %w", currency, err)
		default:
			factor = &f
		}
		c.cache[key] = factor
	}
	if factor == nil {
		return nil, nil
	}
	v := amount.Mul(*factor)
	return &v, nil
}

type coverage struct {
	tariffs, rates map[string]bool
}

func newCoverage() *coverage {
	return &coverage{tariffs: map[string]bool{}, rates: map[string]bool{}}
}

func (c *coverage) missingTariff(month time.Time, what string) {
	c.tariffs[month.Format(monthLayout)+" "+what] = true
}

func (c *coverage) missingRate(day time.Time, currency string) {
	c.rates[day.Format(dayLayout)+" "+currency] = true
}

func (c *coverage) result() Coverage {
	return Coverage{
		MissingTariffs: sortedKeys(c.tariffs),
		MissingRates:   sortedKeys(c.rates),
	}
}

func addTo(values []*decimal.Decimal, k int, v decimal.Decimal) {
	if values[k] == nil {
		values[k] = &v
		return
	}
	sum := values[k].Add(v)
	values[k] = &sum
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func firstOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

func lastDay(month time.Time) time.Time {
	return firstOfMonth(month).AddDate(0, 1, -1)
}

func monthsBetween(a, b time.Time) int {
	return (b.Year()-a.Year())*12 + int(b.Month()) - int(a.Month())
}

func monthRange(from, to time.Time) []time.Time {
	out := make([]time.Time, 0, monthsBetween(from, to)+1)
	for m := from; !m.After(to); m = m.AddDate(0, 1, 0) {
		out = append(out, m)
	}
	return out
}
