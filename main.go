package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jarnoan/vesimittari/csv"
	"github.com/jarnoan/vesimittari/scraper"
	"github.com/jarnoan/vesimittari/updater"
)

type CounterData struct {
	Counter  int
	Date     time.Time
	Customer string
}

func main() {
	var (
		opts        updater.Options
		addCostsCSV string
	)
	flag.StringVar(&addCostsCSV, "add", "", "additional costs CSV file")
	flag.BoolVar(&opts.UpdateMeterReadings, "meter", true, "update meter readings")
	flag.BoolVar(&opts.Verbose, "v", true, "log verbosely")
	flag.Parse()

	acs, err := additionalCosts(addCostsCSV)
	if err != nil {
		log.Fatal("read additional costs csv: %s", err)
	}

	csvf, err := csv.Read(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	scr := scraper.New()
	upd := updater.New(scr, opts)

	if err := upd.Update(csvf, acs); err != nil {
		log.Fatal(err)
	}

	// Write the new data
	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()
	if err := csvf.Write(stdout); err != nil {
		log.Fatal(err)
	}
}

func additionalCosts(filename string) ([]updater.AdditionalCost, error) {
	if filename == "" {
		return nil, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filename, err)
	}
	defer file.Close()

	res, err := csv.ReadAdditionalCosts(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}

	return res, nil
}
