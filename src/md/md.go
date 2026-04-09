package md

import (
	"fmt"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/kvstore"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/teambition/rrule-go"
)

func PrintTasks(tasks []db.Task, title string, verb string, resetKV bool) {
	var tasksClone []db.Task = make([]db.Task, len(tasks))

	copy(tasksClone, tasks)

	store := kvstore.GetInstance(constants.GetPathVariable("APP_KVSTORE_LOCATION"))

	if resetKV {
		store.Reset()
	}

	for id, task := range tasksClone {
		tempId := store.Set(task.ID)

		tasksClone[id].ID = tempId
	}

	markdown := GenerateMarkdown(tasksClone, title, verb)

	var theme string = "dark"

	config := config.LoadConfig()

	if !config.DarkTheme {
		theme = "light"
	}

	out, err := glamour.Render(markdown, theme)

	if err != nil {
		panic(err)
	}

	fmt.Println(out)
}

func GenerateMarkdown(tasks []db.Task, title string, verb string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 📝 %s\n\n", title))
	sb.WriteString("| **🆔 Id** | **📌 Title** | **📅 Start Date** | **🚀 Ocurrences (Max. 3)** |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")

	config := config.LoadConfig()

	for _, task := range tasks {
		startDateStr := ""

		if task.Dtstart != nil {
			startDateStr = task.Dtstart.Format(config.DateTimeFormat)
		}

		occurrences := getOccurrences(task)

		if len(occurrences) == 0 {
			titleCell := formatTitle(task.Title)
			startCell := formatStartDate(startDateStr)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | **No recurrence** |\n", task.ID, titleCell, startCell))
			continue
		}

		// First occurrence row (with id, title and start date)
		titleCell := formatTitle(task.Title)
		startCell := formatStartDate(startDateStr)
		firstOcc := formatOccurrence(1, occurrences[0])
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", task.ID, titleCell, startCell, firstOcc))

		// Remaining occurrences (empty id, title and start date cells)
		for i, occ := range occurrences[1:] {
			occCell := formatOccurrence(i+2, occ)
			sb.WriteString(fmt.Sprintf("| | | | %s |\n", occCell))
		}

		// Empty separator row after recurrences
		sb.WriteString("| | | | |\n")
	}

	sb.WriteString(fmt.Sprintf("\n*Total of `%d` tasks %s.*\n", len(tasks), verb))

	return sb.String()
}

// getOccurrences returns the next N occurrences of the rrule (or empty if nil)
func getOccurrences(task db.Task) []time.Time {
	if task.Rrule == nil {
		return nil
	}

	rrule, err := rrule.StrToRRule(*task.Rrule)

	if err != nil {
		logger.Fatal(err)
	}

	all := rrule.All()

	limit := 3

	if len(all) < limit {
		limit = len(all)
	}

	return all[:limit]
}

// formatTitle wraps title in bold only if NOT isPast
func formatTitle(title string) string {
	return fmt.Sprintf("**%s**", title)
}

// formatStartDate wraps in bold only if NOT isPast
func formatStartDate(date string) string {
	return fmt.Sprintf("**%s**", date)
}

// formatOccurrence formats a single occurrence line
func formatOccurrence(index int, t time.Time) string {
	ordinal := fmt.Sprintf("%dº", index)
	timeStr := fmt.Sprintf("`%s`", t.Format("02/01/2006 15:04"))

	return fmt.Sprintf("**%s** %s", ordinal, timeStr)
}
