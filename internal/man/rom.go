// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

// The ZX81 ROM (edition 3): system variables, error reports and every label
// of EightyOne's symbol file (zx81.edition3.sym, GPL, with the names of Ian
// Logan's "The Complete ZX81 ROM Disassembly"). The descriptions are Go81's.

//go:embed zx81.edition3.sym
var zx81Sym string

type sysvarInfo struct {
	size int
	desc string
}

var sysvars = map[string]sysvarInfo{
	"ERR_NR":  {1, "Código do relatório menos 1; FFh (-1) = nenhum erro. Um RST 08 o põe aqui."},
	"FLAGS":   {1, "Flags do BASIC: bit 1 = impressora, bit 6 = resultado numérico (1) ou texto (0), bit 7 = verificando a sintaxe (0) ou executando (1)."},
	"ERR_SP":  {2, "Endereço do item da pilha da máquina usado como retorno quando há erro."},
	"RAMTOP":  {2, "Primeiro byte depois da área do BASIC. NEW apaga só até aqui: abaixe RAMTOP (POKE 16388/16389 e NEW) para guardar código de máquina no alto da memória."},
	"MODE":    {1, "Cursor especial: K, L, F (função) ou G (gráficos)."},
	"PPC":     {2, "Número da linha sendo executada."},
	"VERSN":   {1, "0: identifica o ZX81. É o primeiro byte de um arquivo .p (o SAVE grava de VERSN até E_LINE)."},
	"E_PPC":   {2, "Número da linha atual (a do cursor de programa >)."},
	"D_FILE":  {2, "Endereço do arquivo de tela: 25 NEWLINEs (76h) com até 32 caracteres entre eles. Vem logo depois do programa."},
	"DF_CC":   {2, "Endereço da posição de impressão (PRINT) no arquivo de tela."},
	"VARS":    {2, "Endereço das variáveis."},
	"DEST":    {2, "Endereço da variável que está recebendo um valor."},
	"E_LINE":  {2, "Endereço da linha em edição; o fim do que o SAVE grava."},
	"CH_ADD":  {2, "Endereço do próximo caractere a ser interpretado."},
	"X_PTR":   {2, "Endereço do caractere antes do marcador de erro de sintaxe (S)."},
	"STKBOT":  {2, "Base da pilha do calculador."},
	"STKEND":  {2, "Topo da pilha do calculador (e começo da memória livre)."},
	"BERG":    {1, "Registrador b do calculador."},
	"MEM":     {2, "Endereço da área de memória do calculador (normalmente MEMBOT)."},
	"SPARE1":  {1, "Não usado."},
	"DF_SZ":   {1, "Número de linhas da parte de baixo da tela (normalmente 2). POKE 16418,0 deixa PRINT usar as 24 linhas."},
	"S_TOP":   {2, "Número da linha no topo da listagem automática."},
	"LAST_K":  {2, "Mostra que teclas estão apertadas."},
	"DBOUNCE": {1, "Estado do antirrepique do teclado."},
	"MARGIN":  {1, "Número de linhas de borda em cima e embaixo da imagem: 55 (PAL) ou 31 (NTSC)."},
	"NXTLIN":  {2, "Endereço da próxima linha do programa a executar. O LOAD continua daqui: um .p com NXTLIN numa linha roda sozinho."},
	"OLDPPC":  {2, "Linha para onde CONT volta."},
	"FLAGX":   {1, "Flags diversos (bit 5 = INPUT, bit 1 = variável de texto...)."},
	"STRLEN":  {2, "Tamanho do texto numa atribuição."},
	"T_ADDR":  {2, "Endereço do próximo item da tabela de sintaxe."},
	"SEED":    {2, "Semente do RND; RAND a muda."},
	"FRAMES":  {2, "Conta os quadros de TV para trás (bit 15 = 1 normalmente). PAUSE usa este contador; 65536-FRAMES dá o tempo desde a última vez que ele foi zerado."},
	"COORDS":  {2, "Coordenadas x (primeiro byte) e y do último ponto do PLOT."},
	"PR_CC":   {1, "Byte menos significativo da posição de LPRINT em PRBUFF."},
	"S_POSN":  {2, "Posição do PRINT: coluna (33 - coluna) e linha (24 - linha)."},
	"CDFLAG":  {1, "Bit 7 = 1 no modo SLOW (computação e tela), 0 no FAST. Bit 6 = SLOW de fato ativo."},
	"PRBUFF":  {33, "Buffer da impressora: 32 caracteres e um NEWLINE."},
	"MEMBOT":  {30, "Memória do calculador: 6 números de 5 bytes."},
	"SPARE2":  {2, "Não usado."},
}

