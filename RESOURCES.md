# Go81 — Recursos fora do GitHub

Este documento lista tudo o que o projeto usa e que **não vai para o GitHub**: material de referência, programas
de terceiros, configurações locais e saídas de compilação. Serve para **remontar o projeto em outro PC** e para
saber **o que levar no backup / pendrive**.

Levantamento de 2026-10-31. Caminho do projeto neste PC: `C:\dos\go81`.

---

## 1. Montar o projeto num PC novo

1. Instalar as ferramentas (seção 4).
2. Clonar o repositório — de preferência no **mesmo caminho** (`C:\dos\go81`), por causa da memória do Claude Code
   (seção 3):
   ```
   git clone https://github.com/wilsonpilon/go81 C:\dos\go81
   ```
3. Copiar do backup as pastas `81emul\` e `resource\` para dentro de `C:\dos\go81\` (seção 2). Elas já estão no
   `.gitignore`.
4. Refazer as configurações locais (seção 3).
5. Conferir:
   ```
   go build -o bin/ ./cmd/go81
   go test ./...
   go run ./test/compat/z80diff      (precisa de 81emul\ e gcc)
   go run ./test/compat/zx81diff     (precisa de 81emul\ e g++)
   ```

---

## 2. Pastas de referência (backup obrigatório ou baixar de novo)

Total: cerca de **950 MB** (`81emul` 309 MB + `resource` 642 MB).

Legenda da última coluna: **guardar** = difícil de achar de novo, manter no backup; **baixar** = dá para obter
de novo pelo endereço indicado (pode ficar fora do pendrive se faltar espaço).

### 2.1 `81emul\` — EightyOne 1.41 (309 MB, 1.159 arquivos) — **guardar**

| Conteúdo | Uso no Go81 |
|---|---|
| `source\Source\` — código C++Builder do EightyOne (Z80, ZX81, fitas, som, depurador…) | Referência da emulação; compilado pelos comparadores `z80diff` e `zx81diff` |
| `source\EightyOne Development Environment.docx`, `source\Installation Files\` | Documentação do ambiente original |
| `exe\` — `EightyOne.exe`, ROMs, documentação de periféricos, exemplos `.p` | ROMs (as do ZX81 já estão copiadas em `assets\roms\`), exemplos |

Origem: <https://sourceforge.net/projects/eightyone-sinclair-emulator/> (versão 1.41, 02/03/2025).
Os comparadores esperam exatamente `81emul\source\Source\` e `81emul\exe\ROMs\`.

### 2.2 `resource\` — material de estudo (642 MB)

| Pasta | Tamanho | Conteúdo | Origem | Backup |
|---|---|---|---|---|
| `asMSX\` | 2,4 MB | Montador asMSX (GPL v3), manual `doc\asmsx.md` — base da sintaxe do asm81 | <https://github.com/Fubukimaru/asMSX> | baixar |
| `pasmo\` | 1 MB | Montador Pasmo (Julián Albo, GPL v3) | <https://pasmo.speccy.org/> (o README cita o endereço antigo `http://www.arrakis.es/~ninsesabe/pasmo/`) | baixar |
| `pasmolib\` | 18 KB | Modelo de programa ZX81 em assembler para o Pasmo (variáveis de sistema, REM, ROM) | origem a confirmar | **guardar** |
| `81asm\` | 14 KB | O mesmo modelo de programa ZX81 em assembler (outra versão) | origem a confirmar | **guardar** |
| `sjasmplus\` | 11 MB | Montador SjASMPlus (BSD) | <https://github.com/z00m128/sjasmplus> | baixar |
| `z88dk\` | 345 MB | Kit de desenvolvimento z88dk (C e assembler, com suporte ao ZX81) | <https://github.com/z88dk/z88dk> — compilações em <http://nightly.z88dk.org/> | baixar |
| `zesarux\` | 120 MB | Emulador ZEsarUX (César Hernández Bañó, GPL v3): código, manual `src\zesarux.pdf` | <https://github.com/chernandezba/zesarux> | baixar |
| `badig\` | 2,7 MB | Basic Dignified Suite (Fred Rique), Python | <https://github.com/farique1/basic-dignified> | baixar |
| `zxtext2p\` | 148 KB | ZXText2P (Chris Cowley, 2002, GPL): fonte C, executáveis, `doc\zxtext2p.html` | site original `http://www.grok.co.uk/` (provavelmente fora do ar) | **guardar** |
| `VB81_RES\` | 8 MB | Código-fonte VB6 do VB81 XuR (Chris Cowley / XavSnap, GPL v2+) | <http://zx81.vb81.free.fr> (não abriu em 2026-10-31); fórum: Sinclair ZX World, tópico 3295 | **guardar** |
| `Vb81_XuR\` | 17 MB | Executável e arquivos do VB81 XuR — **não executar** (relatos de vírus) | mesmo site acima | **guardar** (ou omitir) |
| `superx\` | 296 KB | Super-X 1.22 (MSX): `SUPER-X.DOC` (manual em inglês), `SUPER-X.ASM`, `LOADER.ASM` | origem a confirmar (versão de NYYRIKKI, 2011; comunidade MSX) | **guardar** |
| `docs\` | 137 MB | Livros, manuais e revistas (lista abaixo) | variadas | **guardar** |

#### `resource\docs\` (48 arquivos)

| Arquivo(s) | Conteúdo | Uso |
|---|---|---|
| `The Complete zx81 Rom Disassembly.pdf`, `The Complete Timex TS1000 _ Sinclair ZX81 ROM Disassembly.pdf`, `Assembly Listing of the Operating System of the Sinclair ZX81.pdf` | Desmontagens comentadas da ROM do ZX81 | Símbolos da ROM (asm81, depurador) |
| `zx81 sysvars.pdf` | Variáveis de sistema do ZX81 | Símbolos predefinidos |
| `Sinclair ZX81 BASIC Programming.pdf`, `ZX81_BasicProgramming.pdf`, `ZX81_BASIC_book(1982_Reprinted)(acme).pdf` | Manual de BASIC do ZX81 | ZX Dignified, tokens |
| `ZX81Assembly.pdf`, `fp.pdf`, `hotz2.pdf` | Programação em assembler / ponto flutuante do ZX81 | asm81, depurador |
| `MCoderII.pdf`, `mcoder2.pdf` | Compilador de BASIC MCoder II | Ideias para compilação |
| `CBI.pdf` | CBI (formato de BASIC usado pelo VB81 / ZxToken) | Ferramentas de BASIC |
| `Memo64k.pdf`, `MemoAsm.pdf`, `MemoCalc.pdf`, `MemoText.pdf` | Manuais Memotech (memória 64K, assembler, planilha, editor) | Periféricos Memotech |
| `Manual do Mega Assembler.pdf` | Mega Assembler (MSX, Cibertron, 1987), em português | Editor de assembler e monitor |
| `ZXzine1.pdf` … `ZXzine13.pdf`, `ZXzine_index.pdf`, `ZXzine*.zip`, `zxzine*.zip` | Revista ZXzine (edições 1 a 13) e programas | Referência e testes |
| `zx26.pdf`, `zx26.zip` | Material diverso de ZX81 | Referência |
| `DragonBones.zip` | Arquivo de jogo/demo | Referência |
| `zx81keyboard.jpg` | Foto do teclado do ZX81 | Teclado na tela, ajuda |

> Origens dos documentos: baixados pelo usuário; endereços a completar quando conhecidos.

---

## 3. Configurações locais (fora do Git)

| Item | Onde | O que fazer no PC novo |
|---|---|---|
| Permissões do Claude Code | `C:\dos\go81\.claude\settings.local.json` | Copiar do backup, ou recriar com `/permissions` → *Allow*: `Bash(gh release create:*)`, `Bash("/c/Program Files/GitHub CLI/gh.exe" release create:*)`, `Bash(git push origin:*)` |
| Memória do Claude Code (notas sobre o projeto: processo de release, ambiente de build) | `C:\Users\<usuário>\.claude\projects\C--dos-go81\memory\` | Copiar do backup para o mesmo lugar. O nome da pasta (`C--dos-go81`) vem do caminho do projeto: se o projeto ficar em outro caminho, o nome muda |
| Linha local do `resume.txt` (`claude --resume …`) | `C:\dos\go81\resume.txt` | Só vale neste PC (identifica a sessão do Claude Code); não é enviada ao GitHub |
| Saída da compilação | `C:\dos\go81\bin\` | Não precisa de backup: `go build -o bin/ ./cmd/go81` |
| Login do GitHub (`gh`) e chave SSH (`github-wilson`) | perfil do usuário | Refazer `gh auth login` e configurar a chave SSH do host `github-wilson` em `~/.ssh/config` |

---

## 4. Ferramentas a instalar

| Ferramenta | Versão usada | Para quê | Onde |
|---|---|---|---|
| Go | 1.27 (mínimo 1.25) | Compilar o Go81 | <https://go.dev/dl/> |
| MSYS2 + `mingw-w64-ucrt-x86_64-gcc` | gcc 16.2 | CGO da janela (Fyne) e comparadores com o EightyOne; pôr `C:\msys64\ucrt64\bin` no `PATH` | <https://www.msys2.org/> |
| Git | — | Versionamento | <https://git-scm.com/> |
| GitHub CLI (`gh`) | 2.103 | Páginas de release | <https://cli.github.com/> |
| Python 3 (opcional) | — | Scripts de manutenção da documentação | <https://www.python.org/> |
| Claude Code | — | Assistente de desenvolvimento | <https://claude.com/claude-code> |

Sem gcc ainda é possível compilar e testar sem a janela: `go build -tags nogui ./cmd/go81` e
`go test -tags nogui ./...`.
