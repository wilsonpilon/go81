# Go81 — Outline (documento vivo)

Versão atual: **1.4.1** — *Sinclair · Football Manager · "Janela"*
Última atualização: 2026-10-31

## Resumo

Projeto em fase de fundação. Estudo dos fontes do EightyOne 1.41 concluído em nível de mapeamento
(máquinas, periféricos, formatos e arquivos-fonte de referência). Documentação inicial escrita
e esqueleto Go compilável criado (`go81 --version`). Licença definida como GPL v3 ou posterior.
Escopo fechado: ZX81, ZX80, TS1000, TS1500, Lambda 8300, Ringo R-470, TK85, ZX97 Lite e Jupiter Ace.
Núcleo Z80 pronto e idêntico ao do EightyOne. Máquina ZX81 / TS1000 pronta e idêntica ao `zx81.cpp`
original quadro a quadro; roda pela linha de comando (`go81 run`) e salva a tela em PNG. Carrega e grava
programas `.p` com o *flash load/save* do EightyOne, também idêntico ao original.

Agora também roda numa janela (Fyne), com teclado e menus: primeira versão usável de forma interativa.

**Próximo passo:** completar a janela (SAVE, configuração de hardware, som do ZX81 pelo beeper) → 1.4.2+.

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
| Som (ZON X, Quicksilva) | ⬜ | |
| Fala (Parrot, Sweet Talker, Mageco, S-Pack, Talk-Back) | ⬜ | |
| ZXpand / ZXpand+ | ⬜ | |
| ZX Printer / Centronics | ⬜ | |
| IDE MWCFIde, ZX1541, Larken | ⬜ | |
| Cartuchos ROM, joystick, mouse, modem TS2050 | ⬜ | |
| Testes de compatibilidade com EightyOne | ⬜ | |

### 2. BASIC

| Item | Estado | Observação |
|---|---|---|
| Tabela de tokens ZX81/ZX80 | ⬜ | |
| Lister (`.p` → texto) | ⬜ | |
| Loader (texto → `.p`) | ⬜ | |
| Compilador Dignified (`.dbas` → `.p`) | ⬜ | |
| Editor de BASIC | ⬜ | |

### 3. Montador

| Item | Estado | Observação |
|---|---|---|
| Montador Z80 | ⬜ | |
| Macros / condicionais / include | ⬜ | |
| Saída `.p` com REM / loader | ⬜ | |
| Desmontador | ⬜ | |
| Símbolos da ROM (`.sym`) | ⬜ | |

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

### 7. Interfaces

| Item | Estado | Observação |
|---|---|---|
| CLI: parâmetros e subcomandos | 🟨 | `run`, `roms`, `version`, `help` |
| Shell interativo | ⬜ | |
| TUI (tview) | ⬜ | |
| GUI (Fyne) | 🟨 | Janela, teclado, menus, zoom, pausa, turbo; falta SAVE, hardware, som |
| Editor de textos | ⬜ | |

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

## Pendências e decisões abertas

- **D-04 Porta GoLink**: escolher endereço de E/S sem conflito.
