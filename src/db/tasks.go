package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/costaluu/taskthing/src/logger"
	rrule "github.com/teambition/rrule-go"
)

// ─── CREATE ──────────────────────────────────────────────────────────────────

func CreateTask(db *sql.DB, task *Task) *Task {
	_, err := db.Exec(`
        INSERT INTO tasks (id, active, title, star, dtstart, rrule, until, count, next_occurrence, completed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, boolToInt(task.Active), task.Title, task.Star,
		nullableTime(task.Dtstart),
		task.Rrule,
		nullableTime(task.Until),
		task.Count,
		nullableTime(task.NextOccurrence),
		nullableTime(task.CompletedAt),
	)

	if err != nil {
		logger.Fatal(err)
	}

	return task
}

// ─── READ ─────────────────────────────────────────────────────────────────────

func GetTask(db *sql.DB, id string) (*Task, error) {
	row := db.QueryRow(`
		SELECT
			id, active, title, star, rrule,
			dtstart, until, count, next_occurrence,
			completed_at, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id)

	task, err := scanTask(row.Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetTask: %w", err)
	}

	return task, nil
}

func ListTasks(db *sql.DB, cut string) ([]*Task, []*Task) {
	now := time.Now()

	var cutTime time.Time

	switch cut {
	case "day":
		cutTime = now.AddDate(0, 0, 1)
	case "week":
		cutTime = now.AddDate(0, 0, 7)
	case "month":
		cutTime = now.AddDate(0, 1, 0)
	case "year":
		cutTime = now.AddDate(1, 0, 0)
	default:
		logger.Fatal(fmt.Sprintf("ListTasks: invalid cut %q (expected day, week, month, year)", cut))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
    SELECT
        id, active, title, star, rrule,
        dtstart, until, count, next_occurrence,
        completed_at, created_at, updated_at
    FROM tasks
    WHERE
        active = 1
        AND (
            (rrule IS NOT NULL AND next_occurrence <= ?)

            OR (rrule IS NULL AND dtstart IS NOT NULL AND DATE(dtstart) <= DATE(?))

            OR (rrule IS NULL AND dtstart IS NULL)
        )
    ORDER BY
        COALESCE(next_occurrence, dtstart) ASC NULLS LAST
`,
		cutTime.Format(time.RFC3339),
		cutTime.Format(time.RFC3339),
	)

	if err != nil {
		logger.Error(err)
		os.Exit(0)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)

	if err != nil {
		logger.Fatal(err)
	}

	active := make([]*Task, 0)
	inactive := make([]*Task, 0)

	for _, task := range tasks {
		if task.CompletedAt != nil {
			inactive = append(inactive, task)
		} else {
			active = append(active, task)
		}
	}

	return active, inactive
}

// ─── UPDATE ──────────────────────────────────────────────────────────────────

func UpdateTask(db *sql.DB, task Task) error {
	_, err := db.Exec(`
        UPDATE tasks 
        SET active = ?, 
            title = ?, 
            star = ?, 
            dtstart = ?, 
            rrule = ?, 
            until = ?, 
            count = ?, 
            next_occurrence = ?, 
            completed_at = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ?`,
		boolToInt(task.Active),
		task.Title,
		task.Star,
		nullableTime(task.Dtstart),
		task.Rrule,
		nullableTime(task.Until),
		task.Count,
		nullableTime(task.NextOccurrence),
		nullableTime(task.CompletedAt),
		task.ID, // O ID vai por último para o WHERE
	)

	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}

// ─── DELETE ──────────────────────────────────────────────────────────────────

func DeleteTask(db *sql.DB, id string) bool {
	res, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)

	if err != nil {
		logger.Fatal(err)
	}

	n, _ := res.RowsAffected()

	if n == 0 {
		return false
	}

	return true
}

// ─── COMPLETE ────────────────────────────────────────────────────────────────

func computeNext(task *Task, from time.Time) *time.Time {
	if task.Rrule == nil {
		return nil
	}

	rOption, err := rrule.StrToROption(*task.Rrule)
	if err != nil {
		logger.Fatal(fmt.Sprintf("computeNext: invalid rrule %q: %v", *task.Rrule, err))
	}

	if task.Dtstart != nil {
		rOption.Dtstart = *task.Dtstart
	}

	rule, err := rrule.NewRRule(*rOption)
	if err != nil {
		logger.Fatal(fmt.Sprintf("computeNext: failed to build rrule: %v", err))
	}

	// After retorna a próxima ocorrência estritamente depois de from
	next := rule.After(from, false)
	if next.IsZero() {
		return nil
	}

	return &next
}

func CompleteTask(db *sql.DB, id string) error {
	task, err := GetTask(db, id)
	if err != nil {
		return fmt.Errorf("CompleteTask: %w", err)
	}
	if task == nil {
		return fmt.Errorf("CompleteTask: task %q not found", id)
	}
	if task.CompletedAt != nil {
		return fmt.Errorf("CompleteTask: task %q is already completed", id)
	}

	now := time.Now()

	// task sem rrule: só marca como completa
	if task.Rrule == nil {
		_, err = db.Exec(`
			UPDATE tasks
			SET completed_at = ?, updated_at = ?
			WHERE id = ?
		`, now.Format(time.RFC3339), now.Format(time.RFC3339), id)
		if err != nil {
			return fmt.Errorf("CompleteTask: %w", err)
		}
		return nil
	}

	// task com rrule: marca como completa e avança next_occurrence
	if task.NextOccurrence == nil {
		return fmt.Errorf("CompleteTask: task %q has rrule but no next_occurrence", id)
	}

	next := computeNext(task, *task.NextOccurrence)

	_, err = db.Exec(`
		UPDATE tasks
		SET
			completed_at  = ?,
			next_occurrence = ?,
			updated_at    = ?
		WHERE id = ?
	`,
		now.Format(time.RFC3339),
		nullableTime(next),
		now.Format(time.RFC3339),
		id,
	)
	if err != nil {
		return fmt.Errorf("CompleteTask: %w", err)
	}

	return nil
}

// ─── COMPLETIONS VIEW ────────────────────────────────────────────────────────

// func ListCompletions(db *sql.DB, taskID *string, limit int) ([]*Completion, error) {
// 	args := []any{}
// 	where := []string{}

// 	if taskID != nil {
// 		where = append(where, "c.task_id = ?")
// 		args = append(args, *taskID)
// 	}

// 	q := `SELECT c.id, c.task_id, c.occurrence_date, c.completed_at, c.note FROM completions c`
// 	if len(where) > 0 {
// 		q += " WHERE " + strings.Join(where, " AND ")
// 	}
// 	q += " ORDER BY c.completed_at DESC"
// 	if limit > 0 {
// 		q += fmt.Sprintf(" LIMIT %d", limit)
// 	}

// 	rows, err := db.Query(q, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var out []*Completion
// 	for rows.Next() {
// 		c := &Completion{}
// 		var occStr, doneStr string
// 		var note sql.NullString
// 		if err := rows.Scan(&c.ID, &c.TaskID, &occStr, &doneStr, &note); err != nil {
// 			return nil, err
// 		}
// 		c.OccurrenceDate, _ = time.Parse(time.RFC3339, occStr)
// 		c.CompletedAt, _ = time.Parse(time.RFC3339, doneStr)
// 		if note.Valid {
// 			c.Note = &note.String
// 		}
// 		out = append(out, c)
// 	}
// 	return out, rows.Err()
// }

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func scanTask(scan func(...any) error) (*Task, error) {
	var task Task
	var active, star int
	var dtstart, until, nextOccurrence, completedAt *string
	var createdAt, updatedAt string

	err := scan(
		&task.ID,
		&active,
		&task.Title,
		&star,
		&task.Rrule,
		&dtstart,
		&until,
		&task.Count,
		&nextOccurrence,
		&completedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	task.Active = active == 1
	task.Star = star == 1

	parseOptional := func(s *string) (*time.Time, error) {
		if s == nil || *s == "" {
			return nil, nil
		}

		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, *s, time.Local); err == nil {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("unrecognized time format: %q", *s)
	}

	parseRequired := func(s string) (time.Time, error) {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unrecognized time format: %q", s)
	}

	if task.Dtstart, err = parseOptional(dtstart); err != nil {
		return nil, fmt.Errorf("dtstart: %w", err)
	}
	if task.Until, err = parseOptional(until); err != nil {
		return nil, fmt.Errorf("until: %w", err)
	}
	if task.NextOccurrence, err = parseOptional(nextOccurrence); err != nil {
		return nil, fmt.Errorf("next_occurrence: %w", err)
	}
	if task.CompletedAt, err = parseOptional(completedAt); err != nil {
		return nil, fmt.Errorf("completed_at: %w", err)
	}
	if task.CreatedAt, err = parseRequired(createdAt); err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	if task.UpdatedAt, err = parseRequired(updatedAt); err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}

	var localTime time.Time = *task.Dtstart

	task.Dtstart = &localTime

	return &task, nil
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		task, err := scanTask(rows.Scan)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
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
	return t.Format(time.RFC3339)
}
