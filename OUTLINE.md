# Go81 — Outline (documento vivo)

Versão atual: **1.4.3** — *Sinclair · Football Manager · "Terminal"*
Última atualização: 2026-10-31

## Resumo

**Estado:** alfa. O ZX81 / TS1000 roda numa janela (Fyne, com som), no terminal (tview) e pela linha de
comando, e carrega e grava programas `.p`.

**Compatibilidade:** o núcleo Z80, a máquina ZX81 (vídeo, teclado, temporização), a carga/gravação rápida
de fitas e o som foram portados do EightyOne 1.41 e são comparados automaticamente com o código C/C++
original compilado com gcc (`z80diff`, `zx81diff`): nenhuma diferença nos casos testados. Os resultados
ficam gravados e são conferidos por `go test ./...`.

**Falta (bloco 1.x "Sinclair"):** shell interativo (`go81 shell`) e porta de comunicação (Link).
**Depois:** fitas `.tzx`/`.t81`/`.wav`, depurador, ferramentas de BASIC e assembler, os outros clones
(ZX80, TS1500, Lambda, R-470, TK85, ZX97 Lite, Jupiter Ace) e os periféricos.

**Próximo passo:** `go81 shell` e a primeira versão do Link.

## Legenda

| Símbolo | Significado |
|---|---|
| ⬜ | A fazer |
| 🟨 | Em andamento |
| ✅ | Feito |
| 🧪 | Feito e testado |

## Novas adições

| Data | Versão | Adição |
|---|---|---|
| 2026-10-30 | 1.1.1 | Documentação inicial (SPEC, OUTLINE, RELEASE, MANUAL, README) e esqueleto Go |
| 2026-10-30 | 1.1.2 | Licença GPL v3; ZX80 confirmado; Jupiter Ace incluído no escopo |
| 2026-10-31 | 1.2.1 | Núcleo Z80, harness C do EightyOne, comparador `z80diff`, testes golden e ZEX |
| 2026-10-31 | 1.3.1 | Máquina ZX81, raster de TV, harness C++ do `zx81.cpp`, comparador `zx81diff`, `go81 run`, `go81 roms`, ROMs embutidas |
| 2026-10-31 | 1.3.2 | Player TZX, patches de LOAD/SAVE da ROM, auto-load, `.p`/`.81`/`.p81`, `--save` |
| 2026-10-31 | 1.4.1 | GUI com Fyne (`go81 gui`), teclado, menus, zoom/borda, turbo/pausa, automação `--capture`, build `nogui` |
| 2026-10-31 | 1.4.2 | Som (beeper do EightyOne + oto), configuração do hardware, SAVE pela janela |
| 2026-10-31 | 1.4.3 | TUI (`go81 tui`): tela de texto e pixels, teclado por caractere, automação `--exit-after` |
| 2026-10-31 | — | Registro do montador asm81 (sintaxe asMSX), das referências em `resource/` e do VB81 XuR como inspiração |
| 2026-10-31 | — | Estudo do código-fonte do VB81 XuR (GPL v2+): recursos extraídos para a SPEC §14.2 e a seção 8 |
| 2026-10-31 | — | Estudo do ZEsarUX (GPL v3), parte ZX81: SPEC §14.3, seção 9 e proposta de Link compatível com ZRCP (D-15) |

## Funções e andamento

### 0. Fundação (X=1 Sinclair)

| Item | Estado | Observação |
|---|---|---|
| Estudo do EightyOne (mapa de fontes) | ✅ | Ver SPEC §5.3 |
| SPEC.md / OUTLINE.md / RELEASE.md / MANUAL.md / README.md | ✅ | |
| Esqueleto Go (`go.mod`, `cmd/go81`, `internal/version`) | 🧪 | `go build` e `go81 --version` |
| Definir licença final (D-05) | ✅ | GPL v3 ou posterior |
| CI (GitHub Actions Windows/Linux) | ⬜ | |
| Registro de comandos + `session` | ⬜ | |

