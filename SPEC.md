# Go81 — Especificação

Versão do documento: 1.1.1 · Última revisão: 2026-10-30

Este documento descreve **o que** o Go81 deve ser. O andamento do trabalho fica em [OUTLINE.md](OUTLINE.md).

---

## 1. Visão geral

O Go81 é um ambiente integrado para o ZX81 escrito em Go. Ele reúne em um só executável:

1. Emulador do ZX81 e clones, compatível com o EightyOne 1.41
2. Editor de BASIC embutido
3. Compilador de BASIC em dialeto moderno (inspirado no *Basic Dignified Suite*)
4. Editor de textos
5. Montador (assembler) e desmontador Z80
6. Gerador e conversor de fitas
7. Depurador
8. Porta de comunicação (Link) para automação e troca de dados
9. Três interfaces: **GUI** (Fyne), **TUI** (tview/tcell) e **CLI** (shell próprio + parâmetros)

### Princípios

- **Núcleo sem interface.** Emulador, compiladores e ferramentas são bibliotecas Go puras, sem dependência de GUI/TUI. As interfaces só consomem o núcleo.
- **Tudo automatizável.** Qualquer ação disponível na GUI deve existir também na CLI e no Link.
- **Compatibilidade verificável.** "100% compatível com o EightyOne" significa testes automáticos comparando o estado da máquina (registradores, memória, quadro de vídeo) com o EightyOne.
- **Um único binário** (`go81`), multiplataforma: Windows e Linux (macOS como bônus).

## 2. Escopo

### 2.1 Máquinas emuladas

| Máquina | ROM(s) disponíveis em `81emul/exe/ROMs` | Observação |
|---|---|---|
| Sinclair ZX81 | `zx81.edition1/2/3.rom` | Máquina principal |
| Timex Sinclair TS1000 | `zx81.edition3.rom` | ZX81 americano, 2K RAM |
| Timex Sinclair TS1500 | `ts1500.rom` | 16K, ROM com extras |
| Lambda 8300 / PC8300 | `lambda8300.rom`, `lambda8300colour.rom`, `pc8300timex.rom` | Clone com cor opcional |
| Ringo R-470 | `ringo470.rom` | Clone brasileiro |
| Microdigital TK85 | `tk85.rom` | Clone brasileiro |
| ZX97 Lite | `zx97.rom` | ZX81 expandido (bancos de memória) |
| Sinclair ZX80 | `zx80.rom` | Antecessor do ZX81 — **incluído** por compartilhar o mesmo núcleo no EightyOne (ver decisão D-02) |

ROMs alternativas: `tree-forth.rom`, `zx81-forth.rom`, `zx.aszmic.e04.rom`, `zx.aszmic.e07.rom`, overlays ZXpand (`*.zxpand.ovl`).

### 2.2 Fora do escopo

- ZX Spectrum (todos os modelos), TC2048, TS2068/TC2068
- Jupiter Ace
- Periféricos exclusivos do Spectrum: Interface 1/Microdrive, +3/Beta/Opus/DISCiPLE/+D, divIDE/+3e, Multiface, AY 128K, Spectra, RZX, etc.

### 2.3 Hardware e periféricos do ZX81 (a emular)

Lista tirada das opções do diálogo *Hardware* do EightyOne (`HW_.cpp`, `HW_.dfm`):

| Categoria | Opções |
|---|---|
| Expansão de RAM | Nenhuma, 2K, 4K, 8K, 16K, 32K, 48K; RAM em 8–16K (para WRX/UDG) |
| Alta resolução | Nenhuma, WRX, G007, Memotech HRG, Quicksilva |
| Gerador de caracteres | Sinclair, dk'tronics, Quicksilva, CHR$128, Kayde |
| Cor | Nenhuma, Lambda, Chroma 81 |
| Som | Nenhum, ZON X, Quicksilva |
| Fala | Nenhuma, The Parrot, Sweet Talker, Mageco, S-Pack, Talk-Back |
| Joystick | Nenhum, Cursor, Kempston, Programável, ZXpand |
| Armazenamento | ZXpand / ZXpand+ (SD card), ZX1541, Larken, MWCFIde (IDE/CF) |
| Impressora | ZX Printer, Centronics |
| Cartuchos ROM | Standard 16K, ZXC1, ZXC2, ZXC3, ZXC4, Timex |
| Outros | Kempston Mouse, TS2050 Modem, porta serial, circuitos *Improved Wait* e *M1Not*, ROMs de interface (Memocalc, µSource, Z80 Assembler) |

