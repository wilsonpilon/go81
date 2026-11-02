// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

// The commands of the shell (go81 shell), of the consoles of the window and
// of the terminal (F2), of scripts (.g81) and of the Link. A test checks
// that every registered command has a page here.

var shellPages = []kwPage{
	{name: "help", aliases: []string{"ajuda", "?"}, usage: "help [comando]", summary: "lista os comandos ou explica um deles",
		page: `Sem nada, lista os comandos com uma linha cada. help comando mostra o uso e os outros nomes. Para a página completa: man comando.`, see: []string{"shell man"}},
	{name: "man", usage: "man [seção] nome\nman -k palavra\nman -l [seção]", summary: "o manual do Go81",
		page: `Mostra a página de um comando, de uma palavra do BASIC, de uma variável de sistema, de um rótulo da ROM, de um código de erro ou de um assunto de referência. Sem nada, mostra as seções. Quando o mesmo nome existe em várias seções (PRINT do BASIC e print do shell), a primeira na ordem basic, shell, cli, basic80, sysvar, rom, erro, ref aparece e as outras são citadas no fim; man shell print escolhe a seção.

  man PRINT          man shell load        man D_FILE
  man FP-CALC        man erro 4            man -k fita
  man -l sysvar      man basic80           man memoria

Também na linha de comando (go81 man ...) e na janela (Ajuda → Manual).`},
	{name: "version", usage: "version", summary: "mostra a versão do Go81", page: `Número da versão e codinomes (SPEC §10).`},
	{name: "machine", usage: "machine [perfil=nome] [modelo] [ram=K] [rom=nome] [ntsc=on|off] [wrx=on|off]\n        [sound=on|off] [lowram=on|off] [m1not=on|off] [fpfix=on|off]",
		summary: "mostra a máquina ou troca o modelo e a configuração",
		page: `Sem nada, descreve a máquina. Com opções, monta a nova máquina e a liga (a memória é apagada). As opções valem da esquerda para a direita: o modelo (zx81, ts1000, ts1500, tk85, r470, zx80, ou modelo=nome) traz a ROM, a RAM e o NTSC dele, e o que vier depois vale por cima.

  machine ts1500 ram=16     (16K internos + 16K = 32K)
  machine zx80
  machine rom=zx81.edition1 fpfix=on
  machine perfil=ts1000     (um perfil do config.toml)`,
		see: []string{"ref modelos", "cli go81 config", "shell roms"}},
	{name: "reset", usage: "reset", summary: "desliga e liga (a memória é apagada)", page: `Como desligar o cabo de força: tudo volta ao começo.`},
	{name: "status", usage: "status", summary: "quadro, PC, relógio, fita e teclas pendentes", page: `Um resumo do estado da sessão.`},
	{name: "run", aliases: []string{"wait"}, usage: "run [quadros]", summary: "deixa o ZX81 rodar N quadros",
		page: `Roda N quadros (padrão 50 = 1 s em PAL, 60 por segundo em NTSC) o mais rápido possível. Com o relógio de tempo real ligado (janela, terminal, clock on), espera esse tempo.

Não confundir com RUN do BASIC (man basic RUN): para rodar o programa, typeln R.`,
		see: []string{"shell waitfor", "shell clock"}},
	{name: "pause", usage: "pause", summary: "pausa o relógio de tempo real", page: `O ZX81 para (na janela, F5).`, see: []string{"shell continue"}},
	{name: "continue", aliases: []string{"cont"}, usage: "continue", summary: "retoma o relógio de tempo real", page: `Desfaz o pause.`},
	{name: "clock", aliases: []string{"relogio"}, usage: "clock [on|off]", summary: "relógio de tempo real",
		page: `Com o relógio ligado o ZX81 roda sozinho a 50/60 quadros por segundo (no shell ele começa desligado: o ZX81 só anda com run, type, load...).`},
	{name: "type", usage: "type texto", summary: "digita um texto no ZX81 e espera terminar",
		page: `Cada caractere vira a tecla que o produz (letras, números e sinais; \n = ENTER). No cursor K a primeira letra é a palavra-chave: type P"OI" digita PRINT "OI". O ZX80 tem os sinais em outras teclas e o Go81 usa as dele.

  type 10P"OLA"\n20G10\n`,
		see: []string{"shell typeln", "shell key", "ref teclas"}},
	{name: "typeln", usage: "typeln texto", summary: "como type, com ENTER no fim", page: `typeln R roda o programa.`},
	{name: "key", usage: "key tecla [tecla ...]", summary: "aperta teclas do ZX81 em sequência",
		page: `Cada argumento é uma tecla ou uma combinação com +: letras, números, SPACE, ENTER, SHIFT, ., e SHIFT+tecla.

  key SHIFT+ENTER P     (cursor F e TAB)
  key SHIFT+F ENTER     (FAST)`,
		see: []string{"ref teclas"}},
	{name: "peek", usage: "peek endereço [quantidade]", summary: "lê a memória",
		page: `Um byte, ou um bloco em hexadecimal com o texto do ZX81 ao lado. Números: 16384, $4000, 4000h, 0x4000, 1010%.

  peek $400C 2       (D_FILE)`, see: []string{"shell poke", "basic PEEK"}},
	{name: "poke", usage: "poke endereço valor [valor ...]", summary: "escreve na memória", page: `Escreve bytes a partir do endereço (a ROM continua protegida).`},
	{name: "regs", aliases: []string{"registers"}, usage: "regs", summary: "mostra os registradores do Z80", page: `PC, SP, AF, BC, DE, HL, IX, IY, os alternativos, I, R, modo de interrupção e flags.`},
	{name: "screen", usage: "screen", summary: "a tela como texto (32×24)", page: `Lê o arquivo de tela (D_FILE) e mostra o texto (não funciona com telas de alta resolução).`,
		see: []string{"shell screenshot", "sysvar D_FILE"}},
	{name: "screenshot", usage: "screenshot arquivo [none|small|normal|large|full]", summary: "salva a tela em arquivo",
		page: `O formato vem da extensão: .png, .bmp, .pbm ou .txt. O segundo argumento é a borda.`},
	{name: "load", usage: "load programa [noauto]", summary: "carrega um programa pela fita",
		page: `Põe o programa na fita, digita LOAD "" sozinho (como o auto-load do EightyOne) e espera carregar. Aceita .p, .81, .p81, BASIC em texto (.bas, .txt, .b81, transformado em .p) e, no ZX80, .o e .80. noauto só põe na fita.`,
		see:  []string{"basic LOAD", "shell save", "shell make"}},
	{name: "save", usage: "save arquivo.p [NOME]", summary: "executa SAVE no ZX81 e grava o arquivo",
		page: `Digita SAVE "NOME" (padrão: o nome do arquivo) e grava o que o ZX81 salvou. No ZX80 o SAVE não tem nome e o arquivo é .o.`,
		see:  []string{"basic SAVE", "shell load"}},
	{name: "list", aliases: []string{"listar"}, usage: "list [programa.p|.o]", summary: "lista o programa BASIC como texto",
		page: `Sem nada, o programa do ZX81 (ou do ZX80); com um arquivo, o arquivo. O texto pode voltar a ser .p com make ou load (formato em man ref texto-basic).`,
		see:  []string{"shell make", "ref texto-basic", "cli go81 basic"}},
	{name: "make", usage: "make texto.bas [programa.p] [run=N] [1k]", summary: "transforma um programa em texto num .p",
		page: `Os números recebem o valor calculado pela própria ROM. run=N faz o programa rodar da linha N ao carregar; 1k usa a tela compacta (cabe em 1K).`,
		see:  []string{"ref texto-basic"}},
	{name: "find", aliases: []string{"procurar"}, usage: "find texto", summary: "linhas do programa que contêm o texto",
		page: `Procura no texto da listagem (maiúsculas e minúsculas tanto faz; palavras-chave contam: find GOTO). Só no BASIC do ZX81.`},
	{name: "delete", aliases: []string{"del"}, usage: "delete linha\ndelete de,até", summary: "apaga linhas do programa",
		page: `Apaga uma linha ou um bloco (delete 100,190 ou 100-190) e recarrega o programa pela fita: as variáveis se perdem, como num LOAD. Só no BASIC do ZX81.`},
	{name: "renum", aliases: []string{"renumerar"}, usage: "renum [início [passo [de,até]]]", summary: "renumera o programa e corrige os GOTO",
		page: `Dá às linhas os números início, início+passo... (padrão 10 em 10) e corrige o número escrito depois de GOTO, GOSUB, RUN, LIST e LLIST em todo o programa. Um GOTO para uma linha que não existe passa a apontar para a linha seguinte, como o ZX81 faria. Um número calculado (GOTO A*10, GOSUB VAL "100") não pode ser corrigido: aparece um aviso. Com de,até, só esse bloco é renumerado, e ele não pode passar por cima das outras linhas. O programa é recarregado pela fita (as variáveis se perdem). Só no BASIC do ZX81.

  renum             → 10, 20, 30...
  renum 1000 5      → 1000, 1005, 1010...
  renum 500 10 200,290 → só as linhas de 200 a 290`, see: []string{"shell auto", "basic GOTO", "basic GOSUB"}},
	{name: "auto", usage: "auto [início [passo]]", summary: "numera sozinho as linhas de BASIC digitadas",
		page: `Depois de auto, o que se digita no shell ou no console é BASIC, sem número de linha: o prompt mostra o número da próxima (auto 120>). Começa 10 depois da última linha do programa (ou em início) e anda de passo em passo (padrão 10). Um número no começo da linha troca a numeração a partir dali. Uma linha vazia ou um ponto termina, e só então as linhas entram no programa (recarregado pela fita; as variáveis se perdem). Uma linha com erro não entra e pode ser digitada de novo. Palavras-chave como no texto BASIC: man ref texto-basic.

  auto 100 5
  auto 100> PRINT "OLA"
  auto 105> GOTO 100
  auto 110> .`, see: []string{"shell renum", "shell typeln"}},
	{name: "rem", usage: "rem linha tamanho [byte]", summary: "cria uma linha REM para código de máquina",
		page: `Cria (ou troca) a linha "linha REM" com tamanho bytes iguais a byte (padrão 0; 118 não pode) e diz onde o código começa.

  rem 1 100      → o código começa em 16514 ($4082)`, see: []string{"basic REM", "basic USR"}},
	{name: "record", aliases: []string{"gravar-som"}, usage: "record arquivo.wav\nrecord stop\nrecord", summary: "grava o som num .wav",
		page: `Grava o som pelo vídeo (44,1 kHz, 16 bits, estéreo) do tempo emulado. Sem argumento, mostra a gravação. Com o som pelo vídeo desligado, grava silêncio (e avisa).`},
	{name: "script", aliases: []string{"source"}, usage: "script arquivo.g81", summary: "executa os comandos de um arquivo",
		page: `Uma linha por comando; # começa um comentário. Scripts têm variáveis, expressões e blocos (if, while, repeat): man ref scripts.`, see: []string{"ref scripts"}},
	{name: "echo", usage: "echo texto", summary: "escreve um texto", page: `${nome} é trocado pelo valor da variável.`},
	{name: "sleep", usage: "sleep milissegundos", summary: "espera um tempo", page: `Útil com o relógio de tempo real.`},
	{name: "set", usage: "set [nome valor]", summary: "guarda um valor numa variável", page: `${nome} nos comandos é trocado pelo valor. Sem nada, lista as variáveis.`, see: []string{"ref scripts"}},
	{name: "unset", usage: "unset nome", summary: "apaga uma variável", page: `Apaga uma variável do script.`},
	{name: "calc", usage: "calc nome expressão", summary: "calcula uma expressão e guarda o resultado", page: `calc i i+1. Expressões: man ref expressoes.`, see: []string{"ref expressoes"}},
	{name: "print", usage: "print expressão", summary: "mostra o valor de uma expressão",
		page: `print peek(16396)+256*peek(16397). Não confundir com PRINT do BASIC (man basic PRINT).`, see: []string{"ref expressoes"}},
	{name: "capture", usage: "capture nome comando", summary: "guarda a saída de um comando numa variável", page: `capture tela screen`},
	{name: "assert", usage: "assert expressão", summary: "para o script com erro se a expressão for falsa", page: `Para testes automáticos de programas.`, see: []string{"ref scripts"}},
	{name: "fail", usage: "fail mensagem", summary: "para o script com esta mensagem de erro", page: `O código de saída do go81 shell é de erro.`},
	{name: "waitfor", usage: "waitfor quadros expressão", summary: "roda até a expressão ser verdadeira",
		page: `Roda o ZX81 quadro a quadro até a expressão valer (ou dar erro depois de N quadros).

  waitfor 500 screen("0/0")`, see: []string{"ref expressoes"}},
	{name: "if", usage: "if expressão\n  ...\n[else\n  ...]\nend", summary: "bloco condicional (scripts)", page: `Em scripts .g81.`, see: []string{"ref scripts"}},
	{name: "else", usage: "else", summary: "parte do if", page: `Começa a parte que roda quando a condição do if é falsa.`},
	{name: "end", usage: "end", summary: "fecha if, while ou repeat", page: `Fim de um bloco.`},
	{name: "while", usage: "while expressão\n  ...\nend", summary: "repete enquanto a expressão for verdadeira", page: `Em scripts.`},
	{name: "repeat", usage: "repeat N\n  ...\nend", summary: "repete N vezes", page: `Em scripts.`},
	{name: "break", usage: "break", summary: "sai do while ou repeat", page: `Em scripts.`},
	{name: "link", usage: "link [start [porta] [porta-zrcp|off] | socket arquivo | pipe [nome] |\n      token senha|off | stop | status]",
		summary: "porta de comunicação (outros programas controlam o ZX81)",
		page:    `Liga o servidor do Link: protocolo Go81 (TCP 8181) e modo compatível com o ZRCP do ZEsarUX (TCP 10000), ou socket Unix, ou pipe do Windows. token exige "auth senha" dos clientes. Detalhes no MANUAL §10.`,
		see:     []string{"cli go81 link"}},
	{name: "history", aliases: []string{"historico"}, usage: "history [quantidade]", summary: "os últimos comandos digitados",
		page: `Mostra os últimos comandos do shell interativo (padrão: 20). As setas para cima e para baixo trazem os comandos de volta; o histórico fica guardado entre as sessões.`},
	{name: "pwd", usage: "pwd", summary: "mostra a pasta atual", page: `A pasta onde os arquivos são procurados.`},
	{name: "cd", usage: "cd pasta", summary: "muda a pasta atual", page: `Como no sistema operacional.`},
	{name: "ls", aliases: []string{"dir"}, usage: "ls [pasta]", summary: "lista os arquivos", page: `Os arquivos e pastas.`},
	{name: "roms", usage: "roms", summary: "lista as ROMs embutidas", page: `Todas as ROMs do EightyOne para a família ZX81.`, see: []string{"ref modelos"}},
	{name: "quit", aliases: []string{"exit", "sair"}, usage: "quit", summary: "sai", page: `Sai do shell (no console da janela, fecha a janela).`},
}

func init() {
	providers = append(providers, func() {
		for _, p := range shellPages {
			add(&Topic{Name: p.name, Aliases: p.aliases, Section: "shell", Usage: p.usage, Summary: p.summary, Page: p.page, SeeAlso: p.see})
		}
	})
}
