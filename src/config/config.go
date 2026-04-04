package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/filesystem"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
)

//go:embed default_config.json
var defaultConfigLinuxDarwin []byte

type Config struct {
	DateTimeFormat         string    `json:"date_time_format"`
	DatabaseLocation       string    `json:"database_location"`
	LatestVersion          string    `json:"latest_version"`
	LatestVersionCheckTime time.Time `json:"latest_version_check_time"`
	NeedsUpdate            bool      `json:"needs_update"`
}

func GetDefaultConfig() Config {
	var defaultConfig Config

	err := json.Unmarshal(defaultConfigLinuxDarwin, &defaultConfig)

	if err != nil {
		logger.Fatal("Failed to read default config: " + err.Error())
	}

	return defaultConfig
}

func CreateConfig() {
	var err error
	var defaultConfig Config = GetDefaultConfig()

	if filesystem.FolderExists(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])) {
		err = filesystem.FileDeleteFolder(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])

		if err != nil {
			logger.Fatal("Failed to delete config directory: " + err.Error())
		}
	}

	err = filesystem.FileCreateFolder(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])

	if err != nil {
		logger.Fatal("Failed to create config directory: " + err.Error())
	}

	err = filesystem.FileWriteJSONToFile(constants.OS_CONFIGS["APP_CONFIG_LOCATION"][runtime.GOOS], defaultConfig)

	if err != nil {
		logger.Fatal("Failed to write default config: " + err.Error())
	}
}

func LoadConfig() Config {
	var err error

	var config Config

	err = filesystem.FileReadJSONFromFile(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_CONFIG_LOCATION"][runtime.GOOS]), &config)

	if err != nil {
		logger.Fatal(err)
	}

	return config
}

func SaveConfig(config Config) error {
	var err error

	err = filesystem.FileWriteJSONToFile(utils.ReplaceTildeWithHomeDir(constants.OS_CONFIGS["APP_CONFIG_LOCATION"][runtime.GOOS]), config)

	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
