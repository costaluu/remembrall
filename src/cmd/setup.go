package cmd

import (
	"context"
	"runtime"

	"charm.land/huh/v2"
	spinner "charm.land/huh/v2/spinner"
	"github.com/costaluu/remembrall/src/internal/constants"
	"github.com/urfave/cli/v3"
)

var SetupCommand *cli.Command = &cli.Command{
	Name:  "setup",
	Usage: "makes the initial setup of remembrall, creating necessary files and folders.",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var dbLocation string
		var dbLocationFallback string = constants.OS_CONFIGS["APP_DB_FILE_NAME"][runtime.GOOS]

		huh.NewInput().
			Title("Where would you like to store the database?").
			Placeholder(dbLocationFallback).
			Value(&dbLocation).
			Run()

		if dbLocation == "" {
			dbLocation = dbLocationFallback
		}

		runner := func() {

		}

		_ = spinner.New().
			Title("looking for updates...").
			Type(spinner.Dots).
			Action(runner).
			Run()

		return nil
	},
}
