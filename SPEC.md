# Go81 — Especificação

Versão do documento: 1.10.8 · Última revisão: 2026-11-03

Este documento descreve **o que** o Go81 deve ser. O andamento do trabalho fica em [OUTLINE.md](OUTLINE.md).

---

## 1. Visão geral

O Go81 é um ambiente integrado para o ZX81 escrito em Go. Ele reúne em um só executável:

1. Emulador do ZX81, ZX80, clones (Timex, Lambda, Ringo, TK85, ZX97 Lite) e Jupiter Ace, compatível com o EightyOne 1.41
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
| Timex Sinclair TS1500 | `ts1500.rom` | 16K, ROM com extras; M1NOT sempre; sem cópia da ROM em 32K–48K. **Emulado desde 1.8.0** |
| Lambda 8300 / PC8300 | `lambda8300.rom`, `lambda8300colour.rom`, `pc8300timex.rom` | Clone com cor opcional |
| Ringo R-470 | `ringo470.rom` | Clone brasileiro; 16K; sem RAM em 8K–16K. **Emulado desde 1.8.0** |
| Microdigital TK85 | `tk85.rom` | Clone brasileiro; 16K; ROM de 10K. **Emulado desde 1.8.0** |
| ZX97 Lite | `zx97.rom` | ZX81 expandido (bancos de memória) |
| Sinclair ZX80 | `zx80.rom` | Antecessor do ZX81, mesmo núcleo no EightyOne (D-02); vídeo próprio (`zx80_do_scanline`). **Emulado desde 1.8.1**, idêntico ao EightyOne |
| Jupiter Ace | `jupiterace.rom` | Máquina Forth da Jupiter Cantab, criada por ex-integrantes da equipe do ZX81 (D-08) |

Os modelos Timex compatíveis com a linha ZX80/ZX81 são o **TS1000** e o **TS1500** (o TS2068 é compatível com o Spectrum e fica fora).
Os clones brasileiros e de outros países incluídos são **TK85**, **Ringo R-470** e **Lambda 8300**.

ROMs alternativas: `tree-forth.rom`, `zx81-forth.rom`, `zx.aszmic.e04.rom`, `zx.aszmic.e07.rom`, overlays ZXpand (`*.zxpand.ovl`).

### 2.2 Fora do escopo

- ZX Spectrum (todos os modelos), TC2048, TS2068/TC2068
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

Restrições por máquina seguem o EightyOne: por exemplo, TS1500, R-470 e TK85 não oferecem a expansão de 48K.

### 2.3.1 Jupiter Ace

Opções do EightyOne para o Ace (`Ace/ace.cpp`, `HW_.cpp`):

| Categoria | Opções |
|---|---|
| Memória | Expansões de RAM, incluindo o modo 96K paginado |
| Cor | Nenhuma, ETI |
| Som | Nenhum, Ace User (AY), Boldfield (AY) |
| Joystick | Nenhum, Boldfield, Programável |

O Ace usa **Forth** em vez de BASIC. As ferramentas de BASIC não se aplicam a ele; o editor de textos,
o montador, o depurador, o Link e as fitas funcionam normalmente. Um editor/carregador de Forth para o Ace fica como fase futura.

### 2.4 Formatos de arquivo

| Tipo | Extensões | Ler | Gravar |
|---|---|---|---|
| Programa ZX81 | `.p`, `.81`, `.p81` | ✔ | ✔ |
| Programa ZX80 | `.o`, `.80` | ✔ | ✔ |
| Fita | `.tzx`, `.t81`, `.wav` | ✔ | ✔ |
| Snapshot | `.z81` (formato EightyOne) | ✔ | ✔ |
| Jupiter Ace | `.ace` (snapshot), `.tap` (fita do Ace) | ✔ | ✔ |
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
│   ├── cli/                  # linha de comando (go81 run, roms, ...)
│   ├── session/              # estado central + registro de comandos
│   ├── emu/
│   │   ├── z80/              # núcleo Z80 (porta do núcleo FUSE/EightyOne)
│   │   ├── zx81/             # família ZX81: memória, portas, vídeo, raster de TV
│   │   ├── ace/              # Jupiter Ace (futuro)
│   │   ├── video/            # geração de quadro, hi-res, cor
│   │   ├── audio/            # som, fala
│   │   ├── input/            # teclado, joystick, mouse
│   │   ├── peripherals/      # ram, zxpand, printer, ide, cartuchos, modem...
│   │   └── rom/              # catálogo e carga de ROMs
│   ├── debug/                # breakpoints, histórico, profiler, símbolos
│   ├── media/
│   │   ├── tape/             # player TZX do EightyOne, .p/.81/.p81 (o, 80, t81, tzx, wav: futuro)
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
  *Desde 1.9.5:* `internal/media/z81` lê e grava o formato de `zx81/snap.cpp` ([MACHINE], [CPU], [ZX81],
  [MEMORY] com `*NNNN XX`, [ADVANCED]; nomes antigos HSYNC/ROW; seções desconhecidas puladas; bytes depois de
  FFFFh descartados). O snapshot é tirado entre quadros; a memória vai de ROMTOP+1 até RAMTOP, como no
  EightyOne. `snapshot save|load`, `load x.z81` e Arquivo → Carregar/Salvar snapshot.

### 5.2 Critério de compatibilidade

Para cada máquina e conjunto de periféricos:

1. Mesma ROM, mesma entrada, mesmo número de T-states → mesmo estado de registradores e memória que o EightyOne.
2. Testes Z80 padrão (zexdoc/zexall, testes de flags) passando.
3. Conjunto de programas de referência (jogos, demos hi-res) com quadro de vídeo idêntico pixel a pixel.

Os resultados de referência são gerados com o `81emul/exe/EightyOne.exe` e guardados em `testdata/compat/`.

### 5.2.1 Núcleo Z80 — como a compatibilidade é verificada

- `internal/emu/z80` é uma porta fiel do núcleo do EightyOne (derivado do FUSE). Ele reproduz os resultados,
  as flags (inclusive os bits 3 e 5 não documentados), os T-states, os M-cycles, a contagem de M1 e a
  **ordem exata dos acessos ao barramento**, incluindo qual leitura é "opcode fetch" e qual é "operand read".
  Isso importa no ZX81, porque o hardware de vídeo depende do opcode fetch.
- `test/compat/z80ref/harness.c` compila o núcleo **original** em C (`81emul/source/Source/z80`) com gcc.
- `go run ./test/compat/z80diff` gera casos (todos os opcodes de todos os prefixos com estados aleatórios,
  interrupções, NMI e programas aleatórios longos) e compara os dois núcleos linha a linha, incluindo um
  hash de todos os acessos ao barramento.
- Um subconjunto dos resultados fica em `internal/emu/z80/testdata/eightyone.txt.gz`. O `go test ./...`
  o confere sem precisar de gcc nem do `81emul/`.
- ZEXDOC/ZEXALL rodam com `GO81_LONG=1 go test ./internal/emu/z80/ -run Zex -v`.

**Diferença deliberada:** endereços são sempre reduzidos a 16 bits (ex.: `IX+d` acima de `0xFFFF` dá a volta,
como no Z80 real). No EightyOne esses valores passam ao C sem máscara e podem acessar fora da memória,
o que é comportamento indefinido.

### 5.2.2 Máquina ZX81 — como a compatibilidade é verificada

- `internal/emu/zx81` porta `zx81.cpp` (memória, portas, `opcode_fetch` que gera o vídeo, `do_scanline`
  com NMI, HSYNC, VSYNC e inserção de WAIT) e o `AccurateDraw` de `AccDraw_.cpp` (raster da TV).
- `test/compat/zx81ref/harness.cpp` compila o **`zx81.cpp` original** com g++. A interface VCL e os
  periféricos são substituídos por stubs ("não instalado") em `shim/go81stubs.h`.
- `go run ./test/compat/zx81diff` roda cenários (boot com 1K/16K, PAL/NTSC, WRX, digitação e execução de
  um programa BASIC, FAST/SLOW) e compara, **a cada quadro**, os registradores, os T-states, um hash de
  todas as linhas geradas (pixels e sincronismo), a imagem da TV e toda a memória.
- O resultado fica em `internal/emu/zx81/testdata/eightyone.txt.gz` e é conferido por `go test ./...`.
- Tolerâncias de sincronismo e paleta seguem os valores que a janela *Artifacts* do EightyOne aplica com
  os controles no padrão: VSYNC 283–323 linhas (NTSC: 223–263), HSYNC 390, cores "vibrant".

**Fitas (1.3.2):** o harness também compila o player TZX original (`tzx/*.cpp`), o `rompatch.cpp`
(interceptação de LOAD/SAVE na ROM) e reproduz o gerenciador de fitas e o *auto-load* (`TZXMan.cpp`,
`main_.cpp`). Os cenários `load-*` geram programas `.p` com o `SAVE` da própria ROM, carregam com
*flash load* + *auto-load* e executam `RUN`; tudo idêntico ao EightyOne quadro a quadro.
*Desde 1.10.7:* cenários `load-tzx-16k`, `load-t81-16k` e `load-tzx-by-name`: fitas de vários blocos
escritas pelo Go81 (`tape.SaveTZX`, `tape.SaveT81`) e lidas pelo `TTZXFile::LoadFile` original; o
`LOAD "PLOT"` passa pelo primeiro programa. O harness ganhou `resave=ARQ` (lê a fita e a grava com
`TTZXFile::SaveFile`, de `tzxsave.cpp`): o `zx81diff` confere que o `.tzx` e o `.t81` do Go81 são iguais
byte a byte aos do EightyOne para a mesma fita.

**ZX80 (1.8.1):** o harness ganhou `zx80=1` (máquina `MACHINEZX80`, `zx80_do_scanline`, patches de fita do ZX80,
auto-load com W + ENTER) e o CRC das ROMs igual ao `CRC32Block` do EightyOne (sem a inversão final), que ele
usa para reconhecer a ROM do ZX80. Cenários `zx80-*`: boot (16K, 1K, NTSC), digitação e execução, som pelo
vídeo e um programa gravado com o SAVE do ZX80 e carregado de volta; todos idênticos quadro a quadro.

