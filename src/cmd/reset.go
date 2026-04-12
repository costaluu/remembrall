package cmd

import (
	"context"

	"charm.land/huh/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/urfave/cli/v3"
)

var ResetCommand *cli.Command = &cli.Command{
	Name:  "reset",
	Usage: "reset all configurations, clear everything",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var confirm bool = false

		err :=
			huh.NewConfirm().
				Title("Are you sure? This will reset all configurations and clear all tasks.").
				Affirmative("Yes").
				Negative("No").
				Value(&confirm).
				Run()

		if err != nil {
			logger.Fatal(err.Error())
		}

		if !confirm {
			logger.Info("aborting...")
			return nil
		}

		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		db.TruncateTasks(database)

		config.CreateConfig()

		logger.Success("tasks deleted and default configurations applied!")

		return nil
	},
}
