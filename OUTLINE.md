# Go81 — Outline (documento vivo)

Versão atual: **1.1.1** — *Sinclair · 3D Monster Maze · "Planta baixa"*
Última atualização: 2026-10-30

## Resumo

Projeto em fase de fundação. Estudo dos fontes do EightyOne 1.41 concluído em nível de mapeamento
(máquinas, periféricos, formatos e arquivos-fonte de referência). Documentação inicial escrita
e esqueleto Go compilável criado (`go81 --version`). Nenhuma parte da emulação existe ainda.

**Próximo passo:** núcleo Z80 (`internal/emu/z80`) com testes zexdoc/zexall → versão 1.2.x (*Mazogs*).

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

## Funções e andamento

### 0. Fundação (X=1 Sinclair)

| Item | Estado | Observação |
|---|---|---|
| Estudo do EightyOne (mapa de fontes) | ✅ | Ver SPEC §5.3 |
| SPEC.md / OUTLINE.md / RELEASE.md / MANUAL.md / README.md | ✅ | |
| Esqueleto Go (`go.mod`, `cmd/go81`, `internal/version`) | 🧪 | `go build` e `go81 --version` |
| Definir licença final (D-05) | ⬜ | Pendente de decisão |
| CI (GitHub Actions Windows/Linux) | ⬜ | |
| Registro de comandos + `session` | ⬜ | |

### 1. Emulador

| Item | Estado | Observação |
|---|---|---|
| Núcleo Z80 (todas as instruções, T-states, flags ocultas) | ⬜ | Porta de `z80/*.c` |
| Testes zexdoc/zexall | ⬜ | |
| Máquina ZX81 (ROM, RAM, NMI/HSYNC, vídeo) | ⬜ | `zx81/zx81.cpp` |
| Teclado | ⬜ | |
| Carga rápida (ROM patch) `.p` | ⬜ | `zx81/rompatch.cpp` |
| Snapshot `.z81` | ⬜ | `zx81/snap.cpp` |
| ZX80 | ⬜ | Depende de D-02 |
| TS1000 / TS1500 | ⬜ | |
| Lambda 8300 (+ cor) | ⬜ | |
| Ringo R-470 / TK85 | ⬜ | |
| ZX97 Lite | ⬜ | |
| Hi-res: WRX, G007, Memotech, Quicksilva | ⬜ | |
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
| `.p`, `.81`, `.p81`, `.o`, `.80` | ⬜ | |
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
| CLI: parâmetros e subcomandos | 🟨 | Só `--version`/`--help` |
| Shell interativo | ⬜ | |
| TUI (tview) | ⬜ | |
| GUI (Fyne) | ⬜ | |
| Editor de textos | ⬜ | |

## O que já foi testado

| Data | Versão | Teste | Resultado |
|---|---|---|---|
| 2026-10-30 | 1.1.1 | `go build ./...`, `go vet ./...`, `go test ./...`, `go81 --version` | OK |

## Pendências e decisões abertas

- **D-05 Licença**: EightyOne é GPL v2+; o `LICENSE` atual é BSD-3. Precisa decidir antes de portar código.
- **D-02 ZX80**: confirmar se entra no escopo.
- **D-04 Porta GoLink**: escolher endereço de E/S sem conflito.