**Som (1.4.2):** o beeper (`CSound::Beeper`/`Frame` de `sound/sound.cpp`) foi portado e é comparado
quadro a quadro pelo `zx81diff` (hash do buffer de áudio, 44,1 kHz, 16 bits, estéreo, volume 15/31), nos
cenários `sound-*` com o "Video Sound" ligado.

**Diferenças conhecidas (1.3.1):**
- O EightyOne começa o raster numa linha aleatória (`random(256)`); o Go81 começa na linha 0, para ser
  determinístico. Isso só muda em que ponto do primeiro segundo a imagem se estabiliza.
- Opções de "colourise" (anotações coloridas de sincronismo/NMI) e "dirty display" (ruído/fantasma)
  ainda não existem.
- A porta 0x01 (leitura da configuração do EightyOne) devolve 0 para bytes fora da estrutura `ZX81`.
- Periféricos (ZXpand, Chroma, impressora, som, fala, G007/Memotech/Quicksilva, etc.) virão nas próximas
  versões; até lá se comportam como "não instalados".
- Nome de arquivo sem nenhum caractere do ZX81 (ex.: `!.p`): o EightyOne escreve fora do buffer; o Go81
  recusa o arquivo com uma mensagem de erro.
- O SAVE grava direto em arquivo `.p` (`--save`); o EightyOne grava num bloco da fita do gerenciador.
- Na janela, a gravação rápida (SAVE sem a espera de 5 s) está sempre ligada; no EightyOne ela depende de o
  gerenciador de fitas estar selecionado como saída.

### 5.3 Fonte de referência no EightyOne

| Área | Arquivos em `81emul/source/Source/` |
|---|---|
| Z80 | `z80/*.c` |
| ZX81/ZX80 | `zx81/zx81.cpp`, `zx81/snap.cpp`, `zx81/rompatch.cpp` |
| Jupiter Ace | `Ace/ace.cpp`, `Ace/ace.h` |
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

*Estado em 1.6.3:* o **lister** existe (`internal/basic`): `.p` → texto no formato do zxtext2p, mais `\{XX}` para
bytes que não são texto; a listagem segue a regra de espaços da rotina TOKENS da ROM e é conferida contra o
`LIST` do ZX81 emulado. Usos: `go81 basic list`, `list` (shell, consoles), `view-basic` (ZRCP).

*Desde 1.10.4:* conversor sem o emulador: `go81 txt2p` / `go81 p2txt` (vários arquivos e curingas), `txt2p` /
`p2txt` no shell e nos consoles, Ferramentas → Programa BASIC → Converter na janela (`session.TextToPFile`,
`session.PToTextFile`). Conferido com o `zxtext2p.c` original: os programas sem código de máquina passam;
limites do zxtext2p: `\{XX}`, escapes dentro de REM, `**` e espaços no fim da linha.
*Desde 1.10.6:* `p2txt -legivel` (`basic.TextReadable`): palavras-chave dentro de REM e de strings por
extenso e números com valor oculto diferente só com os dígitos; para leitura, sem garantia de ida e volta.
*Desde 1.10.5:* `""` dentro de uma string vale como a aspa (C0h), como no ZX81 e no `.AZX` do Basic
Dignified; `.azx` é texto BASIC (`load`, janela, `txt2p`); `p2txt` reconhece um arquivo de texto e indica o
`txt2p`.
*Desde 1.7.1:* o **loader** (texto → `.p`, `go81 basic make`, `make`, e `load` de `.bas`/`.txt`).
Os números recebem o valor de 5 bytes calculado pela rotina DEC-TO-FP da própria ROM, executada no Z80
emulado (a ROM trunca; um arredondamento exato daria bytes diferentes em frações). O lister confere cada
linha com o loader e escapa (`\{XX}`) só o que seria lido de outro jeito, então listar e voltar é sem perda
(conferido nos 164 `.p` de exemplo do EightyOne). O `.p` parte de um programa vazio gravado pela ROM (16K,
tela cheia; opção de tela compacta para 1K). Falta o editor.

*Desde 1.7.7:* os primeiros utilitários de §6.1.1 já funcionam no shell, nos consoles e na janela, sobre o
programa do ZX81: `find`, `delete de,até` e `rem linha tamanho [byte]` (`internal/basic/edit.go`). Eles
mudam a lista de linhas e recarregam o programa pela fita (as variáveis se perdem, como num `LOAD`).
*Desde 1.9.3:* `renum [início [passo [de,até]]]` (`internal/basic/renum.go`): renumera o programa ou um bloco,
sem deixar o bloco passar por cima das outras linhas, e reescreve o número (dígitos e os 5 bytes depois de 7Eh)
que vem logo após GOTO, GOSUB, RUN, LIST e LLIST quando ele é constante e fecha a linha; um destino que não
existe passa a ser a linha seguinte (o que o ZX81 faria); números calculados geram aviso. `auto [início
[passo]]`: o shell e os consoles passam a receber linhas de BASIC com o número dado pelo prompt (`auto 120>`)
até uma linha vazia ou um ponto, e então as incluem no programa de uma vez.

*Desde 1.7.10:* tabela de palavras-chave e de caracteres do **ZX80** (`internal/basic/zx80.go`, do
`zx80BasicLister.cpp` do EightyOne e conferida com a tabela da ROM do ZX80) e lister de `.o`/`.80`
(`go81 basic list`, `list`). Linha do ZX80: número (2 bytes), corpo, NEWLINE; números só em dígitos; o
programa vai de 4028h até VARS (4008h).

#### 6.1.1 Editor no estilo do ZX81 (intenção registrada em 2026-11-01)

Um editor de BASIC que funciona **como a edição original do ZX81**: a parte de cima da tela mostra a listagem
(com o cursor de linha `>` na linha atual) e as linhas de baixo são a área de entrada, onde a linha é digitada ou
trazida para edição (EDIT). ENTER manda a linha para o programa: com número, entra no lugar certo da listagem; um
número sozinho apaga a linha; sem número, é um comando direto. Erros de sintaxe ficam marcados com o `S` como no
ZX81 e a linha não entra.

Duas variações, escolhidas pelo usuário:

| Modo | Como se digita |
|---|---|
| **Teclas com mnemônicos** (clássico) | Igual ao ZX81: cursores `K`, `L`, `F`, `G`; uma tecla dá a palavra-chave inteira (`P` = `PRINT` no início da linha), SHIFT/FUNCTION dão os símbolos e funções; o teclado na tela (F3) e a imagem do teclado (F1) servem de guia |
| **Digitação direta** | Letra por letra, como num editor comum (`PRINT`, `GOTO`, `**`…); a linha vira tokens ao entrar, pelo mesmo loader de texto do Go81 (§6.1), e volta a texto pelo lister |

As duas variações trabalham sobre o mesmo programa e podem ser trocadas a qualquer momento. O programa pode ser
mandado para o emulador e trazido dele (`.p` na memória), além de ser aberto e gravado como `.p`/`.bas`.

**Utilitários** (comandos na área de entrada e no menu do editor):

| Utilitário | O que faz |
|---|---|
| `RENUM [início[,passo[,de,até]]]` | Renumera as linhas e corrige `GOTO`, `GOSUB`, `RUN`, `LIST`, `LLIST` com número constante; avisa os desvios calculados (`GOTO A*10`) que não dá para corrigir |
| `AUTO [início[,passo]]` | Numeração automática: cada ENTER já abre a próxima linha numerada; ENTER numa linha vazia sai |
| `DELETE de,até` | Apaga um bloco de linhas |
| `FIND texto` | Procura linhas pelo texto da listagem (palavras-chave incluídas); próxima/anterior ocorrência |
| `REM tamanho[,byte]` | Cria uma linha `REM` com N bytes (para código de máquina), a ser preenchida pelo asm81 (§6.4) ou pelo depurador |
| Ir para a linha / listar faixa | Move a listagem e o cursor de linha |

Outros a avaliar: mover/copiar bloco de linhas (com renumeração), lista de variáveis e de referências
cruzadas (onde cada linha é chamada). (MERGE e a contagem de memória entraram em 1.10.6.)

O editor de assembler numerado no estilo do Mega Assembler (§6.4, D-18) usa a mesma tela (listagem em cima,
entrada embaixo) e os mesmos utilitários de linha (AUTO, RENUM, DELETE, FIND); por isso os dois serão feitos sobre
o mesmo núcleo.

*Desde 1.10.0:* núcleo em `internal/basiced` (estado, teclas, comandos e a tela de 32×24 em códigos do ZX81,
testado sem interface) e a vista na janela (`internal/ui/gui/editor.go`): F7 troca a tela do ZX81 pelo editor
na mesma janela, desenhado com os caracteres da ROM. Modo de digitação direta; comandos RENUM, AUTO, DELETE,
FIND, LIST, EDIT, REM, NEW e HELP; Tab = EDIT; colar várias linhas digita o programa. Ao voltar ao emulador, o
programa alterado é carregado no ZX81 (`Session.SetProgram`).
*Desde 1.10.1:* modo com as teclas do ZX81 (F8, `internal/basiced/keys.go`): cursores K/L/F/G com as tabelas
K-UNSHIFT, K-SHIFT, K-FUNCT e K-GRAPH da ROM, como MODE-SORT; K no começo da linha, depois do número e depois
de THEN; códigos de edição 70h–79h (setas, GRAPHICS, EDIT, NEWLINE, RUBOUT, FUNCTION). A linha é guardada em
códigos do ZX81 e vira texto (`basic.CodesText`) para o mesmo `ParseLine` no ENTER; EDIT traz a linha em
códigos (`basic.TypedCodes`). O mapa do teclado com CTRL mostra o cursor do editor.
*Desde 1.10.2:* a sintaxe é conferida pela ROM (`basic.CheckSyntax`): a linha vai para E_LINE de um ZX81
emulado de rascunho e LINE-SCAN (0CBAh) roda em modo de sintaxe; um erro passa por ERROR-1 (RST 08) com
CH_ADD no caractere do erro, onde o editor põe o cursor S. Comando `RUN [n]`: carrega o programa no ZX81 e
digita RUN.
*Desde 1.10.6:* `LOAD`, `SAVE` e `MERGE "arquivo"` no editor, pela interface `basiced.Files` (a janela usa
`session.ReadProgramLines` / `WriteProgramLines`: `.p`, `.81`, `.p81` e texto, sem o ZX81; sem nome, a
janela pergunta); `MEM` (`Editor.Size`, `Editor.Free`: RAMTOP − (PROG + programa + tela + VARS + E_LINE +
pilha), tela cheia a partir de 4K); rolagem da listagem sem mudar a linha atual (`Editor.Scroll`,
Ctrl+↑↓ e Ctrl+PgUp/PgDn). Menu Exibir → Editor de BASIC com abrir, juntar, gravar e memória.

