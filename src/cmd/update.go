package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	spinner "charm.land/huh/v2/spinner"
	"github.com/costaluu/remembrall/src/internal/constants"
	"github.com/costaluu/remembrall/src/logger"
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

var UpdateCheckCommand *cli.Command = &cli.Command{
	Name:  "check",
	Usage: fmt.Sprintf("check for updates to %s.", constants.APP_NAME),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		currentRaw := cmd.Root().Version

		if currentRaw == "dev" {
			currentRaw = "0.0.0" // fallback para desenvolvimento local
		}

		runner := func() {
			// ── 1. Parse da versão atual ─────────────────────────────────────────
			current, err := parseSemver(currentRaw)

			if err != nil {
				logger.Fatal(err)
			}

			// ── 2. Busca a última release no GitHub ──────────────────────────────
			fmt.Println("Fetching latest release from GitHub...")
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
				logger.Success(fmt.Sprintf("%s is up to date (version %s).\n", constants.APP_NAME, currentRaw))
			}

			// ── 5. Nova versão disponível ─────────────────────────────────────────
			logger.Info(fmt.Sprintf("A new version is avaliable: %s → %s\n", currentRaw, release.TagName))
		}

		_ = spinner.New().
			Title("looking for updates...").
			Type(spinner.Dots).
			Action(runner).
			Run()

		// downloadURL, found := findAssetURL(release)

		// if found {
		// 	logger.Info(fmt.Sprintf("  Download: %s\n", downloadURL))
		// } else {
		// 	logger.Info(fmt.Sprintf("  Release page: %s\n", release.HTMLURL))
		// 	logger.Info(fmt.Sprintf("  (no prebuilt binary found for %s/%s)\n", runtime.GOOS, runtime.GOARCH))
		// }

		return nil
	},
}

var UpdateCommands *cli.Command = &cli.Command{
	Name:     "update",
	Usage:    fmt.Sprintf("command to check and apply updates to %s.", constants.APP_NAME),
	Commands: []*cli.Command{UpdateCheckCommand},
}
