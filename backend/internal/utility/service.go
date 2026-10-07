package utility

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const ServiceElectricity = "electricity"

const dateLayout = "2006-01-02"

type Service struct {
	serviceTypes *ServiceTypeRepository
	addresses    *AddressRepository
	accounts     *AccountRepository
	meters       *MeterRepository
	readings     *ReadingRepository
}

func NewService(
	serviceTypes *ServiceTypeRepository,
	addresses *AddressRepository,
	accounts *AccountRepository,
	meters *MeterRepository,
	readings *ReadingRepository,
) *Service {
	return &Service{
		serviceTypes: serviceTypes,
		addresses:    addresses,
		accounts:     accounts,
		meters:       meters,
		readings:     readings,
	}
}

func (s *Service) ListServiceTypes(ctx context.Context) ([]ServiceType, error) {
	return s.serviceTypes.List(ctx)
}

func (s *Service) CreateAddress(ctx context.Context, address string) (Address, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return Address{}, fmt.Errorf("%w: address is required", ErrInvalidInput)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Address{}, fmt.Errorf("generate uuid: %w", err)
	}

	return s.addresses.Create(ctx, Address{ID: id, Address: address, IsActive: true})
}

func (s *Service) ListAddresses(ctx context.Context) ([]Address, error) {
	return s.addresses.List(ctx)
}

func (s *Service) UpdateAddress(ctx context.Context, id uuid.UUID, address string) (Address, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return Address{}, fmt.Errorf("%w: address is required", ErrInvalidInput)
	}
	return s.addresses.Update(ctx, id, address)
}

func (s *Service) SetAddressActive(ctx context.Context, id uuid.UUID, active bool) error {
	if !active {
		accounts, err := s.accounts.List(ctx)
		if err != nil {
			return err
		}
		for _, a := range accounts {
			if a.AddressID == id && a.IsActive {
				return ErrAddressHasActiveAccounts
			}
		}
	}
	return s.addresses.SetActive(ctx, id, active)
}

func (s *Service) CreateAccount(ctx context.Context, addressID uuid.UUID, service, number string, categoryID *uuid.UUID) (Account, error) {
	service = strings.ToLower(strings.TrimSpace(service))
	number = strings.TrimSpace(number)
	if service == "" {
		return Account{}, fmt.Errorf("%w: service is required", ErrInvalidInput)
	}
	if number == "" {
		return Account{}, fmt.Errorf("%w: account number is required", ErrInvalidInput)
	}

	addr, err := s.addresses.GetByID(ctx, addressID)
	if err != nil {
		return Account{}, err
	}
	if !addr.IsActive {
		return Account{}, ErrAddressInactive
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Account{}, fmt.Errorf("generate uuid: %w", err)
	}

	return s.accounts.Create(ctx, Account{
		ID:                id,
		AddressID:         addressID,
		Service:           service,
		Number:            number,
		ExpenseCategoryID: categoryID,
		IsActive:          true,
	})
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	return s.accounts.List(ctx)
}

func (s *Service) UpdateAccount(ctx context.Context, id uuid.UUID, number string, categoryID *uuid.UUID) (Account, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return Account{}, fmt.Errorf("%w: account number is required", ErrInvalidInput)
	}
	return s.accounts.Update(ctx, id, number, categoryID)
}

func (s *Service) SetAccountActive(ctx context.Context, id uuid.UUID, active bool) error {
	if active {
		acc, err := s.accounts.GetByID(ctx, id)
		if err != nil {
			return err
		}
		addr, err := s.addresses.GetByID(ctx, acc.AddressID)
		if err != nil {
			return err
		}
		if !addr.IsActive {
			return ErrAddressInactive
		}
	}
	return s.accounts.SetActive(ctx, id, active)
}