### 6.2 Compilador ZX Dignified (BASIC moderno → BASIC do ZX81)

Linguagem e compilador **próprios do Go81, escritos do zero em Go** (decisão D-16). Partimos do **zxtext2p**
(texto → `.p`, com rótulos) e chegamos perto do **Basic Dignified Suite**, sem copiar o código de nenhum dos dois
e sem chamá-los como programas externos. Arquivos-fonte: `.dbas` (D-07). Comando previsto:
`go81 basic compile jogo.dbas -o jogo.p` (e `--run` para rodar no emulador).

**Por que o compilador precisa fazer mais que o Dignified:** o BASIC do ZX81 tem **um comando por linha**
(não existe `:`), **não tem `ELSE`**, aceita **um único comando depois do `THEN`**, só permite **uma letra** em
variáveis de texto (`A$`), de laço `FOR` e em matrizes, e usa **1** para verdadeiro. Por isso os blocos de várias
linhas, `ELSE`, laços e funções são compilados em várias linhas com `GOTO`/`GOSUB` gerados automaticamente.

**Base do zxtext2p (formato de texto do BASIC do ZX81)**
- Maiúsculas e minúsculas indiferentes; espaços no início e no fim ignorados; linhas em branco ignoradas.
- Escapes: `\"` (aspas dentro de texto), gráficos de bloco com dois caracteres de "arte ASCII" (`\:'`, `\''`,
  `\..` …), `%` antes de um caractere para vídeo inverso; também caracteres Unicode equivalentes (▘▝▀▖▌▞▛▒ …,
  a mesma tabela da TUI).
- Continuação de linha com `\` no fim; comentários de fonte com `#` (zxtext2p) — no ZX Dignified, `##`.
- Modo clássico (com números de linha) continua aceito: o mesmo compilador serve de *loader* de BASIC.

**Recursos inspirados no Basic Dignified (adaptados ao ZX81)**

| Recurso | Exemplo | Resultado no ZX81 |
|---|---|---|
| Sem números de linha | — | numeração automática (`#start 10`, `#step 10`) |
| Rótulos | `{inicio}` … `goto {inicio}`; `{@}` = a própria linha | `GOTO 10` |
| Rótulos de laço | `laco{ … if x then exit … }` | `GOTO` de volta e de saída |
| Blocos estruturados (novo) | `if … then … else … endif`, `while … wend`, `repeat … until`, `do … loop` | `IF NOT (…) THEN GOTO …` gerados |
| Nomes longos | `vidas`, `nome$` | numéricas: nome longo nativo ou abreviado (opção de economia de memória); texto, `FOR` e matrizes: uma letra atribuída automaticamente; `declare nome:N` força a escolha |
| Definições | `define [VEL][5]`, com argumentos `define [pk][poke 16507,[0]]` | substituição textual |
| `PRINT AT` curto | `[?](x,y) "texto"` | `PRINT AT y,x;"texto"` |
| Proto-funções | `func .soma(a,b)` … `ret a+b`; chamada `r = .soma(1,2)` | `GOSUB` com atribuições antes e depois |
| Operadores | `x++`, `x--`, `+=`, `-=`, `*=`, `/=` | `LET X=X+1` … |
| Booleanos | `true`, `false` | `1`, `0` |
| Comentários | `##` (some), `rem`/`'` (vira `REM`), blocos `###` … `###` | |
| Alternâncias | `#debug print x`, `keep #debug` | linhas incluídas ou removidas |
| Inclusão | `include "rotinas.dbas"` | com espaço de nomes separado |
| `LET` implícito | `x = 1` | `LET X=1` (obrigatório no ZX81) |

**Novidades específicas do ZX81 (além do Dignified)**
- Código de máquina embutido: bloco `asm … endasm` montado pelo **asm81** (§6.4) e colocado num `REM`; os
  rótulos do assembler ficam disponíveis no BASIC (`rand usr {rotina}`).
- Diretivas: `#machine zx81|ts1000|ts1500|lambda|zx80`, `#ram 1k|16k…`, `#autorun {rotulo}` (programa que roda
  sozinho após `LOAD`), `#fast` / `#slow`, `#name "JOGO"` (nome na fita).
- Avisos: tamanho do programa e das variáveis em relação à RAM (crítico no 1K), linha longa demais, variável de
  uma letra usada em dois papéis, `GOTO` para rótulo inexistente.
- Relatórios: mapa linha do ZX81 ↔ linha do fonte, mapa de variáveis, visualização do fluxo (como `-lbr`).
- Erros do ZX81 em execução (ex.: `2/120`) traduzidos para o arquivo e a linha do fonte, via porta de comunicação.
- Saídas: `.p`, `.b81` (listagem em texto), `.wav`; e o caminho inverso — um "dignificador" que transforma uma
  listagem clássica (`.p`) em código ZX Dignified.

Fase futura (opcional): compilação para código de máquina Z80.

**Referências (em `resource/`, só para estudo):** `zxtext2p/` (Chris Cowley, 2002, GPL, baseado no zmakebas de
Russell Marks) e `badig/` (Basic Dignified Suite, Fred Rique "farique", 2018–2022, em Python, módulos MSX e CoCo).

### 6.3 Editor de textos (clone do Norton Editor 2.0)

O editor de textos do Go81 é um **clone do Norton Editor 2.0** do MS-DOS (Symantec, 1990): mesmas teclas, mesmos
menus e mesmas caixas de diálogo, leve e rápido. Esse clone **já existe** no projeto irmão **fwMSX** (do mesmo
autor, em C/C++, licença BSD-3-Clause), em `../fwMSX/src/editor` com a especificação em
`../fwMSX/doc/editor-spec.md`. O Go81 vai **portá-lo para Go com as mesmas funções** (decisão D-19). Como o
código é do próprio autor do Go81, a licença BSD dele não impede a GPL v3 do Go81. Copia-se só o jeito do NE 2
(teclas, menus, diálogos), nunca código ou textos de ajuda da Symantec.

**O que vem do fwMSX** (já feito e testado lá):

- **Teclas do NE 2:**
  - cursor por palavra, tela e arquivo, e rolar sem mover o cursor (Alt-Z/Ctrl-Z);
  - apagar palavra ou trecho da linha (Ctrl/Alt-W, Ctrl/Alt-L, Alt-K) e recuperar o apagado (Ctrl-U);
  - inverter maiúsculas e minúsculas (Ctrl/Alt-A), tecla literal (Ctrl-P) e Insert/Over;
  - as 49 combinações de menu (Alt+letra do menu, depois a letra do item);
  - uma tecla alternativa para cada combinação que o terminal entrega mal.
- **Blocos:** marcadores, copiar, mover, apagar, indentar e desindentar, gravar o bloco em arquivo, e seleção moderna com Shift+setas.
- **Busca:** achar, trocar (interativo ou tudo), continuar para a frente e para trás, ir para a linha, parêntese casado, texto no bloco, diferença entre as duas janelas.
- **Arquivo:** novo, abrir, salvar, salvar como, juntar (Merge), CRLF/LF, fim `Ctrl-Z`, BOM, cópia de segurança (`.{` ou `~`), arquivos maiores que o buffer (Load more, Write to cursor, Top of file).
- **Duas janelas** (F2) com divisor móvel, desfazer/refazer geral e área de transferência do sistema.
- **Text-layout:** formatar parágrafo ou documento, centralizar, word wrap, auto indent, tabulações e quebras de página.
- **Impressão** em PDF ou arquivo, Information, Outline e preview de Markdown (F7).
- **Ajuda F1** paginada, os 5 esquemas de cor do NE e as caixas Display/Mouse/Editor setup.
- **Configuração gravada.** No fwMSX ela fica num arquivo próprio; no Go81 entra no `config.toml`, numa seção `[editor]`.
- **Modo somente leitura** (`view`): o que não é destrutivo funciona; o que é destrutivo aparece desabilitado e a tecla não faz nada.
- **Destaque de sintaxe por dados, completar (Alt-/) e snippets**, com números de linha.
- **Sessão e autosave:** abas sem nome (noname01…) e texto não salvo guardados ao fechar.

**O que muda para o ZX81** (no lugar do que é do MSX):

