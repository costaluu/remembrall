package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"charm.land/huh/v2"
	spinner "charm.land/huh/v2/spinner"
	"github.com/costaluu/remembrall/src/cli/internal/config"
	"github.com/costaluu/remembrall/src/cli/internal/constants"
	"github.com/costaluu/remembrall/src/cli/internal/filesystem"
	"github.com/costaluu/remembrall/src/cli/internal/logger"
	"github.com/emersion/go-autostart"
	"github.com/urfave/cli/v3"
)

func pathsAlreadyInPATH() bool {
	var targets []string

	switch runtime.GOOS {
	case "linux", "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		targets = []string{
			filepath.Join(home, ".local", "bin"),
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return false
		}
		targets = []string{
			filepath.Join(localAppData, "Programs", "remembrall"),
		}
	default:
		return false
	}

	pathEnv := os.Getenv("PATH")
	pathDirs := filepath.SplitList(pathEnv) // usa ":" no unix e ";" no windows automaticamente

	for _, target := range targets {
		found := false
		for _, dir := range pathDirs {
			if filepath.Clean(dir) == filepath.Clean(target) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func isProcessRunning(processName string) bool {
	isRunningLinux := func(name string) bool {
		out, err := exec.Command("pgrep", "-f", name).Output()
		return err == nil && len(out) > 0
	}

	isRunningWindows := func(name string) bool {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name)).Output()
		return err == nil && strings.Contains(string(out), name)
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return isRunningLinux(processName)
	case "windows":
		return isRunningWindows(processName)
	default:
		return false
	}
}

func startDaemon() error {
	logger.Info("starting daemon...")

	if runtime.GOOS != "windows" {
		homeDir, err := os.UserHomeDir()

		if err != nil {
			logger.Fatal("failed to get user config directory: " + err.Error())
		}

		exePath := filepath.Join(homeDir, constants.OS_CONFIGS["APP_DAEMON_LOCATION"][runtime.GOOS])

		cmd := exec.Command(exePath)

		devNull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
		cmd.Stdout = devNull
		cmd.Stderr = devNull
		cmd.Stdin = nil

		// Configuração importante para rodar como daemon em background
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true, // Cria nova sessão (o mais importante para detach)
			// Foreground: false,  // não use isso para background
		}

		if err := cmd.Start(); err != nil {
			logger.Fatal(fmt.Sprintf("falha ao iniciar daemon %s: %w", exePath, err))
		}
	} else {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			logger.Fatal("LOCALAPPDATA environment variable is not set")
		}

		err := exec.Command(filepath.Join(localAppData, "Programs", "remembralld.exe")).Start()

		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to start daemon: %w", err))
		}
	}

	return nil
}

