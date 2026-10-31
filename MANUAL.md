# Go81 — Manual do usuário

Versão do manual: 1.2.1

> O Go81 está em fase inicial. Este manual descreve apenas o que já funciona.
> Itens marcados *(planejado)* estão especificados em [SPEC.md](SPEC.md) mas ainda não existem.

## Sumário

1. [Instalação](#1-instalação)
2. [Linha de comando](#2-linha-de-comando)
3. [Modos de interface](#3-modos-de-interface)
4. [Emulador](#4-emulador-planejado)
5. [Editor de BASIC e compilador Dignified](#5-editor-de-basic-e-compilador-dignified-planejado)
6. [Editor de textos](#6-editor-de-textos-planejado)
7. [Montador Z80](#7-montador-z80-planejado)
8. [Fitas](#8-fitas-planejado)
9. [Depurador](#9-depurador-planejado)
10. [Porta de comunicação (Link)](#10-porta-de-comunicação-link-planejado)
11. [Shell](#11-shell-planejado)
12. [Configuração](#12-configuração-planejado)
13. [Códigos de saída](#13-códigos-de-saída)

## 1. Instalação

Compile a partir do código-fonte (Go 1.23+):

```sh
git clone https://github.com/wilsonpilon/go81
cd go81
go build -o bin/ ./cmd/go81
```

O executável fica em `bin/go81` (Linux) ou `bin\go81.exe` (Windows).

## 2. Linha de comando

### Opções disponíveis

| Opção | Descrição |
|---|---|
| `--version`, `-v` | Mostra a versão completa (número e codinomes) |
| `--help`, `-h` | Mostra a ajuda |

Exemplo:

```
$ go81 --version
go81 1.2.1 — Sinclair · Mazogs · "Núcleo Z80"
```

### Opções planejadas

`--gui`, `--tui`, `--headless`, `--machine`, `--ram`, `--hires`, `--rom`, `--config`, `--profile`,
`--link`, `--frames`, `--screenshot` e os subcomandos `run`, `basic`, `asm`, `tape`, `link`, `shell`.
Ver [SPEC.md §8.3](SPEC.md#83-cli).

## 3. Modos de interface

| Modo | Como abrir | Estado |
|---|---|---|
| GUI (Fyne) | `go81 --gui` | *(planejado)* |
| TUI (tview) | `go81 --tui` | *(planejado)* |
| Shell | `go81 shell` | *(planejado)* |
| CLI direta | `go81 <subcomando>` | parcial |

## 4. Emulador *(planejado)*
## 5. Editor de BASIC e compilador Dignified *(planejado)*
## 6. Editor de textos *(planejado)*
## 7. Montador Z80 *(planejado)*
## 8. Fitas *(planejado)*
## 9. Depurador *(planejado)*
## 10. Porta de comunicação (Link) *(planejado)*
## 11. Shell *(planejado)*
## 12. Configuração *(planejado)*

## 13. Códigos de saída

| Código | Significado |
|---|---|
| 0 | Sucesso |
| 1 | Erro geral |
| 2 | Uso incorreto (parâmetro inválido) |