| fwMSX | Go81 |
|---|---|
| MSX-BASIC, Basic Dignified, KUN… | BASIC do ZX81 em texto (escapes do zxtext2p/EightyOne), ZX Dignified (`.dbas`, §6.2), asm81 (§6.4; a sintaxe do asMSX já tem destaque no fwMSX), scripts `.g81`, Markdown, texto |
| Build: fwasm / Basic Dignified | Build: `go81 basic make` (texto → `.p`), compilador ZX Dignified e asm81 quando existirem; erros levam à linha |
| Run on MSX (F5) | Carregar no emulador e rodar (pela sessão, como o `load`); erro do ZX81 (relatório `n/linha`) volta à linha do fonte |
| Completar com a BIOS/BDOS do MSX | Palavras-chave do ZX81, rotinas e variáveis de sistema da ROM do ZX81, símbolos do asm81 |
| Projeto `.msxproject` (SQLite) | Fica para depois (só se o Go81 ganhar projetos) |
| `man` do fwMSX na ajuda da palavra | Manual do Go81 e ajuda dos comandos do shell |
| Núcleos em C, Fortran e NASM | Tudo em Go puro (a regra das quatro linguagens é só do fwMSX) |

**Arquitetura (a mesma do fwMSX):**

- **Um modelo sem interface** (`internal/editor`) recebe teclas por nome e devolve uma **grade de células**, com o texto, a barra de menus, a barra de estado e os diálogos. Ele é testável sem terminal: sequência de teclas → texto ou grade esperada. Os testes do fwMSX (`tests/editor`, ~3.100 linhas) vêm junto, como casos de teste em Go.
- **Desenhistas finos:**
  - a TUI pinta a grade com tcell, com a cara do NE 2;
  - a janela do Fyne abre o editor numa janela própria, a partir de Ferramentas.
- **Comandos:** `go81 edit <arquivo>` / `go81 view <arquivo>`, e `edit`/`view` no shell e nos consoles.

O fwMSX passou primeiro por uma grade de caracteres na janela e depois trocou por widgets de verdade, a pedido do
usuário. No Go81 essa escolha fica em aberto: a grade no `TextGrid` do Fyne seria idêntica à TUI e mais barata;
os widgets do Fyne dariam menus e diálogos nativos.

**Fases (cada uma uma release, como no fwMSX):**

1. **ED1:** núcleo e testes portados (buffer, teclas, blocos, busca, arquivo, desfazer, somente leitura).
2. **ED2:** tela do NE na TUI (menus, diálogos, duas janelas, mouse, cores, ajuda).
3. **ED3:** janela do Fyne.
4. **ED4:** destaque de sintaxe (BASIC do ZX81 e `.g81` primeiro).
5. **ED5:** completar e snippets.
6. **ED6:** Build/Run ligados ao emulador.
7. **ED7:** Text-layout, impressão e arquivos grandes.

Estimativa: o modelo tem ~10 mil linhas de C++ e vira menos em Go, porque a biblioteca padrão faz o papel dos
núcleos em C/Fortran/NASM.

### 6.4 Montador ZX81 (asm81) e desmontador

O Go81 terá um **montador próprio, especializado no ZX81**, com a **sintaxe do asMSX** (decisão D-12):
tudo que é do MSX sai, e entra tudo que o ZX81 precisa. O asMSX é GPL v3, a mesma licença do Go81, então
podemos aproveitar seu código e seus testes. Comando previsto: `go81 asm fonte.asm`.

**Mantido do asMSX**
- Todas as instruções Z80 oficiais e não documentadas; acumulador implícito (`add b` = `add a,b`).
- Indireção com colchetes `[hl]` (para permitir parênteses nas expressões) e a diretiva `.ZILOG`, que volta
  à notação `(hl)` para código já existente.
- Rótulos globais (`nome:`) e locais (`@@laco:` ou `.laco:`, válidos até o próximo rótulo global); `$` = endereço atual.
- Constantes `.EQU`, variáveis de montagem (`X = X + 1`).
- Números decimais, hexadecimais (`$FF`, `0FFh`, `0xFF`), octais e binários; operadores no estilo C
  (`+ - * / % << >> | & ^ ~ || && == != < <= > >=`); funções `SIN COS TAN ASIN ACOS ATAN SQR SQRT EXP POW LOG LN
  ABS INT`, `PI` e ponto fixo `FIX FIXMUL FIXDIV`.
- Dados: `DB/DEFB/DT/DEFT`, `DW/DEFW`, `DS/DEFS`.
- `.ORG`, `.PHASE`/`.DEPHASE`, `.INCLUDE`, `.INCBIN "arq" [SKIP x] [SIZE y]`, `.SIZE`, `.FILENAME`, `.RANDOM(n)`.
- Montagem condicional (`IF/IFDEF/ELSE/ENDIF`), `REPT/ENDR` e macros.
- Mensagens `.PRINT`, `.PRINTDEC`, `.PRINTHEX`, `.PRINTFIX`, `.PRINTTEXT`.
- `.DEBUG "texto"` e `.BREAK [x]`, ligados ao **depurador do Go81** (no asMSX, ao BlueMSX).
- Comentários `;`, `//` e `/* */`; arquivo de símbolos `.sym`.

**Removido (específico do MSX ou do Spectrum)**
`.BIOS`/`.BIOSVARS` do MSX, `.ROM`, `.MEGAROM`, `.SUBPAGE`, `.SELECT`, `.SEARCH`, `.PAGE`, `.BASIC` (binário MSX),
`.MSXDOS`, `.CALLBIOS`, `.CALLDOS`, `.CAS`/`.WAV` no formato MSX e `.SINCLAIR` (TAP do Spectrum).

**Novo para o ZX81**
- **Símbolos predefinidos:** variáveis de sistema (`ERR_NR`, `FLAGS`, `RAMTOP`, `D_FILE`, `VARS`, `E_LINE`,
  `CDFLAG`, `FRAMES`…), pontos de entrada da ROM (os nomes do *Complete ZX81 ROM Disassembly* e os do livro de
  Toni Baker, como `KSCAN` e `FINDCHAR`), códigos de caracteres e *tokens* do BASIC. Ativados com `.ZX81`
  (ou automaticamente pela máquina escolhida).
- **Texto no conjunto de caracteres do ZX81:** uma diretiva de dados que converte `"OLA"` para os códigos do ZX81
  (com opção de inverter o último caractere, para nomes e strings terminadas).
- **Saída `.p` completa, sem esqueleto manual:** o montador gera as variáveis de sistema, a linha 0 com o código
  escondido num `REM`, a linha com `RAND USR` para iniciar, o arquivo de exibição (cheio ou "colapsado", para 1K)
  e a área de variáveis. Diretivas previstas: tipo de saída (`.P`, `.P81`, `.T81`, `.TZX`, `.WAV`, `.BIN`, `.HEX`),
  `.REMCODE` (código no REM), `.AUTORUN` (início automático via `NXTLIN`), `.DFILE FULL|COLLAPSED`, `.RAMTOP`.
- **Linhas de BASIC no fonte:** inserir linhas de BASIC tokenizadas junto com o código (usa o *loader* de BASIC
  do Go81, o mesmo do editor).
- **Máquina-alvo:** `.MACHINE ZX81|TS1000|TS1500|LAMBDA|ZX80|ACE`, escolhendo símbolos e formato de arquivo.
- **Avisos específicos do ZX81:** código executado acima de 32K (limite do M1NOT), RAM insuficiente para o
  tamanho do programa, uso de registradores reservados pela rotina de vídeo (IX em SLOW, I, R).
- **Integração com o emulador:** `go81 asm fonte.asm --run` monta, carrega e executa; símbolos e `.BREAK` vão para o
  depurador; listagem `.lst`; erros no formato `arquivo:linha:coluna`.
- **Compatibilidade com código existente:** com `.ZILOG`, aceitar o máximo da sintaxe comum do Pasmo, SjASMPlus e
  z88dk (z80asm), para montar fontes antigos sem edição.

**Editor de assembler com linhas numeradas (modo Mega Assembler)** — além de arquivos-fonte comuns, o asm81
terá um editor interativo no estilo do **Mega Assembler** do MSX (Cibertron Software, 1987), usado no shell, na
TUI e no depurador. O fonte é digitado como um programa BASIC, com linhas numeradas
(`10 INICIO: LD HL,D_FILE ; comentário`), e fica guardado no PC (não na memória do ZX81). Comandos:

| Comando | Função |
|---|---|
| `AUTO [início[,passo]]` | Numeração automática enquanto se digita |
| `LIST` / `LLIST [ini[-fim]]` | Lista o fonte na tela / na impressora (ou arquivo) |
| `DELETE ini[-fim]` | Apaga linhas |
| `RENUM [nova[,antiga[,passo]]]` | Renumera |
| `NEW`, `LOAD`, `SAVE`, `MERGE` | Fonte novo, carregar, gravar, intercalar outro fonte |
| `CHANGE 'texto1'texto2` | Troca texto em todo o fonte e lista as linhas alteradas |
| `SEARCH` / `LSEARCH` / `FIND texto` | Procura texto (em qualquer lugar / com saída na impressora / no início da linha) |
| `MAP` | Mostra o tamanho do fonte e a área de memória do código gerado |
| `A [NUPOIRSDH][/deslocamento]` | Monta, com opções: sem números de linha, sem listagem, para a impressora, gera o código, grava em fita, referência cruzada dos rótulos, rótulos em ordem alfabética, deslocamento do código gerado |

Erros de montagem podem ser mostrados também com as **letras do Mega Assembler** no início da linha
(`D` deslocamento relativo fora da faixa, `F` sintaxe, `M` rótulo repetido, `U` rótulo desconhecido, `Q` instrução
inexistente, `O` operando inválido), além da mensagem completa.

**Desmontador:** usado pelo depurador e pela CLI, com os mesmos nomes de símbolos e opção de gerar fonte no
formato do asm81 (remontável).

**Referências na pasta `resource/` (local, fora do Git):** fontes do asMSX, Pasmo (+ `pasmolib`), SjASMPlus e
z88dk; `81asm/` com um exemplo clássico de `.p` em assembler (variáveis de sistema, REM na linha 0, códigos de
caracteres e rotinas da ROM); e, em `docs/`, a desmontagem comentada da ROM, livros de BASIC e de código de
máquina do ZX81 e revistas.

### 6.5 Gerador de fitas

