package cmd

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"charm.land/huh/v2"
	"github.com/costaluu/remembrall/src/cli/internal/config"
	"github.com/costaluu/remembrall/src/cli/internal/constants"
	"github.com/costaluu/remembrall/src/cli/internal/filesystem"
	"github.com/costaluu/remembrall/src/cli/internal/logger"
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
			filepath.Join(localAppData, "Programs", "remembrall"),
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

var SetupCommand *cli.Command = &cli.Command{
	Name:  "setup",
	Usage: "setup the application",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var configPathBase string
		var err error

		configPathBase, err = config.GetConfigPathBase()

		if err != nil {
			logger.Fatal("Failed to get config path: " + err.Error())
		}

		logger.Info("valiting initial setup...")
		logger.Info("Checking config path folder...")

		configFolderExists := filesystem.FolderExists(configPathBase)

		if !configFolderExists {
			logger.Info("Config folder does not exist. Creating config folder...")

			filesystem.FileCreateFolder(path.Join(configPathBase))
		}

		configFileExists := filesystem.FileExists(path.Join(configPathBase, "config.json"))

		if !configFileExists {
			logger.Info("Config file does not exist. Creating config file...")

			config.CreateConfig()
		}

		pathsCorrect := pathsAlreadyInPATH()

		if !pathsCorrect {
			logger.Warning("The installation path is not in your PATH environment variable. Please add it to be able to run remembrall from anywhere.")
			logger.Info("Installation path: " + constants.OS_CONFIGS["APP_BINARY_LOCATION"][runtime.GOOS])
		}

		var timeFormat string = "European Format"

		huh.NewSelect[string]().
			Options(huh.NewOptions("American Format", "European Format")...).
			Value(&timeFormat).
			Title("Select your preferred time format").
			Run()

		SetTimeFormat(timeFormat)

		logger.Success("Time format set to " + timeFormat + ".")

		SetDatabaseLocationCommandAction(ctx, cmd)

		logger.Success("Setup completed successfully!")

		return nil
	},
}