### 2.4 Formatos de arquivo

| Tipo | Extensões | Ler | Gravar |
|---|---|---|---|
| Programa ZX81 | `.p`, `.81`, `.p81` | ✔ | ✔ |
| Programa ZX80 | `.o`, `.80` | ✔ | ✔ |
| Fita | `.tzx`, `.t81`, `.wav` | ✔ | ✔ |
| Snapshot | `.z81` (formato EightyOne) | ✔ | ✔ |
| Listagem BASIC | `.b81`, `.b80`, `.txt` | ✔ | ✔ |
| Fonte Dignified | `.dbas` (proposto) | ✔ | ✔ |
| Fonte assembler | `.asm`, `.z80`, `.s` | ✔ | ✔ |
| Disco/imagem | `.hdf`, `.img` (IDE), cartão SD do ZXpand (pasta) | ✔ | ✔ |
| Compactado | `.zip` (abrir de dentro) | ✔ | — |
| Imagem de tela | `.png`, `.bmp` | — | ✔ |

## 3. Arquitetura

```
             ┌──────────── interfaces ────────────┐
             │   GUI (Fyne)   TUI (tview)   CLI   │
             └──────────────────┬─────────────────┘
                                │  serviços (session)
   ┌────────────┬───────────────┼───────────────┬────────────┐
 emulador    depurador     ferramentas        Link        config
 (machine)    (debug)   (basic, asm, tape,   (TCP/pipe)  (TOML)
   │                      editor, media)
 z80 · video · audio · teclado · periféricos · ROMs
```

- **`session`** é o ponto central: guarda a máquina atual, os arquivos abertos e expõe comandos. GUI, TUI, CLI e Link chamam os mesmos comandos — é isso que garante que tudo seja automatizável.
- **Emulação em goroutine própria**, sincronizada por quadro (50/60 Hz), com mensagens por canal para entrada/saída. Pode rodar sem limite de velocidade (modo turbo) e sem vídeo (headless) para testes e automação.
- **Registro de comandos único**: cada comando (ex.: `load`, `poke`, `run`, `asm`) é declarado uma vez com nome, parâmetros e ajuda. A CLI, o shell, o Link e os menus da GUI são gerados desse registro.

## 4. Estrutura de pastas

```
go81/
├── cmd/
│   └── go81/                 # main: escolhe GUI, TUI, CLI ou shell
├── internal/
│   ├── version/              # número de versão, codinomes
│   ├── session/              # estado central + registro de comandos
│   ├── emu/
│   │   ├── z80/              # núcleo Z80 (porta do núcleo FUSE/EightyOne)
│   │   ├── machine/          # zx81, zx80, ts1000, ts1500, lambda, r470, tk85, zx97lite
│   │   ├── video/            # geração de quadro, hi-res, cor
│   │   ├── audio/            # som, fala
│   │   ├── input/            # teclado, joystick, mouse
│   │   ├── peripherals/      # ram, zxpand, printer, ide, cartuchos, modem...
│   │   └── rom/              # catálogo e carga de ROMs
│   ├── debug/                # breakpoints, histórico, profiler, símbolos
│   ├── media/
│   │   ├── tape/             # p, 81, p81, o, 80, t81, tzx, wav
│   │   └── snapshot/         # z81
│   ├── lang/
│   │   ├── zxbasic/          # tokens, lister e loader do BASIC ZX81/ZX80
│   │   ├── dignified/        # compilador do dialeto moderno → BASIC ZX81
│   │   └── asm/              # montador e desmontador Z80
│   ├── editor/               # buffer de texto compartilhado pelos editores
│   ├── link/                 # porta de comunicação (servidor + protocolo)
│   ├── shell/                # sistema "operacional" de linha de comando
│   ├── config/               # configuração (TOML) e perfis de máquina
│   └── ui/
│       ├── gui/              # Fyne
│       └── tui/              # tview/tcell
├── pkg/
│   └── go81link/             # cliente Go público para o Link
├── assets/
│   ├── roms/                 # ROMs embutidas (go:embed) ou buscadas na pasta do usuário
│   └── fonts/
├── examples/                 # programas exemplo (.dbas, .asm, scripts)
├── testdata/                 # arquivos para testes automáticos
├── test/
│   └── compat/               # testes de compatibilidade com o EightyOne
├── scripts/                  # scripts de build e release
├── docs/                     # material extra (protocolos, formatos)
├── 81emul/                   # referência: EightyOne (source/ e exe/) — não compilado
├── SPEC.md  OUTLINE.md  RELEASE.md  MANUAL.md  README.md
└── go.mod
```

