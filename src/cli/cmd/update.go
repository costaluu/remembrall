package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	spinner "charm.land/huh/v2/spinner"
	"github.com/costaluu/remembrall/src/cli/internal/config"
	"github.com/costaluu/remembrall/src/cli/internal/constants"
	"github.com/costaluu/remembrall/src/cli/internal/logger"
	"github.com/urfave/cli/v3"
)

// githubRelease representa o payload relevante da API do GitHub Releases.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// semver representa uma versão no formato vX.Y.Z.
type semver struct {
	Major, Minor, Patch int
}

// parseSemver faz o parse de uma string no formato "vX.Y.Z" ou "X.Y.Z".
// Retorna erro se o formato for inválido.
func parseSemver(raw string) (semver, error) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "v")
	parts := strings.Split(s, ".")

	if len(parts) != 3 {
		return semver{}, fmt.Errorf("formato de versão inválido: %q (esperado vX.Y.Z)", raw)
	}

	parse := func(part string) (int, error) {
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("componente inválido %q em %q", part, raw)
		}
		return n, nil
	}

	major, err := parse(parts[0])
	if err != nil {
		return semver{}, err
	}
	minor, err := parse(parts[1])
	if err != nil {
		return semver{}, err
	}
	patch, err := parse(parts[2])
	if err != nil {
		return semver{}, err
	}

	return semver{Major: major, Minor: minor, Patch: patch}, nil
}

// isNewer retorna true se `latest` for estritamente maior que `current`.
// Comparação semântica: MAJOR > MINOR > PATCH.
func isNewer(current, latest semver) bool {
	if latest.Major != current.Major {
		return latest.Major > current.Major
	}
	if latest.Minor != current.Minor {
		return latest.Minor > current.Minor
	}
	return latest.Patch > current.Patch
}

// fetchLatestRelease busca a release mais recente no GitHub.
// owner e repo devem corresponder ao repositório onde o pipeline publica.
func fetchLatestRelease(owner, repo string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao contatar o GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("nenhuma release encontrada em %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta inesperada do GitHub: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return &release, nil
}

// assetSuffix retorna o sufixo do binário publicado pelo pipeline para o
// OS/ARCH atual. O pipeline nomeia os artefatos como:
//
//	cli-<tag>-<os>-<arch>        (linux, darwin)
//	cli-<tag>-<os>-<arch>.exe   (windows)
//
// Ref: step "Build binaries" no release.yml.
func assetSuffix() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Normaliza para os nomes usados pelo pipeline
	if arch == "amd64" {
		arch = "amd64"
	}

	suffix := fmt.Sprintf("%s-%s", os, arch)
	if os == "windows" {
		suffix += ".exe"
	}
	return suffix
}

// findAssetURL procura entre os assets da release o binário para o OS/ARCH
// atual e retorna sua URL de download.
func findAssetURL(release *githubRelease) (string, bool) {
	suffix := assetSuffix()

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, suffix) {
			return asset.BrowserDownloadURL, true
		}
	}
	return "", false
}

