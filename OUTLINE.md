# Go81 — Outline (documento vivo)

Versão atual: **1.6.1** — *Sinclair · Defenda · "Shell e Link"*
Última atualização: 2026-10-31

## Resumo

**Estado:** alfa. O ZX81 / TS1000 roda numa janela (Fyne, com som), no terminal (tview) e pela linha de
comando, e carrega e grava programas `.p`. O shell (`go81 shell`) e a porta de comunicação
(`go81 link`, protocolo próprio + ZRCP do ZEsarUX) usam um registro de comandos único.

**Compatibilidade:** o núcleo Z80, a máquina ZX81 (vídeo, teclado, temporização), a carga/gravação rápida
de fitas e o som foram portados do EightyOne 1.41 e são comparados automaticamente com o código C/C++
original compilado com gcc (`z80diff`, `zx81diff`): nenhuma diferença nos casos testados. Os resultados
ficam gravados e são conferidos por `go test ./...`.

**Especificado (1.5.1, sem código ainda):** montador asm81 (sintaxe asMSX + editor estilo Mega Assembler),
ZX Dignified (base zxtext2p + Basic Dignified), depurador estilo Super-X, ideias do VB81 XuR e do ZEsarUX, Link
compatível com o ZRCP (decisões D-12 a D-18). Material de referência listado em `RESOURCES.md`.

**Bloco 1.x "Sinclair" concluído** com a 1.6.1 (shell e Link).
**Depois:** fitas `.tzx`/`.t81`/`.wav`, depurador, ferramentas de BASIC e assembler, os outros clones
(ZX80, TS1500, Lambda, R-470, TK85, ZX97 Lite, Jupiter Ace) e os periféricos.

**Próximo passo:** bloco 2.x "Altwasser": fitas `.tzx`/`.t81`/`.wav` e o depurador, que já poderá ser usado
pelo shell e pelo Link.

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
| 2026-10-31 | — | Especificação do ZX Dignified (base zxtext2p, recursos do Basic Dignified, escrito do zero em Go): SPEC §6.2, D-16 |
| 2026-10-31 | — | Especificação do depurador no estilo do Super-X: SPEC §6.6, D-17 |
| 2026-10-31 | — | Recursos do Mega Assembler (MSX) adaptados ao ZX81: editor numerado e ferramentas de monitor (SPEC §6.4, §6.6, D-18) |
| 2026-10-31 | — | `RESOURCES.md`: o que fica fora do GitHub e como remontar o projeto em outro PC |
| 2026-10-31 | 1.5.1 | Release técnico "Mapa das ferramentas": especificações e estudos acima, `resume.txt` novo |
| 2026-10-31 | 1.6.1 | `internal/session` (registro de comandos), `go81 shell`, `go81 link` (protocolo Go81 + JSON + ZRCP), cliente `pkg/go81link`, `examples/ola.g81` |
| 2026-10-31 | — | Shell com histórico e Tab; GUI e TUI sobre a sessão, com console (F2) e `--link`; Link por stdio, socket Unix e pipe nomeado |

## Funções e andamento

### 0. Fundação (X=1 Sinclair)

