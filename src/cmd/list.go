package cmd

import (
	"context"
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/urfave/cli/v3"
)

type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
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

func newThemeStyles(theme Theme) themeStyles {
	if theme == ThemeLight {
		return themeStyles{
			title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1).Margin(0, 0, 1, 0),
			labelToday:    lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Background(lipgloss.Color("224")).Padding(0, 1),
			labelTomorrow: lipgloss.NewStyle().Foreground(lipgloss.Color("20")).Background(lipgloss.Color("153")).Padding(0, 1),
			labelOther:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("252")).Padding(0, 1),
			labelOverdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("224")).Padding(0, 1),
			labelNone:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("252")).Padding(0, 1),
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
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1).Margin(0, 0, 1, 0),
		labelToday:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.Color("236")).Padding(0, 1),
		labelTomorrow: lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Background(lipgloss.Color("236")).Padding(0, 1),
		labelOther:    lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Background(lipgloss.Color("236")).Padding(0, 1),
		labelOverdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Background(lipgloss.Color("236")).Padding(0, 1),
		labelNone:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236")).Padding(0, 1),
		star:          lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		checkedBox:    lipgloss.NewStyle().Foreground(lipgloss.Color("28")),
		uncheckedBox:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		taskText:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		taskChecked:   lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Strikethrough(true),
		index:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		recurring:     lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
	}
}

func dayLabel(s themeStyles, t *time.Time) lipgloss.Style {
	if t == nil {
		return s.labelNone
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(target.Sub(today).Hours() / 24)

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
	if t == nil {
		return "No date"
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(target.Sub(today).Hours() / 24)

	switch {
	case diff < -1:
		return t.Format("Mon 2 Jan")
	case diff == -1:
		return "Yesterday"
	case diff == 0:
		return "Today"
	case diff == 1:
		return "Tomorrow"
	case diff < 7:
		return t.Format("Mon")
	default:
		return t.Format("Mon 2 Jan")
	}
}

func PrintTasks(tasks []db.Task, theme Theme) {
	s := newThemeStyles(theme)

	fmt.Println()
	fmt.Println(s.title.Render(" Tasks "))

	if len(tasks) == 0 {
		fmt.Println(s.taskText.Render("  No tasks found."))
		fmt.Println()
		return
	}

	for i, task := range tasks {
		index := s.index.Render(fmt.Sprintf("%2d.", i+1))

		var checkbox string
		if !task.Active {
			checkbox = s.checkedBox.Render(" ")
		} else {
			checkbox = s.uncheckedBox.Render(" ")
		}

		dateRef := task.NextOccurrence
		if dateRef == nil {
			dateRef = task.Dtstart
		}
		label := dayLabel(s, dateRef).Render(" " + dayLabelText(dateRef) + " ")

		star := ""
		if task.Star {
			star = " " + s.star.Render("")
		}

		recurring := ""
		if task.Rrule != nil {
			recurring = " " + s.recurring.Render("↻")
		}

		var taskText string
		if !task.Active {
			taskText = s.taskChecked.Render(task.Title)
		} else {
			taskText = s.taskText.Render(task.Title)
		}

		fmt.Printf("  %s %s %s%s %s%s\n", index, checkbox, label, star, taskText, recurring)
		fmt.Println()
	}
}

var ListCommand *cli.Command = &cli.Command{
	Name:  "list",
	Usage: "list tasks",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "all",
			Value: false,
			Usage: "show all tasks including the not active.",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		tasks := db.ListTasks(database, cmd.Bool("all"))
		var realTasks []db.Task = make([]db.Task, 0)

		for _, task := range tasks {
			if task != nil {
				realTasks = append(realTasks, *task)
			}
		}

		config := config.LoadConfig()

		var theme string = string(ThemeDark)

		if !config.DarkTheme {
			theme = string(ThemeLight)
		}

		PrintTasks(realTasks, Theme(theme))

		return nil
	},
}
