package updater

import (
	"fmt"
	"time"

	"github.com/jarnoan/vesimittari/meter"
)

type Data interface {
	MeterRecords() ([]MeterRecord, error)
	SetDate(time.Time)
}

// MeterRecord defines the methods needed from a single meter record.
type MeterRecord interface {
	MeterNumber() (meter.Number, error)
	SiteNumber() (meter.SiteNumber, error)
	AddReading(MeterReading) error
}

type ConsumptionReader interface {
	ReadConsumption(meter.SiteNumber, meter.Number) (MeterReading, error)
}

type MeterReading struct {
	Counter  int
	Date     time.Time
	Customer string
}

//Updater updates the Data.
type Updater struct {
	cr ConsumptionReader
}

// New constructs a new updater.
func New(cr ConsumptionReader) *Updater {
	return &Updater{cr}
}

// Update reads the consumptions and updates the data accordingly.
func (u *Updater) Update(d Data) error {
	// Read the meterings and update the meter records
	mrs, err := d.MeterRecords()
	if err != nil {
		return fmt.Errorf("read meter records: %w", err)
	}

	for _, mr := range mrs {
		num, err := mr.MeterNumber()
		if err != nil {
			return fmt.Errorf("get meter number: %w", err)
		}
		if num == "" {
			continue
		}

		site, err := mr.SiteNumber()
		if err != nil {
			return fmt.Errorf("get site number: %w", err)
		}

		cd, err := u.cr.ReadConsumption(site, num)
		if err != nil {
			return fmt.Errorf("read consumption of meter %s: %w", num, err)
		}
		// fmt.Printf("mr before %+v\n", mr)
		// fmt.Printf("consumption %+v\n", cd)

		if err := mr.AddReading(cd); err != nil {
			return fmt.Errorf("add consumption for meter %s: %w", num, err)
		}
		// fmt.Printf("mr after %+v\n", mr)
	}

	d.SetDate(time.Now())

	return nil
}