| Item | Estado | Observação |
|---|---|---|
| Estudo do EightyOne (mapa de fontes) | ✅ | Ver SPEC §5.3 |
| SPEC.md / OUTLINE.md / RELEASE.md / MANUAL.md / README.md | ✅ | |
| Esqueleto Go (`go.mod`, `cmd/go81`, `internal/version`) | 🧪 | `go build` e `go81 --version` |
| Definir licença final (D-05) | ✅ | GPL v3 ou posterior |
| CI (GitHub Actions Windows/Linux) | 🧪 | Linux sem GUI, Linux completo, Windows + pacote ZIP como artefato |
| Registro de comandos + `session` | 🧪 | Shell, Link, GUI e TUI usam a mesma sessão |

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
| Tabela de tokens ZX81/ZX80 | 🟨 | ZX81 pronta e conferida com a ROM; falta ZX80 |
| Lister (`.p` → texto) | 🧪 | Igual ao `LIST` da ROM (teste diferencial); `go81 basic list`, `list`, `view-basic` |
| Loader (texto → `.p`) | ⬜ | |
| Especificação do ZX Dignified (base zxtext2p + recursos do Basic Dignified) | ✅ | SPEC §6.2, D-16 |
| Formato de texto do ZX81: escapes de gráficos, inverso, Unicode | 🟨 | Formato zxtext2p + `\{XX}` na saída; falta a entrada (loader) |
| Rótulos, numeração automática, rótulos de laço | ⬜ | |
| Blocos `if/else/endif`, `while`, `repeat`, `do/loop` → `GOTO` | ⬜ | Necessário: ZX81 sem `:` e sem `ELSE` |
| Nomes longos (texto, `FOR` e matrizes → uma letra), `declare` | ⬜ | |
| `define` com argumentos, `include`, alternâncias, comentários | ⬜ | |
| Proto-funções, operadores `++`/`+=`, `true`/`false`, `LET` implícito | ⬜ | |
| `asm … endasm` com o asm81, `#autorun`, `#machine`, `#ram` | ⬜ | Depende do asm81 |
| Relatórios (linhas, variáveis, fluxo) e erros do ZX81 mapeados ao fonte | ⬜ | Depende do Link |
| Compilador ZX Dignified (`.dbas` → `.p`), `go81 basic compile` | ⬜ | |
| "Dignificador" (`.p` clássico → `.dbas`) | ⬜ | |
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
| Editor numerado no estilo Mega Assembler (`AUTO`, `RENUM`, `CHANGE`, `SEARCH`, `MERGE`, `MAP`, `A` com opções) | ⬜ | D-18 |
| Erros com as letras do Mega Assembler; referência cruzada e tabela de rótulos ordenada | ⬜ | |

### 4. Mídia e fitas

| Item | Estado | Observação |
|---|---|---|
| `.p`, `.81`, `.p81` | 🧪 | Carga e gravação |
| `.o`, `.80` (ZX80) | ⬜ | |
| `.t81`, `.tzx` | ⬜ | |
| `.wav` (ler e gerar) | ⬜ | |
| Gerenciador de fita | ⬜ | |

### 5. Depurador (estilo Super-X, SPEC §6.6)

| Item | Estado | Observação |
|---|---|---|
| Especificação (comandos do Super-X adaptados ao ZX81) | ✅ | D-17 |
| Motor: breakpoints (exec, leitura, escrita, E/S, condição, contador), passo a passo | ⬜ | Base: `DebugWin` do EightyOne |
| Pular NMI/INT do vídeo; avanço por quadro e por linha de TV | ⬜ | |
| Linha de comando com comandos de duas letras, `?`, repetição, variáveis `@` | ⬜ | |
| Editores de tela cheia: `D` dump, `A` texto ZX81, `I` desmontagem, `M` assembler, `H` bitmaps 8×8 | ⬜ | Alternar entre eles como no Super-X |
| Pilha de navegação (seguir/voltar de `CALL`/`JP`) | ⬜ | |
| Ferramentas: `BT`, `RT` (relocação), `FL`, `CM`, `FD` (inclusive por instrução), `CK`/`TS`, `CL` | ⬜ | |
| `BL`/`BF` (BASIC), breakpoint em linha de BASIC, variáveis de sistema e do BASIC | ⬜ | |
| `SD` super desmontador (asm81, `DEFB`, `DATA`, linha REM) | ⬜ | |
| Notas por endereço com variáveis de sistema e ROM do ZX81 | ⬜ | |
| Histórico, profiler, cobertura de código | ⬜ | |
| Depurador na GUI, na TUI, no shell e pelo Link | ⬜ | |
| Do Mega Assembler: dump com soma por linha (listagens de revista), `DM` com deslocamento, busca com coringa e texto deslocado, `MS`, `SCR` (gráficos), editor de caracteres | ⬜ | D-18 |
| Do Mega Assembler: cabeçalho de fita, verificar fita × memória, cópia de tela por tecla | ⬜ | |

### 6. Link