- Converte entre `.p`, `.81`, `.p81`, `.t81`, `.tzx` e `.wav` (taxa de amostragem configurável).
- Junta vários programas em uma fita, edita blocos e nomes (como o *TZX Manager* do EightyOne).
- *Desde 1.10.6:* comando `tape` (estado, `new`, `insert`, `play`, `stop`, `rewind`) sobre o `tape.Deck`;
  na janela, Arquivo → Abrir fita, Nova fita e Estado da fita.
- *Desde 1.10.7:* `.tzx` (TZX 1.20: bloco geral 19h, pausa, texto; os outros blocos guardados como estão,
  `Block.Raw`) e `.t81` ("EO81", nome de 32 bytes, tamanho decimal de 16, `<Silence>`), portados de
  `tzxload.cpp`/`tzxsave.cpp` (`tape.LoadTape`, `SaveTZX`, `SaveT81`, `MergeBlocks`); `load`, `tape insert`,
  `tape save` (`.tzx`, `.t81`, `.p`, `.p81`); o SAVE do ZX81 entra na fita (`Player.Record`, como
  `RecStopCheck`); Arquivo → Salvar fita; `p2txt` e o editor leem o primeiro programa de uma fita.
- *Desde 1.10.8:* fitas em som (`.wav`), portadas de `wavtape/wavload_.cpp` e `WavFile.cpp`: a leitura roda o
  `GetByte` do EightyOne (a rotina de LOAD da ROM com as mesmas contagens de T-states, limiar 50) bloco a
  bloco (`ConvertNextBlock`, fim pelo E_LINE das variáveis de sistema) → `tape.LoadWAV`; a gravação é o
  `EncodeBlock` (4 ou 8 pulsos por bit, 44.100 Hz, 8 bits) → `tape.SaveWAV`. `tape.WavPlayer` põe o som na
  entrada do `Deck` (TZXin desligado), como o Wave Loader: a ROM lê em tempo real (teste
  `TestROMLoadsWAV`). Comandos `wav2p` e `go81 wav2p`, `load`/`tape insert` de `.wav`, `tape insert x.wav som`,
  `tape save x.wav`; Ferramentas → Carregador WAV e Converter fita WAV em .p.
- Gera WAV para carregar em ZX81 real.

### 6.6 Depurador e monitor (inspirado no Super-X)

O depurador do Go81 segue o **Super-X**, o monitor/depurador de MSX (Romi, 1994; versão 1.22 de NYYRIKKI, 2011;
manual em inglês de JP Grobler): **linha de comando com comandos de duas letras**, editores de tela cheia (dump,
ASCII, desmontagem, montagem, caracteres) entre os quais se alterna rapidamente, e um conjunto grande de
ferramentas de memória (decisão D-17). Tudo o que é do MSX sai; entra o que o ZX81 precisa. Por baixo, o motor
segue o `DebugWin` do EightyOne (breakpoints, histórico, profiler).

O mesmo depurador aparece na **GUI** (janela), na **TUI** (painéis), no **shell** (comandos) e no **Link**
(remoto, inclusive pelo ZRCP).

**Convenções herdadas do Super-X**
- Parâmetros separados por espaço ou vírgula (`D 4000 4100` = `D4000,4100`); `<obrigatório>`, `[opcional]`.
- Números em hexadecimal por padrão; também decimal, binário, negativos, caracteres entre aspas e expressões.
- `?` antes de um comando manda a saída também para arquivo ou para a ZX Printer emulada.
- ENTER numa linha vazia repete o último comando; ESPAÇO pausa uma listagem; ESC interrompe.
- Variáveis de endereço `@0`–`@3` e as automáticas `@B` (início), `@E` (fim), `@S` (início de execução/tamanho),
  preenchidas pelos comandos de carga.
- **Pilha de navegação:** na desmontagem, seta para a direita segue um `CALL`/`JP`, seta para a esquerda volta
  (até 16 níveis).
- Barra de informação: comando atual, endereço do cursor, nível da pilha de navegação, deslocamento global.

**Comandos mantidos (adaptados ao ZX81)**

| Comando | Função |
|---|---|
| `D` | Dump hexadecimal, com edição |
| `A` | Memória como texto — no **conjunto de caracteres do ZX81** (inclusive inverso e gráficos) |
| `I` | Desmontagem com edição; segue saltos; mostra símbolos e notas |
| `M` | Entrada de assembler (via **asm81**) e edição direta de memória; `I n` lista instruções |
| `H` | Editor de caracteres/bitmaps **8×8** (fontes CHR$128, Quicksilva, UDG, telas WRX) |
| `BL` / `BF` | Lista o **BASIC do ZX81** da memória / procura texto no BASIC (com curinga `?`) |
| `BT` | Copia um bloco de memória |
| `RT` | Reloca código de máquina corrigindo os ponteiros internos, listando as mudanças |
| `FL` | Preenche memória |
| `CM` | Compara memória (diferenças ou semelhanças) |
| `FD` | Procura bytes, texto ou **instruções assembler** na memória |
| `CK` / `TS` | Tipo de checksum / checksum de uma área |
| `CL` | Calculadora (hex/dec/bin) |
| `RG` | Mostra e altera registradores (inclusive `IXH`, `IYL`, pares alternativos) |
| `GO` | Executa a partir de um endereço, com ponto de parada |
| `TR` | Executa uma instrução e mostra os registradores (trace) |
| `SD` | "Super desmontador": exporta fonte **asm81** remontável, `DEFB`, linhas `DATA` de BASIC ou **linha REM** do ZX81 |
| `LD` / `SV` | Carrega/salva `.p` e binários |
| `L#` / `S#` | Carga/gravação "crua" de memória em arquivo |
| `iM` `iC` `iS` `iL` | **Notas** por endereço (criar, consultar, salvar, carregar) — já vêm com as variáveis de sistema, as rotinas da ROM e as portas do ZX81 |
| `OF` | Deslocamento global (depurar código fora do endereço real) |
| `PI` / `PO` | Lê/escreve porta de E/S |
| `SF` | Teclas de função programáveis |
| `@` | Mostra/define as variáveis de endereço |
| `CI` | Informações da máquina (modelo, ROM, RAM, periféricos) |
| `TP` | Mostra um arquivo de texto |
| `CS` | Limpa a tela |
| `QT` | Sai do depurador e volta a emular |

**Removido (específico do MSX):** slots e subslots (`#slot`), VRAM, *mapper* de memória (`PP`), troca de CPU
Z80/R800 (`CU`), fontes e textos em japonês (`KL`, `KT`, `KR`), setores de disco (`L%`, `S%`), comandos de
MSX-DOS, exportação X-BASIC e cores do MSX (`CO` vira tema de cores).

**Ferramentas do Mega Assembler incorporadas ao monitor** (adaptadas ao ZX81)

| Comando | Função |
|---|---|
| `C <modo>` | Formato do dump: hexa + texto em 4 ou 16 bytes por linha, ou **hexa com soma de verificação por linha** (com ou sem o endereço na soma) — o formato das listagens de código de máquina publicadas em revistas, para digitar e conferir |
| `DM <end>[,desloc]` | Dump/edição em hexa e texto com **deslocamento de byte** (ver e editar dados "cifrados" somando um valor fixo) |
| `SH [end],b1,,b3…` / `SH [end],'texto` | Busca com **coringa** (`,,` = qualquer byte) e busca de texto que também acha o texto **deslocado** |
| `MS <end>[,desloc],'texto` | Grava um texto na memória, opcionalmente deslocado |
| `SCR <end>,dx,dy[,modo]` | Visualizador gráfico da memória em blocos de 8 bytes (dx × dy), com edição de pontos — para UDG, fontes CHR$128 e gráficos de alta resolução |
| `G <end>[,bp1[,bp2]]` | Executa com até dois pontos de parada rápidos |
| Editor de caracteres | Edita a tabela de caracteres (fontes para CHR$128/Quicksilva/UDG) |
| Cabeçalho de fita | Mostra nome, tamanho e tipo dos blocos da fita (como `CALL HEADER`) |
| Verificar fita | Compara o que está na fita (ou num `.p`/`.wav`) com a memória (como `CALL BVERIFY`) |
| Cópia de tela | Copia a tela para a ZX Printer emulada ou para imagem, com uma tecla de atalho (como `CALL DUMP`/`SETKEY`) |

Sem equivalente no ZX81: `PAGE` (slots), `ZAP` (setores de disco — talvez no futuro para imagens do ZXpand/IDE),
`COPYRV`/`COPYVR` (VRAM), `SETGREY`.

**Novo para o ZX81**
- Breakpoints de execução, leitura, escrita, porta de E/S e **condição** (expressões com registradores e memória),
  com contador de passagens.
- Passo a passo que **pula o tratamento de vídeo** (NMI/INT da geração de tela), como o "Skip NMI/INT" do
  EightyOne; avanço por quadro e por linha de TV.
- **Breakpoint em linha de BASIC** (para quando a variável de sistema `PPC` chega a uma linha).
- Janela das **variáveis de sistema** e das variáveis do BASIC; visualização do arquivo de exibição (D_FILE).
- Histórico de execução, profiler e cobertura de código (EightyOne e ZEsarUX).
- Símbolos do asm81 e do ZX Dignified (rótulos, linhas do fonte) aparecendo na desmontagem.

## 7. Porta de comunicação (Link)

