package constants

import (
	"os"
	"runtime"

	"charm.land/lipgloss/v2"
	"github.com/costaluu/taskthing/src/utils"
)

var (
	APP_NAME     = "taskthing"
	COMMAND      = "tt"
	GITHUB_OWNER = "costaluu"
	GITHUB_REPO  = "taskthing"
)

var OS_CONFIGS = map[string]map[string]string{
	"APP_DIR": {
		"linux":     "~/.config/taskthing",
		"darwin":    "~/.config/taskthing",
		"windows":   "%LOCALAPPDATA%\\taskthing",
		"dev_linux": ".",
	},
	"APP_CONFIG_LOCATION": {
		"linux":     "~/.config/taskthing/config.json",
		"darwin":    "~/.config/taskthing/config.json",
		"windows":   "%LOCALAPPDATA%\\taskthing\\config.json",
		"dev_linux": "./config.json",
	},
	"APP_BINARY_LOCATION": {
		"linux":     "~/.local/bin/taskthing",
		"darwin":    "~/.local/bin/taskthing",
		"windows":   "%LOCALAPPDATA%\\taskthing\\taskthing.exe",
		"dev_linux": "~/.local/bin/taskthing",
	},
	"APP_KVSTORE_LOCATION": {
		"linux":     "~/.config/taskthing/id_store.json",
		"darwin":    "~/.config/taskthing/id_store.json",
		"windows":   "%LOCALAPPDATA%\\taskthing\\id_store.json",
		"dev_linux": "./id_store.json",
	},
}

func GetPathVariable(variable string) string {
	var ostring string = runtime.GOOS

	if os.Getenv("DEV_MODE") == "true" {
		ostring = "dev_" + ostring
	}

	result := utils.ReplaceTildeWithHomeDir(OS_CONFIGS[variable][ostring])

	return result
}

var (
	CheckMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark       = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("⨯")
	InfoMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("31")).SetString("ⓘ")
	WarningMark = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString("⚠")
)