| Item | Estado | Observação |
|---|---|---|
| Servidor TCP / stdio / pipe | 🧪 | TCP (127.0.0.1), `-stdio`, socket Unix, pipe nomeado do Windows (só o usuário) |
| Protocolo texto + JSON | 🧪 | `ok`/`err`, dot-stuffing, JSON de uma linha |
| Eventos assíncronos | ⬜ | |
| GoLink (lado Z80) | ⬜ | D-04 |
| Cliente `pkg/go81link` | 🧪 | |
| Modo compatível com o ZRCP do ZEsarUX (DeZog) | 🟨 | D-15; subconjunto de 12 comandos; faltam breakpoints e passo a passo |

### 7. Interfaces

| Item | Estado | Observação |
|---|---|---|
| CLI: parâmetros e subcomandos | 🟨 | `run`, `gui`, `tui`, `shell`, `link`, `roms`, `version`, `help` |
| Shell interativo | 🧪 | Scripts `.g81`, `-c`, histórico em arquivo, Tab completa comandos/arquivos/opções |
| TUI (tview) | 🧪 | Texto + pixels, teclado, console F2, `--link`; sem som |
| GUI (Fyne) | 🟨 | Janela, teclado, menus, zoom, pausa, turbo, som, hardware, SAVE, console F2, Link pelo menu e `--link`; falta fitas e depurador |
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
| Texto da tela (OCR) e captura do texto impresso pela ROM (RST 10H) | 🟨 | OCR pronto (`screen`, `get-ocr`); falta RST 10H |
| Enviar texto como teclas (arquivo / área de transferência) | 🧪 | `type`/`typeln`, `send-keys-string`, Colar (Ctrl+V) na janela, colar no terminal |
| Conversor de áudio → `.p`/`.p81` com visualização de pulsos/bits/bytes | ⬜ | Fitas |
| Fita visual (forma de onda, avançar/voltar) | ⬜ | Fitas |
| RAM em 2000H/8000H/C000H, até 56 KB; escrita na ROM | ⬜ | Emulador |
| Snapshots `.z81`/`.zsf`, snapshots automáticos e *rewind* | ⬜ | Emulador |
| Depuração reversa, histórico, cobertura de código, *watches* | ⬜ | Depurador |
| Localizador de *cheats* | ⬜ | Depurador |
| Configuração por jogo | ⬜ | Configuração |
| Gravação de vídeo/áudio em arquivo; tela em TXT/PBM/BMP | 🟨 | Tela em PNG/BMP/PBM/TXT pronta (shell, `run`, janela); falta vídeo/áudio |
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
| 2026-10-31 | 1.5.1 | `go vet`, `go test ./...` e `go test -tags nogui ./...` (sem mudanças de código além da versão) | OK |
| 2026-10-31 | 1.6.1 | Comandos do shell (digitar e ver a tela, peek/poke, SAVE → arquivo → LOAD → RUN, scripts, máquina) | OK |
| 2026-10-31 | 1.6.1 | Link por TCP: cliente Go, dot-stuffing, JSON, formatos do ZRCP conferidos com o código do ZEsarUX | OK |
| 2026-10-31 | 1.6.1 | `go81 link` real: programa digitado pela porta aparece na tela; `get-registers` pelo ZRCP | OK |
| 2026-10-31 | — | Tab e histórico (núcleo do console), console da TUI (tela simulada) e da GUI (driver de teste do Fyne) | OK |
| 2026-10-31 | — | `go81 gui -link`: programa digitado por TCP aparece na janela (captura) | OK |
| 2026-10-31 | — | Link por stdio, socket Unix e pipe nomeado: testes em Go, cliente .NET e Python | OK |

## Pendências e decisões abertas

- **D-04 Porta GoLink**: escolher endereço de E/S sem conflito.
- **Go mínimo**: as bibliotecas atuais exigem Go 1.25.
- **VB81 XuR**: o código-fonte VB6 (GPL v2+) está em `resource/VB81_RES`; a lista de recursos foi extraída para a
  SPEC §14.2 e a seção 8 abaixo. O executável não deve ser executado (relatos de vírus); passar antivírus na pasta.
