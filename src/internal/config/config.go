package config

import (
	"runtime"

	"github.com/costaluu/remembrall/src/internal/constants"
	"github.com/costaluu/remembrall/src/internal/filesystem"
)

type Config struct {
	dateTimeFormat   string
	databaseLocation string
}

func CreateConfig() {
	var defaultConfig Config = Config{
		dateTimeFormat:   "2006-01-02 15:04:05",
		databaseLocation: constants.OS_CONFIGS["APP_DB_FILE_NAME"][runtime.GOOS],
	}

	err := filesystem.FileDeleteFolder(constants.OS_CONFIGS["APP_DIR"][runtime.GOOS])

	if err != nil {

	}
}
