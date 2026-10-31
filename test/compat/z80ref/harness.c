/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Reference harness: runs the original EightyOne Z80 core on test cases read
 * from stdin and prints the resulting state, so it can be compared with the
 * Go port (see ../z80diff). The case format and the random generator must
 * match z80diff exactly.
 *
 * Input line : seed mode n b0 b1 b2 b3   (decimal)
 *   mode 0 = one instruction, b0..b3 placed at PC
 *   mode 1 = maskable interrupt, data bus = b0, halted = b1&1
 *   mode 2 = NMI, halted = b1&1
 *   mode 3 = run n instructions on random memory
 * Output line: see print_state().
 */
#include <stdio.h>
#include <string.h>
#include <stdint.h>

#include "config.h"
#include "zx81config.h"
#include "z80.h"

/* Globals the core expects from the rest of EightyOne. */
MACHINE machine;
int tstates;
int RetExecuted;
int StackChange;
int StepOutRequested;

static BYTE mem[65536];
static uint64_t port_state;
static uint64_t log_hash;
static int verbose;

static uint64_t next(uint64_t *s)
{
        uint64_t z = (*s += 0x9E3779B97F4A7C15ULL);
        z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9ULL;
        z = (z ^ (z >> 27)) * 0x94D049BB133111EBULL;
        return z ^ (z >> 31);
}

static void log_event(int kind, int addr, int value, int ts)
{
        unsigned char b[10];
        int i;
        b[0] = (unsigned char)kind;
        b[1] = (unsigned char)addr; b[2] = (unsigned char)(addr >> 8);
        b[3] = (unsigned char)value;
        b[4] = (unsigned char)ts; b[5] = (unsigned char)(ts >> 8);
        b[6] = (unsigned char)(ts >> 16); b[7] = (unsigned char)(ts >> 24);
        for (i = 0; i < 8; i++) {
                log_hash ^= b[i];
                log_hash *= 0x100000001B3ULL;
        }
        if (verbose)
                printf("  %c %04x %02x ts=%d\n", "FORWIO"[kind], addr, value, ts);
}

/* Bus. Addresses are wrapped to 16 bits, as the Go port does. */
static BYTE h_fetch(int a)   { a &= 0xffff; log_event(0, a, mem[a], tstates); return mem[a]; }
static BYTE h_operand(int a) { a &= 0xffff; log_event(1, a, mem[a], tstates); return mem[a]; }
static BYTE h_read(int a)    { a &= 0xffff; log_event(2, a, mem[a], tstates); return mem[a]; }
static void h_write(int a, int d)
{
        a &= 0xffff; d &= 0xff;
        log_event(3, a, d, tstates);
        mem[a] = (BYTE)d;
}
static BYTE h_in(int port, int *ts)
{
        uint64_t s = port_state ^ ((uint64_t)(port & 0xffff) << 32) ^ (uint64_t)(unsigned)*ts;
        BYTE v = (BYTE)next(&s);
        port &= 0xffff;
        log_event(4, port, v, *ts);
        if ((port & 0x0300) == 0x0300) *ts += 1;   /* exercise wait states */
        return v;
}
static void h_out(int port, int d, int *ts)
{
        port &= 0xffff; d &= 0xff;
        log_event(5, port, d, *ts);
        if ((port & 0x0300) == 0x0300) *ts += 2;
}
static int h_contend(int a, int states, int time) { (void)a; (void)states; return time; }

static void setup(uint64_t seed)
{
        uint64_t st = seed, v;
        int i, j;
        for (i = 0; i < 65536; i += 8) {
                v = next(&st);
                for (j = 0; j < 8; j++) mem[i + j] = (BYTE)(v >> (8 * j));
        }
        v = next(&st);
        z80.af.b.h = (BYTE)v;         z80.af.b.l = (BYTE)(v >> 8);
        z80.bc.b.h = (BYTE)(v >> 16); z80.bc.b.l = (BYTE)(v >> 24);
        z80.de.b.h = (BYTE)(v >> 32); z80.de.b.l = (BYTE)(v >> 40);
        z80.hl.b.h = (BYTE)(v >> 48); z80.hl.b.l = (BYTE)(v >> 56);
        v = next(&st);
        z80.af_.w = (WORD)v; z80.bc_.w = (WORD)(v >> 16);
        z80.de_.w = (WORD)(v >> 32); z80.hl_.w = (WORD)(v >> 48);
        v = next(&st);
        z80.ix.w = (WORD)v; z80.iy.w = (WORD)(v >> 16);
        z80.sp.w = (WORD)(v >> 32); z80.pc.w = (WORD)(v >> 48);
        v = next(&st);
        z80.i = (BYTE)v;
        z80.r = (WORD)((v >> 8) & 0x7f);
        z80.r7 = (BYTE)((v >> 16) & 0x80);
        z80.iff1 = (BYTE)((v >> 24) & 1);
        z80.iff2 = (BYTE)((v >> 25) & 1);
        z80.im = (BYTE)(((v >> 26) & 0xff) % 3);
        z80.halted = 0;
        port_state = next(&st);

        StackChange = 0;
        RetExecuted = 0;
        numberOfM1Cycles = 0;
        memset(z80_MCycles(), 0, 11 * sizeof(int));
        log_hash = 0xCBF29CE484222325ULL;
}

