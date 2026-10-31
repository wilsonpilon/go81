<#
.SYNOPSIS
    Compila o Go81 e gera o pacote de distribuição em dist\.

.DESCRIPTION
    1. Confere (e, se faltar, instala com o winget) o Go 1.25+ e o gcc do MSYS2,
       que a janela (Fyne) precisa para compilar.
    2. Baixa e confere os módulos Go (go mod download / verify).
    3. Roda go vet e os testes.
    4. Compila:
         go81.exe        completo (janela, terminal, shell, link)
         go81-nogui.exe  sem janela, para máquinas sem OpenGL ou acesso remoto
    5. Monta dist\go81-X.Y.Z-windows-amd64\ com os executáveis, a documentação,
       a licença e os exemplos, e gera dist\go81-X.Y.Z-windows-amd64.zip e o
       arquivo .sha256 para o release.
    O bin\go81.exe também é atualizado.

.PARAMETER SkipTests
    Não roda go vet nem os testes.

.PARAMETER NoInstall
    Não instala nada: só avisa o que falta.

.PARAMETER NoGui
    Compila só a versão sem janela (não precisa de gcc).

.PARAMETER Msys2
    Pasta do MSYS2 (padrão C:\msys64).

.EXAMPLE
    .\build.ps1
.EXAMPLE
    .\build.ps1 -SkipTests
#>
[CmdletBinding()]
param(
    [switch]$SkipTests,
    [switch]$NoInstall,
    [switch]$NoGui,
    [string]$Msys2 = 'C:\msys64'
)

$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
Set-Location $Root

$MinGo = [version]'1.25'
$Arch = 'windows-amd64'

function Step([string]$msg) {
    Write-Host ''
    Write-Host "==> $msg" -ForegroundColor Cyan
}

function Info([string]$msg) { Write-Host "    $msg" }

function Fail([string]$msg) {
    Write-Host ''
    Write-Host "ERRO: $msg" -ForegroundColor Red
    exit 1
}

# Runs a program and stops the build if it fails.
#   (No param block on purpose: flags like -o reach the program untouched.)
function Run {
    $exe = $args[0]
    $rest = @($args | Select-Object -Skip 1)
    & $exe @rest
    if ($LASTEXITCODE -ne 0) { Fail "$exe $($rest -join ' ') terminou com código $LASTEXITCODE" }
}

# Puts dir first in PATH (so MSYS2's gcc wins over other gcc copies).
function Add-Path([string]$dir) {
    if (Test-Path $dir) {
        $rest = ($env:Path -split ';') | Where-Object { $_ -and ($_.TrimEnd('\') -ne $dir.TrimEnd('\')) }
        $env:Path = (@($dir) + $rest) -join ';'
    }
}

function Install-Winget([string]$id, [string]$what) {
    if ($NoInstall) { Fail "$what não encontrado. Instale-o (winget install --id $id) ou rode sem -NoInstall." }
    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        Fail "$what não encontrado e o winget não está disponível. Instale $what manualmente."
    }
    Info "instalando $what com o winget ($id)..."
    Run winget install --id $id -e --accept-source-agreements --accept-package-agreements
}

# ---------------------------------------------------------------------------
Step 'Ferramentas'

# Go
Add-Path 'C:\Program Files\Go\bin'
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Install-Winget 'GoLang.Go' 'Go'
    Add-Path 'C:\Program Files\Go\bin'
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { Fail 'o Go foi instalado, mas não está no PATH. Abra um novo terminal e rode de novo.' }
}
$goVer = (& go env GOVERSION).Trim()          # ex.: go1.27.0
$goNum = [version](($goVer -replace '^go', '') -replace '[^0-9.].*$', '')
if ($goNum -lt $MinGo) { Fail "o Go81 precisa do Go $MinGo ou mais novo (encontrado $goVer). Atualize: winget upgrade --id GoLang.Go" }
Info "Go: $goVer"

# gcc (só para a janela)
if (-not $NoGui) {
    $ucrt = Join-Path $Msys2 'ucrt64\bin'
    Add-Path $ucrt
    if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
        $bash = Join-Path $Msys2 'usr\bin\bash.exe'
        if (-not (Test-Path $bash)) {
            Install-Winget 'MSYS2.MSYS2' 'MSYS2'
            if (-not (Test-Path $bash)) { Fail "o MSYS2 não está em $Msys2. Use -Msys2 <pasta>." }
        }
        if ($NoInstall) { Fail "gcc não encontrado. No MSYS2: pacman -S mingw-w64-ucrt-x86_64-gcc" }
        Info 'instalando o gcc no MSYS2 (pacman)...'
        Run $bash -lc 'pacman -S --needed --noconfirm mingw-w64-ucrt-x86_64-gcc'
        Add-Path $ucrt
        if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) { Fail "gcc não encontrado em $ucrt" }
    }
    $gcc = (Get-Command gcc).Source
    $env:CC = $gcc
    $gccVer = (& $gcc -dumpfullversion).Trim()
    Info "gcc: $gccVer ($gcc)"
    if ($gcc -notlike "$Msys2*") { Write-Warning "este gcc não é o do MSYS2 ($Msys2); o recomendado é o do MSYS2 UCRT64." }
}