*Estado em 1.6.3:* servidor TCP em `127.0.0.1` com o protocolo de linha (porta 8181: saída + `ok`/`err`,
*dot-stuffing*, pedidos JSON de uma linha `{"cmd":"..."}`), modo ZRCP (porta 10000: `about`, `get-version`,
`get-current-machine`, `get-registers`, `set-register`, `read-memory`, `write-memory`, `send-keys-string`,
`get-ocr`, `view-basic`, `smartload`, `hard-reset-cpu`, `help`, `quit`) e o cliente `pkg/go81link` (MANUAL §10). O protocolo
Go81 também roda pela entrada/saída padrão (`-stdio`), num socket Unix (`-socket`) e num pipe nomeado do
Windows (`-pipe`, restrito ao usuário atual, via `go-winio`). *Desde 1.7.4:* token opcional (`auth <token>`,
`-token`, `[link] token`), eventos assíncronos (`events on`: `! report 0/10`, `! save NOME`, `! load`, também em
JSON; o `report` vem da rotina REPORT da ROM, 06D1H, vigiada só enquanto há ouvintes) e mais comandos ZRCP
(`evaluate`, `save-screen`, `get-os`, `send-keys-ascii`, `reset-cpu`, `enter-cpu-step`/`exit-cpu-step`). Faltam
breakpoints/passo a passo no ZRCP (depurador) e o GoLink do lado Z80.

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
- **Modo compatível com o ZRCP** (ZEsarUX, §14.3, D-15): mesmos nomes e respostas dos comandos mais usados
  (`get-registers`, `read-memory`, `write-memory`, `set-breakpoint`, `cpu-step`, `send-keys-string`, `get-ocr`,
  `view-basic`, `smartload`, `hard-reset-cpu`…), para funcionar com clientes existentes como o DeZog (VS Code).

### 7.2 Programa ZX81 ↔ host (troca de dados)

- Um dispositivo virtual de E/S ("GoLink") que o programa Z80 acessa por `IN`/`OUT` para enviar e receber bytes do host.
- Porta de E/S configurável; o endereço padrão será escolhido no estudo para não colidir com ULA, ZX Printer, ZXpand e demais periféricos (decisão D-04).
- Também disponível via área de memória "caixa de correio" para programas BASIC (`PEEK`/`POKE`).

## 8. Interfaces

### 8.1 GUI (Fyne)

