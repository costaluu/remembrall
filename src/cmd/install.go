package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/huh/v2"
	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/filesystem"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
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

func isProcessRunning(processName string) bool {
	isRunningLinux := func(name string) bool {
		out, err := exec.Command("pgrep", "-f", name).Output()
		return err == nil && len(out) > 0
	}

	isRunningWindows := func(name string) bool {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name)).Output()
		return err == nil && strings.Contains(string(out), name)
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return isRunningLinux(processName)
	case "windows":
		return isRunningWindows(processName)
	default:
		return false
	}
}

var InstallCommand *cli.Command = &cli.Command{
	Name:  "install",
	Usage: "install the application",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		logger.Info("valiting initial setup...")
		logger.Info("checking config path folder...")

		configFolderExists := filesystem.FolderExists(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS]))

		if !configFolderExists {
			logger.Info("config folder does not exist. creating config folder...")

			filesystem.FileCreateFolder(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS]))
		}

		logger.Info("checking config path file...")

		configFileExists := filesystem.FileExists(path.Join(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS]), "config.json"))

		if !configFileExists {
			logger.Info("config file does not exist. creating config file...")

			config.CreateConfig()
		}

		pathsCorrect := pathsAlreadyInPATH()

		if !pathsCorrect {
			logger.Warning("The installation path is not in your PATH environment variable. Please add it to be able to run taskthing from anywhere.")
			logger.Info("Installation path: " + constants.OS_CONFIGS["APP_BINARY_LOCATION"][runtime.GOOS])
		}

		var timeFormat string = "European Format"

		huh.NewSelect[string]().
			Options(huh.NewOptions("American Format", "European Format")...).
			Value(&timeFormat).
			Title("Select your preferred time format").
			Run()

		SetTimeFormat(timeFormat)

		logger.Success("time format set to " + timeFormat + ".")

		SetDatabaseLocationCommandAction(ctx, cmd)

		logger.Info("applying migrations...")

		db.ApplyAllMigrations()

		logger.Success("migrations applied successfully")

		logger.Success("setup completed successfully!")

		return nil
	},
}
