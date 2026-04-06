package main

import (
	"time"

	"github.com/costaluu/taskthing/src/logger"
	"github.com/teambition/rrule-go"
)

var VERSION = "dev"

func main() {
	// database, err := db.Open()

	// if err != nil {
	// 	logger.Fatal(err)
	// }

	test := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.Local)

	rule, _ := rrule.NewRRule(rrule.ROption{
		Freq:    rrule.WEEKLY,
		Count:   10,
		Dtstart: test,
	})

	result := rule.All()

	// rruleString := rule.OrigOptions.RRuleString()
	// count := 10

	// db.CreateTask(database, "teste", &test, &rruleString, nil, &count)

	for _, date := range result {
		logger.Info(date.Local().String())
	}

	// app := &cli.Command{
	// 	Name:    constants.APP_NAME,
	// 	Version: VERSION,
	// 	Authors: []any{
	// 		mail.Address{Name: "costaluu", Address: "costaluu@email.com"},
	// 	},
	// 	Usage: "taskthing is a terminal-based reminders app",
	// 	Commands: []*cli.Command{
	// 		cmd.InstallCommand,
	// 		cmd.UpdateCommands,
	// 		cmd.ResetCommand,
	// 		cmd.ConfigCommands,
	// 	},
	// }

	// if err := app.Run(context.Background(), os.Args); err != nil {
	// 	log.Fatal(err)
	// }
}
