package main

import (
	"fmt"

	"github.com/costaluu/taskthing/src/logger"
	rruleparser "github.com/costaluu/taskthing/src/rrule-parser"
)

var VERSION = "dev"

func main() {
	// database, err := db.Open()

	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// test := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.Local)

	// rule, _ := rrule.NewRRule(rrule.ROption{
	// 	Freq:  rrule.WEEKLY,
	// 	Count: 10,
	// })

	// result := rule.All()

	// rruleString := rule.OrigOptions.RRuleString()
	// count := 10

	// db.CreateTask(database, "teste", &test, &rruleString, nil, &count)

	// for _, date := range result {
	// 	logger.Info(date.Local().String())
	// }

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

	// var ivals []string
	// cmd := &cli.Command{
	// 	Arguments: []cli.Argument{
	// 		&cli.StringArgs{
	// 			Name:        "someint",
	// 			Min:         0,
	// 			Max:         -1,
	// 			Destination: &ivals,
	// 		},
	// 	},
	// 	Action: func(ctx context.Context, cmd *cli.Command) error {
	// 		fmt.Println("We got ", ivals)
	// 		fmt.Println("We got ", len(ivals))
	// 		return nil
	// 	},
	// }

	// if err := cmd.Run(context.Background(), os.Args); err != nil {
	// 	log.Fatal(err)
	// }

	// parser, err := naturaltime.New()
	// if err != nil {
	// 	panic(err)
	// }

	// now := time.Now()

	// // Example 1: Parse a simple date expression
	// date, err := parser.ParseDate("in a year", now)
	// if err != nil {
	// 	panic(err)
	// }
	// if date != nil {
	// 	fmt.Println(date.Local().String())
	// }

	opts, err := rruleparser.ParseText("Every 6 months")

	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(opts)
}
