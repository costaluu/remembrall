package utils

import (
	"os"
	"strings"

	"github.com/costaluu/taskthing/src/logger"
)

func ReplaceTildeWithHomeDir(path string) string {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		logger.Fatal(err)
	}

	return strings.Replace(path, "~", homeDir, 1)
}
