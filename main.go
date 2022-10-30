package main

import (
	"bufio"
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
	scr := scraper.New()
	upd := updater.New(scr)

	csvf, err := csv.Read(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

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
