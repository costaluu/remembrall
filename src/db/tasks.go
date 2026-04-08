package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ─── CREATE ──────────────────────────────────────────────────────────────────

func CreateTask(db *sql.DB, title string, dtstart *time.Time, rruleStr *string, until *time.Time, count *int) (*Task, error) {
	task := &Task{
		ID:      newID(),
		Active:  true,
		Title:   title,
		Dtstart: dtstart,
		Rrule:   rruleStr,
		Until:   until,
		Count:   count,
	}

	next := computeNext(task, time.Now().Add(-time.Second))

	if next == nil && rruleStr == nil {
		task.NextOccurrence = dtstart
	} else {
		task.NextOccurrence = next
	}

	_, err := db.Exec(`
        INSERT INTO tasks (id, active, title, dtstart, rrule, until, count, next_occurrence)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, boolToInt(task.Active), task.Title,
		nullableTime(task.Dtstart),
		task.Rrule,
		nullableTime(task.Until),
		task.Count,
		nullableTime(task.NextOccurrence),
	)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	return task, nil
}

// ─── READ ─────────────────────────────────────────────────────────────────────

func GetTask(db *sql.DB, id string) (*Task, error) {
	row := db.QueryRow(`SELECT id, active, title, dtstart, rrule, until, count, next_occurrence, created_at, updated_at FROM tasks WHERE id = ?`, id)
	return scanTask(row)
}

func ListTasks(db *sql.DB, onlyActive bool) ([]*Task, error) {
	query := `SELECT id, active, title, dtstart, rrule, until, count, next_occurrence, created_at, updated_at FROM tasks`
	if onlyActive {
		query += ` WHERE active = 1`
	}
	query += ` ORDER BY next_occurrence ASC NULLS LAST`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

// DailyTasks returns overdue tasks and tasks due today (by next_occurrence).
func DailyTasks(db *sql.DB) (overdue []*Task, today []*Task, err error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := db.Query(`
		SELECT id, active, title, dtstart, rrule, until, count, next_occurrence, created_at, updated_at
		FROM tasks
		WHERE active = 1 AND next_occurrence IS NOT NULL AND next_occurrence < ?
		ORDER BY next_occurrence ASC`,
		endOfDay.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	all, err := scanTasks(rows)
	if err != nil {
		return nil, nil, err
	}

	for _, task := range all {
		if task.NextOccurrence.Before(startOfDay) {
			overdue = append(overdue, task)
		} else {
			today = append(today, task)
		}
	}

	return overdue, today, nil
}

// ─── UPDATE ──────────────────────────────────────────────────────────────────

func UpdateTask(db *sql.DB, id string, title *string, active *bool, rruleStr *string) (*Task, error) {
	task, err := GetTask(db, id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		task.Title = *title
	}
	if active != nil {
		task.Active = *active
	}
	if rruleStr != nil {
		task.Rrule = rruleStr
	}

	// Recompute next_occurrence
	next := computeNext(task, time.Now().Add(-time.Second))
	task.NextOccurrence = next

	_, err = db.Exec(`
		UPDATE tasks SET title = ?, active = ?, rrule = ?, next_occurrence = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		task.Title, boolToInt(task.Active), task.Rrule, nullableTime(task.NextOccurrence), id,
	)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return task, nil
}

// ─── DELETE ──────────────────────────────────────────────────────────────────

func DeleteTask(db *sql.DB, id string) error {
	res, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task %q not found", id)
	}
	return nil
}

// ─── COMPLETE ────────────────────────────────────────────────────────────────

