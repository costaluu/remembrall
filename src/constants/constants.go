package constants

import (
	"charm.land/lipgloss/v2"
)

var (
	APP_NAME     = "taskthing"
	COMMAND      = "tt"
	GITHUB_OWNER = "costaluu"
	GITHUB_REPO  = "taskthing"
)

var OS_CONFIGS = map[string]map[string]string{
	"APP_DIR": {
		"linux":   "~/.config/taskthing",
		"darwin":  "~/.config/taskthing",
		"windows": "%LOCALAPPDATA%\\taskthing",
	},
	"APP_CONFIG_LOCATION": {
		"linux":   "~/.config/taskthing/config.json",
		"darwin":  "~/.config/taskthing/config.json",
		"windows": "%LOCALAPPDATA%\\taskthing\\config.json",
	},
	"APP_BINARY_LOCATION": {
		"linux":   "~/.local/bin/taskthing",
		"darwin":  "~/.local/bin/taskthing",
		"windows": "%LOCALAPPDATA%\\taskthing\\taskthing.exe",
	},
}

var (
	CheckMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark       = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("⨯")
	InfoMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("31")).SetString("ⓘ")
	WarningMark = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString("⚠")
)
