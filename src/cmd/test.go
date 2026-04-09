package cmd

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/urfave/cli/v3"
)

var TestCommand *cli.Command = &cli.Command{
	Name:  "test",
	Usage: "test",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Margin(1)

		var todayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Background(lipgloss.Color("236"))

		var tomorrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Background(lipgloss.Color("236"))

		var otherDaysStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Background(lipgloss.Color("236"))

		var starStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).SetString("")

		var secondaryCheckedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Strikethrough(true)

		var secondaryTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

		var checkedBox = lipgloss.NewStyle().Foreground(lipgloss.Color("28")).SetString(" ")

		lipgloss.Println(titleStyle.Render(" Tasks "))

		lipgloss.Println(" ", "1.", " ", todayStyle.Render(" Today "), starStyle.Render(), "tarefa teste asoijdoiajdoiasjdoi jaosjd oaijsdoi jasodij aosidjoasijdoiasj do", secondaryTextStyle.Render(" "))

		fmt.Println("")

		lipgloss.Println(" ", "2.", " ", tomorrowStyle.Render(" Tomorrow "), "tarefa teste")

		fmt.Println("")

		lipgloss.Println(" ", "3.", checkedBox.Render(), otherDaysStyle.Render(" Fri "), secondaryCheckedTextStyle.Render("tarefa teste"))

		return nil
	},
}
