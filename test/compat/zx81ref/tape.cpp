/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Tape glue for the ZX81 reference harness. The tape player itself is
 * EightyOne's own code (tzx/*.cpp, zx81/rompatch.cpp); this file provides
 * what those files expect from the rest of EightyOne:
 *  - the filename helpers, copied from EightyOne Utils.cpp
 *    (Copyright (C) 2003-2025 Michael D Wynne, GPL v2 or later);
 *  - the tape manager window logic (TZXMan.cpp ClockTick, Play/Stop buttons)
 *    and WavCInterface.cpp, without the GUI;
 *  - DoAutoLoad from main_.cpp (ZX81 part).
 */
#include <stdio.h>
#include <ctype.h>
#include <string.h>

#include "go81stubs.h"
#include "zx81config.h"
#include "tzxfile.h"
#include "z80.h"

TTZXFile TZXFile;
int FlashLoadable;
TWavLoad wavLoad, *WavLoad = &wavLoad;
TApplication application, *Application = &application;

AnsiString IntToStr(int v) { char b[32]; sprintf(b, "%d", v); return AnsiString(b); }
AnsiString IntToHex(int v, int d) { char b[32]; sprintf(b, "%0*X", d, v); return AnsiString(b); }

/* ---- Utils.cpp ------------------------------------------------------------ */
char ZXCharTable[] = " ----------\'\xA3$:?()><=+-*/;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ";

AnsiString FileNameGetExt(AnsiString Fname)
{
        int i, Pos1, Pos2;
        AnsiString Result;

        Pos1 = Pos2 = 0;
        for (i = 1; i <= Fname.Length(); i++) {
                if (Fname[i] == '\\') Pos1 = i;
                if (Fname[i] == '.') Pos2 = i;
        }
        if (!Pos1) Pos1 = 1;
        if (Pos2 < Pos1) Pos2 = Fname.Length() + 1;
        Result = "";
        for (i = Pos2; i <= Fname.Length(); i++) Result += Fname[i];
        Result = Result.UpperCase();
        return (Result);
}

void ConvertASCIIZX81(AnsiString Str, unsigned char *Output)
{
        char c, *p;
        int i;

        if (!Str.Length()) return;
        i = 1;
        p = ZXCharTable;
        while (i <= Str.Length() && p) {
                c = (char)toupper(Str[i]);
                if (c == '_') c = ' ';
                p = strchr(ZXCharTable, c);
                if (p) *(Output++) = (BYTE)(p - ZXCharTable);
                i++;
        }
        Output--;
        *Output |= 128;
}

int ZX81Strlen(unsigned char *str)
{
        int len = 1;
        while (!((*str) & 128)) { len++; str++; }
        return (len);
}

AnsiString RemoveExt(AnsiString Fname)
{
        AnsiString Ext;
        int len, pos;

        len = Fname.Length();
        if (len < 3) return (Fname);
        pos = len;
        while (pos) {
                if (Fname[pos] == '.') break;
                pos--;
        }
        Ext = Fname.SubString(pos, 1 + len - pos);
        if ((Ext == ".p" || Ext == ".P" || Ext == ".T81" || Ext == ".t81" || Ext == ".P81" || Ext == ".p81"
             || Ext == ".A83" || Ext == ".a83" || Ext == ".o" || Ext == ".O" || Ext == ".81" || Ext == ".80"
             || Ext == ".TZX" || Ext == ".tzx" || Ext == ".TAP" || Ext == ".tap" || Ext == ".WAV" || Ext == ".wav"
             || Ext == ".sna" || Ext == "SNA" || Ext == ".z80" || Ext == ".Z80" || Ext == ".ini" || Ext == ".INI"
             || Ext == ".mdr" || Ext == ".MDR" || Ext == ".mdv" || Ext == ".MDV" || Ext == ".dsk" || Ext == ".DSK"
             || Ext == ".img" || Ext == ".IMG" || Ext == ".opd" || Ext == ".OPD" || Ext == ".opu" || Ext == ".OPU"
             || Ext == ".mgt" || Ext == ".MGT" || Ext == ".trd" || Ext == ".TRD" || Ext == ".hdf" || Ext == ".HDF"
             || Ext == ".b81" || Ext == ".B81" || Ext == ".b82" || Ext == ".B82" || Ext == ".txt" || Ext == ".TXT"
             || Ext == ".bas" || Ext == ".BAS" || Ext == ".web" || Ext == ".WEB" || Ext == ".pdf" || Ext == ".PDF"
             || Ext == ".ref" || Ext == ".REF") && pos > 1)
                Fname = Fname.SubString(1, pos - 1);
        return (Fname);
}

AnsiString GetExt(AnsiString Fname)
{
        AnsiString Ext;
        int len, pos;

        len = Fname.Length();
        if (len < 3) return (Fname);
        pos = len;
        while (pos) {
                if (Fname[pos] == '.') break;
                pos--;
        }
        Ext = Fname.SubString(pos, 1 + len - pos);
        for (pos = 1; pos <= Ext.Length(); pos++) Ext[pos] = (char)toupper(Ext[pos]);
        return (Ext);
}

