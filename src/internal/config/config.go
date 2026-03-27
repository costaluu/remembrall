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
	DateTimeFormat         string
	DatabaseLocation       string
	LastestVersion         string
	LatestVersionCheckTime time.Time
	NeedsUpdate            bool
}

func CreateConfig() {
	var defaultConfig Config = Config{
		DateTimeFormat:         "2006-01-02 15:04:05",
		DatabaseLocation:       constants.OS_CONFIGS["APP_DB_FILE_NAME"][runtime.GOOS],
		LastestVersion:         "0.0.0",
		LatestVersionCheckTime: time.Now(),
		NeedsUpdate:            false,
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