var UpdateApplyCommand *cli.Command = &cli.Command{
	Name:  "apply",
	Usage: fmt.Sprintf("apply the latest update to %s.", constants.APP_NAME),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		UpdateCheckCommandAction(ctx, cmd)

		currentConfig, err := config.LoadConfig()

		if err != nil {
			logger.Fatal("Failed to load config: " + err.Error())
		}

		if currentConfig.NeedsUpdate {
			logger.Info("Fetching latest release from GitHub...")
			release, err := fetchLatestRelease(constants.GITHUB_OWNER, constants.GITHUB_REPO)

			if err != nil {
				logger.Fatal(err)
			}

			runner := func() {
				assetURL, ok := findAssetURL(release)
				if !ok {
					logger.Fatal(fmt.Sprintf("nenhum asset encontrado para %s", assetSuffix()))
				}

				// ── Resolver caminhos por OS ──────────────────────────────────────────────
				var binDir, binaryName string

				switch runtime.GOOS {
				case "linux", "darwin":
					home, err := os.UserHomeDir()
					if err != nil {
						logger.Fatal(fmt.Sprintf("erro ao obter home dir: %w", err))
					}
					binDir = filepath.Join(home, ".local", "bin")
					binaryName = "remembrall"

				case "windows":
					localAppData := os.Getenv("LOCALAPPDATA")
					if localAppData == "" {
						logger.Fatal("variável LOCALAPPDATA não encontrada")
					}
					binDir = filepath.Join(localAppData, "Programs", "remembrall")
					binaryName = "remembrall.exe"

				default:
					logger.Fatal(fmt.Sprintf("sistema operacional não suportado: %s", runtime.GOOS))
				}

				currentBin := filepath.Join(binDir, binaryName)
				oldBin := filepath.Join(binDir, strings.TrimSuffix(binaryName, ".exe")+"-old"+func() string {
					if runtime.GOOS == "windows" {
						return ".exe"
					}
					return ""
				}())

				// ── 1. Renomear binário atual para -old ───────────────────────────────────
				if _, err := os.Stat(currentBin); err == nil {
					// Remove -old anterior se existir
					_ = os.Remove(oldBin)

					if err := os.Rename(currentBin, oldBin); err != nil {
						logger.Fatal(fmt.Sprintf("erro ao renomear binário atual: %w", err))
					}
				}

				// ── 2. Baixar novo binário ────────────────────────────────────────────────
				if err := os.MkdirAll(binDir, 0755); err != nil {
					logger.Fatal(fmt.Sprintf("erro ao criar diretório %s: %w", binDir, err))
				}

				transport := &http.Transport{
					DialContext: (&net.Dialer{
						Timeout: 30 * time.Second, // timeout só na conexão TCP
					}).DialContext,
					TLSHandshakeTimeout: 10 * time.Second,
				}

				client := &http.Client{Transport: transport}

				req, err := http.NewRequest(http.MethodGet, assetURL, nil)
				if err != nil {
					logger.Fatal(fmt.Sprintf("erro ao criar requisição: %w", err))
				}

				resp, err := client.Do(req)
				if err != nil {
					// Rollback: restaura o binário anterior
					_ = os.Rename(oldBin, currentBin)
					logger.Fatal(fmt.Sprintf("erro no download: %w", err))
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					_ = os.Rename(oldBin, currentBin)
					logger.Fatal(fmt.Sprintf("resposta inesperada no download: %s", resp.Status))
				}

				// Escreve em arquivo temporário primeiro para evitar binário corrompido
				tmpFile, err := os.CreateTemp(binDir, "remembrall-update-*")
				if err != nil {
					_ = os.Rename(oldBin, currentBin)
					logger.Fatal(fmt.Sprintf("erro ao criar arquivo temporário: %w", err))
				}
				tmpPath := tmpFile.Name()

				_, err = io.Copy(tmpFile, resp.Body)
				tmpFile.Close()
				if err != nil {
					_ = os.Remove(tmpPath)
					_ = os.Rename(oldBin, currentBin)
					logger.Fatal(fmt.Sprintf("erro ao salvar binário: %w", err))
				}

				// ── 3. Tornar executável (unix) e mover para destino final ────────────────
				if runtime.GOOS != "windows" {
					if err := os.Chmod(tmpPath, 0755); err != nil {
						_ = os.Remove(tmpPath)
						_ = os.Rename(oldBin, currentBin)
						logger.Fatal(fmt.Sprintf("erro ao definir permissões: %w", err))
					}
				}

				if err := os.Rename(tmpPath, currentBin); err != nil {
					_ = os.Remove(tmpPath)
					_ = os.Rename(oldBin, currentBin)
					logger.Fatal(fmt.Sprintf("erro ao mover binário para destino final: %w", err))
				}

				// TODO: Migrações e outras tarefas

			}

			_ = spinner.New().
				Title(fmt.Sprintf("updating %s to lastest version...", constants.APP_NAME)).
				Type(spinner.Dots).
				Action(runner).
				Run()

			logger.Success(fmt.Sprintf("%s updated successfully to version %s!", constants.APP_NAME, currentConfig.LatestVersion))
		}

		return nil
	},
}

func UpdateCheckCommandAction(ctx context.Context, cmd *cli.Command) error {
	currentRaw := cmd.Root().Version

	if currentRaw == "dev" {
		currentRaw = "0.0.0" // fallback para desenvolvimento local
	}

	currentConfig, err := config.LoadConfig()

	if err != nil {
		logger.Fatal("Failed to load config: " + err.Error())
	}

	runner := func() {
		// ── 1. Parse da versão atual ─────────────────────────────────────────
		current, err := parseSemver(currentRaw)

		if err != nil {
			logger.Fatal(err)
		}

		// ── 2. Busca a última release no GitHub ──────────────────────────────
		logger.Info("Fetching latest release from GitHub...")
		release, err := fetchLatestRelease(constants.GITHUB_OWNER, constants.GITHUB_REPO)
		if err != nil {
			logger.Fatal(err)
		}

		// ── 3. Parse da versão remota ────────────────────────────────────────
		latest, err := parseSemver(release.TagName)
		if err != nil {
			logger.Fatal(err)
		}

		// ── 4. Comparação semântica ──────────────────────────────────────────
		if !isNewer(current, latest) {
			logger.Success(fmt.Sprintf("your %s is up to date.\n", constants.APP_NAME))
		}

		// ── 5. Nova versão disponível ─────────────────────────────────────────
		logger.Info(fmt.Sprintf("A new version is avaliable: %s → %s\n", currentRaw, release.TagName))

		currentConfig.LatestVersion = release.TagName
		currentConfig.LatestVersionCheckTime = time.Now()

		if !isNewer(current, latest) {
			currentConfig.NeedsUpdate = false
		} else {
			currentConfig.NeedsUpdate = true
		}

		err = config.SaveConfig(currentConfig)

		if err != nil {
			logger.Fatal("Failed to save config: " + err.Error())
			return
		}
	}

	if currentConfig.LatestVersionCheckTime.Add(12 * time.Hour).After(time.Now()) {
		logger.Info(fmt.Sprintf("Last checked for updates at %s", currentConfig.LatestVersionCheckTime.Format(currentConfig.DateTimeFormat)))
		logger.Info(fmt.Sprintf("Latest version: %s", currentConfig.LatestVersion))
		logger.Info("Checked for updates less than 12 hours ago, skipping check.")
		return nil
	}

	_ = spinner.New().
		Title("looking for updates...").
		Type(spinner.Dots).
		Action(runner).
		Run()

	return nil
}

var UpdateCheckCommand *cli.Command = &cli.Command{
	Name:   "check",
	Usage:  fmt.Sprintf("check for updates to %s.", constants.APP_NAME),
	Action: UpdateCheckCommandAction,
}

var UpdateCommands *cli.Command = &cli.Command{
	Name:     "update",
	Usage:    fmt.Sprintf("command to check and apply updates to %s.", constants.APP_NAME),
	Commands: []*cli.Command{UpdateCheckCommand, UpdateApplyCommand},
}
