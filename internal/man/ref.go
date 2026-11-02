// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

// The command line (go81 <command>) and the reference pages.

var cliPages = []kwPage{
	{name: "go81 gui", aliases: []string{"go81-gui", "gui"}, usage: "go81 gui [opções] [programa]", summary: "abre o ZX81 numa janela",
		page: `A janela com som, menus no formato do EightyOne, console (F2), teclado na tela (F3), tela cheia (F11). Opções de máquina (--model, --rom, --ram, --ntsc, --wrx, --sound, --lowram, --m1not, --fpfix, --profile, --config), --zoom 1 a 8, --border, --console, --link, --no-audio, e para automação --exit-after N --capture arquivo.png. Detalhes: go81 gui -h e MANUAL §4.3.`},
	{name: "go81 tui", aliases: []string{"go81-tui", "tui"}, usage: "go81 tui [opções] [programa]", summary: "abre o ZX81 no terminal",
		page: `A tela do ZX81 em texto 4:3 ou em pixels escalados; funciona por SSH. Mesmas opções de máquina da janela. F2 abre o console.`},
	{name: "go81 run", aliases: []string{"go81-run"}, usage: "go81 run [opções] [programa]", summary: "roda o ZX81 sem janela (automação)",
		page: `Roda N quadros (--frames), pode digitar teclas (--keys), carregar um programa, gravar o que o programa salvar (--save) e salvar a última tela (--screenshot). Não lê o config.toml: dá sempre o mesmo resultado.

  go81 run --frames 300 --keys "P,SHIFT+P,O,I,SHIFT+P,ENTER" --screenshot tela.png`},
	{name: "go81 shell", aliases: []string{"go81-shell"}, usage: "go81 shell [opções] [script.g81]\ngo81 shell -c \"comando; comando\"", summary: "o shell de comandos",
		page: `Interativo (com histórico e Tab), um script ou comandos com -c. Os comandos estão na seção shell (man -l shell).`},
	{name: "go81 link", aliases: []string{"go81-link"}, usage: "go81 link [opções]", summary: "o ZX81 controlado por outros programas",
		page: `Roda o ZX81 com a porta de comunicação ligada (TCP, stdio, socket ou pipe), sem janela. Protocolo Go81 e ZRCP (DeZog). MANUAL §10.`},
	{name: "go81 basic", aliases: []string{"go81-basic"}, usage: "go81 basic list [-o arquivo.txt] programa.p|.o\ngo81 basic make texto.bas [programa.p]", summary: "ferramentas de BASIC",
		page: `list mostra um .p (ou um .o do ZX80) como texto; make transforma texto em .p.`, see: []string{"ref texto-basic"}},
	{name: "go81 config", aliases: []string{"go81-config"}, usage: "go81 config path | show | init | edit", summary: "o arquivo de configuração",
		page: `Mostra onde fica o config.toml, a máquina que ele descreve, cria um modelo ou abre no editor.`, see: []string{"ref config"}},
	{name: "go81 roms", aliases: []string{"go81-roms"}, usage: "go81 roms", summary: "lista as ROMs embutidas", page: `As ROMs do EightyOne para a família ZX81.`},
	{name: "go81 version", aliases: []string{"go81-version"}, usage: "go81 version", summary: "mostra a versão", page: `Também go81 --version.`},
	{name: "go81 help", aliases: []string{"go81-help"}, usage: "go81 help", summary: "lista os comandos", page: `go81 <comando> -h mostra as opções de cada um.`},
	{name: "go81 man", aliases: []string{"go81-man"}, usage: "go81 man [seção] nome\ngo81 man -k palavra\ngo81 man -l [seção]", summary: "o manual", page: `O mesmo man do shell (man man).`},
}

