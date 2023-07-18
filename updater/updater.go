package updater

import (
	"fmt"
	"os"
	"time"

	"github.com/jarnoan/vesimittari/meter"
	"github.com/jarnoan/vesimittari/reference"
)

type Data interface {
	MeterRecords() ([]MeterRecord, error)
	SetDate(time.Time)
}

// MeterRecord defines the methods needed from a single meter record.
type MeterRecord interface {
	MeterNumber() (meter.Number, error)
	SiteNumber() (meter.SiteNumber, error)
	Reference() reference.Number
	AddReading(meter.Reading, reference.Number) error
}

type MeterReader interface {
	ReadMeter(meter.SiteNumber, meter.Number) (meter.Reading, error)
}

type Options struct {
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
func (u *Updater) Update(d Data) error {
	mrs, err := d.MeterRecords()
	if err != nil {
		return fmt.Errorf("read meter records: %w", err)
	}

	// Find the largest reference number
	var lastRef reference.Number
	for _, mr := range mrs {
		if lastRef == "" || mr.Reference() > lastRef {
			lastRef = mr.Reference()
		}
	}
	fmt.Fprintf(os.Stderr, "last reference: %s\n", lastRef)

	// Read the meterings and update the meter records
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

		if u.opts.UpdateMeterReadings {
			cd, err := u.meterReader.ReadMeter(site, num)
			if err != nil {
				return fmt.Errorf("read consumption of meter %s: %w", num, err)
			}
			// fmt.Printf("mr before %+v\n", mr)
			// fmt.Printf("consumption %+v\n", cd)

			var ref reference.Number
			if mr.Reference() != "" {
				ref = lastRef.Next()
				lastRef = ref
			}
			if err := mr.AddReading(cd, ref); err != nil {
				return fmt.Errorf("add consumption for meter %s: %w", num, err)
			}
			// fmt.Printf("mr after %+v\n", mr)
		}
	}

	d.SetDate(time.Now())

	return nil
}