### 1. Emulador

| Item | Estado | Observação |
|---|---|---|
| Núcleo Z80 (todas as instruções, T-states, flags ocultas) | 🧪 | Porta de `z80/*.c`, idêntico ao EightyOne |
| Testes zexdoc/zexall | 🧪 | Mesmos resultados do EightyOne (ver D-09) |
| Comparador com o núcleo C original (`z80diff`) | 🧪 | 116.712 casos sem diferença |
| Máquina ZX81 (ROM, RAM, NMI/HSYNC, vídeo) | 🧪 | `internal/emu/zx81`, idêntica ao `zx81.cpp` |
| Raster de TV (`AccurateDraw`) e paleta | 🧪 | |
| Teclado | 🧪 | `--keys` na CLI |
| "Colourise" e "dirty display" do EightyOne | ⬜ | |
| Carga rápida (ROM patch) `.p` | 🧪 | `zx81/rompatch.cpp`, idêntica ao EightyOne |
| Gravação rápida (SAVE) | ✅ | `--save`; grava `.p` |
| Auto-load (`LOAD ""` automático) | 🧪 | |
| Snapshot `.z81` | ⬜ | `zx81/snap.cpp` |
| ZX80 | ⬜ | |
| TS1000 | 🧪 | ZX81 em NTSC |
| TS1500 | ⬜ | |
| Lambda 8300 (+ cor) | ⬜ | |
| Ringo R-470 / TK85 | ⬜ | |
| ZX97 Lite | ⬜ | |
| Jupiter Ace (RAM/96K, cor ETI, som AY, joystick) | ⬜ | `Ace/ace.cpp` |
| Snapshot `.ace` e fita `.tap` do Ace | ⬜ | |
| Hi-res WRX | 🧪 | |
| Hi-res G007, Memotech, Quicksilva | ⬜ | |
| Geradores de caracteres (dk'tronics, QS, CHR$128, Kayde) | ⬜ | |
| Chroma 81 | ⬜ | |
| Som do ZX81 ("Video Sound", beeper) | 🧪 | Idêntico ao EightyOne |
| Som (ZON X, Quicksilva) | ⬜ | |
| Fala (Parrot, Sweet Talker, Mageco, S-Pack, Talk-Back) | ⬜ | |
| ZXpand / ZXpand+ | ⬜ | |
| ZX Printer / Centronics | ⬜ | |
| IDE MWCFIde, ZX1541, Larken | ⬜ | |
| Cartuchos ROM, joystick, mouse, modem TS2050 | ⬜ | |
| Testes de compatibilidade com EightyOne | 🧪 | `z80diff` e `zx81diff` (14 cenários: vídeo, fita, som) |

### 2. BASIC

| Item | Estado | Observação |
|---|---|---|
| Tabela de tokens ZX81/ZX80 | ⬜ | |
| Lister (`.p` → texto) | ⬜ | |
| Loader (texto → `.p`) | ⬜ | |
| Compilador Dignified (`.dbas` → `.p`) | ⬜ | |
| Editor de BASIC | ⬜ | |

### 3. Montador (asm81 — sintaxe do asMSX, SPEC §6.4)

| Item | Estado | Observação |
|---|---|---|
| Especificação do asm81 (base asMSX, sem MSX, com ZX81) | ✅ | D-12; referências em `resource/` |
| Núcleo: instruções Z80 (inclusive não documentadas), `[ ]` e `.ZILOG` | ⬜ | Portar/inspirar no asMSX (GPL v3) |
| Rótulos locais/globais, `.EQU`, variáveis, expressões e funções | ⬜ | |
| Macros, `REPT`, condicionais, `.INCLUDE`, `.INCBIN`, `.PHASE` | ⬜ | |
| Símbolos ZX81: variáveis de sistema, rotinas da ROM, caracteres, tokens | ⬜ | `resource/81asm`, desmontagem da ROM |
| Texto no conjunto de caracteres do ZX81 | ⬜ | |
| Saída `.p` completa (sysvars, REM, `RAND USR`, D_FILE, auto-run) | ⬜ | |
| Saídas `.p81`, `.t81`, `.tzx`, `.wav`, `.bin`, `.hex`, `.lst`, `.sym` | ⬜ | |
| Linhas de BASIC no fonte | ⬜ | Depende do loader de BASIC |
| Avisos do ZX81 (M1NOT, RAM, registradores do vídeo) | ⬜ | |
| `go81 asm --run` e ligação com o depurador (`.BREAK`, símbolos) | ⬜ | |
| Modo de compatibilidade Pasmo / SjASMPlus / z88dk | ⬜ | |
| Desmontador (remontável, com símbolos) | ⬜ | |

### 4. Mídia e fitas

| Item | Estado | Observação |
|---|---|---|
| `.p`, `.81`, `.p81` | 🧪 | Carga e gravação |
| `.o`, `.80` (ZX80) | ⬜ | |
| `.t81`, `.tzx` | ⬜ | |
| `.wav` (ler e gerar) | ⬜ | |
| Gerenciador de fita | ⬜ | |

### 5. Depurador

| Item | Estado | Observação |
|---|---|---|
| Registradores / desmontagem / memória | ⬜ | |
| Breakpoints (exec, leitura, escrita, E/S, condição) | ⬜ | |
| Histórico e profiler | ⬜ | |

### 6. Link

| Item | Estado | Observação |
|---|---|---|
| Servidor TCP / stdio / pipe | ⬜ | |
| Protocolo texto + JSON | ⬜ | |
| Eventos assíncronos | ⬜ | |
| GoLink (lado Z80) | ⬜ | D-04 |
| Cliente `pkg/go81link` | ⬜ | |
| Modo compatível com o ZRCP do ZEsarUX (DeZog) | ⬜ | D-15 (proposta) |

### 7. Interfaces

| Item | Estado | Observação |
|---|---|---|
| CLI: parâmetros e subcomandos | 🟨 | `run`, `gui`, `tui`, `roms`, `version`, `help` |
| Shell interativo | ⬜ | |
| TUI (tview) | 🧪 | Texto + pixels, teclado; sem som |
| GUI (Fyne) | 🟨 | Janela, teclado, menus, zoom, pausa, turbo, som, hardware, SAVE; falta fitas e depurador |
| Editor de textos | ⬜ | |

### 8. Ferramentas inspiradas no VB81 XuR (SPEC §14.2)

| Item | Estado | Observação |
|---|---|---|
| BASIC: renumerar, REM para código de máquina, números → strings, texto em REM | ⬜ | Bloco BASIC |
| BASIC como texto (ZxToken / CBI), `SAVE "X.TXT"` | ⬜ | Bloco BASIC |
| Importar fontes TASM e ASM2 no asm81 | ⬜ | Montador |
| Desmontador com exportação (asm81, TASM), marcação de dados/texto | ⬜ | Depurador |
| Editor binário, comparar memória, janela de variáveis, memória como texto | ⬜ | Depurador |
| Ferramentas de tela: EasyDraw, importar BMP, salvar tela como SCR/TXT/BAS, UDG | ⬜ | |
| WAV de carga rápida (OTLA) e lenta; decodificar WAV; `.T81`, `.TZX`, `.SBB`, S-record | ⬜ | Fitas |
| Impressora: papel, várias colunas, PDF/PNG; Centronics com buffer | ⬜ | Periféricos |
| Placas: Mageco 16 cores, AGB, Octavia, caneta óptica, cartões de E/S, RomServer | ⬜ | Periféricos |
| Editor/compilador Forth | ⬜ | Futuro (também para o Jupiter Ace) |
| Janela para iniciantes (clicar em comandos), teclas configuráveis, ROMs extras | ⬜ | GUI |

### 9. Ideias do ZEsarUX para o ZX81 (SPEC §14.3)

| Item | Estado | Observação |
|---|---|---|
| Detecção automática de WRX e Chroma81 | ⬜ | |
| Texto da tela (OCR) e captura do texto impresso pela ROM (RST 10H) | ⬜ | Link, TUI, acessibilidade |
| Enviar texto como teclas (arquivo / área de transferência) | ⬜ | Link, shell, GUI |
| Conversor de áudio → `.p`/`.p81` com visualização de pulsos/bits/bytes | ⬜ | Fitas |
| Fita visual (forma de onda, avançar/voltar) | ⬜ | Fitas |
| RAM em 2000H/8000H/C000H, até 56 KB; escrita na ROM | ⬜ | Emulador |
| Snapshots `.z81`/`.zsf`, snapshots automáticos e *rewind* | ⬜ | Emulador |
| Depuração reversa, histórico, cobertura de código, *watches* | ⬜ | Depurador |
| Localizador de *cheats* | ⬜ | Depurador |
| Configuração por jogo | ⬜ | Configuração |
| Gravação de vídeo/áudio em arquivo; tela em TXT/PBM/BMP | ⬜ | GUI / CLI |
| Navegador de jogos online (zx81.nl) | ⬜ | GUI |

## O que já foi testado

| Data | Versão | Teste | Resultado |
|---|---|---|---|
| 2026-10-30 | 1.1.1 | `go build ./...`, `go vet ./...`, `go test ./...`, `go81 --version` | OK |
| 2026-10-30 | 1.1.2 | `go build ./...`, `go vet ./...`, `go test ./...`, `go81 --version` | OK |
| 2026-10-31 | 1.2.1 | `z80diff -per 64 -runs 1000 -runlen 5000` contra o C original | 116.712 casos iguais |
| 2026-10-31 | 1.2.1 | Teste golden (7.296 casos) em `go test ./...` | OK |
| 2026-10-31 | 1.2.1 | ZEXDOC/ZEXALL (`GO81_LONG=1`) | Iguais ao EightyOne |
| 2026-10-31 | 1.2.1 | Benchmark `BenchmarkStep` | ~44 M instruções/s |
| 2026-10-31 | 1.3.1 | `zx81diff` contra o `zx81.cpp` original (7 cenários, 4.226 quadros) | Idênticos |
| 2026-10-31 | 1.3.1 | Golden ZX81 + testes da CLI em `go test ./...` | OK |
| 2026-10-31 | 1.3.1 | Benchmark `BenchmarkFrame` | ~770 quadros/s |
| 2026-10-31 | 1.3.2 | `zx81diff` com fita (11 cenários, incluindo LOAD com auto-load) | Idênticos |
| 2026-10-31 | 1.3.2 | Ciclo SAVE → arquivo → LOAD → RUN pela CLI (teste automático) | OK |
| 2026-10-31 | 1.4.1 | GUI aberta com programa, auto-load e captura da janela (`--capture`) | OK |
| 2026-10-31 | 1.4.1 | Mapeamento de teclas (teste unitário) e build `-tags nogui` | OK |
| 2026-10-31 | 1.4.2 | `zx81diff` com som (14 cenários) | Idênticos |
| 2026-10-31 | 1.4.2 | Fila de áudio e SAVE chegando à janela (testes unitários); janela com som aberta | OK |
| 2026-10-31 | 1.4.3 | TUI em tela simulada do tcell: digitar e rodar programa, cursor, modo pixels | OK |

## Pendências e decisões abertas

- **D-04 Porta GoLink**: escolher endereço de E/S sem conflito.
- **CI**: GitHub Actions para Windows e Linux (a janela precisa de gcc no runner; o resto roda com `-tags nogui`).
- **Go mínimo**: as bibliotecas atuais exigem Go 1.25.
- **VB81 XuR**: o código-fonte VB6 (GPL v2+) está em `resource/VB81_RES`; a lista de recursos foi extraída para a
  SPEC §14.2 e a seção 8 abaixo. O executável não deve ser executado (relatos de vírus); passar antivírus na pasta.
