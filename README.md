# Go81

> Um ambiente completo de desenvolvimento e emulação para o **Sinclair ZX81** e seus clones, escrito em **Go**.

**Versão atual:** `1.1.2` — *Sinclair · 3D Monster Maze · "Escopo e licença"* (somente documentação e esqueleto)

---

## O que é

O Go81 é um "mega emulador" do ZX81 que junta, em um único programa, tudo que um desenvolvedor
de ZX81 precisa:

| Ferramenta | O que faz |
|---|---|
| **Emulador** | ZX81, ZX80, Timex TS1000 e TS1500, Lambda 8300, Ringo R-470, TK85, ZX97 Lite e Jupiter Ace, com meta de 100% de compatibilidade com o EightyOne 1.41 |
| **Editor de BASIC** | Editor embutido com destaque de sintaxe, numeração automática e envio direto para o emulador |
| **Compilador BASIC (Dignified)** | Dialeto moderno, sem números de linha, com rótulos, nomes longos e `include`, que gera BASIC ZX81 tokenizado (`.p`) |
| **Editor de textos** | Editor geral para fontes, notas e scripts |
| **Montador Z80** | Assembler/disassembler com macros, rótulos e saída `.p`, `.bin`, `.hex` e REM-embutido |
| **Gerador de fitas** | Cria e converte `.p`, `.81`, `.p81`, `.tzx`, `.wav` |
| **Depurador** | Breakpoints, passo a passo, memória ao vivo, histórico e profiler |
| **Porta de comunicação (Link)** | Canal TCP/pipe para outros programas controlarem o emulador e trocarem dados com o programa ZX81 em execução |

Tudo pode ser usado de três formas:

- **GUI** — janelas com [Fyne](https://fyne.io)
- **TUI** — interface de terminal com [tview](https://github.com/rivo/tview) / [tcell](https://github.com/gdamore/tcell)
- **CLI** — um pequeno "sistema operacional" de linha de comando, além de muitos parâmetros para automação pelo shell do Windows e do Linux

## Estado do projeto

O projeto está na fase de **estudo e especificação**. Nada é usável ainda. Veja:

- [SPEC.md](SPEC.md) — especificação completa
- [OUTLINE.md](OUTLINE.md) — documento vivo: o que existe, o que foi testado, o que falta
- [RELEASE.md](RELEASE.md) — histórico de versões
- [MANUAL.md](MANUAL.md) — manual do usuário

## Compilando

Requer **Go 1.23 ou mais novo** (desenvolvido com Go 1.27).

```sh
go build -o bin/ ./cmd/go81
./bin/go81 --version
```

A GUI (Fyne) precisará de um compilador C (CGO) quando for adicionada: no Windows, MSYS2/MinGW-w64 ou TDM-GCC;
no Linux, `gcc` e os pacotes de desenvolvimento de OpenGL/X11.

## Ferramentas usadas no desenvolvimento

| Ferramenta | Uso |
|---|---|
| [Go](https://go.dev) | Linguagem do projeto |
| [Fyne](https://fyne.io) | Interface gráfica (GUI) multiplataforma |
| [tview](https://github.com/rivo/tview) | Componentes de interface de terminal (TUI) |
| [tcell](https://github.com/gdamore/tcell) | Acesso de baixo nível ao terminal |
| [Git](https://git-scm.com) / GitHub | Controle de versão |
| [Claude Code](https://claude.com/claude-code) | Assistente de desenvolvimento |

## Projetos estudados e créditos

O Go81 se apoia no trabalho de muitas pessoas. Agradecimentos a:

- **EightyOne** — emulador de referência (versão 1.41). Criado por **Michael D. Wynne**, mantido por
  **Paul Farrow** e **John Stroebel**, entre outros colaboradores. Licença GPL v2 ou posterior.
  Os fontes e executáveis de referência ficam em `81emul/`.
- **Núcleo Z80 do FUSE** — **Philip Kendall** e equipe do Free Unix Spectrum Emulator, base do núcleo Z80 do EightyOne.
- **ZXpand** — **Charlie Robson** (SirMorris), interface de SD card para ZX80/ZX81.
- **FatFs** — **ChaN**, sistema de arquivos FAT usado na emulação do ZXpand.
- **Chroma 81 / Spectra** — **Paul Farrow**.
- **WRX Hi-Res** — **Wilf Rigter**.
- **SP0256** — código de fala de [GmEsoft/SP0256_CTS256A-AL2](https://github.com/GmEsoft/SP0256_CTS256A-AL2).
- **Basic Dignified Suite** — **Fred Rique (farique1)**, inspiração para o dialeto BASIC moderno.
- **Sinclair Research**, **Sir Clive Sinclair**, **Richard Altwasser** e **Steve Vickers** — pelo ZX81, seu hardware e sua ROM.
- **Timex Sinclair**, **Lambda Electronics**, **Microdigital (TK85)** e **Ringo** — pelos clones.
- **Jupiter Cantab**, **Richard Altwasser** e **Steve Vickers** — pelo Jupiter Ace.
- ROMs e documentação de periféricos distribuídos com o EightyOne pertencem a seus respectivos autores.

## Licença

Copyright (C) 2026 Wilson "Barney" Pilon.

O Go81 é software livre sob a **GNU General Public License v3 ou posterior** — veja [LICENSE](LICENSE).
Inclui código portado do EightyOne (GPL v2 ou posterior), com os créditos originais mantidos.
Detalhes em [SPEC.md §13](SPEC.md#13-licenciamento).