# ---------------------------------------------------------------------------
Step 'Módulos Go'
Run go mod download
Run go mod verify

# ---------------------------------------------------------------------------
$verFile = Get-Content (Join-Path $Root 'internal\version\version.go') -Raw
function Get-Const([string]$name) {
    $m = [regex]::Match($verFile, "(?m)^\s*$name\s*=\s*(.+?)\s*(//.*)?$")
    if (-not $m.Success) { Fail "não achei $name em internal\version\version.go" }
    return $m.Groups[1].Value.Trim('"')
}
$Version = '{0}.{1}.{2}' -f (Get-Const 'Major'), (Get-Const 'Minor'), (Get-Const 'Patch')
$Codename = '{0} · {1} · "{2}"' -f (Get-Const 'MajorName'), (Get-Const 'MinorName'), (Get-Const 'Subtitle')
Write-Host ''
Write-Host "Go81 $Version — $Codename" -ForegroundColor Green

# ---------------------------------------------------------------------------
if (-not $SkipTests) {
    Step 'Verificação (go vet e testes)'
    if ($NoGui) {
        $env:CGO_ENABLED = '0'
        Run go vet -tags nogui ./internal/... ./pkg/... ./cmd/...
        Run go test -tags nogui ./...
    } else {
        $env:CGO_ENABLED = '1'
        Run go vet ./...
        Run go test ./...
    }
}

# ---------------------------------------------------------------------------
Step 'Compilação'
$Name = "go81-$Version-$Arch"
$Dist = Join-Path $Root 'dist'
$Pkg = Join-Path $Dist $Name
if (Test-Path $Pkg) { Remove-Item $Pkg -Recurse -Force }
New-Item -ItemType Directory -Force $Pkg | Out-Null

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$ldflags = '-s -w'
$exes = @()
if (-not $NoGui) {
    $env:CGO_ENABLED = '1'
    Info 'go81.exe (completo)'
    Run go build -trimpath -ldflags $ldflags -o (Join-Path $Pkg 'go81.exe') ./cmd/go81
    $exes += 'go81.exe'
}
$env:CGO_ENABLED = '0'
Info 'go81-nogui.exe (sem janela)'
Run go build -trimpath -tags nogui -ldflags $ldflags -o (Join-Path $Pkg 'go81-nogui.exe') ./cmd/go81
$exes += 'go81-nogui.exe'
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue

# Confere a versão gravada nos executáveis.
foreach ($e in $exes) {
    $out = (& (Join-Path $Pkg $e) version) -join ' '
    if ($out -notmatch [regex]::Escape($Version)) { Fail "$e informa '$out', esperado $Version" }
    Info "$e -> $out"
}

