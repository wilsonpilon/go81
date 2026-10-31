# Go81 — Releases

Formato de versão: **X.Y.Z** — veja [SPEC.md §10](SPEC.md#10-versionamento).

- **X** — grande bloco, com nome histórico do ZX81 (ex.: *Sinclair*)
- **Y** — feature ou bloco de programas, com nome de jogo do ZX81 (ex.: *3D Monster Maze*)
- **Z** — ajuste, correção ou compilação, com subtítulo curto

---

## 1.1.1 — Sinclair · 3D Monster Maze · "Planta baixa"

**Data:** 2026-10-30
**Estado:** pré-alfa — nada usável ainda

### Novidades
- Estudo inicial do EightyOne 1.41: máquinas da família ZX81, periféricos, formatos de arquivo e mapa dos fontes de referência.
- Documentação: `SPEC.md`, `OUTLINE.md`, `RELEASE.md`, `MANUAL.md`, `README.md`.
- Esqueleto Go: módulo `github.com/wilsonpilon/go81`, comando `go81` com `--version` e `--help`.

### Conhecido
- Nenhuma emulação implementada.
- Licença final ainda não definida (EightyOne é GPL v2+; ver SPEC §13).
