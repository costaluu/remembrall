package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/costaluu/remembrall/src/internal/constants"
	"github.com/costaluu/remembrall/src/internal/filesystem"
	"github.com/costaluu/remembrall/src/internal/logger"
)

//go:embed default_config_linux_darwin.json
var defaultConfigLinuxDarwin []byte

//go:embed default_config_windows.json
var defaultConfigWindows []byte

type Config struct {
	DateTimeFormat         string    `json:"date_time_format"`
	DatabaseLocation       string    `json:"database_location"`
	LatestVersion          string    `json:"latest_version"`
	LatestVersionCheckTime time.Time `json:"latest_version_check_time"`
	NeedsUpdate            bool      `json:"needs_update"`
}

func GetDefaultConfig() Config {
	var defaultConfig Config

	switch runtime.GOOS {
	case "linux", "darwin":
		err := json.Unmarshal(defaultConfigLinuxDarwin, &defaultConfig)

		if err != nil {
			logger.Fatal("Failed to read default config: " + err.Error())
		}
	case "windows":
		err := json.Unmarshal(defaultConfigWindows, &defaultConfig)

		if err != nil {
			logger.Fatal("Failed to read default config: " + err.Error())
		}
	default:
		logger.Fatal("Unsupported operating system: " + runtime.GOOS)
	}

	return defaultConfig
}

func GetConfigPathBase() (string, error) {
	configDir, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "remembrall"), nil
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "remembrall", "config.json"), nil
}

func CreateConfig() {
	var configPath string
	var err error

	if configPath, err = getConfigPath(); err != nil {
		logger.Fatal("Failed to get config path: " + err.Error())
	}

	var defaultConfig Config

	err = filesystem.FileReadJSONFromFile(configPath, &defaultConfig)

	// Stop daemon from running if config file is missing

	err = filesystem.FileDeleteFolder(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])

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
	var configPath string
	var err error

	if configPath, err = getConfigPath(); err != nil {
		logger.Fatal("Failed to get config path: " + err.Error())
	}

	var config Config

	err = filesystem.FileReadJSONFromFile(configPath, &config)

	if err != nil {
		return Config{}, fmt.Errorf("failed to read default config: %w", err)
	}

	return config, nil
}

func SaveConfig(config Config) error {
	var configPath string
	var err error

	if configPath, err = getConfigPath(); err != nil {
		logger.Fatal("Failed to get config path: " + err.Error())
	}

	err = filesystem.FileWriteJSONToFile(configPath, config)

	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
