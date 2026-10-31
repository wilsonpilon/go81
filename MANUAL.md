# Go81 — Manual do usuário

Versão do manual: 1.4.2

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

A janela (GUI) usa o Fyne, que precisa de um compilador C (CGO):
- **Windows:** MSYS2 com `mingw-w64-ucrt-x86_64-gcc`, e `C:\msys64\ucrt64\bin` no `PATH`.
- **Linux:** `gcc` e os pacotes de desenvolvimento de OpenGL e X11 (ex.: `libgl1-mesa-dev xorg-dev`).

Sem compilador C, gere a versão só de linha de comando: `go build -tags nogui -o bin/ ./cmd/go81`
(os testes: `go test -tags nogui ./...`).

## 2. Linha de comando

```
go81 <comando> [opções]
```

| Comando | Descrição |
|---|---|
| `gui` | Abre o ZX81 numa janela (ver §4.3) |
| `run` | Roda o ZX81 sem janela e pode salvar a tela em PNG (ver §4) |
| `roms` | Lista as ROMs embutidas no executável |
| `version`, `--version`, `-v` | Mostra a versão completa (número e codinomes) |
| `help`, `--help`, `-h` | Mostra a ajuda |

`go81 <comando> -h` mostra as opções de cada comando.

```
$ go81 --version
go81 1.4.2 — Sinclair · Football Manager · "Som e configuração"
```