// CompleteTask marks the current next_occurrence as done, advances to the next one,
// and writes a completion record.
func CompleteTask(db *sql.DB, id string, note *string) (*Completion, error) {
	task, err := GetTask(db, id)
	if err != nil {
		return nil, err
	}
	if task.NextOccurrence == nil {
		return nil, fmt.Errorf("task %q has no pending occurrence", id)
	}

	occDate := *task.NextOccurrence

	// Advance next_occurrence past the one we just completed
	next := computeNext(task, occDate)

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint

	completion := &Completion{
		ID:             newID(),
		TaskID:         id,
		OccurrenceDate: occDate,
		CompletedAt:    time.Now(),
		Note:           note,
	}

	_, err = tx.Exec(`
		INSERT INTO completions (id, task_id, occurrence_date, completed_at, note)
		VALUES (?, ?, ?, ?, ?)`,
		completion.ID, completion.TaskID,
		completion.OccurrenceDate.UTC().Format(time.RFC3339),
		completion.CompletedAt.UTC().Format(time.RFC3339),
		note,
	)
	if err != nil {
		return nil, fmt.Errorf("insert completion: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE tasks SET next_occurrence = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		nullableTime(next), id,
	)
	if err != nil {
		return nil, fmt.Errorf("advance next_occurrence: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return completion, nil
}

// ─── COMPLETIONS VIEW ────────────────────────────────────────────────────────

func ListCompletions(db *sql.DB, taskID *string, limit int) ([]*Completion, error) {
	args := []any{}
	where := []string{}

	if taskID != nil {
		where = append(where, "c.task_id = ?")
		args = append(args, *taskID)
	}

	q := `SELECT c.id, c.task_id, c.occurrence_date, c.completed_at, c.note FROM completions c`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY c.completed_at DESC"
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Completion
	for rows.Next() {
		c := &Completion{}
		var occStr, doneStr string
		var note sql.NullString
		if err := rows.Scan(&c.ID, &c.TaskID, &occStr, &doneStr, &note); err != nil {
			return nil, err
		}
		c.OccurrenceDate, _ = time.Parse(time.RFC3339, occStr)
		c.CompletedAt, _ = time.Parse(time.RFC3339, doneStr)
		if note.Valid {
			c.Note = &note.String
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func scanTask(row *sql.Row) (*Task, error) {
	task := &Task{}
	var activeInt int
	var dtstartStr sql.NullString
	var createdStr, updatedStr string
	var rruleStr sql.NullString
	var untilStr, nextStr sql.NullString
	var count sql.NullInt64

	err := row.Scan(
		&task.ID, &activeInt, &task.Title, &dtstartStr,
		&rruleStr,
		&untilStr, &count, &nextStr,
		&createdStr, &updatedStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, err
	}

	task.Active = activeInt == 1

	if dtstartStr.Valid {
		t, _ := time.Parse(time.RFC3339, dtstartStr.String)
		task.Dtstart = &t
	}

	task.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	task.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	if rruleStr.Valid {
		task.Rrule = &rruleStr.String
	}

	if untilStr.Valid {
		u, _ := time.Parse(time.RFC3339, untilStr.String)
		task.Until = &u
	}
	if count.Valid {
		c := int(count.Int64)
		task.Count = &c
	}
	if nextStr.Valid {
		n, _ := time.Parse(time.RFC3339, nextStr.String)
		task.NextOccurrence = &n
	}
	return task, nil
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var out []*Task
	for rows.Next() {
		task := &Task{}
		var activeInt int
		var dtstartStr, createdStr, updatedStr string
		var rruleStr sql.NullString
		var untilStr, nextStr sql.NullString
		var count sql.NullInt64

		err := rows.Scan(
			&task.ID, &activeInt, &task.Title, &dtstartStr,
			&rruleStr,
			&untilStr, &count, &nextStr,
			&createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}

		task.Active = activeInt == 1

		var dtStart time.Time

		dtStart, _ = time.Parse(time.RFC3339, dtstartStr)

		task.Dtstart = &dtStart
		task.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		task.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

		if rruleStr.Valid {
			task.Rrule = &rruleStr.String
		}

		if untilStr.Valid {
			u, _ := time.Parse(time.RFC3339, untilStr.String)
			task.Until = &u
		}

		if count.Valid {
			c := int(count.Int64)
			task.Count = &c
		}

		if nextStr.Valid {
			n, _ := time.Parse(time.RFC3339, nextStr.String)
			task.NextOccurrence = &n
		}

		out = append(out, task)
	}

	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
