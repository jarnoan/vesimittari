package main

import (
	"bufio"
	"flag"
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
	var opts updater.Options
	flag.BoolVar(&opts.UpdateMeterReadings, "meter", true, "update meter readings")
	flag.BoolVar(&opts.Verbose, "v", true, "log verbosely")
	flag.Parse()

	csvf, err := csv.Read(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	scr := scraper.New()
	upd := updater.New(scr, opts)

	if err := upd.Update(csvf); err != nil {
		log.Fatal(err)
	}

	// Write the new data
	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()
	if err := csvf.Write(stdout); err != nil {
		log.Fatal(err)
	}
}
