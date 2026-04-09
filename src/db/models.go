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

func WithRrule(rrule rrule.RRule) TaskOption {
	var rrulestr = rrule.String()

	next := rrule.After(time.Now(), false)

	return func(t *Task) {
		t.Rrule = &rrulestr
		t.Dtstart = &rrule.OrigOptions.Dtstart
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
		ID:        gen,
		Active:    true,
		Star:      false,
		Title:     "New task",
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, opt := range opts {
		opt(&task)
	}

	if task.Rrule != nil {

	}

	return task
}

// --- COMPLETION ---

type Completion struct {
	ID             string
	TaskID         string
	OccurrenceDate time.Time
	CompletedAt    time.Time
}

type CompletionOption func(*Completion)

// Opções para Completion
func WithCompletionID(id string) CompletionOption {
	return func(c *Completion) { c.ID = id }
}

func WithOccurrenceDate(date time.Time) CompletionOption {
	return func(c *Completion) { c.OccurrenceDate = date }
}

// Construtor de Completion
func NewCompletion(taskID string, opts ...CompletionOption) Completion {
	gen, err := nanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 21)
	if err != nil {
		logger.Fatal(err)
	}

	now := time.Now()

	comp := Completion{
		ID:             gen,
		TaskID:         taskID, // TaskID é obrigatório, por isso está no argumento fixo
		OccurrenceDate: now,
		CompletedAt:    now,
	}

	for _, opt := range opts {
		opt(&comp)
	}

	return comp
}
