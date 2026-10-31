// Comando go81: emulador e ambiente de desenvolvimento para ZX81.
package main

import (
	"fmt"
	"os"

	"github.com/wilsonpilon/go81/internal/version"
)

const usage = `go81 — emulador e ambiente de desenvolvimento ZX81

Uso:
  go81 [opções]

Opções:
  -v, --version   mostra a versão
  -h, --help      mostra esta ajuda

Projeto em fase inicial. Veja SPEC.md e MANUAL.md.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Print(usage)
		return 0
	}
	switch args[0] {
	case "-v", "--version", "version":
		fmt.Println("go81", version.Full())
		return 0
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "go81: opção desconhecida: %s\n\n%s", args[0], usage)
		return 2
	}
}
