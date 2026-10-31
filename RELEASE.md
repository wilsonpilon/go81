# Go81 — Releases

Formato de versão: **X.Y.Z** — veja [SPEC.md §10](SPEC.md#10-versionamento).

- **X** — grande bloco, com nome histórico do ZX81 (ex.: *Sinclair*)
- **Y** — feature ou bloco de programas, com nome de jogo do ZX81 (ex.: *3D Monster Maze*)
- **Z** — ajuste, correção ou compilação, com subtítulo curto

---

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
