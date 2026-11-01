# Go81 — Releases

Formato de versão: **X.Y.Z** — veja [SPEC.md §10](SPEC.md#10-versionamento).

- **X** — grande bloco, com nome histórico do ZX81 (ex.: *Sinclair*)
- **Y** — feature ou bloco de programas, com nome de jogo do ZX81 (ex.: *3D Monster Maze*)
- **Z** — ajuste, correção ou compilação, com subtítulo curto

---

## 1.8.0 — Sinclair · QS Scramble · "Modelos"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Novos modelos**: Timex Sinclair 1500, Microdigital TK85 e Ringo R-470, além do ZX81 e do TS1000. O modelo
  traz a ROM, a memória (16K internos + RAM pack nos três novos) e o NTSC: `--model ts1500`, `machine tk85`,
  `modelo = "r470"` no `config.toml`, ou Opções → Hardware.
- **Diálogo Hardware no formato do EightyOne**: modelos, ROM (lista do modelo ou "…" para um arquivo), RAM pack
  com a RAM total, 60 Hz, RAM em 8K–16K, M1NOT, correção do ponto flutuante, alta resolução; as interfaces que
  ainda não existem (som, cor, fala, joystick, cartucho, ZX Printer, ZXpand, disquete, IDE) dizem "A implementar".
  ZX80, Lambda 8300, ZX97 Lite e Jupiter Ace aparecem como "a implementar".
- Opções novas também na linha de comando e no shell: `--lowram`, `--m1not`, `--fpfix` / `lowram=`, `m1not=`, `fpfix=`.
- **Todas as ROMs do EightyOne** para a família ZX81 embutidas (`go81 roms`): TS1500, TK85, Ringo, ZX80, Lambda
  (3), ZX97, Jupiter Ace, Forth (2) e ASZMIC (2).
- Ajuda → Teclado mostra o teclado do modelo (TS1500, TK85, R-470).
- Registrado na SPEC (§8.4, D-21): o `man` do fwMSX, para os comandos, o BASIC e a ROM do ZX81.

---

## 1.7.11 — Sinclair · Krazy Kong · "Som em WAV"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Gravar o som em WAV**: `record arquivo.wav` / `record stop` no shell e nos consoles, e Ferramentas → *Gravar o
  som em WAV* na janela (a barra de estado mostra `● gravando som`). Grava o som pelo vídeo do tempo emulado
  (44,1 kHz, 16 bits, estéreo); o arquivo fica válido a cada segundo.

---

## 1.7.10 — Sinclair · Krazy Kong · "Tokens do ZX80"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Programas do ZX80**: tabela de palavras-chave e de caracteres do ZX80 (do EightyOne, conferida com a ROM do
  ZX80) e listagem de `.o`/`.80` como texto: `go81 basic list jogo.o`, `list jogo.o` no shell e nos consoles.
  O ZX80 ainda não é emulado.

---

## 1.7.9 — Sinclair · Krazy Kong · "Janelas reunidas"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Exibir → Reunir janelas**: o console, o teclado na tela, as listagens e a ajuda do teclado vão para o centro
  da tela e para a frente (traz de volta uma janela perdida fora da tela).
- **Exibir → Trazer janelas para a frente**: as mesmas janelas passam para a frente da janela principal.

---

## 1.7.8 — Sinclair · Krazy Kong · "Zoom livre"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Zoom definido pelo usuário** (Exibir → Zoom → *Definido pelo usuário…*): qualquer tamanho de 1× a 8×,
  com fração (2,5×). Também `go81 gui --zoom 2.5` e `zoom = 2.5` na seção `[janela]` do `config.toml`.

---

## 1.7.7 — Sinclair · Krazy Kong · "Utilitários de BASIC"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Utilitários de programa BASIC** no shell, nos consoles e na janela (Ferramentas → Programa BASIC):
  `find texto` mostra as linhas com o texto; `delete de,até` apaga um bloco de linhas; `rem linha tamanho
  [byte]` cria uma linha REM para código de máquina e diz o endereço onde ele começa.

---

## 1.7.6 — Sinclair · Krazy Kong · "Menus do EightyOne"

**Data:** 2026-11-01
**Estado:** alfa

