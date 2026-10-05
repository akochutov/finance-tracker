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

func (s *Service) CreateAccount(ctx context.Context, addressID uuid.UUID, service, number string) (Account, error) {
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
		ID:        id,
		AddressID: addressID,
		Service:   service,
		Number:    number,
		IsActive:  true,
	})
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	return s.accounts.List(ctx)
}

func (s *Service) UpdateAccount(ctx context.Context, id uuid.UUID, number string) (Account, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return Account{}, fmt.Errorf("%w: account number is required", ErrInvalidInput)
	}
	return s.accounts.Update(ctx, id, number)
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

func (s *Service) CreateMeter(ctx context.Context, accountID uuid.UUID, serial string, installedOn time.Time, removedOn *time.Time, dual bool) (Meter, error) {
	serial = strings.TrimSpace(serial)
	installedOn, removedOn, err := normalizeMeter(serial, installedOn, removedOn)
	if err != nil {
		return Meter{}, err
	}

	acc, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return Meter{}, err
	}
	if !acc.IsActive {
		return Meter{}, ErrAccountInactive
	}
	if dual && acc.Service != ServiceElectricity {
		return Meter{}, fmt.Errorf("%w: only electricity meters can be dual-tariff", ErrInvalidInput)
	}

	zones := []string{ZoneSingle}
	if dual {
		zones = []string{ZoneDay, ZoneNight}
	}

	meterID, err := uuid.NewV7()
	if err != nil {
		return Meter{}, fmt.Errorf("generate uuid: %w", err)
	}

	registers := make([]Register, 0, len(zones))
	for _, z := range zones {
		regID, err := uuid.NewV7()
		if err != nil {
			return Meter{}, fmt.Errorf("generate uuid: %w", err)
		}
		registers = append(registers, Register{ID: regID, MeterID: meterID, Zone: z})
	}

	return s.meters.Create(ctx, Meter{
		ID:          meterID,
		AccountID:   accountID,
		Serial:      serial,
		InstalledOn: installedOn,
		RemovedOn:   removedOn,
		Registers:   registers,
	})
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

func (s *Service) CreateReadings(ctx context.Context, meterID uuid.UUID, takenOn time.Time, values map[string]decimal.Decimal) ([]Reading, error) {
	if takenOn.IsZero() {
		return nil, fmt.Errorf("%w: reading date is required", ErrInvalidInput)
	}
	takenOn = dateOnly(takenOn)

	meter, err := s.meters.GetByID(ctx, meterID)
	if err != nil {
		return nil, err
	}
	if err := checkInService(meter, takenOn); err != nil {
		return nil, err
	}

	if len(values) != len(meter.Registers) {
		return nil, fmt.Errorf("%w: expected values for %s", ErrInvalidInput, zoneList(meter))
	}

	readings := make([]Reading, 0, len(meter.Registers))
	for _, reg := range meter.Registers {
		value, ok := values[reg.Zone]
		if !ok {
			return nil, fmt.Errorf("%w: missing value for zone %q", ErrInvalidInput, reg.Zone)
		}
		if err := s.checkMonotonic(ctx, reg, takenOn, value, uuid.Nil); err != nil {
			return nil, err
		}

		id, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate uuid: %w", err)
		}
		readings = append(readings, Reading{ID: id, RegisterID: reg.ID, TakenOn: takenOn, Value: value})
	}

	return s.readings.CreateBatch(ctx, readings)
}

func (s *Service) ListReadings(ctx context.Context, meterID uuid.UUID) ([]Reading, error) {
	if _, err := s.meters.GetByID(ctx, meterID); err != nil {
		return nil, err
	}
	return s.readings.ListByMeter(ctx, meterID)
}

func (s *Service) UpdateReading(ctx context.Context, meterID, readingID uuid.UUID, takenOn time.Time, value decimal.Decimal) (Reading, error) {
	if takenOn.IsZero() {
		return Reading{}, fmt.Errorf("%w: reading date is required", ErrInvalidInput)
	}
	takenOn = dateOnly(takenOn)

	meter, rd, reg, err := s.readingOfMeter(ctx, meterID, readingID)
	if err != nil {
		return Reading{}, err
	}
	if err := checkInService(meter, takenOn); err != nil {
		return Reading{}, err
	}
	if err := s.checkMonotonic(ctx, reg, takenOn, value, rd.ID); err != nil {
		return Reading{}, err
	}

	return s.readings.Update(ctx, rd.ID, takenOn, value)
}

func (s *Service) DeleteReading(ctx context.Context, meterID, readingID uuid.UUID) error {
	_, rd, _, err := s.readingOfMeter(ctx, meterID, readingID)
	if err != nil {
		return err
	}
	return s.readings.Delete(ctx, rd.ID)
}

func (s *Service) readingOfMeter(ctx context.Context, meterID, readingID uuid.UUID) (Meter, Reading, Register, error) {
	meter, err := s.meters.GetByID(ctx, meterID)
	if err != nil {
		return Meter{}, Reading{}, Register{}, err
	}
	rd, err := s.readings.GetByID(ctx, readingID)
	if err != nil {
		return Meter{}, Reading{}, Register{}, err
	}
	for _, reg := range meter.Registers {
		if reg.ID == rd.RegisterID {
			return meter, rd, reg, nil
		}
	}
	return Meter{}, Reading{}, Register{}, ErrReadingNotFound
}

func (s *Service) checkMonotonic(ctx context.Context, reg Register, day time.Time, value decimal.Decimal, excludeID uuid.UUID) error {
	if value.IsNegative() {
		return fmt.Errorf("%w: %s value must not be negative", ErrInvalidInput, reg.Zone)
	}

	prev, next, err := s.readings.Neighbors(ctx, reg.ID, day, excludeID)
	if err != nil {
		return err
	}
	if prev != nil && value.LessThan(prev.Value) {
		return fmt.Errorf("%w: %s value %s is below the previous reading %s of %s",
			ErrInvalidInput, reg.Zone, value, prev.Value, prev.TakenOn.Format(dateLayout))
	}
	if next != nil && value.GreaterThan(next.Value) {
		return fmt.Errorf("%w: %s value %s is above the next reading %s of %s",
			ErrInvalidInput, reg.Zone, value, next.Value, next.TakenOn.Format(dateLayout))
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

func zoneList(m Meter) string {
	zones := make([]string, 0, len(m.Registers))
	for _, reg := range m.Registers {
		zones = append(zones, reg.Zone)
	}
	return strings.Join(zones, ", ")
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
