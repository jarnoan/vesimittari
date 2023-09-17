package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jarnoan/vesimittari/updater"
	"github.com/shopspring/decimal"
)

const datefmt = "2.1.2006"

type CSVFile struct {
	headerRow       []string
	meterRows       []MeterRow
	separatorRow    []string
	dateRow         []string
	paymentTimeRow  []string
	mainMeterFeeRow []string
	waterPriceRow   []string
	vatRow          []string
	messageRow      []string
}

func Read(rdr io.Reader) (*CSVFile, error) {
	var res CSVFile
	var err error

	r := csv.NewReader(rdr)

	res.headerRow, err = r.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("read header: %w", err)
	}

	// read meter rows
	for {
		rec, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("read meter row: %w", err)
		}

		if rec[0] == "###" {
			res.separatorRow = rec // separator
			break
		}

		mr := MeterRow{rec}
		res.meterRows = append(res.meterRows, mr)
	}

	// read additional param rows

	// billing date
	res.dateRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read date record: %w", err)
	}

	// paymentTime (days)
	res.paymentTimeRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read payment time record: %w", err)
	}

	// main meter monthly fee, eur / month vat 0%
	res.mainMeterFeeRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read main meter fee record: %w", err)
	}

	// water fee, eur / m3 vat 0%
	res.waterPriceRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read water fee record: %w", err)
	}

	// VAT percentage
	res.vatRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read VAT record: %w", err)
	}

	// message
	res.messageRow, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("read message record: %w", err)
	}

	return &res, nil
}

func (f *CSVFile) MeterRecords() ([]updater.MeterRecord, error) {
	mrs := make([]updater.MeterRecord, len(f.meterRows))
	for i := range f.meterRows {
		mrs[i] = &f.meterRows[i]
	}

	return mrs, nil
}

func (f *CSVFile) CommonVariables() (updater.CommonVariables, error) {
	vat, err := decimal.NewFromString(strings.Replace(f.vatRow[1], ",", ".", 1))
	if err != nil {
		return updater.CommonVariables{}, fmt.Errorf("parse VAT: %w", err)
	}

	mainMeterFee, err := decimal.NewFromString(strings.Replace(f.mainMeterFeeRow[1], ",", ".", 1))
	if err != nil {
		return updater.CommonVariables{}, fmt.Errorf("parse main meter fee: %w", err)
	}

	var meteredCount int64
	for _, mr := range f.meterRows {
		if mnum, _ := mr.MeterNumber(); mnum != "" {
			meteredCount++
		}
	}
	meteredCount-- // less the main meter

	monthly := mainMeterFee.Div(decimal.NewFromInt(meteredCount))

	water, err := decimal.NewFromString(strings.Replace(f.waterPriceRow[1], ",", ".", 1))
	if err != nil {
		return updater.CommonVariables{}, fmt.Errorf("parse water price: %w", err)
	}

	return updater.CommonVariables{
		VAT:        vat,
		MonthlyFee: monthly,
		WaterPrice: water,
	}, nil
}

// Date returns billing date.
func (f *CSVFile) Date() (time.Time, error) {
	res, err := time.Parse(datefmt, f.dateRow[1])
	if err != nil {
		return res, fmt.Errorf("parse date: %w", err)
	}

	return res, nil
}

// PaymentDays returns how many days there's time to pay.
func (f *CSVFile) PaymentDays() (int, error) {
	res, err := strconv.Atoi(f.paymentTimeRow[1])
	if err != nil {
		return res, fmt.Errorf("parse payment time: %w", err)
	}

	return res, nil
}

// Message returns message text.
func (f *CSVFile) Message() (string, error) {
	return f.messageRow[1], nil
}

func (f *CSVFile) SetDate(t time.Time) {
	f.dateRow[1] = t.Format(datefmt)
}

// Write writes the file to the writer.
func (f *CSVFile) Write(wtr io.Writer) error {
	w := csv.NewWriter(wtr)

	if err := w.Write(f.headerRow); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, mr := range f.meterRows {
		if err := w.Write(mr.rec); err != nil {
			return fmt.Errorf("write meter row: %w", err)
		}
	}

	if err := w.Write(f.separatorRow); err != nil {
		return fmt.Errorf("write separator row: %w", err)
	}
	if err := w.Write(f.dateRow); err != nil {
		return fmt.Errorf("write date row: %w", err)
	}
	if err := w.Write(f.paymentTimeRow); err != nil {
		return fmt.Errorf("write payment time row: %w", err)
	}
	if err := w.Write(f.mainMeterFeeRow); err != nil {
		return fmt.Errorf("write main meter fee row: %w", err)
	}
	if err := w.Write(f.waterPriceRow); err != nil {
		return fmt.Errorf("write water fee row: %w", err)
	}
	if err := w.Write(f.vatRow); err != nil {
		return fmt.Errorf("write vat row: %w", err)
	}
	if err := w.Write(f.messageRow); err != nil {
		return fmt.Errorf("write message row: %w", err)
	}

	return nil
}
