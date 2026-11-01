#!/usr/bin/env bash
# Go81 — compila e gera o pacote de distribuição em dist/ (Linux e macOS).
#
#   ./build.sh                 testa, compila go81 e go81-nogui, gera o .tar.gz
#   ./build.sh --skip-tests    sem os testes
#   ./build.sh --nogui         só a versão sem janela (não precisa de gcc)
#   ./build.sh --install-deps  instala gcc e as bibliotecas da janela (apt, dnf ou pacman; usa sudo)
#
# Resultado: dist/go81-X.Y.Z-<sistema>-<arquitetura>/ (executáveis, documentação, exemplos),
# dist/go81-X.Y.Z-<sistema>-<arquitetura>.tar.gz e o .sha256. O bin/ também é atualizado.
set -euo pipefail
cd "$(dirname "$0")"

SKIP_TESTS=0 NOGUI=0 INSTALL=0
for a in "$@"; do
  case "$a" in
    --skip-tests) SKIP_TESTS=1 ;;
    --nogui) NOGUI=1 ;;
    --install-deps) INSTALL=1 ;;
    -h|--help) sed -n '2,10p' "$0"; exit 0 ;;
    *) echo "opção desconhecida: $a" >&2; exit 2 ;;
  esac
done

step() { printf '\n==> %s\n' "$1"; }
fail() { printf '\nERRO: %s\n' "$1" >&2; exit 1; }

step "Ferramentas"
command -v go >/dev/null || fail "Go não encontrado. Instale o Go 1.25 ou mais novo: https://go.dev/dl/"
GOVER=$(go env GOVERSION)
printf '    Go: %s\n' "$GOVER"
MIN=25
MINOR=$(echo "$GOVER" | sed -E 's/^go1\.([0-9]+).*/\1/')
[ "$MINOR" -ge "$MIN" ] || fail "o Go81 precisa do Go 1.$MIN ou mais novo (encontrado $GOVER)"

if [ "$NOGUI" = 0 ]; then
  if [ "$INSTALL" = 1 ]; then
    if command -v apt-get >/dev/null; then
      sudo apt-get update && sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev libasound2-dev
    elif command -v dnf >/dev/null; then
      sudo dnf install -y gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel mesa-libGL-devel wayland-devel libxkbcommon-devel alsa-lib-devel
    elif command -v pacman >/dev/null; then
      sudo pacman -S --needed --noconfirm gcc libxcursor libxrandr libxinerama libxi mesa wayland libxkbcommon alsa-lib
    elif [ "$(uname)" != Darwin ]; then
      fail "não sei instalar as dependências neste sistema; veja o README"
    fi
  fi
  command -v gcc >/dev/null || command -v cc >/dev/null || fail "gcc não encontrado (a janela usa CGO). Use --install-deps ou --nogui."
  printf '    C: %s\n' "$( (gcc --version 2>/dev/null || cc --version) | head -1)"
fi

step "Módulos Go"
go mod download
go mod verify

VFILE=internal/version/version.go
# tr: no Windows (autocrlf) a cópia tem CRLF, e o \r iria junto com o valor
get() { tr -d '\r' < "$VFILE" | sed -nE "s/^[[:space:]]*$1[[:space:]]*=[[:space:]]*\"?([^\"/]*[^\" /])\"?.*$/\1/p" | head -1; }
VERSION="$(get Major).$(get Minor).$(get Patch)"
printf '\nGo81 %s — %s · %s · "%s"\n' "$VERSION" "$(get MajorName)" "$(get MinorName)" "$(get Subtitle)"

if [ "$SKIP_TESTS" = 0 ]; then
  step "Verificação (go vet e testes)"
  if [ "$NOGUI" = 1 ]; then
    CGO_ENABLED=0 go vet -tags nogui ./...
    CGO_ENABLED=0 go test -tags nogui ./...
  else
    go vet ./...
    go test ./...
  fi
fi

step "Compilação"
OS=$(go env GOOS) ARCH=$(go env GOARCH)
NAME="go81-$VERSION-$OS-$ARCH"
PKG="dist/$NAME"
rm -rf "$PKG"
mkdir -p "$PKG" bin
EXES=()
if [ "$NOGUI" = 0 ]; then
  echo "    go81 (completo)"
  CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o "$PKG/go81" ./cmd/go81
  EXES+=(go81)
fi
echo "    go81-nogui (sem janela)"
CGO_ENABLED=0 go build -trimpath -tags nogui -ldflags "-s -w" -o "$PKG/go81-nogui" ./cmd/go81
EXES+=(go81-nogui)
for e in "${EXES[@]}"; do
  out=$("$PKG/$e" version)
  case "$out" in *"$VERSION"*) echo "    $e -> $out" ;; *) fail "$e informa '$out', esperado $VERSION" ;; esac
  cp "$PKG/$e" "bin/$e"
done

step "Pacote"
cp LICENSE README.md MANUAL.md RELEASE.md "$PKG/"
cp -r examples "$PKG/examples"
cat > "$PKG/LEIA-ME.txt" <<EOF
Go81 $VERSION — Sinclair ZX81 e clones, emulador e ambiente de desenvolvimento.
https://github.com/wilsonpilon/go81

Para começar (num terminal, nesta pasta):
  ./go81 gui [programa.p]    abre o ZX81 numa janela
  ./go81 tui [programa.p]    abre o ZX81 no terminal (F1 = ajuda)
  ./go81 shell               shell de comandos (help lista os comandos)
  ./go81 shell examples/ola.g81
  ./go81 help                todos os comandos

go81-nogui é o mesmo programa sem a janela (servidores, acesso remoto).
A janela precisa de OpenGL e das bibliotecas X11/Wayland do sistema.

Manual completo: MANUAL.md. Novidades: RELEASE.md. Licença: GPL v3 ou posterior.
EOF
TAR="dist/$NAME.tar.gz"
rm -f "$TAR"
tar -C dist -czf "$TAR" "$NAME"
if command -v sha256sum >/dev/null; then
  (cd dist && sha256sum "$NAME.tar.gz" > "$NAME.tar.gz.sha256")
else
  (cd dist && shasum -a 256 "$NAME.tar.gz" > "$NAME.tar.gz.sha256")
fi

step "Pronto"
ls -l "$PKG"
printf '\nPacote: %s (%s)\n' "$TAR" "$(du -h "$TAR" | cut -f1)"
cat "$TAR.sha256"