func (s *Service) CreateMeter(ctx context.Context, accountID uuid.UUID, serial, registers string, installedOn time.Time, removedOn *time.Time, initialOn time.Time, initialValues map[string]decimal.Decimal) (Meter, error) {
	serial = strings.TrimSpace(serial)
	installedOn, removedOn, err := normalizeMeter(serial, installedOn, removedOn)
	if err != nil {
		return Meter{}, err
	}

	registers = strings.ToLower(strings.TrimSpace(registers))
	if registers == "" {
		registers = RegistersSingle
	}
	if registers != RegistersSingle && registers != RegistersDayNight {
		return Meter{}, fmt.Errorf("%w: registers must be %q or %q", ErrInvalidInput, RegistersSingle, RegistersDayNight)
	}

	zones := ZonesOf(registers)
	if len(initialValues) == 0 {
		return Meter{}, ErrInitialReadingRequired
	}
	if len(initialValues) != len(zones) {
		return Meter{}, fmt.Errorf("%w: expected initial values for %s", ErrInvalidInput, strings.Join(zones, ", "))
	}

	if initialOn.IsZero() {
		initialOn = installedOn
	}
	initialOn = dateOnly(initialOn)

	acc, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return Meter{}, err
	}
	if !acc.IsActive {
		return Meter{}, ErrAccountInactive
	}

	meter := Meter{AccountID: accountID, Serial: serial, Registers: registers, InstalledOn: installedOn, RemovedOn: removedOn}
	if err := checkInService(meter, initialOn); err != nil {
		return Meter{}, err
	}

	initial := make([]Reading, 0, len(zones))
	for _, z := range zones {
		value, ok := initialValues[z]
		if !ok {
			return Meter{}, fmt.Errorf("%w: missing initial value for %s", ErrInvalidInput, z)
		}
		if value.IsNegative() {
			return Meter{}, fmt.Errorf("%w: initial value must not be negative", ErrInvalidInput)
		}
		id, err := uuid.NewV7()
		if err != nil {
			return Meter{}, fmt.Errorf("generate uuid: %w", err)
		}
		initial = append(initial, Reading{ID: id, Zone: z, TakenOn: initialOn, Value: value, IsInitial: true})
	}

	if meter.ID, err = uuid.NewV7(); err != nil {
		return Meter{}, fmt.Errorf("generate uuid: %w", err)
	}
	return s.meters.Create(ctx, meter, initial)
}

func (s *Service) ListMeters(ctx context.Context) ([]Meter, error) {
	return s.meters.List(ctx)
}

func (s *Service) UpdateMeter(ctx context.Context, id uuid.UUID, serial string, installedOn time.Time, removedOn *time.Time) (Meter, error) {
	serial = strings.TrimSpace(serial)
	installedOn, removedOn, err := normalizeMeter(serial, installedOn, removedOn)
	if err != nil {
		return Meter{}, err
	}

	if _, err := s.meters.GetByID(ctx, id); err != nil {
		return Meter{}, err
	}

	readings, err := s.readings.ListByMeter(ctx, id)
	if err != nil {
		return Meter{}, err
	}
	for _, rd := range readings {
		if rd.TakenOn.Before(installedOn) || (removedOn != nil && rd.TakenOn.After(*removedOn)) {
			return Meter{}, fmt.Errorf("%w: reading of %s falls outside the new service dates",
				ErrInvalidInput, rd.TakenOn.Format(dateLayout))
		}
	}

	return s.meters.Update(ctx, id, serial, installedOn, removedOn)
}

func (s *Service) DeleteMeter(ctx context.Context, id uuid.UUID) error {
	return s.meters.Delete(ctx, id)
}

func (s *Service) CreateReadings(ctx context.Context, takenOn time.Time, values map[uuid.UUID]map[string]decimal.Decimal) ([]Reading, error) {
	if takenOn.IsZero() {
		return nil, fmt.Errorf("%w: reading date is required", ErrInvalidInput)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: no readings to save", ErrInvalidInput)
	}
	takenOn = dateOnly(takenOn)

	readings := make([]Reading, 0, len(values)*2)
	for meterID, byZone := range values {
		meter, err := s.meters.GetByID(ctx, meterID)
		if err != nil {
			return nil, err
		}
		if err := checkInService(meter, takenOn); err != nil {
			return nil, fmt.Errorf("meter %s: %w", meter.Serial, err)
		}

		zones := ZonesOf(meter.Registers)
		if len(byZone) != len(zones) {
			return nil, fmt.Errorf("%w: meter %s: expected values for %s", ErrInvalidInput, meter.Serial, strings.Join(zones, ", "))
		}
		for _, z := range zones {
			value, ok := byZone[z]
			if !ok {
				return nil, fmt.Errorf("%w: meter %s: missing value for %s", ErrInvalidInput, meter.Serial, z)
			}
			if err := s.checkMonotonic(ctx, meter, z, takenOn, value, uuid.Nil, false); err != nil {
				return nil, err
			}
			id, err := uuid.NewV7()
			if err != nil {
				return nil, fmt.Errorf("generate uuid: %w", err)
			}
			readings = append(readings, Reading{ID: id, MeterID: meterID, Zone: z, TakenOn: takenOn, Value: value})
		}
	}

	return s.readings.CreateBatch(ctx, readings)
}

func (s *Service) ListReadings(ctx context.Context, meterID uuid.UUID) ([]Reading, error) {
	if _, err := s.meters.GetByID(ctx, meterID); err != nil {
		return nil, err
	}
	return s.readings.ListByMeter(ctx, meterID)
}

