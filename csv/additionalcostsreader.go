package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/jarnoan/vesimittari/updater"
	"github.com/shopspring/decimal"
)

// ReadAdditionalCosts reads the additional costs from a CSV file.
// The expected columns are description, cost, vat %.
func ReadAdditionalCosts(rdr io.Reader) ([]updater.AdditionalCost, error) {
	r := csv.NewReader(rdr)

	// read header row
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("read header row: %w", err)
	}

	// read additional cost rows
	var res []updater.AdditionalCost
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			// all rows read
			return res, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		cost, err := decimal.NewFromString(row[1])
		if err != nil {
			return nil, fmt.Errorf("read cost column: %w", err)
		}

		vat, err := decimal.NewFromString(row[2])
		if err != nil {
			return nil, fmt.Errorf("read VAT column: %w", err)
		}

		ac := updater.AdditionalCost{
			Description: row[0],
			Cost:        cost,
			VAT:         vat,
		}
		res = append(res, ac)
	}

	return res, nil
}