### Novidades
- **Menus do EightyOne**: Arquivo, Exibir, Máquina, Opções, Ferramentas e Ajuda com as opções do EightyOne; as
  que o Go81 ainda não tem mostram "A implementar".
- **Ajuda → Teclado (F1)**: a imagem do teclado do ZX81 (do TS1000 nas máquinas NTSC), do EightyOne; clique
  duplo dobra o tamanho.
- Novas opções que já funcionam: salvar programa, carregar/salvar bloco de memória, vídeo inverso, esconder a
  barra de estado, reset pela linha RESET, NMI, voltar à máquina de fábrica, perfis no menu, listagem BASIC.
- Hardware e "Salvar esta máquina como padrão" passaram para o menu Opções; o som pelo vídeo também.

---

## 1.7.5 — Sinclair · Krazy Kong · "Janela"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- **Arrastar e soltar**: um programa (`.p`, `.81`, `.p81`, `.bas`, `.txt`) solto na janela é carregado.
- **Abrir recente**: os últimos 8 programas no menu Arquivo (com *Limpar a lista*); Ctrl+O abre programa.
- **Tela cheia** (F11, menu Exibir).
- **Zoom e borda lembrados**: o que se escolhe no menu Exibir vai para a seção `[janela]` do `config.toml`.
- **Teclado do ZX81 na tela** (F3): as 40 teclas com o símbolo do SHIFT e a palavra-chave; SHIFT com trava até a
  próxima tecla; soma-se ao teclado do PC.

---

## 1.7.4 — Sinclair · Krazy Kong · "Scripts e eventos"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- **Scripts com variáveis, expressões e blocos**: `set`/`${nome}`, `calc`, `print`, `capture`, `assert`, `fail`,
  `waitfor`, `if/else/end`, `while/end`, `repeat N/end`, `break`; expressões com `peek()`, `dpeek()`, `reg()`,
  `screen("texto")`, `frames()`… — dá para escrever testes automáticos de programas do ZX81.
- **Link com token e eventos**: `-token` / `[link] token` / `link token` (o cliente envia `auth <token>`);
  `events on` avisa `! report 0/10` (o programa terminou), `! save NOME` e `! load`, também em JSON. Cliente Go:
  `Auth`, `EnableEvents`, `Events`, `WaitEvent`.
- **Mais comandos ZRCP**: `evaluate`/`e`, `save-screen`, `get-os`, `send-keys-ascii`, `reset-cpu`,
  `enter-cpu-step`/`exit-cpu-step` (pausa, com o prompt `command@cpu-step>`).
- **Barra de estado**: modo SLOW/FAST e a tecla sendo digitada (`tecla P (+12)`), no terminal e na janela.
- O pacote Linux também pode ser gerado no WSL (`./build.sh --install-deps`); instruções para compilar no macOS.

## 1.7.3 — Sinclair · Krazy Kong · "Perfis e pacotes"

**Data:** 2026-10-31
**Estado:** alfa — primeiro release com pacotes de Windows e de Linux

### Novidades
- **Configuração**: `config.toml` com a máquina padrão, a janela, as portas do Link e perfis de máquina
  (`[perfis.ts1000]`…). `-profile` e `-config` em `gui`, `tui`, `shell` e `link`; `go81 config path|init|show|edit`;
  `machine perfil=nome` no shell; na janela, Máquina → *Salvar esta máquina como padrão*.
- **Pacote para Linux**: `build.sh` (Linux e macOS, com `--install-deps`) e o `go81-X.Y.Z-linux-amd64.tar.gz`
  gerado pelo CI e anexado aos releases.
- **Colar automático**: na janela (Ctrl+V) e no terminal, um programa BASIC colado é carregado; outro texto é
  digitado. *Colar como teclas* força a digitação.

## 1.7.2 — Sinclair · Krazy Kong · "Programas de 1K"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- **Programas de 1K em máquinas maiores: não havia defeito.** Um `.p` com tela compacta (gravado num ZX81 de
  1K) parecia travar num ZX81 de 16K. A investigação mostrou que o primeiro `CLS` (feito pelo `LIST` e pelo
  `RUN`) expande a tela imprimindo 768 espaços, um a um, e em modo SLOW isso leva cerca de 2 segundos — no
  ZX81 real também. Os 29 programas de 1K dos exemplos do EightyOne mostram a mesma tela em 1K e em 16K.
  Um teste novo garante esse comportamento; a documentação do `-1k` foi corrigida.