func (s *Service) LatestReadings(ctx context.Context) ([]Reading, error) {
	return s.readings.LatestPerMeter(ctx)
}

func (s *Service) UpdateReading(ctx context.Context, meterID, readingID uuid.UUID, takenOn time.Time, value decimal.Decimal) (Reading, error) {
	if takenOn.IsZero() {
		return Reading{}, fmt.Errorf("%w: reading date is required", ErrInvalidInput)
	}
	takenOn = dateOnly(takenOn)

	meter, rd, err := s.readingOfMeter(ctx, meterID, readingID)
	if err != nil {
		return Reading{}, err
	}
	if err := checkInService(meter, takenOn); err != nil {
		return Reading{}, err
	}
	if err := s.checkMonotonic(ctx, meter, rd.Zone, takenOn, value, rd.ID, rd.IsInitial); err != nil {
		return Reading{}, err
	}

	return s.readings.Update(ctx, rd.ID, takenOn, value)
}

func (s *Service) DeleteReading(ctx context.Context, meterID, readingID uuid.UUID) error {
	_, rd, err := s.readingOfMeter(ctx, meterID, readingID)
	if err != nil {
		return err
	}
	if rd.IsInitial {
		return ErrInitialReadingLocked
	}
	return s.readings.Delete(ctx, rd.ID)
}

func (s *Service) readingOfMeter(ctx context.Context, meterID, readingID uuid.UUID) (Meter, Reading, error) {
	meter, err := s.meters.GetByID(ctx, meterID)
	if err != nil {
		return Meter{}, Reading{}, err
	}
	rd, err := s.readings.GetByID(ctx, readingID)
	if err != nil {
		return Meter{}, Reading{}, err
	}
	if rd.MeterID != meter.ID {
		return Meter{}, Reading{}, ErrReadingNotFound
	}
	return meter, rd, nil
}

func (s *Service) checkMonotonic(ctx context.Context, meter Meter, zone string, day time.Time, value decimal.Decimal, excludeID uuid.UUID, isInitial bool) error {
	name := meter.Serial
	if zone != ZoneSingle {
		name += " " + zone
	}
	if value.IsNegative() {
		return fmt.Errorf("%w: meter %s: value must not be negative", ErrInvalidInput, name)
	}

	prev, next, err := s.readings.Neighbors(ctx, meter.ID, zone, day, excludeID)
	if err != nil {
		return err
	}

	if isInitial && prev != nil {
		return fmt.Errorf("%w: meter %s: the initial reading must be the earliest, there is one of %s",
			ErrInvalidInput, name, prev.TakenOn.Format(dateLayout))
	}
	if !isInitial && prev == nil {
		return fmt.Errorf("%w: meter %s: the reading is before the initial reading", ErrInvalidInput, name)
	}
	if prev != nil && value.LessThan(prev.Value) {
		return fmt.Errorf("%w: meter %s: value %s is below the previous reading %s of %s",
			ErrInvalidInput, name, value, prev.Value, prev.TakenOn.Format(dateLayout))
	}
	if next != nil && value.GreaterThan(next.Value) {
		return fmt.Errorf("%w: meter %s: value %s is above the next reading %s of %s",
			ErrInvalidInput, name, value, next.Value, next.TakenOn.Format(dateLayout))
	}
	return nil
}

func normalizeMeter(serial string, installedOn time.Time, removedOn *time.Time) (time.Time, *time.Time, error) {
	if serial == "" {
		return time.Time{}, nil, fmt.Errorf("%w: serial is required", ErrInvalidInput)
	}
	if installedOn.IsZero() {
		return time.Time{}, nil, fmt.Errorf("%w: installation date is required", ErrInvalidInput)
	}
	installedOn = dateOnly(installedOn)

	if removedOn != nil {
		r := dateOnly(*removedOn)
		if r.Before(installedOn) {
			return time.Time{}, nil, fmt.Errorf("%w: meter removed before it was installed", ErrInvalidInput)
		}
		removedOn = &r
	}

	return installedOn, removedOn, nil
}

func checkInService(m Meter, day time.Time) error {
	if day.After(dateOnly(time.Now())) {
		return fmt.Errorf("%w: reading date is in the future", ErrInvalidInput)
	}
	if day.Before(m.InstalledOn) {
		return fmt.Errorf("%w: reading date is before the meter was installed on %s",
			ErrInvalidInput, m.InstalledOn.Format(dateLayout))
	}
	if m.RemovedOn != nil && day.After(*m.RemovedOn) {
		return fmt.Errorf("%w: reading date is after the meter was removed on %s",
			ErrInvalidInput, m.RemovedOn.Format(dateLayout))
	}
	return nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
