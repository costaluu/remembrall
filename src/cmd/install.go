package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"charm.land/huh/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/filesystem"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/urfave/cli/v3"
)

func pathsAlreadyInPATH() bool {
	var targets []string

	switch runtime.GOOS {
	case "linux", "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		targets = []string{
			filepath.Join(home, ".local", "bin"),
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return false
		}
		targets = []string{
			filepath.Join(localAppData, "Programs", "taskthing"),
		}
	default:
		return false
	}

	pathEnv := os.Getenv("PATH")
	pathDirs := filepath.SplitList(pathEnv) // usa ":" no unix e ";" no windows automaticamente

	for _, target := range targets {
		found := false
		for _, dir := range pathDirs {
			if filepath.Clean(dir) == filepath.Clean(target) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

var InstallCommand *cli.Command = &cli.Command{
	Name:  "install",
	Usage: "install the application",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		logger.Info("valiting initial setup...")
		logger.Info("checking config path folder...")

		configFolderExists := filesystem.FolderExists(constants.GetPathVariable("APP_DIR"))

		if !configFolderExists {
			logger.Info("config folder does not exist. creating config folder...")

			filesystem.FileCreateFolder(constants.GetPathVariable("APP_DIR"))
		}

		logger.Info("checking config path file...")

		configFileExists := filesystem.FileExists(constants.GetPathVariable("APP_CONFIG_LOCATION"))

		if !configFileExists {
			logger.Info("config file does not exist. creating config file...")

			config.CreateConfig()
		}

		pathsCorrect := pathsAlreadyInPATH()

		if !pathsCorrect {
			logger.Warning("The installation path is not in your PATH environment variable. Please add it to be able to run taskthing from anywhere.")
			logger.Info("Installation path: " + constants.GetPathVariable("APP_BINARY_LOCATION"))
		}

		var timeFormat string = "European Format"

		huh.NewSelect[string]().
			Options(huh.NewOptions("American Format", "European Format")...).
			Value(&timeFormat).
			Title("Select your preferred time format").
			Run()

		SetTimeFormat(timeFormat)

		var theme constants.Theme = constants.ThemeLight

		huh.NewSelect[string]().
			Options(huh.NewOptions("Dark Theme", "Light Theme")...).
			Value(&timeFormat).
			Title("Select your preferred terminal theme").
			Run()

		if theme == "Light Theme" {
			SetTheme(constants.ThemeLight)
		}

		logger.Success("time format set to " + timeFormat + ".")

		SetDatabaseLocationCommandAction(ctx, cmd)

		logger.Success("setup completed successfully!")

		return nil
	},
}