Observação: o pedido original citava `81emul/src`; no disco os fontes estão em `81emul/source/Source`.

## 5. Emulador

### 5.1 Requisitos

- Z80 com temporização exata em T-states, instruções não documentadas, flags não documentadas (bits 3 e 5) e registrador R.
- Geração de vídeo do ZX81 como no EightyOne: NMI, HSYNC, linhas/quadro reais, *slow*/*fast mode*, tela piscando em `FAST`.
- Seleção de ROM, RAM e periféricos como na seção 2.3.
- Velocidade: 100% (50 Hz PAL / 60 Hz NTSC), turbo, pausa, avanço por quadro.
- Carga rápida (ROM patch, como `rompatch.cpp`) e carga real por áudio (`.wav`/`.tzx`).
- Teclado mapeado (PC → matriz 8×5), com mapa visual e teclas de função configuráveis.
- Snapshots `.z81` compatíveis com o EightyOne.

### 5.2 Critério de compatibilidade

Para cada máquina e conjunto de periféricos:

1. Mesma ROM, mesma entrada, mesmo número de T-states → mesmo estado de registradores e memória que o EightyOne.
2. Testes Z80 padrão (zexdoc/zexall, testes de flags) passando.
3. Conjunto de programas de referência (jogos, demos hi-res) com quadro de vídeo idêntico pixel a pixel.

Os resultados de referência são gerados com o `81emul/exe/EightyOne.exe` e guardados em `testdata/compat/`.

### 5.3 Fonte de referência no EightyOne

| Área | Arquivos em `81emul/source/Source/` |
|---|---|
| Z80 | `z80/*.c` |
| ZX81/ZX80 | `zx81/zx81.cpp`, `zx81/snap.cpp`, `zx81/rompatch.cpp` |
| Configuração | `zx81config.h/.cpp`, `HW_.cpp` |
| Fitas | `tzx/*.cpp`, `wavtape/*.cpp` |
| BASIC | `zx81/zx81BasicLister.cpp`, `zx81/zx81BasicLoader.cpp`, `BasicLoader/`, `BasicLister/` |
| Depurador | `DebugWin/*` |
| ZXpand | `zxpand/*` |
| Impressora | `zxprinter/*` |
| Cor | `Chroma/*` |
| Som/fala | `sound/*`, `SP0256/*` |
| IDE / disco | `ide/*`, `1541/*` |

## 6. Ferramentas de desenvolvimento

### 6.1 Editor de BASIC

- Edição de BASIC ZX81 clássico (com números de linha) e do dialeto Dignified.
- Destaque de sintaxe, palavras-chave em maiúsculas, caracteres gráficos do ZX81 via escapes (`\##`, `\:'` etc., notação compatível com o BASIC Loader do EightyOne).
- Ações: **Enviar para o emulador**, **Listar do emulador** (lister), **Salvar como `.p`**.

### 6.2 Compilador Dignified BASIC (ZX Dignified)

Inspirado no *Basic Dignified Suite* (Fred Rique). Converte código moderno em BASIC ZX81 tokenizado.

| Recurso | Exemplo |
|---|---|
| Sem números de linha | o compilador numera automaticamente |
| Rótulos | `{inicio}` … `GOTO {inicio}` |
| Nomes longos de variáveis | `vidas` → `V` (mapa gerado automaticamente; respeita limites do ZX81) |
| Comentários que não vão para o programa | `' comentário` |
| Constantes e definições | `define [VELOCIDADE] 5` |
| Inclusão de arquivos | `include "rotinas.dbas"` |
| Blocos de várias linhas | continuação com `_` |
| Diretivas | `#machine zx81`, `#ram 16k`, `#autorun {inicio}` |
| Código de máquina embutido | `asm … endasm` → montado dentro de um `REM` |

Saídas: `.p`, `.b81` (listagem texto), mapa de variáveis/rótulos. O compilador emite erros com arquivo:linha:coluna.
Fase futura (opcional): compilação para código de máquina Z80.

### 6.3 Editor de textos

Editor genérico (UTF-8 e conjunto de caracteres ZX81), com busca/substituição, múltiplas abas e destaque para `.asm`, `.dbas`, `.b81`.

### 6.4 Montador e desmontador Z80

- Sintaxe Z80 Zilog padrão, incluindo instruções não documentadas.
- Rótulos locais/globais, expressões, `ORG`, `EQU`, `DB/DW/DS`, `INCLUDE`, `INCBIN`, macros e montagem condicional.
- Símbolos predefinidos da ROM do ZX81 (arquivos `.sym` em `81emul/exe/ROMs`).
- Saídas: `.bin`, `.hex` (Intel HEX), `.p` (com loader BASIC ou dentro de `REM`), lista `.lst`, tabela de símbolos para o depurador.
- Desmontador usado pelo depurador e pela CLI.

### 6.5 Gerador de fitas

- Converte entre `.p`, `.81`, `.p81`, `.t81`, `.tzx` e `.wav` (taxa de amostragem configurável).
- Junta vários programas em uma fita, edita blocos e nomes (como o *TZX Manager* do EightyOne).
- Gera WAV para carregar em ZX81 real.

### 6.6 Depurador

Equivalente ao `DebugWin` do EightyOne: registradores e flags editáveis, desmontagem, breakpoints (execução, leitura, escrita, porta de E/S, condição), passo a passo (*step into/over/out*), histórico de execução, busca de sequências, memória ao vivo, profiler e símbolos.

## 7. Porta de comunicação (Link)

Dois lados:

### 7.1 Host ↔ emulador (automação)

- Transporte: **TCP em `127.0.0.1`** (porta padrão `8181`), **stdin/stdout** (modo `--link stdio`) e *named pipe*/socket Unix.
- Protocolo de linha, legível por humanos (dá para usar com `telnet`/`nc`), com modo JSON opcional:
  ```
  > load "jogo.p"
  < ok
  > peek 16396 2
  < ok 125 64
  > json {"cmd":"regs"}
  < {"ok":true,"pc":1234,"sp":32767,...}
  ```
