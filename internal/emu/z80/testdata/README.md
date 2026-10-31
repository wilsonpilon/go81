# Dados de teste do núcleo Z80

| Arquivo | Origem | Uso |
|---|---|---|
| `eightyone.txt.gz` | Gerado por `go run ./test/compat/z80diff -per 4 -runs 64 -runlen 2000 -seed 81 -golden ...` | Resultados do núcleo Z80 original do EightyOne para 7.296 casos (todos os opcodes, interrupções e programas aleatórios) |
| `zexdoc.com`, `zexall.com` | Z80 Instruction Exerciser, (C) 1994 Frank D. Cringle, GPL v2+ (cópia de github.com/anotherlin/z80emu) | Testes clássicos de instruções Z80 rodando num CP/M mínimo |
