package config

import (
	"fmt"
	"runtime"
	"time"

	"github.com/costaluu/remembrall/src/internal/constants"
	"github.com/costaluu/remembrall/src/internal/filesystem"
	"github.com/costaluu/remembrall/src/internal/logger"
)

type Config struct {
	DateTimeFormat         string    `json:"date_time_format"`
	DatabaseLocation       string    `json:"database_location"`
	LastestVersion         string    `json:"latest_version"`
	LatestVersionCheckTime time.Time `json:"latest_version_check_time"`
	NeedsUpdate            bool      `json:"needs_update"`
}

func CreateConfig() {
	var defaultConfig Config

	switch runtime.GOOS {
	case "windows":
		err := filesystem.FileReadJSONFromFile("src/internal/config/default_config_windows.json", &defaultConfig)

		if err != nil {
			logger.Fatal("Failed to read default config: " + err.Error())
		}
	case "darwin", "linux":
		err := filesystem.FileReadJSONFromFile("src/internal/config/default_config_linux_darwin.json", &defaultConfig)

		if err != nil {
			logger.Fatal("Failed to read default config: " + err.Error())
		}
	default:
		logger.Fatal("Unsupported OS: " + runtime.GOOS)
	}

	// Stop daemon from running if config file is missing

	err := filesystem.FileDeleteFolder(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])

	if err != nil {
		logger.Fatal("Failed to delete config directory: " + err.Error())
	}

	err = filesystem.FileCreateFolder(constants.OS_CONFIGS["APP_CONFIG_FILE_NAME"][runtime.GOOS])

	if err != nil {
		logger.Fatal("Failed to create config directory: " + err.Error())
	}

	err = filesystem.FileWriteJSONToFile(constants.OS_CONFIGS["APP_CONFIG_FILE_NAME"][runtime.GOOS], defaultConfig)

	if err != nil {
		logger.Fatal("Failed to write default config: " + err.Error())
	}
}

func LoadConfig() (Config, error) {
	var config Config

	err := filesystem.FileReadJSONFromFile(constants.OS_CONFIGS["APP_CONFIG_FILE_NAME"][runtime.GOOS], &config)

	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	return config, nil
}

func SaveConfig(config Config) error {
	err := filesystem.FileWriteJSONToFile(constants.OS_CONFIGS["APP_CONFIG_FILE_NAME"][runtime.GOOS], config)

	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
