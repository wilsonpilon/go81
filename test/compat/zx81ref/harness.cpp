/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Reference harness for the ZX81 machine: runs EightyOne's zx81.cpp (with the
 * peripherals stubbed out) and a copy of its AccurateDraw raster logic, and
 * prints one line per emulated frame so the Go port can be compared with it.
 *
 * Usage: zx81ref ROM FRAMES [options] < keys
 *   options: wrx=0|1  ram=K (RAM pack size in K, default 16)  ntsc=0|1
 *            tape=FILE (.p/.81/.p81, inserted with flash load and auto start)
 *            autoload=1 (reset and type LOAD "" like EightyOne's auto-load)
 *            dump=N   (write frame N of the display to frameN.pgm)
 * keys (stdin): lines "frame row bits", setting ZXKeyboard[row]=bits from
 *               that frame on (bits as in EightyOne: bit set = key down).
 * Output per frame:
 *   f=N pc=.. af=.. bc=.. de=.. hl=.. ix=.. iy=.. sp=.. i=.. r=.. ts=..
 *   line=<hash of every scanline> disp=<hash of last completed display>
 *   mem=<hash of 0-64K> ry=RasterY
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <vector>

#include "go81stubs.h"
#include "zx81config.h"
#include "zx81.h"
#include "z80.h"

/* ---- globals normally defined elsewhere in EightyOne --------------------- */
ZX81 zx81;
ZX97 zx97;
SPECTRUM spectrum;
TV tv;
MACHINE machine;
MOUSE mouse;
ROMCARTRIDGE romcartridge;
EMULATOR emulator;
unsigned char zx97bankmem[16 * 16384];
const char *romsFolder = "";
const char *replacementRomsFolder = "";
const char *fdcRomsFolder = "";
const char *ideRomsFolder = "";

bool directMemoryAccess;
int lastMemoryReadAddrLo, lastMemoryWriteAddrLo;
int lastMemoryReadAddrHi, lastMemoryWriteAddrHi;
int lastMemoryReadValueLo, lastMemoryWriteValueLo;
int lastMemoryReadValueHi, lastMemoryWriteValueHi;
long noise;
int RasterX = 0, RasterY = 0;
int VSYNC_TOLLERANCEMIN = 290;   /* set in main() as Artifacts_.cpp does */
int VSYNC_TOLLERANCEMAX = 340;
static int HSYNC_TOLLERANCE = 405;
int StackChange, RetExecuted, StepOutRequested;

extern BYTE memory[];
extern void InitPatches(int machineType);
extern bool InsertTape(const char *path, bool autoLoad);
extern void DoAutoLoad(void);
extern int AutoLoadCount;
#include "tzxfile.h"
extern TTZXFile TZXFile;
extern BYTE ZXKeyboard[8];
extern int frametstates;

/* ---- stubs ---------------------------------------------------------------- */
TLiveMemoryWindow lmw, *LiveMemoryWindow = &lmw;
TBasicLister bl = {false}, *BasicLister = &bl;
TForm1 form, *Form1 = &form;
TSound Sound;
TSP0256 sp0256_AL2;
TDigitalker Digitalker;
int RomCartridgeCapacity = 0;
TJoystickBinding JoystickLeft1, JoystickRight1, JoystickUp1, JoystickDown1, JoystickFire1;

AnsiString LowerCase(const AnsiString &a) { std::string s = a.s; for (auto &c : s) c = (char)tolower(c); return AnsiString(s); }
bool FileExists(const AnsiString &) { return true; }
AnsiString getMachineRoot(AnsiString fullRomName)
{
        std::string s = fullRomName.s;
        size_t p = s.find_last_of("/\\");
        if (p != std::string::npos) s = s.substr(p + 1);
        p = s.find('.');
        return AnsiString(s.substr(0, p));
}
AnsiString PrependFolder(AnsiString folder, char *romFile) { return folder + romFile; }
unsigned char d8255_read(int) { return 0; }
void d8255_write(int, unsigned char) {}
void d8255_reset() {}
void d8251reset() {}
void d8251writeDATA(unsigned char) {}
void d8251writeCTRL(unsigned char) {}
unsigned char d8251readDATA() { return 0; }
unsigned char d8251readCTRL() { return 0; }
void ZXPrinterReset() {}
void ZXPrinterWritePort(unsigned char) {}
unsigned char ZXPrinterReadPort(unsigned char d) { return d; }
void ZXPrinterClockTick(int) {}
void IECClockTick(int) {}
void IECAssertReset(int) {} void IECReleaseReset(int) {}
void IECAssertATN(int) {} void IECReleaseATN(int) {}
void IECAssertClock(int) {} void IECReleaseClock(int) {}
void IECAssertData(int) {} void IECReleaseData(int) {}
int IECIsReset() { return 0; } int IECIsATN() { return 0; } int IECIsClock() { return 0; } int IECIsData() { return 0; }
void floppy_init() {}
void ATA_Reset() {} void ATA_SetMode(int) {}
void ATA_WriteRegister(int, int) {} int ATA_ReadRegister(int) { return 0xff; }
void InitialiseRomCartridge() {}
void ResetRomCartridge() {}
bool ReadRomCartridge(int, unsigned char *) { return false; }
bool WriteRomCartridge(int, unsigned char *) { return false; }
void InitialiseChroma() {}
void DisableChroma() {}
bool ChromaRAMWrite(int, unsigned char, unsigned char *, unsigned char *) { return false; }
bool ChromaRAMRead(int, unsigned char *, unsigned char *) { return false; }
bool ChromaIOWrite(int, unsigned char) { return false; }
bool ChromaIORead(int, unsigned char *) { return false; }
void FetchChromaColour(int, unsigned char, int, unsigned char *) {}
void SetChromaColours() {}
void GetChromaColours(int *, int *) {}
int GetChromaBorderColour() { return 15; }
void InitialiseJoysticks() {}
unsigned char ReadJoystick1() { return 0xff; }
unsigned char ReadJoystick1_Left() { return 0xff; }
unsigned char ReadJoystick1_Right() { return 0xff; }
unsigned char ReadJoystick1_Up() { return 0xff; }
unsigned char ReadJoystick1_Down() { return 0xff; }
unsigned char ReadJoystick1_Fire() { return 0xff; }
unsigned char ReadJoystick1_RightUpDownFire() { return 0xff; }
void LogOutAccess(int, BYTE) {}
void LogInAccess(int, BYTE) {}
void ResetLastIOAccesses() {}
void DebugUpdate() {}
extern "C" int CRC32Block(BYTE *mem, int len)
{
        /* Same CRC as EightyOne's CRC32Block (crc32.c): only used to tell ROMs apart. */
        unsigned crc = 0xffffffff;
        for (int i = 0; i < len; i++) {
                crc ^= mem[i];
                for (int k = 0; k < 8; k++) crc = (crc >> 1) ^ (0xEDB88320 & (0 - (crc & 1)));
        }
        return (int)(~crc & 0xffff);
}
extern "C" void P3DriveMachineHasInitialised(void) {}

static char romFile[1024];
extern "C" int memory_load(char *, int address, int length, int)
{
        FILE *f = fopen(romFile, "rb");
        if (!f) { fprintf(stderr, "cannot open %s\n", romFile); exit(2); }
        int n = (int)fread(memory + address, 1, length, f);
        fclose(f);
        return n;
}
extern "C" int memory_device_rom_load(char *, int, int) { return 0; }
int memoryLoadToAddress(char *, void *, int) { return 0; }
int font_load(const char *, char *, int) { return 0; }

/* zx81 machine hooks (as wired by EightyOne's HW_.cpp) */
extern void zx81_initialise();
extern int zx81_do_scanline(SCANLINE *);
extern void zx81_writebyte(int, int);
extern void zx81_setbyte(int, int);
extern BYTE zx81_readbyte(int);
extern BYTE zx81_readoperandbyte(int);
extern BYTE zx81_getbyte(int);
extern BYTE zx81_opcode_fetch(int);
extern void zx81_writeport(int, int, int *);
extern BYTE zx81_readport(int, int *);
extern int zx81_contend(int, int, int);

/* ---- AccurateDraw (from AccDraw_.cpp, GDI mode, no TV effects) ------------ */
static const int TVW = 520, TVH = 380, HSYNC_MINLEN = 10, VSYNC_MINLEN = 350;
static BYTE frameBuf[TVW * TVH];
static BYTE dispBuf[TVW * TVH];
static int destRow = 0;
static int ScanLen;
static uint64_t lineHash, dispHash;

static uint64_t fnv(uint64_t h, const BYTE *p, size_t n)
{
        for (size_t i = 0; i < n; i++) { h ^= p[i]; h *= 0x100000001B3ULL; }
        return h;
}

static void CompleteFrame()
{
        int x = RasterX, y = RasterY;
        while (y < TVH) {          /* EightyOne clears to WinB; we clear the rest */
                for (; x < TVW; x++) frameBuf[y * TVW + x] = 0;
                x = 0; y++;
        }
        memcpy(dispBuf, frameBuf, sizeof frameBuf);
        dispHash = fnv(0xCBF29CE484222325ULL, dispBuf, sizeof dispBuf);
}

static void AccurateDraw(SCANLINE *Line)
{
        int i, c;
        lineHash = fnv(lineHash, Line->scanline, Line->scanline_len);
        int meta[3] = {Line->scanline_len, Line->sync_len, Line->sync_type};
        lineHash = fnv(lineHash, (BYTE *)meta, sizeof meta);

        for (i = 0; i < Line->scanline_len; i++) {
                c = Line->scanline[i];
                if (destRow < TVH && RasterX < TVW) frameBuf[destRow * TVW + RasterX] = (BYTE)c;
                RasterX += 1;
                if (RasterX >= ScanLen) {
                        RasterX = 0;
                        destRow++;
                        RasterY += 1;
                        if (RasterY >= TVH) {
                                i = Line->scanline_len + 1;
                                Line->sync_type = SYNCTYPEV;
                                if (Line->sync_len < HSYNC_MINLEN) Line->sync_len = HSYNC_MINLEN;
                        }
                }
        }
        if (Line->sync_len < HSYNC_MINLEN) Line->sync_type = 0;
        if (Line->sync_type) {
                if (RasterX > HSYNC_TOLLERANCE) {
                        RasterX = 0;
                        RasterY += 1;
                        destRow++;
                }
                if (RasterY >= TVH || RasterY >= VSYNC_TOLLERANCEMAX ||
                    (Line->sync_len > VSYNC_MINLEN && RasterY > VSYNC_TOLLERANCEMIN)) {
                        CompleteFrame();
                        RasterX = RasterY = 0;
                        destRow = 0;
                }
        }
}

static void writePGM(const char *name)
{
        FILE *f = fopen(name, "wb");
        fprintf(f, "P5\n%d %d\n255\n", TVW, TVH);
        for (int i = 0; i < TVW * TVH; i++) fputc(dispBuf[i], f);
        fclose(f);
}

int main(int argc, char **argv)
{
        if (argc < 3) { fprintf(stderr, "usage: zx81ref ROM FRAMES [wrx=0|1 ram=K ntsc=0|1 dump=N]\n"); return 2; }
        strncpy(romFile, argv[1], sizeof romFile - 1);
        int frames = atoi(argv[2]);
        int wrx = 0, ramK = 16, ntsc = 0, dump = -1, autoload = 0, sound = 0;
        char tape[1024] = "";
        for (int a = 3; a < argc; a++) {
                sscanf(argv[a], "autoload=%d", &autoload);
                if (!strncmp(argv[a], "tape=", 5)) strncpy(tape, argv[a] + 5, sizeof tape - 1);
                sscanf(argv[a], "wrx=%d", &wrx);
                sscanf(argv[a], "ram=%d", &ramK);
                sscanf(argv[a], "ntsc=%d", &ntsc);
                sscanf(argv[a], "dump=%d", &dump);
                sscanf(argv[a], "sound=%d", &sound);
        }

        struct Ev { int frame, row, bits; };
        std::vector<Ev> evs;
        { Ev e; while (scanf("%d %d %d", &e.frame, &e.row, &e.bits) == 3) evs.push_back(e); }

        /* Stock ZX81 as EightyOne's HW dialog configures it. */
        emulator.machine = MACHINEZX81;
        emulator.romcrc = -1;
        emulator.simpleghost = 1;
        strcpy(machine.CurRom, "zx81.edition3.rom");
        machine.protectROM = 1;
        machine.NTSC = (CFGBYTE)ntsc;
        machine.clockspeed = 3250000;
        machine.tperscanline = 207;
        machine.scanlines = ntsc ? 262 : 310;
        machine.tperframe = machine.tperscanline * machine.scanlines;
        machine.fps = ntsc ? 60 : 50;
        machine.colour = COLOURDISABLED;
        machine.baseRamSize = 1;
        zx81.RAMTOP = ramK ? (ramK << 10) + 16383 : (1 << 10) + 16383;
        if (zx81.RAMTOP > 65535) zx81.RAMTOP = 65535;
        zx81.m1not = 32768;
        zx81.maxireg = 0x1F;
        zx81.truehires = wrx ? HIRESWRX : HIRESDISABLED;
        zx81.chrgen = CHRGENSINCLAIR;
        zx81.chromaMode = 0x0F;   /* load_config default */
        zx81.vsyncsound = (CFGBYTE)sound;
        Sound.Init(machine.fps);

        machine.initialise = zx81_initialise;
        machine.do_scanline = zx81_do_scanline;
        machine.writebyte = zx81_writebyte;
        machine.setbyte = zx81_setbyte;
        machine.readbyte = zx81_readbyte;
        machine.readoperandbyte = zx81_readoperandbyte;
        machine.getbyte = zx81_getbyte;
        machine.opcode_fetch = zx81_opcode_fetch;
        machine.readport = zx81_readport;
        machine.writeport = zx81_writeport;
        machine.contendmem = zx81_contend;
        machine.contendio = zx81_contend;

        /* TArtifacts::StandardPalette with the default slider positions
           (VBias = 0, VGain = 0) overrides the AccDraw defaults. */
        VSYNC_TOLLERANCEMIN = 283;
        VSYNC_TOLLERANCEMAX = VSYNC_TOLLERANCEMIN + 40;
        HSYNC_TOLLERANCE = 390;
        if (machine.NTSC) { VSYNC_TOLLERANCEMIN -= 60; VSYNC_TOLLERANCEMAX -= 60; }

        machine.initialise();
        InitPatches(MACHINEZX81);
        if (tape[0] && !InsertTape(tape, autoload != 0)) {
                fprintf(stderr, "cannot load tape %s\n", tape);
                return 2;
        }
        ScanLen = machine.tperscanline * 2;

        static SCANLINE line;
        memset(&line, 0, sizeof line);
        int borrow = 0;
        size_t ev = 0;
        for (int f = 0; f < frames; f++) {
                while (ev < evs.size() && evs[ev].frame <= f) {
                        ZXKeyboard[evs[ev].row & 7] = (BYTE)evs[ev].bits;
                        ev++;
                }
                lineHash = 0xCBF29CE484222325ULL;
                Sound.Frame(false);
                uint64_t sndHash = fnv(0xCBF29CE484222325ULL, (BYTE *)Sound.Buffer, Sound.FrameSize * Sound.m_Channels * sizeof(short));
                if (AutoLoadCount) DoAutoLoad();
                frametstates = 0;
                int j = machine.tperframe + borrow;
                while (j > 0) {
                        j -= machine.do_scanline(&line);
                        AccurateDraw(&line);
                }
                borrow = j;
                if (f == dump) {
                        char name[64];
                        sprintf(name, "frame%d.pgm", f);
                        writePGM(name);
                }
                printf("f=%d pc=%04x af=%04x bc=%04x de=%04x hl=%04x ix=%04x iy=%04x sp=%04x i=%02x r=%02x ts=%d "
                       "line=%016llx disp=%016llx mem=%016llx ry=%d tape=%d/%d snd=%016llx\n",
                       f, z80.pc.w, z80.af.w, z80.bc.w, z80.de.w, z80.hl.w, z80.ix.w, z80.iy.w, z80.sp.w,
                       z80.i, (z80.r7 & 0x80) | (z80.r & 0x7f), frametstates,
                       (unsigned long long)lineHash, (unsigned long long)dispHash,
                       (unsigned long long)fnv(0xCBF29CE484222325ULL, memory, 65536), RasterY,
                       (int)TZXFile.Playing, TZXFile.CurBlock, (unsigned long long)sndHash);
        }
        return 0;
}