- Comandos: todos os do registro de comandos (`load`, `save`, `reset`, `type`, `key`, `run`, `pause`, `step`, `peek`, `poke`, `regs`, `screen`, `screenshot`, `break`, `wait`, `asm`, `compile` …).
- Eventos assíncronos: breakpoint atingido, programa terminou, tela mudou, dado recebido do ZX81.
- Segurança: só escuta em `localhost` por padrão; token opcional.
- Cliente Go pronto em `pkg/go81link`.

### 7.2 Programa ZX81 ↔ host (troca de dados)

- Um dispositivo virtual de E/S ("GoLink") que o programa Z80 acessa por `IN`/`OUT` para enviar e receber bytes do host.
- Porta de E/S configurável; o endereço padrão será escolhido no estudo para não colidir com ULA, ZX Printer, ZXpand e demais periféricos (decisão D-04).
- Também disponível via área de memória "caixa de correio" para programas BASIC (`PEEK`/`POKE`).

## 8. Interfaces

### 8.1 GUI (Fyne)

Janela principal com a tela do emulador; menus Arquivo, Máquina, Ferramentas, Depurar, Exibir, Ajuda; janelas para editor BASIC, editor de textos, montador, gerenciador de fita, depurador, configuração de hardware e teclado visual. Zoom, tela cheia, filtros de imagem.

### 8.2 TUI (tview/tcell)

Mesma organização da GUI em painéis de texto. A tela do ZX81 é desenhada com caracteres (blocos Unicode para os gráficos 2×2 do ZX81; modo de meio-bloco para hi-res aproximado). Funciona via SSH.

### 8.3 CLI

Dois usos:

**a) Parâmetros na chamada** — execução única, ideal para scripts:
```
go81 --gui jogo.p                       # abre a GUI carregando um programa
go81 --tui --machine ts1500 --ram 16k
go81 run jogo.p --headless --frames 500 --screenshot fim.png
go81 basic compile fonte.dbas -o jogo.p
go81 asm rotina.asm -o rotina.p --rem
go81 tape convert jogo.p -o jogo.wav --rate 44100
go81 link --port 8181 --headless
go81 --version
```

**b) Shell interativo** (`go81 shell`) — um pequeno sistema operacional:
```
go81> machine zx81 ram=16k hires=wrx
go81> cd ~/zx81/projetos
go81> edit jogo.dbas
go81> compile jogo.dbas
go81> load jogo.p
go81> run
go81> peek 16396 2
go81> script build.g81
```
Inclui histórico, autocompletar, variáveis, scripts (`.g81`) e códigos de saída padronizados para o shell do Windows e Linux.

