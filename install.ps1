# ─── Configuração ─────────────────────────────────────────────────────────────
$OWNER       = "costaluu"
$REPO        = "remembrall"
$INSTALL_DIR = Join-Path $env:APPDATA "remembrall"
$BIN_DIR     = Join-Path $LOCALAPPDATA "Programs" "remembrall" "bin"
$API_URL     = "https://api.github.com/repos/$OWNER/$REPO/releases/latest"
$RAW_BASE    = "https://raw.githubusercontent.com/$OWNER/$REPO/master"
# ──────────────────────────────────────────────────────────────────────────────

$ErrorActionPreference = "Stop"

function Log     { param($msg) Write-Host "  -> $msg" -ForegroundColor Cyan }
function Success { param($msg) Write-Host "  v  $msg" -ForegroundColor Green }
function Warn    { param($msg) Write-Host "  !  $msg" -ForegroundColor Yellow }
function Die     { param($msg) Write-Host "`n  X  ERRO: $msg" -ForegroundColor Red; exit 1 }

# ── 1. Detectar ARCH ──────────────────────────────────────────────────────────
function Get-Arch {
  $arch = $env:PROCESSOR_ARCHITECTURE
  switch ($arch) {
    "AMD64" { return "amd64" }
    "ARM64" { return "arm64" }
    default { Die "Arquitetura não suportada: $arch" }
  }
}

# ── 2. Buscar URL do asset na última release ──────────────────────────────────
function Get-AssetUrl {
  param([string]$arch)

  Log "Consultando última release em $OWNER/$REPO..."

  $headers = @{
    "Accept"               = "application/vnd.github+json"
    "X-GitHub-Api-Version" = "2022-11-28"
  }

  try {
    $release = Invoke-RestMethod -Uri $API_URL -Headers $headers
  } catch {
    Die "Falha ao contatar a API do GitHub: $_"
  }

  $tag    = $release.tag_name
  $suffix = "windows-$arch.exe"

  Log "Última versão: $tag"

  $asset = $release.assets | Where-Object { $_.name -like "*$suffix" } | Select-Object -First 1

  if (-not $asset) {
    Die "Nenhum asset encontrado para $suffix na release $tag."
  }

  Log "Asset encontrado: $($asset.name)"
  return $asset.browser_download_url
}

# ── 3. Criar estrutura de diretórios ──────────────────────────────────────────
function New-Dirs {
  Log "Criando estrutura em $INSTALL_DIR..."

  New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null

  Success "Diretórios prontos."
}

# ── 4. Baixar e instalar o binário ────────────────────────────────────────────
function Install-Binary {
  param([string]$assetUrl)

  $dest = Join-Path $BIN_DIR "remembrall.exe"

  Log "Baixando binário..."

  try {
    Invoke-WebRequest -Uri $assetUrl -OutFile $dest -UseBasicParsing
  } catch {
    Die "Falha no download do binário: $_"
  }

  Success "Binário instalado em $dest."
}

# ── 5. Baixar config padrão ───────────────────────────────────────────────────
function Install-Config {
  $configFile = Join-Path $INSTALL_DIR "config.json"
  $configUrl  = "$RAW_BASE/src/internal/config/default_config_windows.json"

  if (Test-Path $configFile) {
    Warn "config.json já existe — mantendo o arquivo atual."
    return
  }

  Log "Baixando configuração padrão..."

  try {
    Invoke-WebRequest -Uri $configUrl -OutFile $configFile -UseBasicParsing
  } catch {
    Die "Falha no download do config.json: $_"
  }

  Success "config.json criado em $configFile."
}

# ── 6. Atualizar PATH do usuário ──────────────────────────────────────────────
function Update-Path {
  $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $entries     = $currentPath -split ";" | Where-Object { $_ -ne "" }

  if ($entries -contains $BIN_DIR) {
    Warn "$BIN_DIR já está no PATH — nada alterado."
    return $false
  }

  $newPath = ($entries + $BIN_DIR) -join ";"
  [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

  # Atualiza a sessão atual também
  $env:PATH = "$env:PATH;$BIN_DIR"

  Success "PATH atualizado."
  return $true
}

# ── Main ──────────────────────────────────────────────────────────────────────
function Main {
  Write-Host ""
  Write-Host "  Remembrall Installer" -ForegroundColor White
  Write-Host ""

  $arch     = Get-Arch
  $assetUrl = Get-AssetUrl -arch $arch

  New-Dirs
  Install-Binary  -assetUrl $assetUrl
  Install-Config
  $pathUpdated = Update-Path

  Write-Host ""
  Write-Host "  Instalacao concluida!" -ForegroundColor Green
  Write-Host ""

  if ($pathUpdated) {
    Write-Host "  O PATH foi atualizado para o seu usuario." -ForegroundColor Cyan
    Write-Host "  Abra um novo terminal para as alteracoes terem efeito." -ForegroundColor Cyan
    Write-Host ""
  }

  Write-Host "  Depois e so rodar:"
  Write-Host ""
  Write-Host "    remembrall setup" -ForegroundColor Cyan
  Write-Host "    rem setup" -ForegroundColor Cyan
  Write-Host ""
}

Main