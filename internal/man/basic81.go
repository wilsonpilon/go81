// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

import (
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

// The BASIC of the ZX81 (ROM edition 3). Texts written for Go81 from the
// ZX81 manual's rules; the key of each keyword comes from the ROM tables.

type kwPage struct {
	name, usage, summary, page string
	aliases, see               []string
}

var basic81Pages = []kwPage{
	// --- commands ---
	{name: "PRINT", usage: "PRINT [item] [; | ,] [item] ...\nPRINT AT linha,coluna; item\nPRINT TAB coluna; item",
		summary: "escreve números e textos na tela",
		page: `Escreve cada item na posição de impressão. Um ; entre itens não deixa espaço; uma vírgula pula para a coluna 0 ou 16; sem ; ou , no fim, a próxima impressão vai para a linha de baixo. AT linha,coluna (0-21, 0-31) e TAB coluna mudam a posição.

Números saem com até 8 algarismos significativos (1E+12 em notação científica). Textos entre aspas; "" dentro de um texto é uma aspa.

  10 PRINT "OLA"; " MUNDO"
  20 PRINT AT 10,5;"MEIO", 1/3
  30 PRINT TAB 20;"*"

Erros: 5 (a tela encheu: use SCROLL ou CLS), B (AT ou TAB fora da tela).`,
		see: []string{"basic LPRINT", "basic SCROLL", "basic CLS", "basic AT", "basic TAB"}},
	{name: "LPRINT", usage: "LPRINT [item] [; | ,] ...", summary: "como PRINT, mas na impressora (ZX Printer)",
		page: `Igual a PRINT, mas o texto vai para o buffer da impressora (PRBUFF), que é impresso a cada linha completa. AT linha,coluna usa só a coluna. Sem impressora ligada, nada acontece.`,
		see:  []string{"basic LLIST", "basic COPY", "sysvar PRBUFF"}},
	{name: "LIST", usage: "LIST [linha]", summary: "mostra o programa a partir de uma linha",
		page: `Limpa a tela e lista o programa a partir da linha dada (ou do começo). A linha atual (o cursor de programa >) passa a ser essa.

No Go81, o comando list do shell lista o programa como texto, sem mexer na tela.`,
		see: []string{"basic LLIST", "shell list", "sysvar E_PPC"}},
	{name: "LLIST", usage: "LLIST [linha]", summary: "lista o programa na impressora", page: `Como LIST, mas na ZX Printer.`,
		see: []string{"basic LIST", "basic LPRINT"}},
	{name: "RUN", usage: "RUN [linha]", summary: "apaga as variáveis e roda o programa",
		page: `Faz CLEAR (apaga as variáveis), CLS e GOTO para a linha dada (ou a primeira). Termina com o relatório 0 se o programa chegar ao fim.

  RUN
  RUN 100`,
		see: []string{"basic GOTO", "basic CLEAR", "basic CONT", "shell run"}},
	{name: "GOTO", aliases: []string{"GO TO"}, usage: "GOTO expressão", summary: "desvia para uma linha",
		page: `Continua o programa na linha dada; se ela não existir, na seguinte. A expressão pode ser calculada (GOTO 100+10*X), mas o renumerador não consegue corrigir desvios calculados.

  10 PRINT "A";
  20 GOTO 10`,
		see: []string{"basic GOSUB", "basic RUN", "basic IF"}},
	{name: "GOSUB", aliases: []string{"GO SUB"}, usage: "GOSUB expressão", summary: "chama uma sub-rotina",
		page: `Guarda a linha seguinte na pilha do GOSUB e desvia para a linha dada; RETURN volta. As sub-rotinas podem se chamar umas às outras enquanto houver memória.

  10 GOSUB 100
  20 STOP
  100 PRINT "SUB-ROTINA"
  110 RETURN

Erro 4: memória cheia (GOSUB demais sem RETURN).`,
		see: []string{"basic RETURN", "basic GOTO"}},
	{name: "RETURN", usage: "RETURN", summary: "volta de uma sub-rotina",
		page: `Volta para a linha depois do último GOSUB. Erro 7: RETURN sem GOSUB.`, see: []string{"basic GOSUB"}},
	{name: "IF", usage: "IF condição THEN comando", summary: "executa um comando se a condição for verdadeira",
		page: `Se a condição for verdadeira (diferente de 0), executa o comando depois de THEN; senão, passa para a próxima linha. O ZX81 não tem ELSE nem vários comandos por linha: use GOTO.

  10 IF A>10 THEN PRINT "GRANDE"
  20 IF A$="S" THEN GOTO 100

Condições: = <> < > <= >=, com AND, OR e NOT. Verdadeiro vale 1, falso vale 0.`,
		see: []string{"basic THEN", "basic AND", "basic OR", "basic NOT", "ref operadores"}},
	{name: "THEN", usage: "IF condição THEN comando", summary: "parte do IF", page: `Separa a condição do comando no IF.`, see: []string{"basic IF"}},
	{name: "FOR", usage: "FOR v = início TO fim [STEP passo]", summary: "começa um laço",
		page: `A variável de uma letra v recebe o início; NEXT v soma o passo (1 se não houver STEP) e volta enquanto v não passar do fim. Se o laço não deve rodar nenhuma vez, pula para depois do NEXT.

  10 FOR I=1 TO 10 STEP 2
  20 PRINT I
  30 NEXT I`,
		see: []string{"basic NEXT", "basic TO", "basic STEP"}},
	{name: "TO", usage: "FOR v = a TO b\nA$(início TO fim)", summary: "fim do FOR; fatia de um texto",
		page: `No FOR, o valor final. Num texto, uma fatia: A$(2 TO 4) são o 2o, 3o e 4o caracteres; A$( TO 3) os três primeiros; A$(5 TO ) do 5o ao fim.`,
		see:  []string{"basic FOR", "ref textos"}},
	{name: "STEP", usage: "FOR v = a TO b STEP passo", summary: "passo do FOR", page: `O valor somado a cada NEXT; pode ser negativo (laço para trás) ou fracionário.`,
		see: []string{"basic FOR"}},
	{name: "NEXT", usage: "NEXT v", summary: "fim de um laço FOR",
		page: `Soma o passo à variável v e volta para o comando depois do FOR, se ainda não passou do fim. Erro 1: NEXT sem FOR (a variável não é de controle de laço). Erro 2: variável não encontrada.`,
		see:  []string{"basic FOR"}},
	{name: "LET", usage: "LET variável = expressão", summary: "dá valor a uma variável",
		page: `Cria ou muda uma variável. No ZX81 o LET é obrigatório (é a tecla L no cursor K).

Nomes: numéricas começam com letra e podem ter letras e números (LET TOTAL1=0); textos são uma letra e $ (LET A$="OI"); variáveis de controle de FOR são uma letra; matrizes vêm do DIM.

  10 LET X=X+1
  20 LET N$="ZX81"`,
		see: []string{"basic DIM", "ref variaveis"}},
	{name: "DIM", usage: "DIM letra(n1[,n2...])\nDIM letra$(n1[,n2...],tamanho)", summary: "cria uma matriz",
		page: `Cria (ou recria, apagando a antiga) uma matriz de números (zeros) ou de textos de tamanho fixo (espaços). Os índices começam em 1. O nome é uma letra.

  10 DIM A(10)
  20 DIM T$(5,8)
  30 LET A(3)=7

Erro 3: índice fora da matriz. Erro 4: sem memória.`,
		see: []string{"basic LET"}},
	{name: "INPUT", usage: "INPUT variável", summary: "lê um valor digitado",
		page: `Para o programa e espera o usuário digitar um valor (número, ou texto entre aspas já postas para variáveis de texto) e ENTER. STOP como resposta para o programa (erro D). Não pode ser usado como comando direto (erro 8).

  10 INPUT N
  20 INPUT N$`,
		see: []string{"basic INKEY$"}},
	{name: "REM", usage: "REM texto", summary: "comentário",
		page: `Tudo depois de REM é ignorado. No ZX81, uma linha REM no começo do programa é o lugar de sempre para código de máquina: o primeiro byte depois do REM da linha 1 fica em 16514 (4082h).

No Go81: rem 1 100 cria essa linha com 100 bytes.`,
		see: []string{"basic USR", "shell rem"}},
	{name: "STOP", usage: "STOP", summary: "para o programa (relatório 9)",
		page: `Para o programa com o relatório 9; CONT continua da linha seguinte.`, see: []string{"basic CONT", "erro 9"}},
	{name: "CONT", aliases: []string{"CONTINUE"}, usage: "CONT", summary: "continua depois de um STOP ou BREAK",
		page: `Continua o programa de onde ele parou (STOP, BREAK, tela cheia). Se a tela encheu (5), continua com a tela limpa.`,
		see:  []string{"basic STOP", "sysvar OLDPPC"}},
	{name: "NEW", usage: "NEW", summary: "apaga o programa e as variáveis",
		page: `Apaga tudo até RAMTOP, como ligar a máquina de novo (a memória acima de RAMTOP fica: é ali que se guarda código de máquina que deve sobreviver).`,
		see:  []string{"sysvar RAMTOP", "shell reset"}},
	{name: "CLEAR", usage: "CLEAR", summary: "apaga as variáveis",
		page: `Apaga todas as variáveis e libera a memória delas. RUN faz CLEAR antes de começar.`, see: []string{"basic RUN"}},
	{name: "CLS", usage: "CLS", summary: "limpa a tela",
		page: `Limpa a tela e volta a posição de impressão para o canto de cima. Com 4K ou mais o arquivo de tela é expandido para 24 linhas cheias (com pouca memória ele é "comprimido").`,
		see:  []string{"basic SCROLL", "sysvar D_FILE", "ref tela"}},
	{name: "SCROLL", usage: "SCROLL", summary: "sobe a tela uma linha",
		page: `Sobe toda a tela uma linha e põe a posição de impressão no começo da última. Útil para não parar com o erro 5.

  10 SCROLL
  20 PRINT RND
  30 GOTO 10`,
		see: []string{"basic CLS", "erro 5"}},
	{name: "PLOT", usage: "PLOT x,y", summary: "acende um ponto (64×44)",
		page: `Acende o "pixel" (um quarto de caractere) na coluna x (0-63) e linha y (0-43, de baixo para cima). Usa os caracteres gráficos de blocos.

  10 FOR X=0 TO 63
  20 PLOT X,22
  30 NEXT X

Erro B: coordenada fora da tela.`,
		see: []string{"basic UNPLOT"}},
	{name: "UNPLOT", usage: "UNPLOT x,y", summary: "apaga um ponto", page: `Apaga o ponto x,y (como PLOT).`, see: []string{"basic PLOT"}},
	{name: "PAUSE", usage: "PAUSE n", summary: "espera n quadros ou uma tecla",
		page: `Espera n quadros de TV (50 por segundo em PAL, 60 em NTSC) ou até uma tecla ser apertada. PAUSE 32768 ou mais espera para sempre (só a tecla). No modo FAST a tela aparece durante a pausa.

Dica do ZX81: depois de PAUSE, faça POKE 16437,255 (há um defeito na ROM que pode estragar o programa).`,
		see: []string{"basic INKEY$", "sysvar FRAMES"}},
	{name: "POKE", usage: "POKE endereço, valor", summary: "escreve um byte na memória",
		page: `Escreve o valor (0-255) no endereço (0-65535). A ROM (0-8191) não muda. Erro B: valor ou endereço fora da faixa.

  POKE 16418,0      (DF_SZ = 0: usa as 24 linhas da tela)`,
		see: []string{"basic PEEK", "shell poke", "sysvar"}},
	{name: "RAND", aliases: []string{"RANDOMIZE", "RANDOMISE"}, usage: "RAND [n]", summary: "começa os números aleatórios",
		page: `Muda a semente de RND (variável SEED). RAND 0 ou RAND sozinho usam o contador de quadros (FRAMES), dando uma sequência imprevisível; RAND n repete sempre a mesma sequência.

RAND USR endereço é o jeito comum de chamar código de máquina.`,
		see: []string{"basic RND", "basic USR", "sysvar SEED"}},
	{name: "FAST", usage: "FAST", summary: "modo rápido: sem tela durante o cálculo",
		page: `Desliga o gerador de NMI: o processador não gasta tempo desenhando a tela, que só aparece quando o programa espera uma tecla (INPUT, PAUSE, fim do programa). Cerca de 4 vezes mais rápido que SLOW.`,
		see:  []string{"basic SLOW", "sysvar CDFLAG"}},
	{name: "SLOW", usage: "SLOW", summary: "modo de computação e tela",
		page: `Liga o gerador de NMI: a tela fica sempre visível, mesmo com o programa rodando. É o modo do ZX81 ao ligar (o ZX80 só funciona como no FAST).`,
		see:  []string{"basic FAST"}},
	{name: "SAVE", usage: "SAVE \"nome\"", summary: "grava o programa na fita",
		page: `Grava o programa e as variáveis (de 16393 até o começo da área de edição) com o nome dado. O nome não pode ser vazio (erro F). SAVE numa linha do programa e GOTO depois fazem um programa que roda sozinho ao carregar.

No Go81 o SAVE abre a janela para escolher o arquivo .p; no shell: save arquivo.p NOME.`,
		see: []string{"basic LOAD", "shell save", "erro F"}},
	{name: "LOAD", usage: "LOAD \"nome\"\nLOAD \"\"", summary: "carrega um programa da fita",
		page: `Procura na fita o programa com esse nome (LOAD "" carrega o primeiro) e o carrega no lugar do atual. BREAK interrompe.

No Go81: load arquivo.p digita LOAD "" sozinho e carrega o arquivo pela fita.`,
		see: []string{"basic SAVE", "shell load"}},
	{name: "COPY", usage: "COPY", summary: "copia a tela para a impressora", page: `Imprime a tela inteira na ZX Printer.`,
		see: []string{"basic LPRINT", "basic LLIST"}},
	// --- functions ---
	{name: "ABS", usage: "ABS número", summary: "valor absoluto", page: `ABS -3.5 é 3.5.`},
	{name: "SGN", usage: "SGN número", summary: "sinal: -1, 0 ou 1", page: `SGN -7 é -1; SGN 0 é 0; SGN 2 é 1.`},
	{name: "INT", usage: "INT número", summary: "parte inteira (arredonda para baixo)",
		page: `INT 3.7 é 3; INT -3.7 é -4 (arredonda para baixo, não para zero).

  LET D=INT (RND*6)+1    (um dado)`, see: []string{"basic RND"}},
	{name: "SQR", usage: "SQR número", summary: "raiz quadrada", page: `SQR 16 é 4. Erro A: número negativo.`},
	{name: "EXP", usage: "EXP número", summary: "e elevado ao número", page: `EXP 1 é 2.7182818.`, see: []string{"basic LN"}},
	{name: "LN", usage: "LN número", summary: "logaritmo natural", page: `LN 2.7182818 é 1. Erro A: número zero ou negativo.`, see: []string{"basic EXP"}},
	{name: "SIN", usage: "SIN ângulo", summary: "seno (em radianos)", page: `Ângulos em radianos: graus*PI/180.`, see: []string{"basic COS", "basic TAN", "basic PI"}},
	{name: "COS", usage: "COS ângulo", summary: "cosseno (em radianos)", page: `Ângulos em radianos.`, see: []string{"basic SIN"}},
	{name: "TAN", usage: "TAN ângulo", summary: "tangente (em radianos)", page: `Ângulos em radianos.`, see: []string{"basic SIN"}},
	{name: "ASN", usage: "ASN número", summary: "arco seno", page: `O ângulo (radianos) cujo seno é o número (-1 a 1).`},
	{name: "ACS", usage: "ACS número", summary: "arco cosseno", page: `O ângulo (radianos) cujo cosseno é o número (-1 a 1).`},
	{name: "ATN", usage: "ATN número", summary: "arco tangente", page: `O ângulo (radianos) cuja tangente é o número. 4*ATN 1 é PI.`},
	{name: "PI", usage: "PI", summary: "3.1415927", page: `O número pi.`},
	{name: "RND", usage: "RND", summary: "número aleatório entre 0 e 1",
		page: `Um número pseudo-aleatório de 0 até menos de 1. INT (RND*N) dá um inteiro de 0 a N-1. A sequência depende de SEED (RAND).`,
		see:  []string{"basic RAND", "basic INT"}},
	{name: "PEEK", usage: "PEEK endereço", summary: "lê um byte da memória",
		page: `O byte (0-255) no endereço. Endereços de 2 bytes: PEEK a+256*PEEK (a+1).

  PRINT PEEK 16396+256*PEEK 16397    (o endereço do arquivo de tela, D_FILE)`,
		see: []string{"basic POKE", "shell peek", "sysvar"}},
	{name: "USR", usage: "USR endereço", summary: "chama código de máquina",
		page: `Executa a rotina de código de máquina no endereço e devolve o valor do par BC no RET. Usos comuns: RAND USR 16514 (descarta o valor), PRINT USR 16514.`,
		see:  []string{"basic REM", "shell rem", "ref memoria"}},
	{name: "LEN", usage: "LEN texto", summary: "tamanho de um texto", page: `LEN "ZX81" é 4.`},
	{name: "CODE", usage: "CODE texto", summary: "código do primeiro caractere",
		page: `O código ZX81 (0-255, não ASCII) do primeiro caractere; 0 para texto vazio. CODE "A" é 38.`, see: []string{"basic CHR$", "ref caracteres"}},
	{name: "CHR$", usage: "CHR$ código", summary: "o caractere de um código", page: `CHR$ 38 é "A". Os códigos 64-127 e 192-255 dão palavras-chave ou "?".`,
		see: []string{"basic CODE", "ref caracteres"}},
	{name: "VAL", usage: "VAL texto", summary: "calcula uma expressão escrita num texto",
		page: `Calcula o texto como expressão numérica: VAL "2*3" é 6. Erro C: o texto não é uma expressão válida.`, see: []string{"basic STR$"}},
	{name: "STR$", usage: "STR$ número", summary: "o número como texto", page: `STR$ 12.5 é "12.5" (como PRINT o escreveria).`, see: []string{"basic VAL"}},
	{name: "INKEY$", usage: "INKEY$", summary: "a tecla apertada agora",
		page: `O caractere da tecla apertada no momento, ou "" se nenhuma. Não espera: use num laço. No modo FAST, a tela some a cada leitura.

  10 IF INKEY$="" THEN GOTO 10
  20 PRINT INKEY$`,
		see: []string{"basic INPUT", "basic PAUSE"}},
	{name: "AT", usage: "PRINT AT linha,coluna; ...", summary: "posição de impressão", page: `Linha 0-21 e coluna 0-31. Erro B: fora da tela.`, see: []string{"basic PRINT"}},
	{name: "TAB", usage: "PRINT TAB coluna; ...", summary: "vai até uma coluna", page: `Avança até a coluna (0-31, módulo 32), pulando de linha se preciso.`, see: []string{"basic PRINT"}},
	{name: "NOT", usage: "NOT valor", summary: "negação lógica", page: `NOT 0 é 1; NOT de qualquer outro número é 0.`, see: []string{"ref operadores"}},
	{name: "AND", usage: "a AND b", summary: "e lógico",
		page: `Com números: a AND b é a se b for verdadeiro (diferente de 0), senão 0. Com texto: A$ AND b é A$ se b for verdadeiro, senão "".`, see: []string{"ref operadores"}},
	{name: "OR", usage: "a OR b", summary: "ou lógico", page: `a OR b é 1 se b for verdadeiro, senão a.`, see: []string{"ref operadores"}},
}

// keywordCode finds the token of a keyword.
func keywordCode(name string, tokens map[byte]string) (byte, bool) {
	for c, n := range tokens {
		if n == name {
			return c, true
		}
	}
	return 0, false
}

func init() {
	providers = append(providers, func() {
		for _, p := range basic81Pages {
			t := &Topic{Name: p.name, Aliases: p.aliases, Section: "basic", Usage: p.usage, Summary: p.summary, Page: p.page, SeeAlso: p.see}
			if c, ok := keywordCode(p.name, basic.Tokens); ok {
				if k := zx81KeyMap[c]; k != "" {
					t.Page = strings.TrimRight(t.Page, "\n") + "\n\nTECLA: " + k + "."
				}
			}
			add(t)
		}
	})
}
