# Go81 — Releases

Formato de versão: **X.Y.Z** — veja [SPEC.md §10](SPEC.md#10-versionamento).

- **X** — grande bloco, com nome histórico do ZX81 (ex.: *Sinclair*)
- **Y** — feature ou bloco de programas, com nome de jogo do ZX81 (ex.: *3D Monster Maze*)
- **Z** — ajuste, correção ou compilação, com subtítulo curto

---

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