- Gancho por instrução na máquina (`StepHook`), base para o depurador.
- **Terminal em 4:3**: o modo texto passa a usar duas células por caractere do ZX81 (letras de largura
  dupla do Unicode, gráficos em meios-blocos), chegando perto da proporção da TV; o modo pixels ocupa o
  espaço disponível em 4:3 (resolução inteira a partir de 132×98 células) e não apaga os traços finos
  nem o texto em inverso quando precisa reduzir. F9 alterna texto 4:3 / pixels 4:3 / texto compacto; em
  terminais estreitos o texto compacto entra sozinho.
- README com capturas de tela (janela, terminal, shell e console), geradas pelo próprio programa.

## 1.7.1 — Sinclair · Krazy Kong · "Do texto ao .p"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- **Loader de BASIC** (texto → `.p`): `go81 basic make jogo.bas [-o jogo.p] [-run N] [-1k]`, `make` no shell e
  nos consoles; `load jogo.bas` (também `.txt`), Abrir e **Colar programa BASIC** na janela, `go81 run jogo.bas`,
  `smartload` do ZRCP.
- Os números recebem o valor de 5 bytes calculado pela **rotina DEC-TO-FP da própria ROM**, executada no Z80
  emulado: o `.p` sai igual ao que o ZX81 grava quando a linha é digitada (conferido digitando 30 números).
- O lister ficou sem perda: escapa só o que o loader leria de outro jeito (variável `LN`, `POKE`...) e escreve
  o valor oculto quando ele não é o dos dígitos. Ida e volta idêntica nos 164 `.p` de exemplo do EightyOne.
- Linhas 0 a 16383, comentários `#`, continuação com `\`, `GO TO`/`GO SUB`/`RANDOMIZE`/`CONTINUE`.

### Conhecido
- Um `.p` com tela compacta (de 1K) parece travar ao ser carregado em 16K (corrigido na 1.7.2: era só a
  espera de cerca de 2 s do primeiro `CLS`, igual ao ZX81 real).

## 1.6.3 — Sinclair · Defenda · "Documentação em dia"

**Data:** 2026-10-31
**Estado:** alfa — **release de documentação**: o programa é o mesmo da 1.6.2 (só o número de versão muda)

### Novidades
- `resume.txt` reescrito: tudo o que funciona hoje (emulação, interfaces, shell, Link, BASIC,
  distribuição), estrutura do código, como compilar no Windows e no Linux e próximos passos.
- SPEC: estado atual da janela (menus Editar e Ferramentas, console), do terminal, do Link (`view-basic`),
  da CLI (`basic list`) e do lister de BASIC (§6.1).
- OUTLINE: resumo e próximo passo atualizados (loader de BASIC e bloco 2.x).
- README: exemplos com `gui -link` e `basic list`; MANUAL: exemplos com a versão atual.

## 1.6.2 — Sinclair · Defenda · "Console, canais e listagem"

**Data:** 2026-10-31
**Estado:** alfa — fecha o nível 1 do plano (itens rápidos)

### Novidades
- **Shell interativo** com histórico (↑ ↓, guardado entre execuções; comando `history`) e Tab que completa
  comandos, arquivos e opções.
- **GUI e TUI sobre a mesma sessão** do shell e do Link:
  - console de comandos (F2) dentro da janela e do terminal, com histórico e Tab;
  - `go81 gui -link` e `go81 tui -link`: a janela/terminal controlados de fora;
  - menu Ferramentas na janela: console e porta de comunicação.
- **Link por outros canais**, além do TCP: `go81 link -stdio` (entrada/saída padrão, para quem abre o go81
  como processo), `-socket <arquivo>` (socket Unix) e `-pipe <nome>` (pipe nomeado do Windows, só o
  usuário atual). No shell: `link socket`, `link pipe`. O cliente `pkg/go81link` conecta com
  `pipe:<nome>` e `unix:<arquivo>`, e `NewClient` usa qualquer canal.
- **Lister de BASIC**: `go81 basic list jogo.p [-o jogo.txt]`, `list` no shell e nos consoles, `view-basic`
  no ZRCP. Texto no formato do zxtext2p (gráficos `\' `, inverso `%A`, `\"`, `#` = £) mais `\{XX}` para
  bytes que não são texto. A listagem é conferida contra o `LIST` da própria ROM em todas as palavras-chave.