var refPages = []kwPage{
	{name: "intro", summary: "como o manual está organizado",
		page: `O manual do Go81 tem uma página para cada comando (seções shell e cli), cada palavra-chave do BASIC do ZX81 (basic) e do ZX80 (basic80), cada variável de sistema (sysvar), cada rótulo da ROM do ZX81 (rom), cada código de erro (erro) e alguns assuntos (ref).

  man -l ref        os assuntos
  man memoria       o mapa da memória
  man caracteres    o conjunto de caracteres
  man teclas        nomes das teclas para key e --keys`},
	{name: "memoria", aliases: []string{"memória", "mapa"}, summary: "o mapa da memória do ZX81",
		page: `  0000-1FFF  ROM (8K); 1E00-1FFF é o conjunto de caracteres
  2000-3FFF  livre (RAM em 8K-16K com a opção lowram; cópia da ROM sem ela)
  4000-407C  variáveis de sistema (man -l sysvar)
  407D       o programa BASIC (a primeira linha)
  D_FILE     o arquivo de tela (logo depois do programa)
  VARS       as variáveis do BASIC
  E_LINE     a linha em edição; depois, a pilha do calculador
  STKEND     memória livre
  RAMTOP     fim da área do BASIC (NEW não apaga acima)
  8000-FFFF  cópias da memória; o código acima de 32K só roda com M1NOT

Com 1K a memória vai até 43FFh; com 16K, até 7FFFh.`,
		see: []string{"sysvar D_FILE", "sysvar RAMTOP", "basic REM"}},
	{name: "tela", aliases: []string{"display"}, summary: "como o ZX81 guarda a tela",
		page: `O arquivo de tela (D_FILE) começa com um NEWLINE (76h) e tem 24 linhas, cada uma terminada por NEWLINE. Com 4K ou mais as linhas têm 32 caracteres (tela cheia, 793 bytes); com pouca memória as linhas podem ser curtas (tela "comprimida": 25 bytes vazia). As duas últimas linhas (DF_SZ) são da edição.

A imagem é feita pelo Z80: ele "executa" o arquivo de tela acima de 32K; cada caractere vira um NOP e o hardware busca os 8 pontos dele no conjunto de caracteres (registrador I = 1Eh).`,
		see: []string{"sysvar D_FILE", "sysvar DF_SZ", "rom CHAR-SET"}},
	{name: "caracteres", aliases: []string{"charset", "conjunto de caracteres"}, summary: "o conjunto de caracteres do ZX81",
		page: `  0 espaço   1-10 gráficos de blocos   11 "   12 £   13 $   14 :   15 ?
  16 (   17 )   18 >   19 <   20 =   21 +   22 -   23 *   24 /   25 ;   26 ,   27 .
  28-37 os algarismos 0-9   38-63 as letras A-Z
  64-127 e 192-255: palavras-chave e códigos de controle (não são caracteres)
  128-191: os mesmos 0-63 em vídeo inverso

Não é ASCII: CODE "A" é 38. No texto do Go81 os gráficos são escritos com \ e dois caracteres (man texto-basic).`,
		see: []string{"basic CODE", "basic CHR$", "ref caracteres80"}},
	{name: "caracteres80", summary: "o conjunto de caracteres do ZX80",
		page: `  0 espaço   1 "   2-11 gráficos   12 £   13 $   14 :   15 ?   16 (   17 )
  18 -   19 +   20 *   21 /   22 =   23 >   24 <   25 ;   26 ,   27 .
  28-37 os algarismos   38-63 as letras

Os sinais estão em outra ordem que no ZX81 (o Go81 lista e mostra a tela do ZX80 com esta tabela).`},
	{name: "teclas", aliases: []string{"teclado", "keys"}, summary: "nomes das teclas para key, --keys e scripts",
		page: `Letras A-Z, algarismos 0-9, SPACE, ENTER, SHIFT e . (ponto). Combinações com +: SHIFT+P. O ZX81 tem 40 teclas numa matriz de 8 meias-filas lidas pela porta FEh.

  SHIFT+ENTER    cursor F (funções)
  SHIFT+9        cursor G (gráficos)
  SHIFT+0        RUBOUT (apaga)
  SHIFT+1        EDIT
  SHIFT+5..8     setas
  SPACE          BREAK (durante um programa)

No PC: as letras, números e sinais vão para a tecla do ZX81 que os produz; Backspace = RUBOUT; setas = setas. Ajuda → Teclado (F1) mostra a imagem do teclado.`,
		see: []string{"shell key", "rom K-UNSHIFT"}},
	{name: "texto-basic", aliases: []string{"formato de texto", "bas"}, summary: "o formato de texto dos programas BASIC",
		page: `O mesmo do zxtext2p (e do BASIC Loader do EightyOne): uma linha do programa por linha, palavras-chave por extenso, números normais.

  \' \. \: \'' ...   gráficos de blocos (\ e dois caracteres)
  %A                  A em vídeo inverso
  \"                  aspas dentro de um texto
  #                   a libra (£)
  \{XX}               um byte em hexadecimal (código de máquina num REM, ou o que não é texto)

O Go81 lista e carrega sem perda: list mostra, make e load voltam ao .p.`,
		see: []string{"shell list", "shell make"}},
	{name: "variaveis", aliases: []string{"variáveis"}, summary: "as variáveis do BASIC do ZX81",
		page: `Numéricas: letra seguida de letras e algarismos (A, TOTAL1). Texto: uma letra e $ (A$). Matrizes: uma letra, com DIM. Controle de FOR: uma letra. Os números têm 5 bytes (expoente e 4 de mantissa: cerca de 9,5 algarismos).`,
		see:  []string{"basic LET", "basic DIM", "sysvar VARS"}},
	{name: "operadores", aliases: []string{"<=", ">=", "<>", "**", "="}, summary: "operadores e prioridades",
		page: `  **            potência (6 ** 2 = 36)                 prioridade 10
  - (unário)    negativo                                9
  * /           multiplicação, divisão                  8
  + -           soma, subtração (+ junta textos)        6
  = < > <= >= <>  comparações (verdadeiro = 1)          5
  NOT           negação                                 4
  AND           e                                       3
  OR            ou                                      2

Funções (SIN, PEEK...) têm prioridade maior que todos: SIN X*2 é (SIN X)*2.`,
		see: []string{"basic AND", "basic OR", "basic NOT"}},
	{name: "textos", aliases: []string{"strings", "fatias"}, summary: "textos e fatias no BASIC do ZX81",
		page: `Um texto vai entre aspas; "" é uma aspa. + junta textos. A$(n) é o enésimo caractere; A$(a TO b) uma fatia; LEN dá o tamanho; VAL e STR$ convertem; CODE e CHR$ trabalham com códigos.`},
	{name: "expressoes", aliases: []string{"expressões"}, summary: "expressões do shell (print, calc, if, waitfor)",
		page: `Números (decimal, $hex, 4000h, 0x hex, 1010% binário), textos entre aspas, ${variável}, operadores + - * / % == != < <= > >= && || ! (também and), e funções:

  peek(a) dpeek(a)     memória (byte, palavra)
  reg("hl") pc() sp()  registradores do Z80
  screen("texto")      1 se o texto está na tela
  frames()             quadros rodados
  len(s)               tamanho de um texto

Veja o MANUAL §11.1.`,
		see: []string{"shell print", "shell calc", "shell waitfor"}},
	{name: "scripts", aliases: []string{"g81"}, summary: "scripts .g81 e testes automáticos",
		page: `Um script tem um comando por linha; # é comentário. Blocos: if/else/end, while/end, repeat N/end, break. Variáveis com set e ${nome}. assert e fail tornam o script um teste.

  # testa um programa
  load jogo.p
  typeln R
  waitfor 500 screen("FIM")
  assert peek(16507)=0

Rode com go81 shell teste.g81 (código de saída 0 = passou).`,
		see: []string{"shell script", "ref expressoes"}},
	{name: "modelos", aliases: []string{"maquinas", "máquinas"}, summary: "os modelos que o Go81 emula",
		page: `  zx81    Sinclair ZX81 (padrão)      ROM zx81.edition3 (1 e 2, Forth, ASZMIC)   1K
  ts1000  Timex Sinclair 1000         zx81.edition3                              2K, NTSC
  ts1500  Timex Sinclair 1500         ts1500                                     16K, NTSC
  tk85    Microdigital TK85           tk85 (10K)                                 16K, NTSC
  r470    Ringo R-470                 ringo470                                   16K, NTSC
  zx80    Sinclair ZX80               zx80 (ou zx81.edition1)                    1K

O RAM pack substitui a memória interna (ZX81, TS1000, ZX80) ou soma a ela (16K internos). Lambda 8300, ZX97 Lite e Jupiter Ace ficam para depois.`,
		see: []string{"shell machine", "ref zx80"}},
	{name: "zx80", summary: "o que muda no ZX80",
		page: `Vídeo sem gerador de NMI (a tela some com o programa rodando e com tecla apertada), BASIC de inteiros (man -l basic80), outro conjunto de caracteres (man caracteres80), sinais em outras teclas, programas .o gravados sem nome. No cursor K: O = PRINT, W = LOAD, E = SAVE, R = RUN.`,
		see:  []string{"basic80", "ref caracteres80"}},
	{name: "config", aliases: []string{"config.toml"}, summary: "o arquivo de configuração",
		page: `Seções [maquina] (modelo, rom, ram, ntsc, wrx, som, ram_8k_16k, m1not, correcao_pf), [janela] (zoom, borda), [link] (porta, zrcp, token), [perfis.nome] e [jogos."arquivo.p"] (a máquina de cada jogo: man jogo). As opções da linha de comando valem por cima. go81 config path mostra onde ele fica.`,
		see:  []string{"cli go81 config", "shell machine"}},
}

func init() {
	providers = append(providers, func() {
		for _, p := range cliPages {
			add(&Topic{Name: p.name, Aliases: p.aliases, Section: "cli", Usage: p.usage, Summary: p.summary, Page: p.page, SeeAlso: p.see})
		}
		for _, p := range refPages {
			add(&Topic{Name: p.name, Aliases: p.aliases, Section: "ref", Usage: p.usage, Summary: p.summary, Page: p.page, SeeAlso: p.see})
		}
	})
}
