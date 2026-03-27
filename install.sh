#!/usr/bin/env bash
set -euo pipefail

# ─── Configuração ─────────────────────────────────────────────────────────────
OWNER="costaluu"
REPO="remembrall"
BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/remembrall"
RAW_BASE="https://raw.githubusercontent.com/$OWNER/$REPO/master"
API_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
# ──────────────────────────────────────────────────────────────────────────────

# Cores
BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RESET='\033[0m'

log()     { echo -e "${CYAN}  →${RESET} $1"; }
success() { echo -e "${GREEN}  ✔${RESET} $1"; }
warn()    { echo -e "${YELLOW}  !${RESET} $1"; }
die()     { echo -e "\n  ✖ ERRO: $1" >&2; exit 1; }

# ── 1. Detectar OS e ARCH ─────────────────────────────────────────────────────
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      die "Sistema operacional não suportado: $OS" ;;
  esac

  case "$ARCH" in
    x86_64)         ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)              die "Arquitetura não suportada: $ARCH" ;;
  esac

  ASSET_SUFFIX="${OS}-${ARCH}"
  log "Plataforma detectada: $OS/$ARCH"
}

# ── 2. Buscar URL do asset na última release ──────────────────────────────────
fetch_asset_url() {
  log "Consultando última release em $OWNER/$REPO..."

  RELEASE_JSON="$(curl -fsSL \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "$API_URL")" || die "Falha ao contatar a API do GitHub."

  TAG="$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [ -n "$TAG" ] || die "Não foi possível determinar a tag da release."
  log "Última versão: $TAG"

  ASSET_URL="$(echo "$RELEASE_JSON" \
    | grep '"browser_download_url"' \
    | grep "$ASSET_SUFFIX" \
    | head -1 \
    | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')"

  [ -n "$ASSET_URL" ] || die "Nenhum asset encontrado para $ASSET_SUFFIX na release $TAG."
  log "Asset encontrado: $(basename "$ASSET_URL")"
}

# ── 3. Criar estrutura de diretórios ──────────────────────────────────────────
create_dirs() {
  log "Criando estrutura em $CONFIG_DIR..."
  mkdir -p "$CONFIG_DIR"
  success "Diretórios prontos."
}

# ── 4. Baixar e instalar o binário ────────────────────────────────────────────
install_binary() {
  BINARY_PATH="$BIN_DIR/remembrall"

  log "Baixando binário..."
  curl -fsSL "$ASSET_URL" -o "$BINARY_PATH" || die "Falha no download do binário."
  chmod +x "$BINARY_PATH"
  success "Binário instalado em $BINARY_PATH."
}

# ── 5. Baixar config padrão ───────────────────────────────────────────────────
install_config() {
  CONFIG_FILE="$CONFIG_DIR/config.json"
  CONFIG_URL="$RAW_BASE/src/internal/config/default_config_linux_darwin.json"

  if [ -f "$CONFIG_FILE" ]; then
    warn "config.json já existe — mantendo o arquivo atual."
    return
  fi

  log "Baixando configuração padrão..."
  curl -fsSL "$CONFIG_URL" -o "$CONFIG_FILE" || die "Falha no download do config.json."
  success "config.json criado em $CONFIG_FILE."
}

# ── 6. Atualizar PATH ─────────────────────────────────────────────────────────
update_path() {
  EXPORT_LINE="export PATH=\"$BIN_DIR:\$PATH\""

  # Detectar shell rc
  SHELL_NAME="$(basename "${SHELL:-bash}")"
  case "$SHELL_NAME" in
    zsh)  RC_FILE="$HOME/.zshrc" ;;
    bash) RC_FILE="${BASH_ENV:-$HOME/.bashrc}" ;;
    fish) RC_FILE="$HOME/.config/fish/config.fish"
          EXPORT_LINE="fish_add_path $BIN_DIR" ;;
    *)    RC_FILE="$HOME/.profile" ;;
  esac

  if echo "$PATH" | grep -q "$BIN_DIR"; then
    warn "$BIN_DIR já está no PATH desta sessão."
    return
  fi

  if grep -qF "$BIN_DIR" "$RC_FILE" 2>/dev/null; then
    warn "$BIN_DIR já está em $RC_FILE — nada alterado."
    return
  fi

  echo -e "\n# Remembrall\n$EXPORT_LINE" >> "$RC_FILE"
  success "PATH atualizado em $RC_FILE."
  RC_UPDATED=true
}

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${BOLD}  🔔 Remembrall Installer${RESET}"
  echo ""

  detect_platform
  fetch_asset_url
  create_dirs
  install_binary
  install_config
  RC_UPDATED=false
  update_path

  echo ""
  echo -e "${GREEN}${BOLD}  ✔ Instalação concluída!${RESET}"
  echo ""

  if [ "$RC_UPDATED" = true ]; then
    echo -e "  Para começar a usar ${BOLD}agora mesmo${RESET}, rode:"
    echo ""
    echo -e "    ${CYAN}source $RC_FILE${RESET}"
    echo ""
    echo -e "  Ou simplesmente ${BOLD}feche e abra o terminal${RESET}."
  fi

  echo -e "  Depois é só rodar:"
  echo ""
  echo -e "    ${CYAN}remembrall setup${RESET}  ou  ${CYAN}rem setup${RESET}"
  echo ""
}

main