- **Tela em PNG, BMP, PBM ou TXT** (pela extensão) em `screenshot`, `go81 run --screenshot` e na janela.
- **Colar texto** na janela (Editar → Colar, Ctrl+V): o texto é digitado no ZX81; Copiar tela como texto
  (Ctrl+C).
- **CI no GitHub Actions**: testes em Linux (sem e com GUI) e Windows a cada push; o pacote ZIP do
  Windows fica como artefato.

### Conhecido
- O caminho inverso (texto → `.p`, o "loader") ainda não existe: é o próximo passo do BASIC.
- Colar texto digita tecla por tecla, como no teclado: palavras-chave por extenso não viram tokens
  (no cursor K, `P` é PRINT).
- Histórico e Tab do shell interativo só num terminal de verdade (em pipes e scripts, não).

## 1.6.1 — Sinclair · Defenda · "Shell e Link"

**Data:** 2026-10-31
**Estado:** alfa — fecha o bloco 1.x "Sinclair"

### Novidades
- **Registro de comandos único** (`internal/session`): cada comando é declarado uma vez, com uso e ajuda em
  português, e funciona igual no shell, em scripts e na porta de comunicação.
- **`go81 shell`**: shell interativo (`go81>`), scripts `.g81` (param no primeiro erro) e `-c "cmd; cmd"`.
  Comandos: `machine`, `reset`, `status`, `run`, `clock`/`pause`/`continue`, `type`/`typeln`, `key`, `peek`,
  `poke`, `regs`, `screen`, `screenshot`, `load`, `save`, `script`, `echo`, `sleep`, `cd`/`pwd`/`ls`, `roms`,
  `version`, `link`, `help`, `quit`. Números como no ZEsarUX: `16384`, `$4000`, `0x4000`, `4000H`, `1010%`.
- **`go81 link`**: porta de comunicação em `127.0.0.1`, com o ZX81 em tempo real.
  - Protocolo Go81 (porta 8181): uma linha por comando, resposta com `ok` / `err <mensagem>`, e pedidos JSON.
  - Modo **ZRCP** compatível com o ZEsarUX (porta 10000): `get-registers`, `read-memory`, `write-memory`,
    `set-register`, `send-keys-string`, `get-ocr`, `smartload`, `hard-reset-cpu` e outros, com as mesmas
    mensagens do original.
  - `go81 shell -link` junta os dois: shell e porta na mesma máquina.
- **Cliente Go** `pkg/go81link` (`Dial`, `Cmd`).
- Os comandos de digitação esperam o ZX81 terminar de ligar.
- Exemplo `examples/ola.g81`; MANUAL §10 e §11 escritos.

### Conhecido
- Sem histórico nem autocompletar no shell interativo.
- O ZRCP ainda não tem breakpoints nem passo a passo (vêm com o depurador).
- Link só por TCP; stdio/pipe, eventos e o GoLink do lado Z80 ficam para depois.

## 1.5.1 — Sinclair · Black Crystal · "Mapa das ferramentas"

**Data:** 2026-10-31
**Estado:** alfa — **release técnico**: documentação e especificação, sem executável novo

### Novidades
- Especificação do **montador asm81** (SPEC §6.4): sintaxe do asMSX sem a parte do MSX, símbolos do ZX81
  (variáveis de sistema, rotinas da ROM, caracteres, tokens), saída `.p` completa automática, e editor de assembler
  com linhas numeradas no estilo do **Mega Assembler**.
- Especificação do **ZX Dignified** (SPEC §6.2): BASIC moderno para o ZX81, com base no **zxtext2p** e recursos
  do **Basic Dignified Suite**, adaptado às limitações do ZX81 (blocos estruturados compilados para `GOTO`).
- Especificação do **depurador** (SPEC §6.6) no estilo do **Super-X**, com ferramentas do Mega Assembler e o que o
  ZX81 precisa (pular a rotina de vídeo, breakpoint em linha de BASIC, variáveis de sistema).
