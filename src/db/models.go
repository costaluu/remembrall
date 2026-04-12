package db

import (
	"time"

	"github.com/costaluu/taskthing/src/logger"
	nanoid "github.com/matoous/go-nanoid/v2"
	rrule "github.com/teambition/rrule-go"
)

// --- TASK ---

type Task struct {
	ID             string
	Active         bool
	Title          string
	Star           bool
	Dtstart        *time.Time
	Rrule          *string
	Until          *time.Time
	Count          *int
	NextOccurrence *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
}

type TaskOption func(*Task)

func WithTitle(title string) TaskOption {
	return func(t *Task) { t.Title = title }
}

func WithActive(active bool) TaskOption {
	return func(t *Task) { t.Active = active }
}

func WithStar(star bool) TaskOption {
	return func(t *Task) { t.Star = star }
}

func WithDtstart(dtstart *time.Time) TaskOption {
	return func(t *Task) { t.Dtstart = dtstart }
}

func WithCompletedAt(completedAt *time.Time) TaskOption {
	return func(t *Task) { t.CompletedAt = completedAt }
}

func WithRrule(rrule rrule.RRule) TaskOption {
	var rrulestr = rrule.String()
	var dtStartLocal time.Time = rrule.OrigOptions.Dtstart.Local()

	next := rrule.After(dtStartLocal, true)

	return func(t *Task) {
		t.Rrule = &rrulestr
		t.Dtstart = &dtStartLocal
		t.Until = &rrule.OrigOptions.Until
		t.Count = &rrule.OrigOptions.Count

		if next.IsZero() {
			t.NextOccurrence = nil
		} else {
			t.NextOccurrence = &next
		}
	}
}

func WithCreatedAt(createdAt time.Time) TaskOption {
	return func(t *Task) { t.CreatedAt = createdAt }
}

func WithUpdatedAt(updatedAt time.Time) TaskOption {
	return func(t *Task) { t.UpdatedAt = updatedAt }
}

// Construtor de Task
func NewTask(opts ...TaskOption) Task {
	gen, err := nanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 21)
	if err != nil {
		logger.Fatal(err)
	}

	now := time.Now()

	task := Task{
		ID:          gen,
		Active:      true,
		Star:        false,
		Title:       "New task",
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: nil,
	}

	for _, opt := range opts {
		opt(&task)
	}

	if task.Rrule != nil {

	}

	return task
}
