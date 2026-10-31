/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Runs a CP/M .COM program (e.g. zexall.com) on the original EightyOne Z80
 * core with a minimal BDOS, to know which tests EightyOne itself passes.
 * Usage: cpm program.com
 */
#include <stdio.h>

#include "config.h"
#include "zx81config.h"
#include "z80.h"

MACHINE machine;
int tstates;
int RetExecuted;
int StackChange;
int StepOutRequested;

static BYTE mem[65536];

static BYTE rd(int a) { return mem[a & 0xffff]; }
static void wr(int a, int d) { mem[a & 0xffff] = (BYTE)d; }
static BYTE in(int p, int *ts) { (void)p; (void)ts; return 0xff; }
static void out(int p, int d, int *ts) { (void)p; (void)d; (void)ts; }
static int contend(int a, int s, int t) { (void)a; (void)s; return t; }

int main(int argc, char **argv)
{
        FILE *f;
        if (argc < 2 || !(f = fopen(argv[1], "rb"))) {
                fprintf(stderr, "usage: cpm program.com\n");
                return 2;
        }
        fread(mem + 0x100, 1, 0xfe00, f);
        fclose(f);
        mem[5] = 0xc9; mem[6] = 0x00; mem[7] = 0xf0;

        machine.readbyte = rd; machine.readoperandbyte = rd; machine.opcode_fetch = rd;
        machine.writebyte = wr; machine.readport = in; machine.writeport = out;
        machine.contendmem = contend; machine.contendio = contend;
        z80_init();
        z80_reset();
        z80.pc.w = 0x100;
        z80.sp.w = 0xf000;

        for (;;) {
                if (z80.pc.w == 0) break;
                if (z80.pc.w == 5) {
                        if (z80.bc.b.l == 2) putchar(z80.de.b.l);
                        if (z80.bc.b.l == 9) {
                                int a;
                                for (a = z80.de.w; mem[a & 0xffff] != '$'; a++) putchar(mem[a & 0xffff]);
                        }
                        fflush(stdout);
                }
                z80_do_opcode();
        }
        return 0;
}