- Estudo do **VB81 XuR** (código VB6, GPL v2+) e do **ZEsarUX** (GPL v3): ideias de ferramentas e proposta de porta
  de comunicação compatível com o **ZRCP** (SPEC §14).
- Decisões D-12 a D-18; novas seções no OUTLINE; créditos no README.
- `RESOURCES.md`: material de referência e configurações fora do GitHub, com origem e plano de backup.
- `resume.txt` reescrito.

### Conhecido
- Nenhuma mudança no emulador; use o executável da 1.4.3 ou compile a partir do código.

## 1.4.3 — Sinclair · Football Manager · "Terminal"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- Interface de terminal com tview 0.42.0 / tcell 2.13.10: `go81 tui [programa.p]`, funciona por SSH.
- Tela de texto (32×24, decodificada do arquivo de exibição) com gráficos de bloco em quadrantes Unicode;
  modo pixels (F9) com a imagem real em meia resolução.
- Teclado por caractere com tradução de símbolos para SHIFT e Alt+tecla = SHIFT+tecla.
- F1 ajuda, F5 pausa, F6 turbo, F8 reinicia, F10 sai; `--exit-after` para automação.
- Testes com a tela simulada do tcell (digitar e rodar um programa, cursor, modo pixels).

### Conhecido
- Sem som no terminal.
- Alta resolução só no modo pixels, em meia resolução.

## 1.4.2 — Sinclair · Football Manager · "Som e configuração"

**Data:** 2026-10-31
**Estado:** alfa

### Novidades
- Som: beeper do EightyOne portado ("Video Sound": o sinal de sincronismo vira áudio), saída pela placa de
  som com oto 3.5.1. Desligado por padrão, como no EightyOne; liga em Máquina → Som ou `go81 gui --sound`.
- O som é comparado com o original quadro a quadro (3 cenários novos no `zx81diff`, 14 no total).
- Janela de configuração do hardware: ROM, memória (1K a 48K), NTSC, WRX e som.
- SAVE pela janela: quando o programa executa `SAVE "NOME"`, a janela oferece gravar `NOME.p`.
- `go81 gui --sound`, `--no-audio`.

### Conhecido
- No turbo o som fica mudo.
- Ainda sem gerenciador de fitas, depurador e outros periféricos.

## 1.4.1 — Sinclair · Football Manager · "Janela"

**Data:** 2026-10-31
**Estado:** alfa — primeira versão usável de forma interativa

### Novidades
- Interface gráfica com Fyne 2.8.2: `go81 gui [programa.p]`.
- Tela do ZX81 em tempo real (50 Hz PAL / 60 Hz NTSC), zoom 1–4× e tamanhos de borda.
- Teclado do PC mapeado por posição (Shift = SHIFT do ZX81), com atalhos para RUBOUT, cursor e EDIT.
- Menus: abrir programa (auto-load), salvar tela em PNG, reiniciar, pausar (F5), turbo 10× (F6), ajuda do teclado.
- Automação: `--exit-after N --capture janela.png`.
- Build sem GUI e sem compilador C: `-tags nogui`.

### Conhecido
- Sem som.
- Sem SAVE pela janela (use `go81 run --save`), sem gerenciador de fitas e sem depurador ainda.
- O teclado segue a posição das teclas do ZX81, não os símbolos do teclado do PC.

## 1.3.2 — Sinclair · Flight Simulation · "LOAD e SAVE"

**Data:** 2026-10-31
**Estado:** pré-alfa — ZX81 pela linha de comando, com carga e gravação de programas

### Novidades
- Carga de programas `.p`, `.81` e `.p81` com o *flash load* do EightyOne: player TZX (`internal/media/tape`)
  portado de `tzx/*.cpp`, interceptação da ROM portada de `rompatch.cpp` e *auto-load* (`LOAD ""` digitado
  sozinho) como no `main_.cpp`.
- `SAVE` rápido: `go81 run --save arquivo.p` grava o que o programa salvar.
- `go81 run programa.p` (ou `--tape`), `--autoload`, `--save`.
- Comparação com o EightyOne ampliada: o harness compila também o player TZX e o `rompatch.cpp` originais;
  11 cenários (incluindo carga com auto-load em PAL e NTSC e execução do programa carregado), sem diferenças.

### Conhecido
- `.tzx`, `.t81`, `.wav` e ZX80 ainda não.
- Diferenças em relação ao EightyOne em SPEC §5.2.2.

