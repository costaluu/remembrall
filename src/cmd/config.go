package cmd

import (
	"context"
	"fmt"
	"os"

	"charm.land/huh/v2"
	spinner "charm.land/huh/v2/spinner"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
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
	defaultConfig := config.GetDefaultConfig()

	oldDatabaseConnection, err := db.Open()

	if err != nil {
		logger.Fatal(err)
	}

	var newLocation string = defaultConfig.DatabaseLocation

	var databaseLocationOption string = "Local"

	huh.NewSelect[string]().
		Options(huh.NewOptions("Local", "Remote (LibSQL)")...).
		Value(&databaseLocationOption).
		Title("Where's your database ?").
		Run()

	if databaseLocationOption == "Remote (LibSQL)" {
		var baseURL string = ""
		var jwtToken string = ""

		for {
			baseURL = ""
			jwtToken = ""

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Database base URL").
						Prompt(">").
						Value(&baseURL),
				),
				huh.NewGroup(
					huh.NewInput().
						Title("Database Token").
						Prompt(">").
						Value(&jwtToken),
				),
			)

			err := form.Run()

			if err != nil {
				logger.Fatal(err)
			}

			if baseURL != "" || jwtToken != "" {
				break
			}
		}

		newLocation = fmt.Sprintf("%s/?authToken=%s", baseURL, jwtToken)
	}

	if os.Getenv("DEV_MODE") == "true" {
		currentConfig.DatabaseLocation = "./dev.db"
	} else {
		currentConfig.DatabaseLocation = newLocation
	}

	err = config.SaveConfig(currentConfig)

	if err != nil {
		logger.Fatal("Failed to save config: " + err.Error())
	}

	logger.Success("database location set to: " + currentConfig.DatabaseLocation)

	newDatabaseConnection, err := db.Open()

	if err != nil {
		logger.Fatal(err)
	}

	var isOldDatabaseEmpty bool = !db.SchemaMigrationsExists(oldDatabaseConnection)
	var isNewDatabaseEmpty bool = !db.SchemaMigrationsExists(newDatabaseConnection)

	if isOldDatabaseEmpty {
		logger.Info("applying migrations...")

		db.ApplyAllMigrations()

		logger.Success("migrations applied successfully")
	} else {
		if isNewDatabaseEmpty {
			var migrateEverything bool

			err = huh.NewConfirm().
				Title("New database is empty, do you want to migrate your data? (recommended)").
				Affirmative("Yes").
				Negative("No").
				Value(&migrateEverything).
				Run()

			if migrateEverything {
				db.ApplyAllMigrations()

				runner := func() {
					db.MigrateDatabases(oldDatabaseConnection, newDatabaseConnection)
				}

				_ = spinner.New().
					Title("migrating data...").
					Type(spinner.Dots).
					Action(runner).
					Run()

				logger.Success("database migrated!")
			}
		} else {
			logger.Info("applying peding database migrations...")

			appliedAnyMigration := db.ApplyMigrations()

			if !appliedAnyMigration {
				logger.Info("no migrations to apply")
			}
		}
	}

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