AnsiString RemovePath(AnsiString Str)
{
        int p;
        if (Str[Str.Length()] == '\\') Str = "";
        do {
                p = Str.AnsiPos("\\");
                if (p) {
                        p++;
                        Str = Str.SubString(p, 1 + Str.Length() - p);
                }
        } while (p);
        return (Str);
}

/* ---- TZXMan.cpp (tape manager, without the window) ------------------------ */
static bool playBtnDown = false, stopBtnDown = true;
static bool autoStart = true;           /* AutoStartBtn default: down */

static void PlayBtnClick()
{
        if (TZXFile.Tape[TZXFile.CurBlock].BlockID == TZX_BLOCK_PAUSE) TZXFile.CurBlock++;
        if (!TZXFile.Start()) return;
        if (playBtnDown) return;
        playBtnDown = true;
        stopBtnDown = false;
}

static void StopBtnClick()
{
        if (stopBtnDown) return;
        TZXFile.Stop(false);
        playBtnDown = false;
}

static void TZX_ClockTick(int TStates, bool ZX81)
{
        static bool Previous = true;

        if (ZX81 && WavLoad->IgnoreZX81) return;
        if (TZXFile.Playing != Previous) {
                Previous = TZXFile.Playing;
                playBtnDown = Previous;
                stopBtnDown = !Previous;
        }
        if (!TZXFile.Playing) return;
        if (TZXFile.ClockTick(TStates)) {
                int a = TZXFile.CurBlock + 1;
                int rowCount = TZXFile.Blocks + 2;
                if (a >= rowCount) a = rowCount - 1;
                TZXFile.CurBlock = a - 1;
        }
}

/* ---- WavCInterface.cpp ----------------------------------------------------- */
int IsFlashSaveable(void) { return emulator.TZXout ? 1 : 0; }

int GetEarState()
{
        if (emulator.TZXin) return (TZXFile.GetEarState());
        return 0; /* WAV player not playing */
}

void WavClockTick(int TStates, int MicState) { TZX_ClockTick(TStates, true); }

void WavStop()
{
        if (playBtnDown && autoStart) StopBtnClick();
}

void WavStart()
{
        if (emulator.TZXin) {
                if (autoStart) PlayBtnClick();
        }
}

void WavStartRec() {}
void WavRecordByte(int) {}

/* ---- Tape insertion and DoAutoLoad (main_.cpp, ZX81 part) ------------------ */
int AutoLoadCount = 0;
extern BYTE ZXKeyboard[8];

bool InsertTape(const char *path, bool autoLoad)
{
        TZXFile.FlashLoad = true;      /* FlashLoadBtn default: down */
        TZXFile.AutoStart = true;
        if (!TZXFile.LoadFile(AnsiString(path), false)) return false;
        TZXFile.CurBlock = 0;          /* TZX->UpdateTable(true): Row = 1 */
        emulator.TZXin = 1;            /* InTZXManagerClick */
        emulator.TZXout = 1;           /* OutTZXManagerClick */
        if (autoLoad) AutoLoadCount = 2;
        return true;
}

/* Keys as EightyOne's PC keyboard mapping sets them. */
static void key(int row, int bit, bool down)
{
        if (down) ZXKeyboard[row] |= (BYTE)(1 << bit);
        else ZXKeyboard[row] &= (BYTE)~(1 << bit);
}

#define AUTOINC(i) (340 + i * 10)
extern bool zx80rom;
void DoAutoLoad(void)
{
        /* A ZX80 with its ROM types W (LOAD) and ENTER first; the ZX81 keys
           that follow are pressed too, as in EightyOne's DoAutoLoad. */
        bool zx80 = (emulator.machine == MACHINEZX80) && zx80rom;
        switch (AutoLoadCount) {
        case 2:                                 /* ResetZX811Click */
                memset(ZXKeyboard, 0, 8);
                z80_reset();
                break;
        case AUTOINC(0): if (zx80) key(2, 1, true); else key(6, 3, true); break;   /* W / J (LOAD) */
        case AUTOINC(1): if (zx80) key(2, 1, false); else key(6, 3, false); break;
        case AUTOINC(2): if (zx80) key(6, 0, true); else key(0, 0, true); break;   /* ENTER / SHIFT */
        case AUTOINC(3): if (zx80) key(6, 0, false); else key(5, 0, true); break;  /* ENTER / P -> " */
        case AUTOINC(4): key(5, 0, false); break;
        case AUTOINC(5): key(5, 0, true); break;
        case AUTOINC(6): key(5, 0, false); break;
        case AUTOINC(7): key(0, 0, false); break;
        case AUTOINC(8): key(6, 0, true); break;   /* ENTER */
        case AUTOINC(9): key(6, 0, false); break;
        default: break;
        }
        machine.NTSC ? AutoLoadCount++ : AutoLoadCount += 2;
        if (AutoLoadCount == 1000) AutoLoadCount = 0;
}

/* Larken floppy (not fitted). */
extern "C" int LarkenLoadTrack(int, int, unsigned char *) { return 1; }
extern "C" void LarkenSaveTrack(int, int, unsigned char *) {}
