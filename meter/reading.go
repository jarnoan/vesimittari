package meter

import "time"

type Reading struct {
	Counter  int
	Date     time.Time
	Customer string
}