var SetupCommand *cli.Command = &cli.Command{
	Name:  "setup",
	Usage: "setup the application",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var configPathBase string
		var err error

		configPathBase, err = config.GetConfigPathBase()

		if err != nil {
			logger.Fatal("Failed to get config path: " + err.Error())
		}

		logger.Info("valiting initial setup...")
		logger.Info("checking config path folder...")

		configFolderExists := filesystem.FolderExists(configPathBase)

		if !configFolderExists {
			logger.Info("config folder does not exist. creating config folder...")

			filesystem.FileCreateFolder(path.Join(configPathBase))
		}

		logger.Info("checking config path file...")

		configFileExists := filesystem.FileExists(path.Join(configPathBase, "config.json"))

		if !configFileExists {
			logger.Info("config file does not exist. creating config file...")

			config.CreateConfig()
		}

		configDir, err := os.UserHomeDir()

		if err != nil {
			logger.Fatal("failed to get user config directory: " + err.Error())
		}

		logger.Info("checking daemon binary...")

		daemonExists := filesystem.FileExists(path.Join(configDir, constants.OS_CONFIGS["APP_DAEMON_LOCATION"][runtime.GOOS]))

		if !daemonExists {
			logger.Warning("daemon binary not found!")

			release, err := fetchLatestRelease(constants.GITHUB_OWNER, constants.GITHUB_REPO)

			if err != nil {
				logger.Fatal(err)
			}

			_, remembralldURL, ok := FindAssetURLs(release)

			if !ok {
				logger.Fatal(fmt.Sprintf("nenhum asset encontrado para %s", assetSuffix()))
			}

			runner := func() {
				home, err := os.UserHomeDir()
				if err != nil {
					logger.Fatal(fmt.Sprintf("erro ao obter home dir: %w", err))
				}

				binDir := filepath.Join(home, ".local", "bin")

				transport := &http.Transport{
					DialContext: (&net.Dialer{
						Timeout: 30 * time.Second, // timeout só na conexão TCP
					}).DialContext,
					TLSHandshakeTimeout: 10 * time.Second,
				}

				client := &http.Client{Transport: transport}

				req, err := http.NewRequest(http.MethodGet, remembralldURL, nil)
				if err != nil {
					logger.Fatal(fmt.Sprintf("erro ao criar requisição: %w", err))
				}

				resp, err := client.Do(req)
				if err != nil {
					logger.Fatal(fmt.Sprintf("erro no download: %w", err))
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					logger.Fatal(fmt.Sprintf("resposta inesperada no download: %s", resp.Status))
				}

				tmpFile, err := os.Create(path.Join(binDir, "remembralld"))

				if err != nil {
					logger.Fatal(fmt.Sprintf("erro ao criar arquivo temporário: %w", err))
				}
				tmpPath := tmpFile.Name()

				_, err = io.Copy(tmpFile, resp.Body)
				tmpFile.Close()
				if err != nil {
					logger.Fatal(fmt.Sprintf("erro ao salvar binário: %w", err))
				}

				if runtime.GOOS != "windows" {
					if err := os.Chmod(tmpPath, 0755); err != nil {
						logger.Fatal(fmt.Sprintf("erro ao definir permissões: %w", err))
					}
				}
			}

			_ = spinner.New().
				Title("daemon binary not found. creating daemon binary...").
				Type(spinner.Dots).
				Action(runner).
				Run()

			// Executar daemon

			startDaemon()

			logger.Success("daemon started successfully!")
		}

		logger.Info("checking if daemon is running...")

		if !isProcessRunning("remembralld") {
			logger.Warning("daemon is not running!")

			err := startDaemon()

			if err != nil {
				logger.Fatal("failed to start daemon: " + err.Error())
			}

			logger.Success("daemon started successfully!")
		}

		pathsCorrect := pathsAlreadyInPATH()

		if !pathsCorrect {
			logger.Warning("The installation path is not in your PATH environment variable. Please add it to be able to run remembrall from anywhere.")
			logger.Info("Installation path: " + constants.OS_CONFIGS["APP_BINARY_LOCATION"][runtime.GOOS])
		}

		var executeDaemonOnStartup bool = true

		huh.NewConfirm().
			Title("Do you want to execute the remembrall daemon on startup? (recommended)").
			Affirmative("Yes!").
			Negative("No.").
			Value(&executeDaemonOnStartup).
			Run()

		if executeDaemonOnStartup {
			home, err := os.UserHomeDir()

			if err != nil {
				logger.Fatal(err)
			}

			app := &autostart.App{
				Name:        fmt.Sprintf("%s daemon", constants.APP_NAME),
				DisplayName: fmt.Sprintf("%s daemon", constants.APP_NAME),
				Exec:        []string{home, constants.OS_CONFIGS["APP_DAEMON_LOCATION"][runtime.GOOS]},
			}

			if app.IsEnabled() {
				logger.Info("remembrall daemon is already set to execute on startup.")
			} else {
				if err := app.Enable(); err != nil {
					logger.Fatal(fmt.Sprintf("failed to enable autostart: %w", err))
				}
				logger.Success("remembrall daemon set to execute on startup successfully!")
			}
		}

		var timeFormat string = "European Format"

		huh.NewSelect[string]().
			Options(huh.NewOptions("American Format", "European Format")...).
			Value(&timeFormat).
			Title("Select your preferred time format").
			Run()

		SetTimeFormat(timeFormat)

		logger.Success("Time format set to " + timeFormat + ".")

		SetDatabaseLocationCommandAction(ctx, cmd)

		logger.Success("setup completed successfully!")

		return nil
	},
}