// romDescs describe the most useful ROM entry points.
var romDescs = map[string]string{
	"START":     "Início ao ligar (RST 00): desliga as interrupções, zera o A e vai para o teste de memória (RAM-CHECK).",
	"ERROR-1":   "RST 08: relata um erro. O byte depois do RST é o código do relatório menos 1; vai para ERR_NR e o programa para.",
	"PRINT-A":   "RST 10: imprime o caractere (código ZX81) do registrador A na posição atual (tela ou impressora).",
	"GET-CH":    "RST 18: devolve em A o caractere apontado por CH_ADD, pulando espaços.",
	"NEXT-CH":   "RST 20: avança CH_ADD e devolve o próximo caractere.",
	"FP-CALC":   "RST 28: entra no calculador de ponto flutuante; os bytes seguintes são literais (operações) até END-CALC (38h).",
	"END-CALC":  "Literal 38h do calculador: volta ao código Z80.",
	"BC-SPACES": "RST 30: reserva BC bytes no espaço de trabalho; devolve o endereço em DE.",
	"INTERRUPT": "RST 38 / interrupção (modo IM 1): conta as linhas de um caractere durante a geração da imagem.",
	"NMI":       "Rotina do NMI: no modo SLOW, conta as linhas de borda entre as imagens e devolve o controle ao programa.",
	"K-UNSHIFT": "Tabela do teclado: o código de cada tecla sem SHIFT (40 bytes na ordem da matriz).",
	"K-SHIFT":   "Tabela do teclado com SHIFT (sinais e algumas palavras-chave).",
	"K-FUNCT":   "Tabela das funções (cursor F, depois de SHIFT+ENTER).",
	"K-GRAPH":   "Tabela dos gráficos (cursor G).",
	"TOKENS":    "Tabela das palavras-chave (cada uma com o último caractere em vídeo inverso). Também o nome da rotina que imprime uma palavra-chave.",
	"LOAD/SAVE": "Início das rotinas de fita.",
	"SLOW/FAST": "Ajusta o modo (CDFLAG) e liga ou desliga o gerador de NMI.",
	"DISPLAY-1": "Começo da rotina de tela do modo SLOW (o laço de geração da imagem).",
	"KEYBOARD":  "Lê o teclado: devolve em HL o estado das 8 meias-filas (nome de Toni Baker: KSCAN).",
	"SET-FAST":  "Passa para o modo FAST.",
	"SAVE":      "Comando SAVE: grava de VERSN (4009h) até E_LINE.",
	"LOAD":      "Comando LOAD.",
	"NEW":       "Comando NEW.",
	"RAM-CHECK": "Teste da memória ao ligar: acha o fim da RAM e põe em RAMTOP.",
	"INITIAL":   "Inicialização das variáveis de sistema depois do teste de memória.",
	"DECODE":    "Converte o valor do teclado (de KEYBOARD) no código do caractere (nome de Toni Baker: FINDCHAR).",
	"PRINT-AT":  "Muda a posição de impressão: B = linha, C = coluna (como PRINT AT).",
	"LINE-ADDR": "Acha o endereço da linha de número HL no programa.",
	"CLS":       "Limpa a tela (e a expande para 24 linhas se houver memória).",
	"MAKE-ROOM": "Abre BC bytes de espaço na memória, ajustando os ponteiros.",
	"RECLAIM-1": "Devolve memória (apaga um trecho), ajustando os ponteiros.",
	"PLOT/UNP":  "PLOT e UNPLOT.",
	"STK-TO-BC": "Tira dois números da pilha do calculador para B e C.",
	"STK-TO-A":  "Tira um número da pilha do calculador para A.",
	"SCROLL":    "Comando SCROLL.",
	"LINE-SCAN": "Interpreta uma linha (verificação de sintaxe ou execução).",
	"FIND-INT":  "Tira um inteiro (0-65535) da pilha do calculador para BC.",
	"FAST":      "Comando FAST.",
	"SLOW":      "Comando SLOW.",
	"PAUSE":     "Comando PAUSE.",
	"BREAK-1":   "Testa a tecla BREAK (SPACE): carry desligado se apertada.",
	"SCANNING":  "Calcula uma expressão do BASIC e deixa o resultado na pilha do calculador.",
	"LOOK-VARS": "Procura uma variável na área de variáveis.",
	"LET":       "Comando LET (atribuição).",
	"DEC-TO-FP": "Converte um número escrito em decimal para ponto flutuante (5 bytes). O Go81 usa esta rotina, rodando no Z80, para gerar os números do .p.",
	"STACK-A":   "Põe o valor de A na pilha do calculador.",
	"STACK-BC":  "Põe o valor de BC na pilha do calculador.",
	"FP-TO-BC":  "Tira um número da pilha do calculador para BC (arredondado).",
	"FP-TO-A":   "Tira um número da pilha do calculador para A.",
	"PRINT-FP":  "Imprime o número do topo da pilha do calculador.",
	"ADDITION":  "Soma de ponto flutuante.",
	"MULTIPLY":  "Multiplicação de ponto flutuante.",
	"DIVISION":  "Divisão de ponto flutuante.",
	"CALCULATE": "O calculador: executa os literais depois de RST 28.",
	"PEEK":      "Função PEEK.",
	"USR":       "Função USR: chama o código de máquina; o BC no RET é o resultado.",
	"CHAR-SET":  "O conjunto de caracteres: 64 caracteres de 8 bytes (1E00h-1FFFh). O registrador I vale 1Eh para a tela usar esta tabela.",
}

