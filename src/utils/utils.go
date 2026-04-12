package utils

import (
	"os"
	"regexp"
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

func IsNumeric(s string) bool {
	return regexp.MustCompile(`^[0-9]+$`).MatchString(s)
}

func TrimSpaces(s string) string {
	// Trim leading and trailing spaces
	s = strings.TrimSpace(s)

	// Replace multiple consecutive spaces with a single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return s
}