*Estado em 1.6.3:* janela com a tela do ZX81 a 50/60 Hz, som, teclado do PC mapeado por posição, menus
Arquivo (abrir programa, salvar tela em PNG/BMP/PBM/TXT, sair), Editar (colar texto como teclas, copiar tela como
texto), Máquina (configuração do hardware, reiniciar, pausar, turbo, som), Exibir (zoom 1–4×, borda),
Ferramentas (console de comandos F2, ligar/desligar o Link) e Ajuda. SAVE feito pelo programa abre a janela de
gravação do `.p`. Roda sobre a mesma sessão do shell; `--link` deixa a janela ser controlada de fora. Opções de
automação `--exit-after` e `--capture`.
*Desde 1.7.5:* arrastar e soltar programas na janela; Arquivo → Abrir recente (8 últimos, guardados nas
preferências do Fyne); tela cheia (F11); zoom e borda do menu gravados em `[janela]` do `config.toml`; teclado do
ZX81 na tela (F3: 40 teclas com símbolo do SHIFT e palavra-chave, tecla presa enquanto o mouse está apertado,
mínimo de 0,1 s, SHIFT com trava até a próxima tecla), somado ao teclado do PC.
*Desde 1.7.6:* menus no formato do EightyOne (Arquivo, Exibir, Máquina, Opções, Ferramentas, Ajuda; mais
Editar), com todas as opções do EightyOne para o ZX81 e clones; as que ainda não existem abrem o aviso "A
implementar". Novas: salvar programa, carregar/salvar bloco de memória, vídeo inverso, barra de estado
opcional, reset pela linha RESET (só o Z80), NMI, máquina de fábrica, perfis no menu, listagem BASIC numa
janela, Ajuda → Teclado (F1) com a imagem do teclado do EightyOne (ZX81 ou TS1000; clique duplo dobra o
tamanho). Atalhos do Go81 mantidos. Ficaram de fora as opções só do ZX Spectrum (teclado numérico do 128,
Interface 1, exemplos do Spectrum) e as ocultas no EightyOne (RZX, Memotext, tópicos de ajuda).
*Desde 1.7.8:* zoom definido pelo usuário (Exibir → Zoom): qualquer valor de 1 a 8, com fração; o mesmo em
`--zoom` e na chave `zoom` de `[janela]`.
*Desde 1.8.0:* Opções → Hardware no formato do EightyOne (modelos, ROM com "…", RAM pack e RAM total, 60 Hz,
RAM em 8K–16K, M1NOT, correção do ponto flutuante, alta resolução; as interfaces que faltam dizem "A
implementar"); Ajuda → Teclado mostra o teclado do modelo (TS1500, TK85, R-470).
*Desde 1.10.0:* F7 troca a tela pelo editor de BASIC (§6.1.1) na mesma janela; F8 escolhe o modo do editor
(1.10.1). *Desde 1.10.3:* F5 (pausa) e F6 (turbo) funcionam pelo teclado: o Fyne só dispara atalhos de menu
com Ctrl/Alt, então as teclas F passam pelo tratamento de teclas da janela.
*Desde 1.9.1:* mapa do teclado sobre a tela (F4, Exibir → Mapa do teclado): a imagem do teclado do modelo numa
camada por cima da tela, esmaecida (transparência 25/50/75%), opaca ou só enquanto SHIFT estiver apertado;
a escolha fica nas preferências do Fyne. O editor de BASIC no estilo ZX81 (§6.1.1) vai usar o mesmo mapa.
*Desde 1.9.2:* CTRL apertado, com o mapa ligado, mostra as teclas do cursor atual (K, L, F ou G): o cursor é
lido de MODE (4006h) e FLAGS (4001h) como em OUT-CURS (079Dh); o que cada tecla dá sai das tabelas K-UNSHIFT,
K-SHIFT, K-FUNCT e K-GRAPH da ROM, como em MODE-SORT (04DFh); o desenho usa os caracteres da ROM (1E00h).
*Desde 1.7.9:* Exibir → Reunir janelas (as janelas secundárias vão para o centro da tela, à frente; o Fyne
não deixa escolher outra posição) e Trazer janelas para a frente.
*Desde 1.7.11:* Ferramentas → Gravar o som em WAV (comando `record`): o som de cada quadro emulado vai para
um `.wav` de 44,1 kHz, 16 bits, estéreo (`internal/media/wav`); o cabeçalho é atualizado a cada segundo.

*Planejado:*

Menu Depurar; janelas para editor BASIC, editor de textos, montador, gerenciador de fita e depurador. Filtros de imagem.

#### 8.1.1 Temas que lembram os micros (intenção registrada em 2026-11-02)

A janela (e depois os editores, o depurador e a TUI) poderá vestir um tema inspirado no micro emulado (D-23):

| Tema | Inspiração |
|---|---|
| ZX81 | Preto fosco com as letras das teclas em branco e os símbolos do SHIFT em vermelho; a faixa do logotipo |
| ZX80 | Branco/creme com o teclado azul e preto do ZX80 |
| TS1000 / TS1500 | O prateado e preto da Timex, com o vermelho dos símbolos |
| TK85 / Ringo R-470 | As cores das caixas e teclados da Microdigital e da Ringo |
| Lambda 8300, ZX97, Jupiter Ace | Quando essas máquinas forem emuladas |
| Neutro | O tema claro/escuro do Fyne, como hoje (padrão) |

O que muda com o tema:
- cores e fonte da interface: a fonte do próprio micro (lida da ROM) nos títulos, nas barras de estado e no mapa do teclado;
- a moldura da tela;
- o mapa do teclado (§8.1, F4/CTRL) nas cores do teclado do modelo.

Regras:
- o tema pode acompanhar o modelo escolhido (opção "automático") ou ficar fixo;
- tudo isso só mexe na aparência, sem efeito na emulação;
- as imagens dos teclados continuam as do EightyOne;
- marcas e logotipos não são reproduzidos, só as cores e o estilo.

Quando forem estudadas as diferenças de teclado entre as ROMs (TS1500, TK85, R-470), o mapa do cursor (CTRL)
passará a ler as tabelas da ROM do modelo em vez das do ZX81.

### 8.2 TUI (tview/tcell)

*Estado em 1.4.3:* `go81 tui`. Modo texto: decodifica o arquivo de exibição (D_FILE) em 32×24 caracteres,
com os gráficos de bloco do ZX81 em caracteres de quadrante Unicode e o vídeo inverso como cores invertidas.
Modo pixels (F9): a imagem real da TV em meia resolução (64×48 células de quadrante), para alta resolução.
Teclado por caractere (o terminal só informa a tecla pressionada): cada tecla vira um toque curto na matriz
do ZX81, e os símbolos viram a combinação SHIFT correspondente. Funciona via SSH e sem CGO.
*Desde 1.6.2:* console de comandos (F2) com os comandos do shell, histórico e Tab; `--link`.
*Desde 1.7.2:* três modos de tela (F9): texto 4:3 (letras Unicode de largura dupla e gráficos em meios-blocos,
duas células por caractere), pixels 4:3 escalados ao espaço (resolução inteira a partir de 132×98 células) e
texto compacto (terminais estreitos).

*Planejado:*

Mesma organização da GUI em painéis de texto. A tela do ZX81 é desenhada com caracteres (blocos Unicode para os gráficos 2×2 do ZX81; modo de meio-bloco para hi-res aproximado). Funciona via SSH.

### 8.4 Manual embutido (`man`) — registrado em 2026-11-01, feito na 1.9.0

*Desde 1.10.3:* páginas `ref editor` (comandos e teclas do editor de BASIC) e `ref atalhos` (teclas da janela).
*Desde 1.9.0:* `internal/man` com as seções basic, shell, cli, basic80, sysvar, rom, erro e ref (~900
páginas, quase todas rótulos da ROM). As teclas das palavras-chave saem das tabelas K-SHIFT e K-FUNCT da
própria ROM; os rótulos vêm do `zx81.edition3.sym` do EightyOne. Testes garantem que todo comando do shell e
da CLI e toda palavra-chave tem página. Comando `man` no shell e nos consoles, `go81 man`, e Ajuda → Manual
embutido na janela.

O Go81 terá um `man` como o do **fwMSX** (`../fwMSX/src/man`, do mesmo autor; especificação em
`../fwMSX/doc/man-spec.md`), portado para Go (decisão D-21): tópicos com nome, apelidos, categoria, linha de uso,
resumo e página completa; `man <tópico>`, `man -k <palavra>` (procura), `man -l [categoria]` (lista);
`help <comando>` continua curto e `help` sozinho mostra os comandos agrupados. As páginas são dados (texto em
Go), sem depender de arquivos externos. No terminal, a página passa por um paginador; nos consoles da janela e
do terminal, aparece numa janela de texto.

Conteúdo, nesta ordem:

1. **Todos os comandos** do Go81: os subcomandos da CLI (`go81 run`, `gui`, `basic list`…) e os comandos do
   shell, dos consoles e do Link (`load`, `find`, `record`…), com exemplos.
2. **Palavras-chave do BASIC** do ZX81 e do ZX80: sintaxe, o que fazem, a tecla e o modo do cursor (K, L, F, G)
   onde estão, os erros que causam, exemplos.
3. **A ROM do ZX81 original** (edição 3): rótulos e pontos de entrada interessantes (rotinas de
   impressão, teclado, fita, calculadora, `RST`), as variáveis de sistema e os códigos de erro — a base é a
   desmontagem comentada da ROM e os símbolos do EightyOne (`zx81.edition3.sym`). As particularidades das ROMs
   dos outros modelos (TS1500, TK85, R-470, Lambda, ZX80) ficam para depois.

O depurador e o asm81 vão usar os mesmos dados (nomes de rotinas e variáveis de sistema).

### 8.3 CLI

*Estado em 1.6.3:* já existem `go81 run`, `gui`, `tui`, `shell`, `link`, `basic list`, `roms`, `version` e `help` (ver MANUAL §2–4,
§10, §11). O shell tem o registro de comandos único (`internal/session`), scripts `.g81`, `-c`, histórico
(arquivo em `os.UserConfigDir()/go81/shell_history`) e Tab que completa comandos, arquivos e opções (biblioteca
`liner`). A GUI e a TUI rodam sobre a mesma sessão e têm um console (F2) com esses comandos; `--link` as
deixa ser controladas de fora. *Desde 1.7.4:* variáveis (`set`, `${nome}`, `calc`, `capture`), expressões com
funções do ZX81 (`peek`, `reg`, `screen`…), blocos `if/else/while/repeat/end/break` em scripts, `assert`, `fail`
e `waitfor`. Os comandos `asm`, `tape` e o resto de `basic` são planejados.

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

*Estado em 1.7.3:* arquivo TOML em `%APPDATA%\go81\config.toml` (Windows) ou `~/.config/go81/config.toml`
(Linux), ou `$GO81_CONFIG`; seções `[maquina]`, `[janela]`, `[link]` e `[perfis.<nome>]`; `-config` e `-profile` em
`gui`, `tui`, `shell` e `link` (opções digitadas valem por cima); `go81 config path|init|show|edit`;
`machine perfil=` no shell; "Salvar esta máquina como padrão" na janela. O `go81 run` não lê o arquivo
(automação reprodutível).
*Desde 1.9.4:* configuração por jogo (`internal/session/gamecfg.go`): `<programa>.toml` ao lado do programa
ou `[jogos."nome.p"]`/`[jogos.nome]` no config.toml, com as chaves de `[maquina]` e `perfil`; o `load` e a
abertura de arquivos na janela trocam a máquina antes de carregar e, num programa sem configuração, voltam à
máquina de antes do primeiro jogo configurado. `jogo [salvar|apagar]` no shell; Arquivo → Gravar a máquina
para este jogo.

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
- Testes Z80 (zexdoc/zexall) em `internal/emu/z80` (longos: `GO81_LONG=1`).
- Testes de compatibilidade com o EightyOne em `test/compat` (`z80ref` = núcleo original em C, `z80diff` = comparador),
  com resultados de referência em `testdata/` de cada pacote.
- Testes "dourados" (golden files) para compilador, montador, lister e conversões de fita.
- CI no GitHub Actions para Windows e Linux.

## 12. Decisões

| ID | Decisão | Estado |
|---|---|---|
| D-01 | Um binário `go81` com subcomandos; GUI/TUI/CLI no mesmo executável | Aceita |
| D-02 | Incluir o ZX80 | Aceita (1.1.2) |
| D-03 | Família Spectrum (incluindo TS2068/TC2048/TC2068) fora do escopo | Aceita |
| D-04 | Endereço de porta padrão do GoLink (lado Z80) | Em estudo |
| D-05 | Licença do projeto: GPL v3 ou posterior | Aceita (1.1.2) |
| D-06 | Configuração em TOML | Proposta |
| D-07 | Extensão `.dbas` para fontes Dignified | Proposta |
| D-08 | Incluir o Jupiter Ace | Aceita (1.1.2) |
| D-10 | ROMs da Sinclair embutidas no executável (`assets/roms`, `go:embed`); a Amstrad permite distribuí-las para uso em emuladores | Aceita (1.3.1) |
| D-11 | A GUI (Fyne) exige CGO e um compilador C; `-tags nogui` gera um executável só com CLI, sem gcc | Aceita (1.4.1) |
| D-12 | Montador próprio (asm81) com a sintaxe do asMSX, sem a parte do MSX e com tudo do ZX81 (§6.4) | Aceita (registrada em 1.4.3; implementação futura) |
| D-13 | VB81 XuR (código VB6, GPL v2+) como fonte de ideias de **ferramentas** (§14.2); emulação segue o EightyOne; o executável não deve ser usado (relatos de vírus) | Aceita |
| D-14 | `resource/` (material de referência de terceiros, ~500 MB) fica fora do Git, como `81emul/` | Aceita |
| D-16 | Compilador ZX Dignified escrito do zero em Go, baseado no zxtext2p e próximo do Basic Dignified Suite, sem copiar código nem chamar programas externos (§6.2) | Aceita |
| D-17 | Depurador/monitor no estilo do Super-X (comandos de duas letras, editores de tela cheia), sem o que é do MSX e com o que o ZX81 precisa (§6.6) | Aceita |
| D-18 | Incorporar do Mega Assembler (MSX) o editor de assembler com linhas numeradas e as ferramentas de monitor que faltam no Super-X (§6.4, §6.6) | Aceita |
| D-19 | O editor de textos é o clone do Norton Editor 2.0 já feito no fwMSX (C/C++, do mesmo autor), portado para Go com as mesmas funções; o que é do MSX vira ZX81 (§6.3) | Aceita (registrada em 2026-11-01) |
| D-21 | Manual embutido (`man`) portado do fwMSX: todos os comandos, as palavras-chave do BASIC do ZX81 e do ZX80 e os rótulos/pontos de entrada da ROM do ZX81 original (§8.4) | Aceita (registrada em 2026-11-01) |
| D-22 | Todas as ROMs do EightyOne para a família ZX81 vão embutidas no executável (Timex, Microdigital, Ringo, Lambda, Jupiter Cantab e as alternativas), como o EightyOne as distribui | Aceita (1.8.0) |
| D-20 | Editor de BASIC no estilo do ZX81 (listagem em cima, entrada embaixo), em dois modos (teclas com mnemônicos e digitação direta), com utilitários RENUM, AUTO, DELETE, FIND e criador de REM (§6.1.1) | Aceita; feita na 1.10.0–1.10.2 (na mesma janela do emulador, F7) |
| D-23 | Temas visuais que lembram cada micro (ZX81, ZX80, Timex, Microdigital, Ringo…), automáticos pelo modelo ou fixos; só aparência (§8.1.1) | Aceita (registrada em 2026-11-02) |
| D-15 | O Link tem um modo compatível com o ZRCP do ZEsarUX (telnet, porta 10000), além do protocolo próprio, para aproveitar ferramentas existentes como o DeZog | Aceita (subconjunto desde 1.6.1) |
| D-09 | O núcleo Z80 reproduz também os desvios do EightyOne em relação ao Z80 real (flags do DAA, falta de MEMPTR nas flags de `BIT`). Um modo "Z80 exato" pode ser opção futura | Aceita (1.2.1) |

## 13. Licenciamento

O Go81 é distribuído sob a **GNU General Public License v3 ou posterior** (arquivo [LICENSE](LICENSE)).

- O EightyOne é **GPL v2 ou posterior** (cabeçalho de `zx81/zx81.cpp`; núcleo Z80 do FUSE também GPL v2+).
  A cláusula "ou posterior" permite portar esse código para o Go81 sob GPL v3.
- Arquivos portados mantêm os avisos de copyright originais (Michael D Wynne, Philip Kendall etc.) e recebem uma linha
  indicando a porta para Go.
- Arquivos novos levam o cabeçalho:
  ```go
  // Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
  // SPDX-License-Identifier: GPL-3.0-or-later
  ```
- Bibliotecas usadas (Fyne BSD-3, tview MIT, tcell Apache-2.0) são compatíveis com a GPL v3.
- As ROMs têm direitos próprios (a Amstrad permite a distribuição das ROMs Sinclair para uso em emuladores; ROMs de terceiros variam) e ficam separadas do código.

## 14. Referências e inspirações

### 14.1 Material local (`resource/`, fora do Git)

| Pasta | Conteúdo | Uso no Go81 |
|---|---|---|
| `asMSX/` | Montador asMSX (GPL v3), manual em `doc/asmsx.md` | Base da sintaxe do asm81 (§6.4) |
| `pasmo/`, `pasmolib/` | Montador Pasmo (Julián Albo, GPL v3) | Compatibilidade de sintaxe, testes |
| `sjasmplus/` | Montador SjASMPlus (aprisobal e colaboradores, BSD) | Compatibilidade de sintaxe, testes |
| `z88dk/` | Kit de desenvolvimento C/assembler z88dk, com suporte ao ZX81 | Formatos, bibliotecas, compatibilidade com z80asm |
| `81asm/` | Exemplo de programa ZX81 em assembler (variáveis de sistema, REM, caracteres, ROM) | Símbolos predefinidos e geração de `.p` |
| `VB81_RES/`, `Vb81_XuR/` | Código-fonte VB6 (GPL v2+) e executável do VB81 XuR (não executar) | Ideias de ferramentas (§14.2) |
| `zesarux/` | ZEsarUX (César Hernández Bañó, GPL v3): código C, manual e lista de recursos | ZRCP, ferramentas de fita, depurador (§14.3) |
| `zxtext2p/` | ZXText2P (Chris Cowley, GPL): texto → `.p`, com rótulos e escapes de gráficos | Base do formato de texto do ZX Dignified (§6.2) |
| `badig/` | Basic Dignified Suite (Fred Rique), Python, módulos MSX e CoCo | Modelo de recursos do ZX Dignified (§6.2) |
| `superx/` | Super-X 1.22 (monitor/depurador de MSX): manual `.DOC` em inglês e fonte em assembly | Modelo do depurador (§6.6) |
| `docs/Manual do Mega Assembler.pdf` | Mega Assembler (MSX, Cibertron Software, 1987), manual em português | Editor de assembler numerado e ferramentas de monitor (§6.4, §6.6) |
| `docs/` | Desmontagens da ROM, livros de BASIC e de código de máquina, manuais (Memotech, MCoder), revistas ZXzine, foto do teclado | Símbolos da ROM, BASIC, documentação de periféricos |

### 14.2 VB81 XuR

**O que é:** emulador e ambiente de ZX81 para Windows, escrito em Visual Basic 6. O **VB81** original é de
**Chris Cowley** (Grok Developments, 1999–2002); o **XuR** é a continuação de **XavSnap (Xavier Martin)**,
2005–2025, com versão final 1.4 de 28/08/2025. Site: `http://zx81.vb81.free.fr` (não abriu na consulta de
2026-10-31); fórum: Sinclair ZX World, tópico 3295.

**Material local:** `resource/VB81_RES/` traz o **código-fonte VB6** (projeto `VB81.VBP`: 60 formulários,
45 módulos, classes, imagens dos caracteres do ZX81) e `resource/Vb81_XuR/` o executável e o histórico
`VB81_Changes.TXT`. O código é **GPL v2 ou posterior** (cabeçalhos de `modZX81.bas`, `modZ80.bas` etc.), então
pode ser estudado e portado para o Go81 com os créditos mantidos, como o EightyOne.

**Cuidados:** há relatos de vírus no executável. Ele **não deve ser executado** (nem os `.exe`/`.bat` que vêm
junto); só os textos e fontes são lidos. Recomenda-se passar um antivírus na pasta.

**Papel no Go81:** fonte de **ideias para ferramentas**, não de emulação. O próprio autor explica que o vídeo do
VB81 é uma "pseudo-emulação" que lê o D_FILE em vez de emular a ULA, e que programas com rotinas de vídeo
avançadas não rodam nele. A emulação continua seguindo o EightyOne.

**Recursos do VB81 XuR que interessam ao Go81** (extraídos do histórico e da lista de formulários):

| Área | Recursos | Onde entra no Go81 |
|---|---|---|
| Ferramentas de BASIC | Renumerar linhas; inserir REM grande para código de máquina (`[SPH:]`, `[SPD:]`); converter números em strings para economizar memória; texto em REM com 6 bits por caractere (`[TX6:]`); carregar/salvar BASIC como texto (ZxToken, `.TXT`/`.BAS`, com códigos CBI); `SAVE "X.TXT"` direto do BASIC | Bloco BASIC (lister/loader, ferramentas) |
| Assembler | Montador compatível com TASM; o montador ASM2 (Artic) com exportação para a linha REM; editor de REM | asm81 (modo de importação TASM) |
| Desmontador e depurador | Desmontador com saída `.TXT`, `.ASM` (ASM2) e `.TAS` (TASM, com rótulos de ROM/RAM, marcação de dados, texto ZX81 e ASCII); breakpoints; "Load & Break"/"ASM Break" | Depurador e desmontador |
| Memória | Editor binário (digitação em caracteres ZX81); comparar memória; transferir/carregar blocos; janela de variáveis do BASIC; visualizador de memória em texto | Depurador / ferramentas de memória |
| Tela (D_FILE) | EasyDraw (editar a tela com o mouse); importar BMP e "digitalizar" imagens para a tela; salvar a tela como `.SCR`, `.TXT` (com PRINTs) ou `.BAS`; captura ASCII; "ASM screen shot" comprimido; fábrica de UDG | Ferramentas de tela |
| Fitas e arquivos | `.P`, `.81`, `.T81` (fitas completas), `.TZX`, `.Z81`, `.SBB`; WAV de carga lenta e **WAV de carga rápida (OTLA, de Francisco Villa)**; `SAVE "X.WAV"`/`"X.FL"`; conversor e decodificador de WAV; fita virtual; explorador de arquivos com prévia do D_FILE; exportação Motorola S-record | Gerador e gerenciador de fitas |
| Impressora | ZX Printer com fundo de papel e listagem em várias colunas; exportação PDF, GIF, JPEG, PNG; Centronics (Memotech, AGB, 8255) com buffer bruto | Periféricos (impressoras) |
| Placas gráficas | Memotech HRG, G007, Quicksilva (HRG e UDG), WRX16, AGB UDG, Mageco 16 cores, pseudo-hi-res | Periféricos (hi-res) |
| Som e fala | AY-3-8912 (ZON X/VTR, Quicksilva, AGB, placas do usuário), Octavia, Mageco SP0256 | Periféricos (som/fala) |
| Discos e E/S | Larken (com gerenciador de discos), Mageco, ZXpand, conversor CopyQM; placas 8255, 8×8 E/S, 16/16 E/S de Wilf Rigter, Mageco 8 E/S, RomServer (porta $1F), caneta óptica Mageco | Periféricos |
| Forth | Editor e compilador Forth para ZX81 (de Dominique Contant), arquivos `.4TT`/`.4TB` | Futuro (útil também para o Jupiter Ace) |
| Interface | Janela para iniciantes que clica em comandos/caracteres do ZX81; teclas configuráveis e "digitação automática"; gerenciador de ROMs extras por segmento de memória; ROM e teclado franceses; área de transferência (copiar/colar texto convertendo caracteres); tela cheia; captura com F11; abrir `.p` pela linha de comando | GUI / shell |

Fontes adicionais: [Emulation General Wiki — ZX81](https://emulation.gametechwiki.com/index.php/ZX81),
[t2e.pl — notas de versão](https://www.t2e.pl/News/wiadomosci-oficjalne-releasy.322/zx81-vb81-xur---zx81-emulator-1.4-030819.20156),
[Sinclair ZX World](https://sinclairzxworld.com/viewtopic.php?p=35097).

### 14.3 ZEsarUX

**O que é:** emulador de muitas máquinas Sinclair e compatíveis, de **César Hernández Bañó**, em C, **GPL v3**
(mesma licença do Go81). Código em `resource/zesarux/` (local); manual em `src/zesarux.pdf`, lista de recursos em
`src/FEATURES`. Interessa só a parte do ZX80/ZX81 (`src/machines/zx8081.c`) e as ferramentas em volta.

**Recursos para o ZX81 que interessam ao Go81:**

| Área | Recurso no ZEsarUX | Onde entra no Go81 |
|---|---|---|
| Automação | **ZRCP** — protocolo de comando remoto por telnet (porta 10000), ~175 comandos: registradores, ler/escrever memória e portas, `disassemble`/`assemble`, breakpoints com condições, passo a passo, `send-keys-string`, `get-ocr` (texto da tela), `view-basic`, `save-screen`, `smartload`, snapshots, `hard-reset-cpu`, `evaluate`, histórico da CPU | **Link** (§7): oferecer um modo compatível com o ZRCP, para que ferramentas existentes (ex.: depurador **DeZog** do VS Code) funcionem com o Go81 (D-15) |
| Vídeo | Detecção automática de WRX e de Chroma81; "real video" também nos drivers de texto | Emulador / TUI |
| Texto | Captura do texto impresso pela ROM (*print char traps*, RST 10H) e OCR da tela; drivers `stdout`, `curses`, `simpletext` | Link (`get-ocr`), TUI, acessibilidade |
| Teclado | Enviar texto como teclas a partir de arquivo ou da área de transferência; teclado na tela (F8) | Link, shell, GUI |
| Fitas | Carga real de WAV/RWA/SMP; **conversor de áudio para `.p`/`.p81`** com visualização dos pulsos, bits e bytes e opção de inverter o sinal; "Visual Real Tape" (forma de onda, avançar/voltar); *traps* de fita ligáveis | Gerador e gerenciador de fitas |
| Memória | RAM de 8K em 2000H, 16K em 8000H/C000H, até 56 KB; permitir escrita na ROM | Emulador / configuração |
| Depurador | Depuração reversa, histórico da CPU, breakpoints com expressões, *watches*, cobertura de código, log de transações, visualização de memória lida/escrita | Depurador |
| BASIC | Listagem do BASIC do ZX81 na memória (`view-basic`) | Bloco BASIC / Link |
| Snapshots | `.Z81`, `.ZSF`; snapshots automáticos em RAM e **rewind** | Emulador |
| Periféricos | ZXpand com pasta do PC como raiz do SD; Quicksilva e ZON-X81; joysticks (incl. ZXpand) com botões ligados a ações | Periféricos |
| Arquivos | Tela salva como TXT, PBM, BMP, SCR; gravação de áudio e vídeo em arquivo; navegador online de jogos de ZX81 (`www.zx81.nl`) | GUI / CLI |
| Configuração | Opções por linha de comando, arquivo de configuração e **configuração por jogo** | Configuração (§9) |
| Outros | Localizador de *cheats* na memória (contadores de vidas, energia) | Depurador |