func init() {
	providers = append(providers, func() {
		for _, line := range strings.Split(zx81Sym, "\n") {
			f := strings.Fields(strings.TrimSpace(line))
			if len(f) != 2 {
				continue
			}
			addr, err := strconv.ParseUint(f[0], 16, 16)
			if err != nil {
				continue
			}
			name := f[1]
			hex := fmt.Sprintf("%04Xh (%d)", addr, addr)
			if sv, ok := sysvars[name]; ok {
				add(&Topic{Name: name, Section: "sysvar",
					Summary: fmt.Sprintf("variável de sistema em %d (%04Xh), %d byte(s)", addr, addr, sv.size),
					Page: fmt.Sprintf("Endereço: %s. Tamanho: %d byte(s).\n\n%s\n\nLer no BASIC: PEEK %d (2 bytes: PEEK %d+256*PEEK %d). No shell: peek %d %d.",
						hex, sv.size, sv.desc, addr, addr, addr+1, addr, sv.size),
					SeeAlso: []string{"ref memoria", "man -l sysvar"}})
				continue
			}
			if addr >= 0x4000 {
				continue
			}
			t := &Topic{Name: name, Section: "rom", Page: "Endereço na ROM: " + hex + "."}
			if d, ok := romDescs[name]; ok {
				t.Summary = fmt.Sprintf("%04Xh: %s", addr, firstSentence(d))
				t.Page += "\n\n" + d
			} else {
				t.Summary = fmt.Sprintf("%04Xh", addr)
			}
			add(t)
		}
		addErrors()
	})
}

func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return strings.TrimSuffix(s[:i], ":")
	}
	return strings.TrimSuffix(s, ".")
}

// The reports, shown as "código/linha" at the bottom of the screen.
var errors81 = []struct{ code, desc, more string }{
	{"0", "terminou bem", "O programa chegou ao fim, ou um comando direto foi executado."},
	{"1", "NEXT sem FOR", "A variável do NEXT não é de controle de um FOR (há uma variável comum com esse nome)."},
	{"2", "variável não definida", "Usou uma variável antes do LET, do INPUT ou do DIM; ou uma variável de controle sem o FOR."},
	{"3", "índice fora da matriz", "Um índice maior que o tamanho do DIM, ou fatia de texto fora do texto."},
	{"4", "sem memória", "A memória acabou (programa, variáveis, GOSUB demais...)."},
	{"5", "tela cheia", "Não há mais lugar na tela; CONT continua com a tela limpa. SCROLL evita o problema."},
	{"6", "estouro aritmético", "Um resultado maior que cerca de 1E38."},
	{"7", "RETURN sem GOSUB", "Um RETURN a mais."},
	{"8", "INPUT como comando", "INPUT só pode ser usado dentro de um programa."},
	{"9", "STOP", "O comando STOP. CONT continua na linha seguinte."},
	{"A", "argumento inválido", "Por exemplo SQR de número negativo ou LN de zero."},
	{"B", "inteiro fora da faixa", "Um número inteiro fora do permitido: POKE, PLOT, AT, CHR$, endereço..."},
	{"C", "VAL inválido", "O texto do VAL não é uma expressão numérica válida."},
	{"D", "BREAK", "A tecla BREAK interrompeu o programa (ou STOP como resposta a um INPUT). CONT continua."},
	{"E", "não usado", "Este código não é usado pelo ZX81."},
	{"F", "nome vazio", "SAVE com o nome vazio."},
}

func addErrors() {
	for _, e := range errors81 {
		add(&Topic{Name: e.code, Aliases: []string{"erro " + e.code, "relatorio " + e.code}, Section: "erro",
			Summary: "ZX81: " + e.desc,
			Page:    e.more + "\n\nO ZX81 mostra o relatório como código/linha embaixo da tela, ex.: " + e.code + "/20 (o código " + e.code + " na linha 20).",
			SeeAlso: []string{"sysvar ERR_NR", "rom ERROR-1"}})
	}
	add(&Topic{Name: "zx80", Aliases: []string{"erros do zx80"}, Section: "erro", Summary: "os relatórios do ZX80",
		Page: `O ZX80 usa os códigos 0 a 9 com os mesmos sentidos do ZX81 (0 terminou, 1 NEXT sem FOR, 2 variável não encontrada, 3 índice, 4 sem memória, 5 tela cheia, 6 estouro, 7 RETURN sem GO SUB, 8 INPUT como comando, 9 STOP), mostrados como código/linha. Ele não tem os códigos de letra do ZX81.`})
}
