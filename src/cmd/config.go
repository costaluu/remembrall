package cmd

import (
	"context"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/urfave/cli/v3"
)

func SetTheme(theme constants.Theme) {
	currentConfig := config.LoadConfig()

	if theme == constants.ThemeDark {
		currentConfig.DarkTheme = true
	} else {
		currentConfig.DarkTheme = false
	}

	err := config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}
}

var DarkThemeCommand *cli.Command = &cli.Command{
	Name:  "set-dark-theme",
	Usage: "set the time format to european (DD/MM/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTheme(constants.ThemeDark)

		logger.Success("dark theme activated.")

		return nil
	},
}

var LightThemeCommand *cli.Command = &cli.Command{
	Name:  "set-light-theme",
	Usage: "set the time format to american (MM/DD/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTheme(constants.ThemeLight)

		logger.Success("light theme activated.")

		return nil
	},
}

func SetTimeFormat(timeFormat string) {
	currentConfig := config.LoadConfig()

	if timeFormat == "American Format" {
		currentConfig.DateTimeFormat = "01/02/2006 15:04:05"
	} else if timeFormat == "European Format" {
		currentConfig.DateTimeFormat = "02/01/2006 15:04:05"
	}

	err := config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}
}

var EuropeanTimeFormatCommand *cli.Command = &cli.Command{
	Name:  "set-european-time",
	Usage: "set the time format to european (DD/MM/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTimeFormat("European Format")

		logger.Success("Time format set to european (DD/MM/YYYY HH:mm:ss).")

		return nil
	},
}

var AmericanTimeFormatCommand *cli.Command = &cli.Command{
	Name:  "set-american-time",
	Usage: "set the time format to american (MM/DD/YYYY HH:mm:ss)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		SetTimeFormat("American Format")

		logger.Success("Time format set to american (MM/DD/YYYY HH:mm:ss).")

		return nil
	},
}

func SetDatabaseLocationCommandAction(ctx context.Context, cmd *cli.Command) error {
	currentConfig := config.LoadConfig()
	var newLocation string
	var oldLocation string = currentConfig.DatabaseLocation

	if currentConfig.DatabaseLocation == "" {
		defaultConfig := config.GetDefaultConfig()
		oldLocation = defaultConfig.DatabaseLocation
	}

	huh.NewInput().
		Title("Where's your database ?").
		Prompt("?").
		Placeholder(oldLocation).
		Value(&newLocation).
		Run()

	if newLocation == "" {
		currentConfig.DatabaseLocation = oldLocation
	} else {
		currentConfig.DatabaseLocation = newLocation
	}

	if os.Getenv("DEV_MODE") == "true" {
		currentConfig.DatabaseLocation = "./dev.db"
	}

	err := config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}

	logger.Success("database location set to: " + currentConfig.DatabaseLocation)

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
	Commands: []*cli.Command{AmericanTimeFormatCommand, EuropeanTimeFormatCommand, DarkThemeCommand, LightThemeCommand, SetDatabaseLocationCommand},
}
