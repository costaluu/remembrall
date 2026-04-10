package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/kvstore"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/urfave/cli/v3"
)

type themeStyles struct {
	title         lipgloss.Style
	labelToday    lipgloss.Style
	labelTomorrow lipgloss.Style
	labelOther    lipgloss.Style
	labelOverdue  lipgloss.Style
	labelNone     lipgloss.Style
	star          lipgloss.Style
	checkedBox    lipgloss.Style
	uncheckedBox  lipgloss.Style
	taskText      lipgloss.Style
	taskChecked   lipgloss.Style
	index         lipgloss.Style
	recurring     lipgloss.Style
}

func newThemeStyles(theme constants.Theme) themeStyles {
	if theme == constants.ThemeLight {
		return themeStyles{
			title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1),
			labelToday:    lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Background(lipgloss.Color("224")),
			labelTomorrow: lipgloss.NewStyle().Foreground(lipgloss.Color("20")).Background(lipgloss.Color("153")),
			labelOther:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("252")),
			labelOverdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("224")),
			labelNone:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("252")),
			star:          lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
			checkedBox:    lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
			uncheckedBox:  lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
			taskText:      lipgloss.NewStyle().Foreground(lipgloss.Color("232")),
			taskChecked:   lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Strikethrough(true),
			index:         lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
			recurring:     lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		}
	}

	// dark (default)
	return themeStyles{
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228")).Background(lipgloss.Color("63")).Padding(0, 1).Margin(1),
		labelToday:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.Color("236")),
		labelTomorrow: lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Background(lipgloss.Color("236")),
		labelOther:    lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Background(lipgloss.Color("236")),
		labelOverdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Background(lipgloss.Color("236")),
		labelNone:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236")),
		star:          lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		checkedBox:    lipgloss.NewStyle().Foreground(lipgloss.Color("28")),
		uncheckedBox:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		taskText:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		taskChecked:   lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Strikethrough(true),
		index:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		recurring:     lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
	}
}

func dayLabel(s themeStyles, t *time.Time) lipgloss.Style {
	if t == nil {
		return s.labelNone
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(math.Round(target.Sub(today).Hours() / 24))

	switch {
	case diff < 0:
		return s.labelOverdue
	case diff == 0:
		return s.labelToday
	case diff == 1:
		return s.labelTomorrow
	default:
		return s.labelOther
	}
}

func dayLabelText(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(math.Round(target.Sub(today).Hours() / 24))

	if diff == -1 {
		return "Yesterday"
	}

	if target.Year() > today.Year() {
		return t.Format("2006")
	}

	diff = int(math.Abs(float64(diff)))

	switch {
	case diff == 0:
		return "Today"
	case diff == 1:
		return "Tomorrow"
	case diff <= 7:
		return t.Format("Mon")
	case diff > 30:
		return t.Format("Jan")
	default:
		return t.Format("Mon 2 Jan")
	}
}

func PrintTasks(tasks []db.Task, completedTasks []db.Task, showCompleted bool, cut string, theme constants.Theme) {
	s := newThemeStyles(theme)

	fmt.Println()
	fmt.Println(s.title.Render(" Tasks "))

	if len(tasks) == 0 {
		if cut == "day" {
			fmt.Println(s.taskText.Render(fmt.Sprintf("  no tasks for today")))
		} else if cut == "week" {
			fmt.Println(s.taskText.Render(fmt.Sprintf("  no tasks for this week")))
		} else if cut == "month" {
			fmt.Println(s.taskText.Render(fmt.Sprintf("  no tasks for this month")))
		} else {
			fmt.Println(s.taskText.Render(fmt.Sprintf("  no tasks for this year")))
		}

		return
	}

	store := kvstore.GetInstance(constants.GetPathVariable("APP_KVSTORE_LOCATION"))

	store.Reset()

	if showCompleted {
		for _, task := range completedTasks {
			checkbox := s.checkedBox.Render("")

			taskText := s.taskChecked.Render(task.Title)

			star := ""

			if task.Star {
				star = " " + s.star.Render("")
			}

			fmt.Printf("  %s%s %s (%s)\n", checkbox, star, taskText, s.labelToday.Render(task.CompletedAt.Local().String()))
			fmt.Println()
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color("246")).SetString("___")

		fmt.Println(style.Render())
		fmt.Println()
	}

	for i, task := range tasks {
		index := s.index.Render(fmt.Sprintf("%2d.", i+1))

		var checkbox string

		checkbox = s.uncheckedBox.Render("")

		dateRef := task.NextOccurrence

		if dateRef == nil {
			dateRef = task.Dtstart
		}

		label := dayLabel(s, dateRef).Render(" " + dayLabelText(dateRef) + " ")

		if dateRef == nil || dateRef.IsZero() {
			label = ""
		}

		star := ""

		if task.Star {
			star = " " + s.star.Render("")
		}

		recurring := ""
		if task.Rrule != nil {
			recurring = " " + s.recurring.Render("")
		}

		var taskText string
		taskText = s.taskText.Render(task.Title)

		fmt.Printf("  %s %s %s%s %s%s\n", index, checkbox, label, star, taskText, recurring)
		fmt.Println()

		store.Set(task.ID)
	}
}

var ListCommand *cli.Command = &cli.Command{
	Name:  "list",
	Usage: "list tasks",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "cut",
			Value: "day",
			Usage: "show tasks in the cut provided. (day|week|month|year) defaults to day.",
		},
		&cli.BoolFlag{
			Name:  "completed",
			Value: false,
			Usage: "show completed tasks",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		cut := cmd.String("cut")

		if !slices.Contains([]string{"day", "week", "month", "year"}, cut) {
			logger.Warning("invalid option use day|week|month|year")
			os.Exit(0)
		}

		activeTasks, inativeTasks := db.ListTasks(database, cut)

		var realActiveTasks []db.Task = make([]db.Task, 0)
		var realInactiveTasks []db.Task = make([]db.Task, 0)

		for _, task := range activeTasks {
			if task != nil {
				realActiveTasks = append(realActiveTasks, *task)
			}
		}

		for _, task := range inativeTasks {
			if task != nil {
				realInactiveTasks = append(realInactiveTasks, *task)
			}
		}

		config := config.LoadConfig()

		var theme string = string(constants.ThemeDark)

		if !config.DarkTheme {
			theme = string(constants.ThemeLight)
		}

		PrintTasks(realActiveTasks, realInactiveTasks, cmd.Bool("completed"), cut, constants.Theme(theme))

		return nil
	},
}
