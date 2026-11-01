// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

import (
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

// The 4K BASIC of the ZX80. Only what differs from the ZX81 is explained
// at length; each page points to the ZX81 one.

const zx80Note = `O BASIC do ZX80 só tem números inteiros de -32768 a 32767 (a divisão / descarta a fração) e não tem PLOT, UNPLOT, PAUSE, SCROLL, FAST, SLOW, LPRINT, LLIST nem COPY. A tela some enquanto um programa roda e enquanto uma tecla está apertada.`

var basic80Pages = []kwPage{
	{name: "PRINT", usage: "PRINT item [; | ,] ...", summary: "escreve na tela", page: `Como no ZX81, mas só com números inteiros e sem AT nem TAB.`, see: []string{"basic PRINT"}},
	{name: "LIST", usage: "LIST [linha]", summary: "lista o programa", page: `Lista o programa a partir da linha. No Go81: list (shell) lista o programa do ZX80 como texto.`, see: []string{"shell list"}},
	{name: "RUN", usage: "RUN [linha]", summary: "apaga as variáveis e roda", page: `Como no ZX81.`, see: []string{"basic RUN"}},
	{name: "GO TO", aliases: []string{"GOTO"}, usage: "GO TO expressão", summary: "desvia para uma linha", page: `Como o GOTO do ZX81.`, see: []string{"basic GOTO"}},
	{name: "GO SUB", aliases: []string{"GOSUB"}, usage: "GO SUB expressão", summary: "chama uma sub-rotina", page: `Como o GOSUB do ZX81.`, see: []string{"basic80 RETURN"}},
	{name: "RETURN", usage: "RETURN", summary: "volta de uma sub-rotina", page: `Erro 7: RETURN sem GO SUB.`, see: []string{"basic80 GO SUB"}},
	{name: "IF", usage: "IF condição THEN comando", summary: "comando condicional",
		page: `Como no ZX81. AND, OR e NOT trabalham bit a bit nos números de 16 bits; verdadeiro vale -1 (todos os bits 1).`, see: []string{"basic IF"}},
	{name: "THEN", usage: "IF condição THEN comando", summary: "parte do IF", page: `Separa a condição do comando.`},
	{name: "FOR", usage: "FOR v = início TO fim", summary: "começa um laço",
		page: `A variável de uma letra vai do início ao fim, de 1 em 1 (o ZX80 não tem STEP).`, see: []string{"basic80 NEXT"}},
	{name: "TO", usage: "FOR v = a TO b", summary: "fim do FOR", page: `O valor final do laço.`},
	{name: "NEXT", usage: "NEXT v", summary: "fim de um laço", page: `Erro 1: NEXT sem FOR.`, see: []string{"basic80 FOR"}},
	{name: "LET", usage: "LET variável = expressão", summary: "dá valor a uma variável",
		page: `Variáveis numéricas com nome de letras e números; textos com uma letra e $.`, see: []string{"basic LET"}},
	{name: "DIM", usage: "DIM letra(n)", summary: "cria uma matriz de números", page: `Uma dimensão só, índices de 0 a n.`},
	{name: "INPUT", usage: "INPUT variável", summary: "lê um valor digitado", page: `Como no ZX81. Erro 8: INPUT como comando direto.`},
	{name: "REM", usage: "REM texto", summary: "comentário", page: `Ignorado na execução.`},
	{name: "STOP", usage: "STOP", summary: "para o programa (relatório 9)", page: `CONTINUE continua.`},
	{name: "CONTINUE", aliases: []string{"CONT"}, usage: "CONTINUE", summary: "continua depois de STOP", page: `Como o CONT do ZX81.`},
	{name: "NEW", usage: "NEW", summary: "apaga o programa", page: `Recomeça com a memória vazia.`},
	{name: "CLEAR", usage: "CLEAR", summary: "apaga as variáveis", page: `Como no ZX81.`},
	{name: "CLS", usage: "CLS", summary: "limpa a tela", page: `Como no ZX81.`},
	{name: "POKE", usage: "POKE endereço, valor", summary: "escreve um byte", page: `Endereços acima de 32767 são escritos como números negativos (endereço-65536).`},
	{name: "RANDOMISE", aliases: []string{"RANDOMIZE", "RAND"}, usage: "RANDOMISE [n]", summary: "começa os números aleatórios", page: `Muda a semente de RND.`},
	{name: "SAVE", usage: "SAVE", summary: "grava o programa na fita (sem nome)",
		page: `O ZX80 grava sem nome tudo de 4000h até E_LINE: o arquivo .o. No Go81: save arquivo.o.`, see: []string{"shell save"}},
	{name: "LOAD", usage: "LOAD", summary: "carrega o programa da fita",
		page: `Carrega o próximo programa da fita. No Go81: load arquivo.o (digita LOAD e ENTER sozinho).`, see: []string{"shell load"}},
	// functions (typed letter by letter on the ZX80)
	{name: "PEEK", usage: "PEEK(endereço)", summary: "lê um byte", page: `As funções do ZX80 são digitadas letra por letra e levam parênteses.`},
	{name: "USR", usage: "USR(endereço)", summary: "chama código de máquina", page: `Devolve o par HL no RET (no ZX81 é BC).`},
	{name: "RND", usage: "RND(n)", summary: "número aleatório de 1 a n", page: `Diferente do ZX81: devolve um inteiro de 1 a n.`},
	{name: "ABS", usage: "ABS(n)", summary: "valor absoluto", page: `Inteiro.`},
	{name: "CHR$", usage: "CHR$(código)", summary: "caractere de um código", page: `O conjunto de caracteres do ZX80 é diferente do ZX81 (man ref caracteres80).`},
	{name: "CODE", usage: "CODE(texto)", summary: "código do primeiro caractere", page: `Código do ZX80.`},
	{name: "STR$", usage: "STR$(n)", summary: "o número como texto", page: `Inteiro.`},
	{name: "AND", usage: "a AND b", summary: "e bit a bit", page: `Trabalha nos 16 bits: 12 AND 10 é 8. Com comparações (verdadeiro = -1) funciona como o e lógico.`},
	{name: "OR", usage: "a OR b", summary: "ou bit a bit", page: `Trabalha nos 16 bits: 12 OR 10 é 14.`},
	{name: "NOT", usage: "NOT a", summary: "inverte os bits", page: `NOT 0 é -1; NOT -1 é 0.`},
	{name: "TL$", usage: "TL$(texto)", summary: "o texto sem o primeiro caractere", page: `Só no ZX80: TL$("ZX80") é "X80".`},
}

func init() {
	providers = append(providers, func() {
		for _, p := range basic80Pages {
			t := &Topic{Name: p.name, Aliases: p.aliases, Section: "basic80", Usage: p.usage, Summary: p.summary,
				Page: strings.TrimRight(p.page, "\n") + "\n\n" + zx80Note, SeeAlso: p.see}
			if c, ok := keywordCode(p.name, basic.ZX80Tokens); ok {
				if k := zx80KeyMap[c]; k != "" {
					t.Page += "\n\nTECLA: " + k + "."
				}
			}
			add(t)
		}
	})
}