# Os executáveis só podem depender de DLLs do Windows.
$objdump = Get-Command objdump -ErrorAction SilentlyContinue
if ($objdump) {
    foreach ($e in $exes) {
        $dlls = & $objdump.Source -p (Join-Path $Pkg $e) | Select-String 'DLL Name: (.+)$' | ForEach-Object { $_.Matches[0].Groups[1].Value.Trim() }
        $extra = $dlls | Where-Object { $_ -notmatch '^(api-ms-win-.*|KERNEL32|USER32|GDI32|SHELL32|OPENGL32|ADVAPI32|WS2_32|WINMM|OLE32|OLEAUT32|IMM32|COMDLG32|DWMAPI|NTDLL|WINSPOOL\.DRV|USERENV|CRYPT32|BCRYPT|SECUR32|IPHLPAPI|DNSAPI|POWRPROF|VERSION|SHLWAPI|SETUPAPI|UXTHEME|MSVCRT|UCRTBASE)(\.dll)?$' }
        if ($extra) { Fail "$e depende de DLLs que não vêm com o Windows: $($extra -join ', ')" }
    }
    Info 'dependências: só DLLs do Windows'
}

# Atualiza o bin\ local.
New-Item -ItemType Directory -Force (Join-Path $Root 'bin') | Out-Null
foreach ($e in $exes) { Copy-Item (Join-Path $Pkg $e) (Join-Path $Root "bin\$e") -Force }
Info 'bin\ atualizado'

# ---------------------------------------------------------------------------
Step 'Pacote'
foreach ($f in 'LICENSE', 'README.md', 'MANUAL.md', 'RELEASE.md') {
    Copy-Item (Join-Path $Root $f) $Pkg
}
Copy-Item (Join-Path $Root 'examples') (Join-Path $Pkg 'examples') -Recurse

$guiLine = if ($NoGui) { '' } else { "  go81 gui [programa.p]     abre o ZX81 numa janela`r`n" }
$leiame = @"
Go81 $Version — $Codename
Emulador e ambiente de desenvolvimento para o Sinclair ZX81 e clones.
https://github.com/wilsonpilon/go81

Para começar (num terminal, nesta pasta):
$guiLine  go81 tui [programa.p]     abre o ZX81 no terminal (F1 = ajuda)
  go81 shell                shell de comandos (help lista os comandos)
  go81 shell examples\ola.g81
  go81 link [programa.p]    porta de comunicação (127.0.0.1:8181 e ZRCP 10000)
  go81 help                 todos os comandos

go81-nogui.exe é o mesmo programa sem a janela, para computadores sem OpenGL
ou acesso remoto.

Manual completo: MANUAL.md. Novidades: RELEASE.md.
Licença: GPL v3 ou posterior (LICENSE).
"@
Set-Content -Path (Join-Path $Pkg 'LEIA-ME.txt') -Value $leiame -Encoding UTF8

$Zip = Join-Path $Dist "$Name.zip"
if (Test-Path $Zip) { Remove-Item $Zip -Force }
Compress-Archive -Path $Pkg -DestinationPath $Zip -CompressionLevel Optimal
$hash = (Get-FileHash $Zip -Algorithm SHA256).Hash.ToLower()
Set-Content -Path "$Zip.sha256" -Value "$hash  $Name.zip" -Encoding ASCII

# ---------------------------------------------------------------------------
Step 'Pronto'
Get-ChildItem $Pkg -Recurse -File | ForEach-Object {
    Info ('{0,12:N0}  {1}' -f $_.Length, $_.FullName.Substring($Dist.Length + 1))
}
$zipSize = (Get-Item $Zip).Length
Write-Host ''
Write-Host ("Pacote: {0} ({1:N1} MB)" -f $Zip, ($zipSize / 1MB)) -ForegroundColor Green
Write-Host "SHA256: $hash"
