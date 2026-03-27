package cmd

import (
	"context"
	"fmt"

	"charm.land/huh/v2"
	"github.com/costaluu/remembrall/src/cli/internal/config"
	"github.com/costaluu/remembrall/src/cli/internal/constants"
	"github.com/costaluu/remembrall/src/cli/internal/logger"
	"github.com/urfave/cli/v3"
)

func SetTimeFormat(timeFormat string) {
	currentConfig, err := config.LoadConfig()

	if err != nil {
		logger.Fatal("Failed to get config: " + err.Error())
	}

	if timeFormat == "American Format" {
		currentConfig.DateTimeFormat = "01/02/2006 15:04:05"
	} else if timeFormat == "European Format" {
		currentConfig.DateTimeFormat = "02/01/2006 15:04:05"
	}

	err = config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}
}

var EuropeanTimeFormatCommand *cli.Command = &cli.Command{
	Name:  "european-time",
	Usage: "set the time format to european (DD/MM/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTimeFormat("European Format")

		logger.Success("Time format set to european (DD/MM/YYYY HH:mm:ss).")

		return nil
	},
}

var AmericanTimeFormatCommand *cli.Command = &cli.Command{
	Name:  "american-time",
	Usage: "set the time format to american (MM/DD/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTimeFormat("American Format")

		logger.Success("Time format set to american (MM/DD/YYYY HH:mm:ss).")

		return nil
	},
}

var SetTimeFormatCommand *cli.Command = &cli.Command{
	Name:     "set-time-format",
	Usage:    "set the time format for the application",
	Commands: []*cli.Command{EuropeanTimeFormatCommand, AmericanTimeFormatCommand},
}

func SetDatabaseLocationCommandAction(ctx context.Context, cmd *cli.Command) error {
	currentConfig, err := config.LoadConfig()

	if err != nil {
		logger.Fatal("Failed to get config: " + err.Error())
	}

	var newLocation string
	var oldLocation string

	if currentConfig.DatabaseLocation == "" {
		defaultConfig := config.GetDefaultConfig()
		oldLocation = defaultConfig.DatabaseLocation
	}

	huh.NewInput().
		Title("What's the new database location?").
		Prompt("?").
		Placeholder(oldLocation).
		Value(&newLocation).
		Run()

	if newLocation == "" {
		currentConfig.DatabaseLocation = oldLocation
	} else {
		currentConfig.DatabaseLocation = newLocation
	}

	err = config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}

	logger.Success("Database location set to: " + currentConfig.DatabaseLocation)

	return nil
}

var SetDatabaseLocationCommand *cli.Command = &cli.Command{
	Name:   "set-database-location",
	Usage:  "set the location of the database file",
	Action: SetDatabaseLocationCommandAction,
}

var ConfigCommands *cli.Command = &cli.Command{
	Name:     "config",
	Usage:    fmt.Sprintf("command to check and apply updates to %s.", constants.APP_NAME),
	Commands: []*cli.Command{SetTimeFormatCommand, SetDatabaseLocationCommand},
}
