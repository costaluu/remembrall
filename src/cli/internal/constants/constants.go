package constants

import (
	"runtime"

	"charm.land/lipgloss/v2"
)

var OS_CONFIGS = map[string]map[string]string{
	"APP_DIR": {
		"windows": "%APPDATA%\\remembrall",
		"darwin":  "/.config/remembrall",
		"linux":   "/.config/remembrall",
	},
	"APP_CONFIG_FILE_NAME": {
		"windows": "%APPDATA%\\remembrall\\config.json",
		"darwin":  "/.config/remembrall/config.json",
		"linux":   "/.config/remembrall/config.json",
	},
	"APP_DB_FILE_NAME": {
		"windows": "%APPDATA%\\remembrall\\remembrall.db",
		"darwin":  "/.config/remembrall/remembrall.db",
		"linux":   "/.config/remembrall/remembrall.db",
	},
	"APP_BINARY_LOCATION": {
		"windows": "%LOCALAPPDATA%\\Programs\\remembrall.exe",
		"darwin":  "/.local/bin/remembrall",
		"linux":   "/.local/bin/remembrall",
	},
	"APP_DAEMON_LOCATION": {
		"windows": "%LOCALAPPDATA%\\Programs\\remembralld.exe",
		"darwin":  "/.local/bin/remembralld",
		"linux":   "/.local/bin/remembralld",
	},
}

var (
	APP_NAME                = "remembrall"
	COMMAND                 = "rb"
	GITHUB_OWNER            = "costaluu"
	GITHUB_REPO             = "remembrall"
	APP_ACCENT_COLOR        = "#f97900"
	APP_ACCENT_COLOR_DARKER = "#c76000"
	APP_CONFIG_FILE_NAME    = OS_CONFIGS["APP_CONFIG_FILE_NAME"][runtime.GOOS]
)

var (
	CheckMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark       = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("⨯")
	InfoMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("31")).SetString("ⓘ")
	WarningMark = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString("⚠")
)
