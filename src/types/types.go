package types

import (
	"time"

	"github.com/teambition/rrule-go"
)

type Candidate struct {
	Title   string
	Star    bool
	Rrule   *rrule.RRule
	Dtstart *time.Time
	IsPast  bool
}
