package db

import "time"

type Task struct {
	ID             string
	Active         bool
	Title          string
	Dtstart        *time.Time
	Rrule          *string
	Until          *time.Time
	Count          *int
	NextOccurrence *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Completion struct {
	ID             string
	TaskID         string
	OccurrenceDate time.Time
	CompletedAt    time.Time
	Note           *string
}
