<#
.SYNOPSIS
    Gera as capturas de tela do README em docs\img.

.DESCRIPTION
    Tudo sai do próprio Go81, com o programa examples\demo.bas:
      gui.png          a janela (captura real, go81 gui --capture)
      gui-console.png  o console de comandos da janela (driver de teste do Fyne)
      gui-teclado.png  o teclado do ZX81 na tela (driver de teste do Fyne)
      gui-mapa.png     o mapa do teclado esmaecido sobre a tela (driver de teste do Fyne)
      gui-mapa-cursor.png  CTRL apertado: as teclas do cursor K (driver de teste do Fyne)
      gui-editor.png   o editor de BASIC no lugar da tela (F7) (driver de teste do Fyne)
      gui-editor-teclas.png  o editor com as teclas do ZX81 e o mapa do cursor F (driver de teste do Fyne)
      gui-editor-erro.png    erro de sintaxe achado pela ROM, com o cursor S (driver de teste do Fyne)
      tui.png          o terminal com o console aberto (tela simulada do tcell)
      shell.png        uma sessão do go81 shell
    Precisa do gcc do MSYS2 (a janela usa CGO) e de uma área de trabalho.
#>
$ErrorActionPreference = 'Stop'
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root
$out = Join-Path $Root 'docs\img'
New-Item -ItemType Directory -Force $out | Out-Null
$msys = 'C:\msys64\ucrt64\bin'
if (Test-Path $msys) { $env:Path = "$msys;$env:Path" }

$tmp = Join-Path ([IO.Path]::GetTempPath()) 'go81-screenshots'
New-Item -ItemType Directory -Force $tmp | Out-Null
$exe = Join-Path $tmp 'go81.exe'
$demo = Join-Path $tmp 'demo.p'

function Run { $exe0 = $args[0]; $rest = @($args | Select-Object -Skip 1); & $exe0 @rest; if ($LASTEXITCODE -ne 0) { throw "$exe0 falhou ($LASTEXITCODE)" } }

Write-Host '==> compila'
Run go build -o $exe ./cmd/go81
Run $exe basic make examples/demo.bas -o $demo -run 10

Write-Host '==> janela'
Run $exe gui -no-audio -zoom 2 -exit-after 1600 -capture (Join-Path $out 'gui.png') $demo

Write-Host '==> terminal, shell e console'
$env:GO81_SCREENSHOTS = $out
Run go test -count=1 -run Screenshot ./internal/ui/tui ./internal/shell ./internal/ui/gui
Remove-Item Env:GO81_SCREENSHOTS

Get-ChildItem $out -Filter *.png | ForEach-Object { Write-Host ('    {0,8:N0}  {1}' -f $_.Length, $_.Name) }

