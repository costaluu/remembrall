package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/filesystem"
	"github.com/costaluu/taskthing/src/logger"
)

//go:embed default_config.json
var defaultConfigLinuxDarwin []byte

type Config struct {
	DateTimeFormat         string    `json:"date_time_format"`
	DatabaseLocation       string    `json:"database_location"`
	LatestVersion          string    `json:"latest_version"`
	LatestVersionCheckTime time.Time `json:"latest_version_check_time"`
	NeedsUpdate            bool      `json:"needs_update"`
	DarkTheme              bool      `json:"dark_theme"`
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

	if os.Getenv("DEV_MODE") != "true" && filesystem.FolderExists(constants.GetPathVariable("APP_DIR")) {
		err = filesystem.FileDeleteFolder(constants.GetPathVariable("APP_DIR"))

		if err != nil {
			logger.Fatal("Failed to delete config directory: " + err.Error())
		}
	}

	if os.Getenv("DEV_MODE") != "true" {
		err = filesystem.FileCreateFolder(constants.GetPathVariable("APP_DIR"))

		if err != nil {
			logger.Fatal("Failed to create config directory: " + err.Error())
		}
	}

	err = filesystem.FileWriteJSONToFile(constants.GetPathVariable("APP_CONFIG_LOCATION"), defaultConfig)

	if err != nil {
		logger.Fatal("Failed to write default config: " + err.Error())
	}
}

func LoadConfig() Config {
	var err error

	var config Config

	err = filesystem.FileReadJSONFromFile(constants.GetPathVariable("APP_CONFIG_LOCATION"), &config)

	if err != nil {
		logger.Fatal(err)
	}

	return config
}

func SaveConfig(config Config) error {
	var err error

	err = filesystem.FileWriteJSONToFile(constants.GetPathVariable("APP_CONFIG_LOCATION"), config)

	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