## 9. Configuração

- Arquivo TOML em `%APPDATA%\go81\config.toml` (Windows) ou `~/.config/go81/config.toml` (Linux); `--config` para outro arquivo.
- Perfis de máquina nomeados (`--profile zx81-wrx`).
- Pasta de ROMs configurável; por padrão procura em `assets/roms` e na pasta do usuário.

## 10. Versionamento

Formato **X.Y.Z**, começando em **1.1.1**.

| Nível | Quando muda | Nome |
|---|---|---|
| **X** (major) | Ao fechar um grande bloco (emulador, editor, depurador…) | Algo importante na história do ZX81: modelos, pessoas ou empresas |
| **Y** (minor) | A cada nova feature / bloco de programas concluído | Jogos famosos do ZX81 |
| **Z** (patch) | A cada ajuste, bug corrigido ou compilação | Subtítulo curto do trabalho do momento |

Quando X muda, Y e Z voltam para 1. Quando Y muda, Z volta para 1.
Forma completa: `1.1.1 — Sinclair · 3D Monster Maze · "Planta baixa"`.

### Nomes reservados para X

| X | Nome | Bloco planejado |
|---|---|---|
| 1 | Sinclair | Fundação + núcleo Z80 + emulador ZX81 |
| 2 | Altwasser | Ferramentas BASIC (editor, lister, loader, Dignified) |
| 3 | Vickers | Montador / desmontador |
| 4 | Timex | Clones e periféricos |
| 5 | Nine Tiles | Depurador |
| 6 | Lambda | Fitas e mídia |
| 7 | ZXpand | Link e automação |
| 8 | Ferranti | Interfaces finais (GUI/TUI completas) |

(Richard Altwasser projetou o hardware do ZX81; Steve Vickers escreveu a ROM; Nine Tiles foi a empresa da ROM; a Ferranti fabricou a ULA.)

### Nomes para Y (jogos do ZX81, usados em ordem)

3D Monster Maze, Mazogs, Flight Simulation, Football Manager, Black Crystal, Defenda, Krazy Kong, QS Scramble,
1K ZX Chess, Rocket Man, Galaxians, Space Raiders, Fungaloids, City Patrol, Forty Niner, Invasion Force,
Asteroids, Frogger, Pacman, Meteor Storm.

O número de versão vive em `internal/version` e é mostrado por `go81 --version`.

## 11. Testes

- Testes unitários Go (`go test ./...`) em cada pacote.
- Testes Z80 (zexdoc/zexall e afins) em `test/`.
- Testes de compatibilidade com o EightyOne (`test/compat`), com estados de referência em `testdata/compat`.
- Testes "dourados" (golden files) para compilador, montador, lister e conversões de fita.
- CI no GitHub Actions para Windows e Linux.

## 12. Decisões

| ID | Decisão | Estado |
|---|---|---|
| D-01 | Um binário `go81` com subcomandos; GUI/TUI/CLI no mesmo executável | Aceita |
| D-02 | Incluir o ZX80 (mesmo núcleo do ZX81 no EightyOne) | Proposta — confirmar |
| D-03 | Jupiter Ace e família Spectrum fora do escopo | Aceita |
| D-04 | Endereço de porta padrão do GoLink (lado Z80) | Em estudo |
| D-05 | Licença final do projeto (ver seção 13) | **Pendente** |
| D-06 | Configuração em TOML | Proposta |
| D-07 | Extensão `.dbas` para fontes Dignified | Proposta |

## 13. Licenciamento

O EightyOne é **GPL v2 ou posterior** (cabeçalho de `zx81/zx81.cpp`; núcleo Z80 do FUSE também GPL v2+).
Portar seu código para Go cria uma **obra derivada**, que precisa ser distribuída sob GPL.
O `LICENSE` atual do repositório é **BSD 3-Clause**, que não pode ser aplicado ao código derivado.

Opções:

1. **Adotar GPL v3 (ou v2+) para todo o projeto** — mais simples; permite portar código do EightyOne livremente. *Recomendada.*
2. Manter BSD e escrever o emulador do zero, usando o EightyOne só como referência de comportamento (*clean room*) — muito mais trabalho e difícil de provar.

As ROMs têm direitos próprios (a Amstrad permite a distribuição das ROMs Sinclair para uso em emuladores; ROMs de terceiros variam) e ficam separadas do código.
