package cmd

import (
	"context"

	"charm.land/huh/v2"
	"github.com/costaluu/remembrall/src/internal/config"
	"github.com/costaluu/remembrall/src/internal/logger"
	"github.com/urfave/cli/v3"
)

var ResetCommand *cli.Command = &cli.Command{
	Name:  "reset",
	Usage: "reset all configurations, clear everything",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var confirm bool = false

		err :=
			huh.NewConfirm().
				Title("Are you sure? This will reset all configurations and clear everything.").
				Affirmative("Yes").
				Negative("No.").
				Value(&confirm)

		if err != nil {
			logger.Fatal(err.Error())
		}

		if !confirm {
			logger.Info("Reset cancelled.")
			return nil
		}

		config.CreateConfig()

		// Handle daemon

		logger.Success("All configurations reset and everything cleared.")

		return nil
	},
}