static uint64_t state_hash(void)
{
        uint64_t h = 0xCBF29CE484222325ULL;
        unsigned v[20];
        int i;
        v[0] = z80.af.w; v[1] = z80.bc.w; v[2] = z80.de.w; v[3] = z80.hl.w;
        v[4] = z80.af_.w; v[5] = z80.bc_.w; v[6] = z80.de_.w; v[7] = z80.hl_.w;
        v[8] = z80.ix.w; v[9] = z80.iy.w; v[10] = z80.sp.w; v[11] = z80.pc.w;
        v[12] = z80.i; v[13] = z80.r; v[14] = z80.r7;
        v[15] = z80.iff1 | z80.iff2 << 1 | z80.im << 2 | z80.halted << 4;
        v[16] = (unsigned)tstates; v[17] = (unsigned)StackChange;
        v[18] = (unsigned)numberOfM1Cycles; v[19] = (unsigned)RetExecuted;
        for (i = 0; i < 20; i++) {
                h ^= v[i];
                h *= 0x100000001B3ULL;
        }
        return h;
}

static void print_state(int mode, int ts)
{
        printf("%04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x "
               "i=%02x r=%04x r7=%02x iff=%d%d im=%d h=%d ts=%d sc=%d ret=%d",
               z80.af.w, z80.bc.w, z80.de.w, z80.hl.w,
               z80.af_.w, z80.bc_.w, z80.de_.w, z80.hl_.w,
               z80.ix.w, z80.iy.w, z80.sp.w, z80.pc.w,
               z80.i, z80.r, z80.r7, z80.iff1, z80.iff2, z80.im, z80.halted,
               ts, StackChange, RetExecuted);
        printf(" m1=%d", numberOfM1Cycles);
        if (mode == 0) {
                int *m = z80_MCycles(), i;
                printf(" mc=");
                for (i = 0; i < 11 && m[i]; i++) printf("%s%d", i ? "," : "", m[i]);
                printf(" io=%d", z80_InputOutputMCycle());
        }
        printf(" log=%016llx\n", (unsigned long long)log_hash);
}

int main(int argc, char **argv)
{
        char line[256];
        verbose = argc > 1 && strcmp(argv[1], "-v") == 0;

        machine.writebyte = h_write;
        machine.readbyte = h_read;
        machine.readoperandbyte = h_operand;
        machine.opcode_fetch = h_fetch;
        machine.writeport = h_out;
        machine.readport = h_in;
        machine.contendmem = h_contend;
        machine.contendio = h_contend;
        z80_init();

        while (fgets(line, sizeof line, stdin)) {
                unsigned long long seed;
                int mode, n, b[4], ts = 0, i;
                if (sscanf(line, "%llu %d %d %d %d %d %d", &seed, &mode, &n,
                           &b[0], &b[1], &b[2], &b[3]) != 7)
                        continue;
                setup(seed);
                switch (mode) {
                case 0:
                        for (i = 0; i < 4; i++) mem[(z80.pc.w + i) & 0xffff] = (BYTE)b[i];
                        ts = z80_do_opcode();
                        break;
                case 1:
                        z80.halted = b[1] & 1;
                        tstates = 0;
                        ts = z80_interrupt(b[0]);
                        break;
                case 2:
                        z80.halted = b[1] & 1;
                        tstates = 0;
                        ts = z80_nmi();
                        break;
                case 3: {
                        uint64_t acc = 0xCBF29CE484222325ULL;
                        for (i = 0; i < n; i++) {
                                ts += z80_do_opcode();
                                acc ^= state_hash();
                                acc *= 0x100000001B3ULL;
                        }
                        log_hash ^= acc;
                        break;
                }
                }
                print_state(mode, ts);
                fflush(stdout);
        }
        return 0;
}