## 1.3.1 — Sinclair · Flight Simulation · "Máquina ZX81"

**Data:** 2026-10-31
**Estado:** pré-alfa — ZX81 roda pela linha de comando, sem janela

### Novidades
- Máquina ZX81 / TS1000 (`internal/emu/zx81`), portada do `zx81.cpp` do EightyOne: mapa de memória com RAM packs
  de 1K a 48K, sombras de RAM/ROM, RAM em 8–16K, geração de vídeo pelo Z80 (opcode fetch acima de M1NOT),
  gerador de NMI, HSYNC/VSYNC, inserção de WAIT, teclado, PAL/NTSC e alta resolução WRX.
- Raster de TV igual ao `AccurateDraw` do EightyOne, com as tolerâncias de sincronismo e a paleta padrão dele.
- Comparação quadro a quadro com o `zx81.cpp` original compilado com g++ (`test/compat/zx81ref` +
  `go run ./test/compat/zx81diff`): 7 cenários, 4.226 quadros, sem nenhuma diferença.
- Teste golden em `go test ./...`.
- Comando `go81 run`: roda o ZX81 sem janela, digita teclas e salva a tela em PNG.
- Comando `go81 roms`; ROMs ZX81 (edições 1, 2 e 3) embutidas no executável.
- Velocidade: cerca de 770 quadros por segundo (≈ 15× um ZX81 real).

### Conhecido
- Ainda sem janela, som e carga de programas (`.p`).
- Periféricos e outros clones ainda não emulados.
- Diferenças em relação ao EightyOne listadas em SPEC §5.2.2.

## 1.2.1 — Sinclair · Mazogs · "Núcleo Z80"

**Data:** 2026-10-31
**Estado:** pré-alfa — núcleo Z80 pronto, ainda sem máquina emulada

### Novidades
- Núcleo Z80 em Go (`internal/emu/z80`), porta fiel do núcleo do EightyOne: resultados, flags, T-states,
  M-cycles e ordem dos acessos ao barramento idênticos.
- Comparação automática com o núcleo C original (`test/compat/z80ref` + `go run ./test/compat/z80diff`):
  116.712 casos e 5 milhões de instruções aleatórias sem nenhuma diferença.
- Teste golden com 7.296 casos gravados do EightyOne, rodando em `go test ./...`.
- ZEXDOC/ZEXALL: mesmos resultados (e mesmos CRCs) do EightyOne.
- Velocidade: cerca de 44 milhões de instruções por segundo (≈ 90× um ZX81 real).

### Conhecido
- Por fidelidade ao EightyOne, o núcleo herda seus desvios do Z80 real: flags do `DAA` (falha no ZEXDOC)
  e flags não documentadas de `BIT` sem MEMPTR (2 grupos do ZEXALL).
- Endereços acima de 64K (ex.: `IX+d`) dão a volta em 16 bits; no EightyOne isso é comportamento indefinido.

## 1.1.2 — Sinclair · 3D Monster Maze · "Escopo e licença"

**Data:** 2026-10-30
**Estado:** pré-alfa — nada usável ainda

### Mudanças
- Licença trocada de BSD 3-Clause para **GPL v3 ou posterior**, permitindo portar código do EightyOne (GPL v2+).
- Escopo fechado: ZX81, ZX80, Timex TS1000 e TS1500, Lambda 8300, Ringo R-470, TK85, ZX97 Lite e **Jupiter Ace** (novo).
- Cabeçalho SPDX nos arquivos Go.

## 1.1.1 — Sinclair · 3D Monster Maze · "Planta baixa"

**Data:** 2026-10-30
**Estado:** pré-alfa — nada usável ainda

### Novidades
- Estudo inicial do EightyOne 1.41: máquinas da família ZX81, periféricos, formatos de arquivo e mapa dos fontes de referência.
- Documentação: `SPEC.md`, `OUTLINE.md`, `RELEASE.md`, `MANUAL.md`, `README.md`.
- Esqueleto Go: módulo `github.com/wilsonpilon/go81`, comando `go81` com `--version` e `--help`.

### Conhecido
- Nenhuma emulação implementada.
- Licença final ainda não definida (resolvido em 1.1.2).
