package updater

import (
	"fmt"
	"log"
	"time"

	"github.com/jarnoan/vesimittari/meter"
	"github.com/jarnoan/vesimittari/reference"
	"github.com/shopspring/decimal"
)

type Data interface {
	MeterRecords() ([]MeterRecord, error)
	SetDate(time.Time)
	CommonVariables() (CommonVariables, error)
}

// CommonVariables contains general values needed for fee calculations.
type CommonVariables struct {
	VAT        decimal.Decimal // %
	MonthlyFee decimal.Decimal // € without tax (päämittarin kuukausimaksu)
	WaterPrice decimal.Decimal // €/m³ without tax
}

// AdditionalCost is some cost that is shared between all shareholders, like insurance.
type AdditionalCost struct {
	Description string
	VAT         decimal.Decimal // %
	Cost        decimal.Decimal // € without tax
}

// MeterRecord defines the methods needed from a single meter record.
type MeterRecord interface {
	Name() string
	MeterNumber() (meter.Number, error)
	SiteNumber() (meter.SiteNumber, error)
	Reference() reference.Number
	AddReading(meter.Reading) error
	UpdateBilling(reference.Number, CommonVariables, []AdditionalCost) error
}

type MeterReader interface {
	ReadMeter(meter.SiteNumber, meter.Number) (meter.Reading, error)
}

type Options struct {
	Verbose             bool
	UpdateMeterReadings bool
}

//Updater updates the Data.
type Updater struct {
	meterReader MeterReader
	opts        Options
}

// New constructs a new updater.
func New(mr MeterReader, opts Options) *Updater {
	return &Updater{mr, opts}
}

// Update reads the consumptions and updates the data accordingly.
func (u *Updater) Update(d Data, acs []AdditionalCost) error {
	mrs, err := d.MeterRecords()
	if err != nil {
		return fmt.Errorf("read meter records: %w", err)
	}

	cv, err := d.CommonVariables()
	if err != nil {
		return fmt.Errorf("get common variables: %w", err)
	}

	var lastRef reference.Number
	for _, mr := range mrs {
		if lastRef == "" || mr.Reference() > lastRef {
			lastRef = mr.Reference()
		}
	}

	// Calculate additional costs per member
	memberCount := decimal.NewFromInt(int64(len(mrs) - 1)) // exclude main meter
	acsPerMember := make([]AdditionalCost, len(acs))
	for i, c := range acs {
		acsPerMember[i] = c
		acsPerMember[i].Cost = c.Cost.DivRound(memberCount, 2)
	}
	if u.opts.Verbose {
		log.Printf("common variables: %+v", cv)
		log.Printf("additional costs: %+v", acsPerMember)
		log.Printf("last reference: %s\n", lastRef)
	}

	// Read the meterings and update the meter records
	for i, mr := range mrs {
		num, err := mr.MeterNumber()
		if err != nil {
			return fmt.Errorf("get meter number: %w", err)
		}

		site, err := mr.SiteNumber()
		if err != nil {
			return fmt.Errorf("get site number: %w", err)
		}

		if u.opts.UpdateMeterReadings && num != "" {
			log.Printf("reading meter for %s", mr.Name())
			r, err := u.meterReader.ReadMeter(site, num)
			if err != nil {
				return fmt.Errorf("read meter %s: %w", num, err)
			}

			if err := mr.AddReading(r); err != nil {
				return fmt.Errorf("add reading for meter %s: %w", num, err)
			}
		}

		if i > 0 { // skip the main meter row
			ref := lastRef.Next()
			lastRef = ref

			if err := mr.UpdateBilling(ref, cv, acsPerMember); err != nil {
				return fmt.Errorf("update billing for %s: %w", mr.Name(), err)
			}
		}
	}

	d.SetDate(time.Now())

	return nil
}
