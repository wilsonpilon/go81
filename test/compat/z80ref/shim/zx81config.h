/* Go81 — stand-in for EightyOne's zx81config.h: only what the Z80 core uses.
 * SPDX-License-Identifier: GPL-3.0-or-later */
#ifndef GO81_SHIM_ZX81CONFIG_H
#define GO81_SHIM_ZX81CONFIG_H
#include "config.h"

typedef struct {
        void (*writebyte)(int Address, int Data);
        BYTE (*readbyte)(int Address);
        BYTE (*readoperandbyte)(int Address);
        BYTE (*opcode_fetch)(int Address);
        void (*writeport)(int Address, int Data, int *tstates);
        BYTE (*readport)(int Address, int *tstates);
        int (*contendmem)(int Address, int states, int time);
        int (*contendio)(int Address, int states, int time);
} MACHINE;

extern MACHINE machine;
extern int tstates;

#define readbyte(Addr) (machine.readbyte(Addr))
#define readoperandbyte(Addr) (machine.readoperandbyte(Addr))
#define writebyte(Addr,Data) (machine.writebyte(Addr,Data))
#define writeport(Addr,Data,tstates) (machine.writeport(Addr,Data,tstates))
#define readport(Addr, tstates) (machine.readport(Addr, tstates))
#define opcode_fetch(Addr) (machine.opcode_fetch(Addr))
#endif