Planejado: `--gui`, `--tui`, `--config`, `--profile`, `--link` e os comandos `basic`, `asm`, `tape`,
`link` e `shell`. Ver [SPEC.md §8.3](SPEC.md#83-cli).

## 3. Modos de interface

| Modo | Como abrir | Estado |
|---|---|---|
| GUI (Fyne) | `go81 gui` | disponível |
| TUI (tview) | `go81 --tui` | *(planejado)* |
| Shell | `go81 shell` | *(planejado)* |
| CLI direta | `go81 <subcomando>` | parcial |

## 4. Emulador

O emulador roda numa janela (`go81 gui`) ou sem janela, pela linha de comando (`go81 run`).
Máquina disponível: **ZX81 / TS1000**.

### 4.1 `go81 run`

```
go81 run [opções] [programa.p]
```

| Opção | Padrão | Descrição |
|---|---|---|
| `programa.p` ou `--tape` | — | Programa a carregar da fita: `.p`, `.81` ou `.p81` (ver §8) |
| `--autoload` | ligado | Com um programa: reinicia e digita `LOAD ""` sozinho, como o EightyOne |
| `--save` | — | Grava num arquivo `.p` o que o programa salvar com `SAVE` |
| `--rom` | `zx81.edition3` | ROM embutida (`go81 roms`) ou caminho de um arquivo `.rom` |
| `--frames` | `250` | Quadros a emular (50 por segundo em PAL, 60 em NTSC) |
| `--ram` | `16` | RAM pack em K: `0` (só o 1K interno), `2`, `4`, `8`, `16`, `32` ou `48` |
| `--ntsc` | desligado | ZX81 americano / TS1000 (60 Hz, 262 linhas) |
| `--wrx` | desligado | Alta resolução WRX |
| `--keys` | — | Teclas a digitar, separadas por vírgula (ver §4.2) |
| `--keys-start` | `150` | Quadro em que a digitação começa (o ZX81 leva cerca de 2 s para ligar) |
| `--key-hold` | `6` | Quadros com cada tecla pressionada |
| `--key-gap` | `6` | Quadros entre teclas |
| `--enter-gap` | `60` | Quadros extras depois de `ENTER`, enquanto a ROM redesenha a tela |
| `--screenshot` | — | Salva a última tela num arquivo PNG |
| `--border` | `normal` | Tamanho da imagem: `none` (256×192), `small` (264×200), `normal` (320×240), `large` (400×300), `full` (414×313) |

Exemplo, que digita um programa, roda e salva a tela:

```
go81 run --frames 600 --keys "1,0,P,SHIFT+P,O,L,A,SPACE,M,U,N,D,O,SHIFT+P,ENTER,R,ENTER" --screenshot tela.png
```

O ZX81 usa palavras-chave de uma tecla: no início da linha (cursor `K`), `P` vira `PRINT` e `R` vira `RUN`.

Carregar um programa e rodá-lo (o auto-load leva cerca de 5 s; o `RUN` é digitado depois):

```
go81 run --frames 700 --keys-start 450 --keys "R,ENTER" --screenshot tela.png jogo.p
```

Digitar um programa e gravá-lo com `SAVE "OI"`:

```
go81 run --frames 900 --keys "1,0,P,SHIFT+P,O,I,SHIFT+P,ENTER,S,SHIFT+P,O,I,SHIFT+P,ENTER" --save oi.p
```

### 4.2 Nomes das teclas

Letras `A`–`Z`, dígitos `0`–`9`, `SHIFT`, `ENTER`, `SPACE` e `.`. Teclas pressionadas juntas usam `+`:
`SHIFT+P` é `"`, `SHIFT+F` é `FAST`, `SHIFT+D` é `SLOW`.
### 4.3 `go81 gui`

```
go81 gui [opções] [programa.p]
```

Abre o ZX81 numa janela. Com um programa, ele é carregado sozinho (auto-load).

| Opção | Padrão | Descrição |
|---|---|---|
| `--rom`, `--ram`, `--ntsc`, `--wrx` | — | Como em `go81 run` |
| `--sound` | desligado | Liga o som pelo sinal de vídeo ("Video Sound" do EightyOne) |
| `--no-audio` | — | Não abre a placa de som |
| `--zoom` | `2` | Tamanho da janela: 1 a 4 |
| `--border` | `normal` | Borda: `none`, `small`, `normal`, `large`, `full` |
| `--exit-after` | — | Automação: fecha depois de N quadros |
| `--capture` | — | Automação: com `--exit-after`, salva a imagem da janela num PNG |

**Menus:**
- *Arquivo* — Abrir programa (`.p`, `.81`, `.p81`, carregado com auto-load), Salvar tela como PNG, Sair.
- *Máquina* — Configuração do hardware (ROM, memória, NTSC, WRX, som; aplica reiniciando), Reiniciar
  (desligar e ligar), Pausar/continuar (F5), Turbo 10× (F6), Som (liga/desliga).

**Gravar programas:** quando o programa no ZX81 executa `SAVE "NOME"`, a janela pergunta onde gravar
o arquivo, já sugerindo `NOME.p`.

**Som:** o ZX81 não tem alto-falante; o "som" é o sinal de vídeo/sincronismo levado ao áudio, como faziam
os jogos que usavam o rádio ou a TV. Por isso ele vem desligado (como no EightyOne): ligue em
Máquina → Som ou com `--sound`. No modo turbo o som fica mudo.
- *Exibir* — Zoom 1× a 4×; tamanho da borda.
- *Ajuda* — Teclado, Sobre.

**Teclado:** as teclas do PC ficam na mesma posição das do ZX81, e o Shift do PC é o SHIFT do ZX81
(Shift+P = `"`). Atalhos: Backspace = RUBOUT, setas = cursor, vírgula = `,`, Esc = EDIT.
A barra inferior mostra a máquina, os quadros por segundo e se a fita está tocando.

## 5. Editor de BASIC e compilador Dignified *(planejado)*
## 6. Editor de textos *(planejado)*
## 7. Montador Z80 *(planejado)*
## 8. Fitas

Por enquanto: programas `.p`, `.81` e `.p81`, com a carga e a gravação rápidas do EightyOne
(*flash load* / *flash save*): a ROM é interceptada e os bytes passam direto, sem som.

- `.p` e `.81`: o nome do programa vem do nome do arquivo (letras, dígitos, espaço e `'$:?()><=+-*/;,.`;
  `_` vira espaço). O `LOAD ""` aceita qualquer nome.
- `.p81`: o nome já está dentro do arquivo.

Planejado: `.tzx`, `.t81`, `.wav`, `.o`/`.80` (ZX80), gerenciador de fitas com vários blocos.
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
