# ==============================================================================
# Script para rodar o Robô Jurídico (Go API + Painel PHP/Filament) sem Docker
# ==============================================================================

$ErrorActionPreference = "Stop"
$RootDir = $PSScriptRoot
$GoDir = Join-Path $RootDir "api-juridica-go"
$PainelDir = Join-Path $RootDir "painel"
$DadosDir = Join-Path $PainelDir "dados"
$DbPath = Join-Path $DadosDir "api.db"

Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host " Iniciando Robô Jurídico Localmente (Sem Docker)" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan

# 1. Garantir diretórios necessários (SQLite, Cache, Storage do Laravel)
$Dirs = @(
    $DadosDir,
    (Join-Path $PainelDir "storage\framework\cache\data"),
    (Join-Path $PainelDir "storage\framework\sessions"),
    (Join-Path $PainelDir "storage\framework\views"),
    (Join-Path $PainelDir "storage\logs"),
    (Join-Path $PainelDir "bootstrap\cache")
)
foreach ($d in $Dirs) {
    if (-not (Test-Path $d)) {
        New-Item -ItemType Directory -Path $d -Force | Out-Null
    }
}

# 2. Garantir arquivo .env no painel
$EnvFile = Join-Path $PainelDir ".env"
$EnvExample = Join-Path $PainelDir ".env.example"
if (-not (Test-Path $EnvFile) -and (Test-Path $EnvExample)) {
    Copy-Item $EnvExample $EnvFile
}

# 3. Garantir arquivo SQLite
if (-not (Test-Path $DbPath)) {
    New-Item -ItemType File -Path $DbPath -Force | Out-Null
}

# 4. Verificar dependências do Composer e rodar migrations/seeders do Painel
Push-Location $PainelDir
try {
    if (-not (Test-Path (Join-Path $PainelDir "vendor\autoload.php"))) {
        Write-Host "[1/4] Instalando dependências do PHP (Composer)..." -ForegroundColor Yellow
        composer install --no-interaction --prefer-dist --ignore-platform-reqs
    } else {
        Write-Host "[1/4] Dependências do Composer já instaladas." -ForegroundColor Green
    }

    Write-Host "[2/4] Executando migrations, seeders e publicando assets do Filament..." -ForegroundColor Yellow
    php artisan migrate --force
    php artisan db:seed --force
    php artisan filament:assets
} finally {
    Pop-Location
}

# 5. Iniciar a API em Go (Porta 8000) em uma nova janela ou processo em background
Write-Host "[3/4] Compilando e iniciando a API Go na porta 8000..." -ForegroundColor Yellow
$Env:PORT = "8000"
$Env:DB_PATH = $DbPath
$Env:API_BANCO = $DbPath
$Env:LOG_LEVEL = "info"
$Env:API_CHAVES = "painel-interno:teste-chave-secreta-longa-com-mais-de-32-caracteres"

Push-Location $GoDir
try {
    $BinDir = Join-Path $GoDir "bin"
    if (-not (Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    }
    $GoExe = Join-Path $BinDir "api.exe"
    go build -o $GoExe ./cmd/api
} finally {
    Pop-Location
}

$GoProcess = Start-Process -FilePath $GoExe -WorkingDirectory $GoDir -NoNewWindow -PassThru

# 6. Iniciar o Painel Administrativo PHP/Filament (Porta 8080)
Write-Host "[4/4] Iniciando o Painel Administrativo na porta 8080..." -ForegroundColor Green
Write-Host ""
Write-Host "  -> API Go:        http://localhost:8000/saude" -ForegroundColor Cyan
Write-Host "  -> Painel Admin:  http://localhost:8080/admin" -ForegroundColor Cyan
Write-Host "  Pressione Ctrl+C para encerrar ambos os serviços." -ForegroundColor Gray
Write-Host ""

Push-Location $PainelDir
try {
    php artisan serve --host=127.0.0.1 --port=8080
} finally {
    Pop-Location
    if ($GoProcess -and -not $GoProcess.HasExited) {
        Write-Host "Encerrando API Go..." -ForegroundColor Yellow
        Stop-Process -Id $GoProcess.Id -Force -ErrorAction SilentlyContinue
    }
}
