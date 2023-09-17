package csv

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jarnoan/vesimittari/meter"
	"github.com/jarnoan/vesimittari/reference"
	"github.com/jarnoan/vesimittari/updater"
	"github.com/shopspring/decimal"
)

const (
	colName = iota
	colBankAccount
	colPhone
	colEmail
	colStreetAddress
	colPostalCode
	colCity
	colPropertyID
	colTenants
	colPermanentResidency
	colJoinDate
	colLeaveDate
	colSite
	colMeter
	colPrevCounter
	colPrevDate
	colCounter
	colDate
	colCheck
	colConsumption
	colWaterFeeWithoutTax
	colWaterTax
	colWaterFeeWithTax
	colMonths
	colBasicFeeWithoutTax
	colBasicFeeTax
	colBasicFeeWithTax
	colExtraDescription
	colExtraCost
	colTotal
	colReference
)

type MeterRow struct {
	rec []string // raw csv records
}

func (r *MeterRow) Name() string {
	return r.rec[colName]
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

func (r *MeterRow) AddReading(rdg meter.Reading) error {
	prevCounter, err := strconv.Atoi(r.rec[colCounter])
	if err != nil {
		return fmt.Errorf("parse previous counter: %w", err)
	}
	cons := rdg.Counter - prevCounter

	r.rec[colPrevCounter] = r.rec[colCounter]
	r.rec[colPrevDate] = r.rec[colDate]
	r.rec[colCounter] = strconv.Itoa(rdg.Counter)
	r.rec[colDate] = rdg.Date.Format(datefmt)
	r.rec[colCheck] = rdg.Customer
	r.rec[colConsumption] = strconv.Itoa(cons)

	return nil
}

var hundred = decimal.NewFromInt(100)

func (r *MeterRow) UpdateBilling(ref reference.Number, cv updater.CommonVariables, acs []updater.AdditionalCost) error {
	var total decimal.Decimal

	// Basic fee and consumption are billed only from members who have a water meter
	if r.rec[colConsumption] != "" {
		months, err := r.months()
		if err != nil {
			return fmt.Errorf("get months: %w", err)
		}

		basicFeeWithoutTax := cv.MonthlyFee.Mul(decimal.NewFromInt(int64(months)))
		basicFeeTax := basicFeeWithoutTax.Mul(cv.VAT.Div(hundred))
		basicFeeWithTax := basicFeeWithoutTax.Add(basicFeeTax)
		total = total.Add(basicFeeWithTax)

		r.rec[colMonths] = strconv.Itoa(months)
		r.rec[colBasicFeeWithoutTax] = decimalToString(basicFeeWithoutTax)
		r.rec[colBasicFeeTax] = decimalToString(basicFeeTax)
		r.rec[colBasicFeeWithTax] = decimalToString(basicFeeWithTax)

		cons, err := strconv.ParseInt(r.rec[colConsumption], 10, 64)
		if err != nil {
			return fmt.Errorf("parse consumption: %w", err)
		}

		waterFeeWithoutTax := decimal.NewFromInt(int64(cons)).Mul(cv.WaterPrice)
		waterTax := waterFeeWithoutTax.Mul(cv.VAT.Div(hundred))
		waterFeeWithTax := waterFeeWithoutTax.Add(waterTax)
		total = total.Add(waterFeeWithTax)

		r.rec[colWaterFeeWithoutTax] = decimalToString(waterFeeWithoutTax)
		r.rec[colWaterTax] = decimalToString(waterTax)
		r.rec[colWaterFeeWithTax] = decimalToString(waterFeeWithTax)
	}

	// Additional costs are billed from all members
	var acsTotal decimal.Decimal
	for _, ac := range acs {
		acTax := ac.Cost.Mul(ac.VAT.Div(hundred))
		acsTotal = acsTotal.Add(ac.Cost).Add(acTax)
	}
	r.rec[colExtraCost] = decimalToString(acsTotal)
	total = total.Add(acsTotal)

	r.rec[colTotal] = decimalToString(total)
	r.rec[colReference] = string(ref)

	return nil
}

func (r *MeterRow) months() (int, error) {
	prevDate, err := time.Parse(datefmt, r.rec[colPrevDate])
	if err != nil {
		return 0, fmt.Errorf("parse previous date: %w", err)
	}

	meterDate, err := time.Parse(datefmt, r.rec[colDate])
	if err != nil {
		return 0, fmt.Errorf("parse meter date: %w", err)
	}

	mos := meterDate.Month() - prevDate.Month()
	if mos < 0 {
		mos += 12
	}

	return int(mos), nil
}

func decimalToString(d decimal.Decimal) string {
	return strings.Replace(d.StringFixed(2), ".", ",", 1)
}
