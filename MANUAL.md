# Go81 — Manual do usuário

Versão do manual: 1.7.6

> O Go81 está em fase alfa: o ZX81 já roda numa janela, no terminal e pela linha de comando.
> Itens marcados *(planejado)* estão especificados em [SPEC.md](SPEC.md) mas ainda não existem.

## Sumário

1. [Instalação](#1-instalação)
2. [Linha de comando](#2-linha-de-comando)
3. [Modos de interface](#3-modos-de-interface)
4. [Emulador](#4-emulador) — `run`, `gui`, `tui`
5. [BASIC](#5-basic)
6. [Editor de textos](#6-editor-de-textos-planejado)
7. [Montador Z80](#7-montador-z80-planejado)
8. [Fitas](#8-fitas)
9. [Depurador](#9-depurador-planejado)
10. [Porta de comunicação (Link)](#10-porta-de-comunicação-link)
11. [Shell](#11-shell)
12. [Configuração](#12-configuração)
13. [Códigos de saída](#13-códigos-de-saída)

## 1. Instalação

**Pacote pronto (Windows):** baixe `go81-X.Y.Z-windows-amd64.zip` na página de releases do GitHub,
descompacte e rode `go81.exe` (ou `go81-nogui.exe`, sem janela) num terminal. Não precisa instalar nada.

**Pacote gerado por você (Windows):** na pasta do projeto, `.\build.ps1`. O script instala o que faltar
(Go e gcc do MSYS2, pelo `winget`), baixa os módulos, roda os testes, compila os dois executáveis e gera
`dist\go81-X.Y.Z-windows-amd64\`, o `.zip` e o `.sha256`; o `bin\` também é atualizado. Opções:
`-SkipTests`, `-NoInstall` (só avisa o que falta), `-NoGui` (sem gcc), `-Msys2 <pasta>`.

**Pacote pronto (Linux):** `go81-X.Y.Z-linux-amd64.tar.gz` nos releases: `tar xzf` e `./go81`. A janela
precisa das bibliotecas de OpenGL/X11 do sistema (já instaladas num Linux com área de trabalho);
`go81-nogui` funciona em qualquer lugar.

**Pacote gerado por você (Linux e macOS):** `./build.sh` faz o mesmo que o `build.ps1` e gera
`dist/go81-X.Y.Z-<sistema>-<arquitetura>.tar.gz` e o `.sha256`. Opções: `--skip-tests`, `--nogui`,
`--install-deps` (instala gcc e as bibliotecas da janela com apt, dnf ou pacman, usando sudo).

**Linux pelo WSL (Windows):** num Ubuntu do WSL, na pasta do projeto, `./build.sh --install-deps` instala
o que falta e gera o pacote Linux em `dist/`, sem sair do Windows.

**macOS (só instruções; não há pacote pronto):** instale as ferramentas de linha de comando da Apple
(`xcode-select --install`) e o Go 1.25 ou mais novo (https://go.dev/dl/), baixe o código e rode `./build.sh`.
O resultado é `dist/go81-X.Y.Z-darwin-<arm64|amd64>.tar.gz`. A janela usa o OpenGL do próprio macOS; sem
compilador C, `./build.sh --nogui`.

**Manualmente** (Go 1.25 ou mais novo):

```sh
git clone https://github.com/wilsonpilon/go81
cd go81
go build -o bin/ ./cmd/go81
```

O executável fica em `bin/go81` (Linux) ou `bin\go81.exe` (Windows).

A janela (GUI) usa o Fyne, que precisa de um compilador C (CGO):
- **Windows:** MSYS2 com `mingw-w64-ucrt-x86_64-gcc`, e `C:\msys64\ucrt64\bin` no `PATH`.
- **Linux:** `gcc` e os pacotes de desenvolvimento de OpenGL e X11 (ex.: `libgl1-mesa-dev xorg-dev`).

Sem compilador C, gere a versão sem janela (linha de comando e terminal): `go build -tags nogui -o bin/ ./cmd/go81`
(os testes: `go test -tags nogui ./...`).

## 2. Linha de comando

```
go81 <comando> [opções]
```

| Comando | Descrição |
|---|---|
| `gui` | Abre o ZX81 numa janela (ver §4.3) |
| `tui` | Abre o ZX81 no terminal (ver §4.4) |
| `run` | Roda o ZX81 sem janela e pode salvar a tela (PNG, BMP, PBM ou texto) (ver §4) |
| `basic list` | Lista um programa `.p` como texto (ver §5) |
| `shell` | Shell de comandos: interativo, script ou `-c` (ver §11) |
| `link` | Porta de comunicação: outros programas controlam o ZX81 (ver §10) |
| `roms` | Lista as ROMs embutidas no executável |
| `version`, `--version`, `-v` | Mostra a versão completa (número e codinomes) |
| `help`, `--help`, `-h` | Mostra a ajuda |

`go81 <comando> -h` mostra as opções de cada comando.

```
$ go81 --version
go81 1.7.6 — Sinclair · Krazy Kong · "Menus do EightyOne"
```

Planejado: `--gui`, `--tui`, `--config`, `--profile`, os comandos `asm` e `tape` e mais ferramentas em `basic`. Ver [SPEC.md §8.3](SPEC.md#83-cli).

## 3. Modos de interface

| Modo | Como abrir | Estado |
|---|---|---|
| GUI (Fyne) | `go81 gui` | disponível |
| TUI (tview) | `go81 tui` | disponível |
| Shell | `go81 shell` | disponível |
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
| `--screenshot` | — | Salva a última tela; o formato vem da extensão: `.png`, `.bmp`, `.pbm` (preto e branco) ou `.txt` (32×24 caracteres) |
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
| `--console` | — | Abre também o console de comandos |
| `--link` | — | Liga a porta de comunicação (opções `--port` e `--zrcp` como em `go81 link`) |
| `--exit-after` | — | Automação: fecha depois de N quadros |
| `--capture` | — | Automação: com `--exit-after`, salva a imagem da janela num PNG |

**Menus:** seguem os do EightyOne (Arquivo, Exibir, Máquina, Opções, Ferramentas, Ajuda), mais o Editar. As
opções do EightyOne que o Go81 ainda não tem estão nos menus e mostram um aviso **"A implementar"**.

- *Arquivo* — Abrir programa (Ctrl+O; `.p`, `.81`, `.p81` ou BASIC em texto `.bas`/`.txt`, carregado com
  auto-load), **Abrir recente** (os últimos 8 programas abertos; *Limpar a lista*), **Salvar programa** (o que
  está no ZX81, num `.p`), **Carregar/Salvar bloco de memória** (um arquivo binário num endereço; endereço e
  tamanho em decimal, `$hex` ou `0x` hex), Salvar tela (PNG, BMP, PBM ou TXT, pela extensão), Sair. Também dá
  para **arrastar um programa** do Explorer (ou do gerenciador de arquivos) e soltar na janela: ele é carregado
  como pelo menu. *A implementar:* fitas, drives, snapshots.
- *Editar* — **Colar** (Ctrl+V): se o texto da área de transferência é um programa BASIC (todas as linhas
  começam com um número), ele vira um `.p` (§5.2) e é carregado; senão, é digitado no ZX81 (uma linha =
  ENTER; caracteres que o ZX81 não tem são ignorados). *Colar como teclas* força a digitação. Copiar tela como
  texto (Ctrl+C). O resultado aparece por alguns segundos na barra de baixo.
- *Exibir* — Zoom (1× a 4×, **Tela cheia** F11); Tamanho da borda (os dois ficam gravados na seção `[janela]`
  do arquivo de configuração, §12, e a próxima janela abre do mesmo jeito); **Vídeo inverso**; **Barra de
  estado** (mostra/esconde); **Teclado do ZX81 na tela** (F3). *A implementar:* zoom definido pelo usuário,
  anotações, reunir janelas, tela de abertura.
- *Máquina* — **Reiniciar (linha RESET)**: reinicia só o Z80, como o botão de reset (a memória fica);
  **Desligar e ligar**; Pausar/continuar (F5); **NMI** (gera uma interrupção não mascarável); Turbo 10× (F6).
  *A implementar:* joysticks, Memotech, Quicksilva, Chroma, Spectra, divIDE/ZXCF, proteção da RAM de 8–16K.
- *Opções* — Configuração: Salvar esta máquina como padrão (§12), Voltar à máquina de fábrica (ZX81 16K,
  edição 3, PAL) e os **perfis** do arquivo de configuração; Som pelo vídeo (liga/desliga); **Hardware…** (ROM,
  memória, NTSC, WRX, som; aplica reiniciando). *A implementar:* fita, portas, tela, teclado, som, velocidade,
  emulação de TV.
- *Ferramentas* — Console de comandos (F2); **Listagem BASIC** (o programa do ZX81 numa janela, com botão
  Copiar); ligar, desligar e ver o estado da porta de comunicação. *A implementar:* memória ao vivo,
  depurador, ZX Printer, saída de som, gerenciador de fitas, carregador WAV.
- *Ajuda* — **Teclado** (F1): a imagem do teclado do ZX81 (do TS1000 nas máquinas NTSC), tirada do
  EightyOne; **clique duas vezes na imagem** para dobrar o tamanho (e de novo para voltar). Manual, exemplos,
  histórico de versões e atualizações abrem a página do projeto no GitHub. Sobre.

As teclas de atalho são as do Go81 (F1 teclado, F2 console, F3 teclado na tela, F5 pausa, F6 turbo, F11 tela
cheia), não as do EightyOne.

**Gravar programas:** quando o programa no ZX81 executa `SAVE "NOME"`, a janela pergunta onde gravar
o arquivo, já sugerindo `NOME.p`.

**Som:** o ZX81 não tem alto-falante; o "som" é o sinal de vídeo/sincronismo levado ao áudio, como faziam
os jogos que usavam o rádio ou a TV. Por isso ele vem desligado (como no EightyOne): ligue em
Opções → Som pelo vídeo ou com `--sound`. No modo turbo o som fica mudo.

**Console de comandos (F2):** uma janela com os comandos do shell (§11) agindo no ZX81 da janela:
`peek`, `poke`, `regs`, `type`, `load`, `save`, `link start`… Tab completa, ↑ ↓ trazem comandos anteriores.
O `save` do console grava direto no arquivo pedido (a janela de gravação não abre).

**Teclado:** as teclas do PC ficam na mesma posição das do ZX81, e o Shift do PC é o SHIFT do ZX81
(Shift+P = `"`). Atalhos: Backspace = RUBOUT, setas = cursor, vírgula = `,`, Esc = EDIT.
A barra inferior mostra a máquina, os quadros por segundo e se a fita está tocando.

**Teclado na tela (F3):** as 40 teclas do ZX81 desenhadas como no original, com o símbolo do SHIFT em vermelho
em cima e a palavra-chave do cursor K embaixo. A tecla fica apertada enquanto o botão do mouse estiver
pressionado (um clique rápido vale pelo menos 0,1 s, para o ZX81 ler). **SHIFT** fica preso com um clique e
solta sozinho depois da próxima tecla (ou com outro clique): SHIFT e depois P dá `"`. As teclas do PC também
funcionam com essa janela em foco, e as duas formas se somam.

### 4.4 `go81 tui`

```
go81 tui [opções] [programa.p]
```

Abre o ZX81 no terminal (também funciona por SSH). Colar texto na tela do ZX81 funciona como o *Colar* da
janela: um programa BASIC é carregado, outro texto é digitado (o resultado aparece na linha de baixo).
Opções `--profile`, `--config` (§12), `--rom`, `--ram`, `--ntsc`, `--wrx`, `--sound`
como em `go81 run`; `--console` abre com o console visível; `--link` liga a porta de comunicação
(`--port`, `--zrcp`); `--exit-after N` sai depois de N quadros e imprime a tela de texto (automação).

A célula do terminal é cerca de duas vezes mais alta que larga, então há três modos de tela (F9 alterna):

| Modo | Como fica | Precisa de |
|---|---|---|
| **Texto 4:3** (padrão) | Cada caractere do ZX81 ocupa duas células: letras e números nas formas de **largura dupla** do Unicode (`Ｅ`), gráficos de bloco como meios-blocos (`▀▄█`) — o desenho fica exato e a tela, perto do 4:3 da TV | 68 colunas |
| **Pixels 4:3** | A imagem real da TV, com 2×2 pontos por célula, ocupando o espaço disponível em 4:3. A partir de 132×98 células é a resolução inteira (exata); menor que isso é uma redução, boa para gráficos e alta resolução | — |
| **Texto compacto** | Um caractere por célula (a tela fica "em pé"); usado sozinho quando o terminal tem menos de 68 colunas | 36 colunas |

Use um terminal com fonte que tenha os caracteres de largura dupla e de bloco (o Windows Terminal e os
terminais do Linux têm).

**Teclado:** digite normalmente. Letras e números viram as teclas do ZX81; os símbolos viram a combinação
SHIFT certa (`"` = SHIFT+P, `+` = SHIFT+K, `=` = SHIFT+L, `( ) $ : ; ? / * < > , -`). Enter = ENTER,
Backspace = RUBOUT, setas = cursor, Esc = EDIT, Alt+tecla = SHIFT+tecla (Alt+F = FAST, Alt+D = SLOW,
Alt+A = STOP, Alt+9 = GRAPHICS).

| Tecla | Ação |
|---|---|
| F1 | Ajuda |
| F2 | Console de comandos (abre / fecha; Esc também fecha) |
| F5 | Pausa / continua |
| F6 | Turbo 10× |
| F8 | Reinicia (desliga e liga) |
| F9 | Tela: texto 4:3 / pixels 4:3 / texto compacto |
| F10 ou Ctrl+C | Sai |

Como o terminal não informa quando uma tecla é solta, cada tecla é "apertada" por alguns quadros; depois
de ENTER há uma pausa de pouco mais de 1 s enquanto a ROM redesenha a tela. Sem som no terminal.

**Barra de baixo:** máquina, modo **SLOW** ou **FAST** (em amarelo) e, enquanto teclas são digitadas pelo
Go81 (colar, `type`, Link), a tecla apertada e quantas faltam (`tecla P (+12)`).

**Console (F2):** um painel na parte de baixo com os comandos do shell (§11), no ZX81 que está na tela.
Enquanto ele está aberto, o teclado vai para o console; Tab completa, ↑ ↓ trazem comandos anteriores,
Esc fecha. O ZX81 continua rodando enquanto o comando trabalha.

## 5. BASIC

### 5.1 Listar um programa como texto

```
go81 basic list jogo.p               # mostra na tela
go81 basic list jogo.p -o jogo.txt   # grava num arquivo
```

No shell e nos consoles: `list` lista o programa que está no ZX81; `list jogo.p` lista um arquivo. Pelo ZRCP:
`view-basic`.

A listagem sai igual à do `LIST` do ZX81 (os espaços em volta das palavras-chave seguem a mesma regra da
ROM; isso é conferido automaticamente contra o ZX81 emulado). O texto usa o formato do **zxtext2p**, para
poder ser editado e, no futuro, voltar a ser um `.p`:

| No ZX81 | No texto |
|---|---|
| Palavras-chave (`PRINT`, `GOTO`, `**`, `<=`…) | Por extenso |
| Gráficos de bloco | `\` + dois caracteres: `\' ` `\ '` `\''` `\. ` `\: ` `\.'` `\:'` `\##` `\,,` `\~~` e os inversos `\::` `\.:` `\:.` `\..` `\':` `\ :` `\'.` `\ .` `\@@` `\;;` `\!!` |
| Vídeo inverso | `%` + o caractere (`%A`) |
| Aspas dentro de uma string | `\"` |
| Libra (£) | `#` |
| Espaço fora de strings (guardado na linha) | `\  ` |
| Byte que não é texto, ou palavra-chave dentro de REM ou de string | `\{XX}` (hexadecimal) — extensão do Go81 |

Com o `\{XX}`, nada se perde: código de máquina numa linha REM aparece como `\{..}`. Um programa com uma
linha "impossível" (tamanho errado, truque de alguns programas) é listado até ela, com um aviso.

Para que o texto volte **exatamente** ao mesmo programa, o lister confere cada linha com o loader (§5.2) e
escapa só o que seria lido de outro jeito: uma variável chamada `LN` ou `POKE` (`LET \{35}OKE=6000`), uma
palavra-chave grudada numa letra, dígitos sem o valor oculto. Quando o valor de 5 bytes guardado não é o dos
dígitos (truque de economia de memória), ele aparece depois do número: `0\{7E}\{81}\{00}\{00}\{00}\{00}`.
Conferido nos 164 programas `.p` de exemplo do EightyOne: listar e voltar dá os mesmos bytes.

### 5.2 Texto → programa `.p`

```
go81 basic make jogo.bas                 # grava jogo.p
go81 basic make jogo.bas -o outro.p -run 10   # roda sozinho a partir da linha 10 ao carregar
go81 basic make jogo1k.bas -1k           # tela compacta, para ZX81 de 1K
```

No shell e nos consoles: `make jogo.bas [programa.p] [run=N] [1k]`. E o `load` aceita o texto direto:
`load jogo.bas` (também `.txt` e `.b81`), assim como a janela (Arquivo → Abrir, Editar → **Colar programa
BASIC**), o terminal, `go81 run jogo.bas` e o `smartload` do ZRCP.

Regras do texto (as mesmas do zxtext2p, mais as extensões do §5.1):

- Uma linha do programa por linha: número (0 a 16383) e o conteúdo. Linhas em branco e linhas que começam
  com `#` são ignoradas; uma linha terminada em `\` continua na seguinte.
- Fora de strings e de REM, os espaços só separam palavras. Palavras-chave por extenso, em maiúsculas ou
  minúsculas (`print`, `PRINT`); também `GO TO`, `GO SUB`, `RANDOMIZE`, `CONTINUE`. Um nome de variável
  termina num espaço: `IF A THEN`, `A$(X TO Y)`.
- Cada número recebe o valor oculto de 5 bytes calculado pela **própria rotina da ROM** (DEC-TO-FP, rodada
  no Z80 emulado): o `.p` sai igual ao que o ZX81 grava quando você digita a linha (a ROM trunca: `0.1` é
  `7D 4C CC CC CC`, e `0.5` é `7F 7F FF FF FF`).
- O `.p` parte de um programa vazio gravado pela ROM (16K: tela cheia; `-1k`: tela compacta).
- Erros dizem a linha do texto: `linha 2: o caractere '!' não existe no ZX81`.

Tela compacta (`-1k`) ou cheia (padrão): os dois funcionam em qualquer memória. Com a tela compacta o `.p`
fica menor (cabe num ZX81 de 1K), mas numa máquina de 4K ou mais o primeiro `CLS` (que o `LIST` e o `RUN`
fazem) expande a tela um espaço por vez e leva **cerca de 2 segundos** — o ZX81 real faz o mesmo. A tela
cheia evita essa espera.

### 5.3 Editor de BASIC e compilador Dignified *(planejado)*
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
## 10. Porta de comunicação (Link)

A porta de comunicação deixa outros programas (editores, scripts de teste, IDEs) controlarem o ZX81.
O ZX81 roda em tempo real e cada conexão manda **comandos do shell** (§11).

```
go81 link                      # Go81 em 127.0.0.1:8181 e ZRCP em 127.0.0.1:10000
go81 link jogo.p               # carrega um programa antes
go81 link -port 9000 -zrcp -1  # outra porta, sem ZRCP
go81 link -stdio               # protocolo pela entrada/saída padrão (sem TCP)
go81 link -pipe go81           # pipe nomeado do Windows: \\.\pipe\go81
go81 link -socket /tmp/go81.sock -port -1   # só socket Unix
go81 shell -link               # shell e link juntos, na mesma máquina
go81 gui -link                 # a janela, controlada de fora (também go81 tui -link)
```

**Canais** (podem ser usados juntos; todos falam o protocolo Go81, e o ZRCP só por TCP):

| Canal | Opção | Comando no shell | Quem pode conectar |
|---|---|---|---|
| TCP | `-port`, `-zrcp` (`-1` desliga) | `link start [porta] [porta-zrcp\|off]` | Só o próprio computador (`127.0.0.1`) |
| Entrada/saída padrão | `-stdio` | — | O programa que abriu o `go81` |
| Socket Unix | `-socket <arquivo>` | `link socket <arquivo>` | Quem pode abrir o arquivo (Linux e Windows 10+) |
| Pipe nomeado (Windows) | `-pipe <nome>` | `link pipe [nome]` | Só o usuário atual |

`link stop` fecha todos; `link status` mostra os abertos. Com `-stdio`, a saída padrão é só do protocolo
(as mensagens vão para a saída de erro), as portas TCP só ligam se pedidas, e o `go81` termina quando a
entrada acaba ou chega `quit`. Exemplo em Python:

```python
import subprocess
p = subprocess.Popen(["go81", "link", "-stdio"], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
saida, _ = p.communicate('typeln 10P"OLA"' + chr(10) + "screen" + chr(10) + "quit" + chr(10))
```

No Windows, o pipe abre como arquivo em qualquer linguagem (C#: `NamedPipeClientStream(".", "go81")`;
Python: `open(r"\\.\pipe\go81", "r+b", buffering=0)`).

### 10.1 Protocolo Go81 (porta 8181)

Ao conectar, o servidor manda uma linha `go81 link X.Y.Z`. Depois, cada linha enviada é um comando
(vários podem ir na mesma linha, separados por `;`). A resposta é a saída do comando seguida de uma linha
`ok` ou `err <mensagem>`:

```
$ telnet 127.0.0.1 8181
go81 link 1.7.6
typeln 10P"OLA"
ok
peek 16396 2
400C: 7D 40                                            ..
ok
pokee 1
err comando desconhecido: pokee (digite help)
```

Se uma linha da saída começar com `ok`, `err` ou `.`, ela recebe um `.` na frente (o cliente tira).
`quit` responde `ok tchau` e fecha só a conexão; o servidor continua.

**Token (senha):** com `-token segredo` (ou `token = "segredo"` em `[link]` no `config.toml`, ou
`link token segredo` no shell), a saudação termina em ` auth` e o cliente precisa enviar `auth segredo` antes
de qualquer comando. Vale para TCP, socket e pipe; o ZRCP não tem senha (como no ZEsarUX) — desligue-o com
`-zrcp -1` se precisar.

**Eventos:** `events on` faz o servidor avisar, entre as respostas, o que acontece no ZX81, em linhas que
começam com `!` (linhas da saída que começam com `!` recebem o `.` na frente):

```
events on
eventos ligados
ok
typeln R
! report 0/10        ← o programa terminou (relatório 0, linha 10)
ok
! save JOGO          ← o programa gravou com SAVE "JOGO"
! load               ← a fita terminou de carregar
```

Pedido em JSON (`{"cmd":"events on"}`), os eventos vêm em JSON: `{"event":"report","data":"0/10"}`.
`events off` desliga. No cliente Go: `c.Auth(token)`, `c.EnableEvents()`, `c.Events()` e `c.WaitEvent()`.

**JSON:** uma linha que começa com `{` é um pedido JSON e recebe uma linha JSON de volta:

```
{"cmd":"peek 0"}
{"ok":true,"output":"0000: D3 (211)\n"}
{"cmd":"xyz"}
{"ok":false,"output":"","error":"comando desconhecido: xyz (digite help)"}
```

**Cliente em Go:** o pacote `github.com/wilsonpilon/go81/pkg/go81link`:

```go
c, err := go81link.Dial("127.0.0.1:8181") // ou "pipe:go81", "unix:/tmp/go81.sock"
defer c.Close()
c.Cmd(`typeln 10P"OLA"`) // P = PRINT
tela, err := c.Cmd("screen")
```

### 10.2 Modo ZRCP (porta 10000, compatível com o ZEsarUX)

Um subconjunto do protocolo de comando remoto do ZEsarUX, com as mesmas mensagens, para aproveitar
ferramentas feitas para ele. Cada resposta termina com o prompt `command> `.

| Comando | O que faz |
|---|---|
| `about`, `get-version`, `get-current-machine` (`gcm`) | Identificação (`ZX81` ou `TS1000`) |
| `get-registers` (`gr`) | Registradores no formato do ZEsarUX (`MEMPTR` sempre `0000`: não é emulado, como no EightyOne) |
| `set-register` (`sr`) `REG=valor` | Muda um registrador (PC, SP, AF, BC, DE, HL, IX, IY, AF'…, A, F, B…, I, R) |
| `read-memory endereço [tamanho]` | Bytes em hexadecimal, sem espaços |
| `write-memory` (`wm`) `endereço v1 v2…` | Escreve bytes |
| `send-keys-string pausa texto` | Digita o texto (pausa em ms por tecla) |
| `get-ocr` | Texto da tela, 24 linhas (gráficos de bloco viram `#`) |
| `view-basic` | Listagem do programa BASIC (como `list`) |
| `smartload arquivo` (`sl`) | Carrega um programa (como `load`) |
| `hard-reset-cpu` | Desliga e liga |
| `evaluate` (`e`) `expressão` | Calcula uma expressão do Go81 (`e peek(16396)`, `e reg("HL")+1`, §11.1) |
| `save-screen arquivo` | Salva a tela (`.png`, `.bmp`, `.pbm`, `.txt`) |
| `send-keys-ascii pausa código…` | Digita teclas pelos códigos ASCII (13 = ENTER) |
| `enter-cpu-step` (`encs`) / `exit-cpu-step` (`ecs`) | Pausa / retoma o ZX81 (o prompt vira `command@cpu-step>`); o passo a passo vem com o depurador |
| `reset-cpu`, `get-os` | Reinicia; sistema operacional |
| `help [comando]`, `quit` / `exit` / `logout` | Ajuda e saída |

Os números seguem as regras do ZEsarUX e do shell: decimal, `$4000`, `0x4000`, `4000H`, `1010%`.

## 11. Shell

O shell é um pequeno sistema de comandos para o ZX81. Os mesmos comandos funcionam no shell interativo,
em scripts e pela porta de comunicação.

```
go81 shell                                  # interativo (prompt go81>)
go81 shell build.g81                        # executa um script; para no primeiro erro (código 1)
go81 shell -c "load jogo.p; run 100; screenshot fim.png"
go81 gui                                    # F2 abre os mesmos comandos dentro da janela
go81 shell -ram 2 -ntsc                     # escolhe a máquina (também -rom, -wrx, -sound)
echo "peek 16396 2" | go81 shell            # comandos pela entrada padrão
```

**No modo interativo:** ↑ ↓ trazem comandos anteriores (o histórico fica em
`%APPDATA%\go81\shell_history` no Windows e `~/.config/go81/shell_history` no Linux; `history` lista);
Tab completa nomes de comandos, arquivos (`load`, `save`, `script`, `cd`…) e opções (`machine ram=`,
`link start`, bordas do `screenshot`); com várias opções, Tab de novo mostra todas. Ctrl+C apaga a
linha; Ctrl+D (Ctrl+Z e Enter no Windows) sai.

No shell, o ZX81 só anda quando um comando o faz andar (`run`, `type`, `load`…): os resultados são sempre
iguais. Com `clock on` (ou `-link`) ele roda sozinho em tempo real. Os comandos de digitação esperam o
ZX81 terminar de ligar.

| Comando | O que faz |
|---|---|
| `help [comando]` | Lista os comandos ou explica um deles |
| `machine [perfil=nome] [zx81\|ts1000] [ram=K] [ntsc=on\|off] [wrx=on\|off] [sound=on\|off] [rom=nome]` | Mostra ou troca a máquina (desliga e liga); `perfil=` usa um perfil da configuração (§12) |
| `reset` | Desliga e liga |
| `status` | Quadro, PC, relógio, fita, teclas pendentes |
| `run [quadros]` (`wait`) | Deixa o ZX81 rodar (padrão 50 = 1 s em PAL) |
| `clock [on\|off]`, `pause`, `continue` | Relógio de tempo real |
| `type texto` | Digita o texto (`\n` = ENTER); letras na posição do cursor K viram comandos (P = PRINT) |
| `typeln texto` | Como `type`, com ENTER no fim |
| `key tecla…` | Aperta teclas, ex.: `key SHIFT+1 ENTER` |
| `peek endereço [quantidade]` | Lê a memória (hexadecimal e texto do ZX81) |
| `poke endereço valor…` | Escreve na memória (a ROM fica protegida) |
| `regs` | Registradores do Z80 |
| `screen` | Tela como texto (32×24) |
| `screenshot arquivo [none\|small\|normal\|large\|full]` | Salva a tela: `.png`, `.bmp`, `.pbm` ou `.txt` (pela extensão) |
| `list [programa.p]` | Lista o programa BASIC do ZX81 (ou de um arquivo) como texto (§5) |
| `load programa [noauto]` | Carrega da fita (digita `LOAD ""` sozinho e espera); aceita `.p`, `.81`, `.p81` e BASIC em texto (`.bas`, `.txt`) |
| `make texto.bas [programa.p] [run=N] [1k]` | Transforma um programa em texto num `.p` (§5.2) |
| `save programa.p [NOME]` | Executa `SAVE "NOME"` no ZX81 e grava o arquivo |
| `script arquivo.g81` (`source`) | Executa um script |
| `echo texto`, `sleep ms` | Mensagem, espera |
| `cd`, `pwd`, `ls` (`dir`) | Pastas e arquivos |
| `roms`, `version` | ROMs embutidas, versão |
| `link [start\|stop\|status]` | Porta de comunicação (§10) |
| `history [quantidade]` | Últimos comandos digitados (↑ ↓ os trazem de volta; Tab completa comandos, arquivos e opções) |
| `quit` (`exit`, `sair`) | Sai |

Argumentos com espaço vão entre aspas: `save "meu jogo.p" JOGO`. Em scripts, linhas começadas por `#` são
comentários. Exemplo (`examples/ola.g81`):

```
# digita, roda e grava um programa
run 150
typeln 10P"OLA DO GO81"
typeln R
screenshot ola.png
save ola.p OLA
```

Em `type`/`typeln`, cada caractere é uma tecla, como no teclado real: no cursor K uma letra é um comando
(`P` = PRINT, `R` = RUN, `L` = LET, `G` = GOTO), então `10P"OLA"` vira `10 PRINT "OLA"`. Para colar um programa
escrito por extenso, use `load programa.bas` ou o *Colar* da janela (§5.2).

### 11.1 Variáveis, expressões e blocos

Os scripts (e o shell interativo, menos os blocos) têm variáveis, expressões e blocos — o bastante para
automatizar testes de programas do ZX81:

```
# testa um programa: carrega, roda e confere a tela e a memória
load jogo.bas
typeln R
waitfor 500 screen("0/")          # espera o relatório
assert screen("FIM")              # para com erro se não aparecer
set limite 10
calc i 0
while i < limite
  poke 30000+i i*2
  calc i i + 1
end
if peek(30009) == 18
  echo ok: ${i} bytes escritos
else
  fail memória errada
end
repeat 3
  key SPACE
end
```

| Comando | O que faz |
|---|---|
| `set nome valor` | Guarda um texto numa variável; `set` sozinho lista as variáveis; `unset nome` apaga |
| `${nome}` | Em qualquer comando, é trocado pelo valor (`$4000` continua sendo hexadecimal) |
| `calc nome expressão` | Calcula e guarda (`calc i i + 1`) |
| `print expressão` | Mostra o valor (`print dpeek(16396)`) |
| `capture nome comando` | Guarda a saída de um comando (`capture v regs`) |
| `assert expressão` | Para o script com erro se for falsa |
| `fail mensagem` | Para o script com esta mensagem |
| `waitfor quadros expressão` | Roda o ZX81 até a expressão ficar verdadeira (erro se passar do limite) |
| `if expr` … `else` … `end` | Condição (só em scripts) |
| `while expr` … `end` | Repete enquanto for verdadeira; `break` sai (só em scripts) |
| `repeat N` … `end` | Repete N vezes (só em scripts) |

**Expressões:** números (`16384`, `$4000`, `0x4000`, `'A'`), textos entre aspas, nomes de variáveis, operadores
`+ - * / %`, comparações `== != < <= > >=`, `&& || !` e parênteses. Funções: `peek(a)`, `dpeek(a)` (16 bits),
`reg("HL")`, `pc()`, `sp()`, `frames()`, `screen("texto")` (verdadeiro se o texto está na tela) e `len("texto")`.
Um erro mostra o arquivo e a linha (`teste.g81:12: assert falhou: peek(30009) == 18`).
## 12. Configuração

Um arquivo de texto (TOML) guarda a máquina de sempre, o tamanho da janela, as portas do Link e **perfis**
de máquina com nome. As opções digitadas na linha de comando valem por cima dele.

```
go81 config init      # cria o arquivo com um exemplo comentado
go81 config path      # onde ele fica: %APPDATA%\go81\config.toml ou ~/.config/go81/config.toml
go81 config show      # a máquina que vale agora e os perfis
go81 config edit      # abre no editor (o de $EDITOR, ou o do sistema)
```

```toml
[maquina]           # a máquina de sempre
rom  = "zx81.edition3"
ram  = 16           # 0 (1K), 2, 4, 8, 16, 32 ou 48
ntsc = false
wrx  = false
som  = false

[janela]
zoom  = 2
borda = "normal"

[link]
porta = 8181
zrcp  = 10000

[perfis.ts1000]
ram  = 2
ntsc = true
```

- **Perfis:** `go81 gui -profile ts1000`, `go81 tui -profile wrx`… (também `shell` e `link`). No shell e nos
  consoles: `machine perfil=ts1000` (e mais opções por cima: `machine perfil=ts1000 ram=16`).
- **Outro arquivo:** `-config caminho.toml`, ou a variável de ambiente `GO81_CONFIG`.
- **Pela janela:** Opções → Configuração → *Salvar esta máquina como padrão* grava a seção `[maquina]`, e o zoom e a borda
  escolhidos no menu Exibir gravam a seção `[janela]` (o resto do arquivo e os comentários ficam como estão).
  `--zoom` e `--border` na linha de comando continuam valendo por cima.
- Uma opção escrita errado é avisada (`opções desconhecidas: maquina.ramm`).
- O `go81 run` **não** lê o arquivo: ele é a ferramenta de automação e dá sempre o mesmo resultado.

## 13. Códigos de saída

| Código | Significado |
|---|---|
| 0 | Sucesso |
| 1 | Erro geral |
| 2 | Uso incorreto (parâmetro inválido) |
