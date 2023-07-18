package csv

import (
	"fmt"
	"strconv"

	"github.com/jarnoan/vesimittari/meter"
	"github.com/jarnoan/vesimittari/reference"
)

const (
	colName        = 0
	colSite        = 14
	colMeter       = 15
	colPrevCounter = 16
	colPrevDate    = 17
	colCounter     = 18
	colDate        = 19
	colCheck       = 20
	colConsumption = 21
	colReference   = 31
)

type MeterRow struct {
	rec []string // raw csv records
}

func (r *MeterRow) MeterNumber() (meter.Number, error) {
	return meter.Number(r.rec[colMeter]), nil
}

func (r *MeterRow) SiteNumber() (meter.SiteNumber, error) {
	return meter.SiteNumber(r.rec[colSite]), nil
}

func (r *MeterRow) Reference() reference.Number {
	return reference.Number(r.rec[colReference])
}

func (r *MeterRow) AddReading(rdg meter.Reading, ref reference.Number) error {
	prevCounter, err := strconv.Atoi(r.rec[colCounter])
	if err != nil {
		return fmt.Errorf("parse previous counter: %w", err)
	}
	cons := rdg.Counter - prevCounter

	r.rec[colPrevCounter] = r.rec[colCounter]
	r.rec[colPrevDate] = r.rec[colDate]
	r.rec[colReference] = string(ref)
	r.rec[colCounter] = strconv.Itoa(rdg.Counter)
	r.rec[colDate] = rdg.Date.Format(datefmt)
	r.rec[colCheck] = rdg.Customer
	r.rec[colConsumption] = strconv.Itoa(cons)

	return nil
}
