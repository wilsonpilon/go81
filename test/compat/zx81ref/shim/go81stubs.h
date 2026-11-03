/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Stand-ins for the VCL and for EightyOne's peripherals, so zx81.cpp can be
 * compiled outside C++Builder. Every peripheral behaves as "not fitted".
 * All the empty shim headers include this file.
 */
#ifndef GO81_STUBS_H
#define GO81_STUBS_H

#include <string>
#include <string.h>

/* ---- VCL ---------------------------------------------------------------- */
typedef unsigned char  BYTE;
typedef signed char    SBYTE;
typedef unsigned short WORD;
typedef unsigned long  DWORD;
namespace Forms { struct TMonitor; }
/* VCL AnsiString: 1-based indexing, as in C++Builder. */
class AnsiString {
public:
        std::string s;
        AnsiString() {}
        AnsiString(const char *p) : s(p ? p : "") {}
        AnsiString(const std::string &x) : s(x) {}
        AnsiString &operator=(const char *p) { s = p ? p : ""; return *this; }
        const char *c_str() const { return s.c_str(); }
        int Length() const { return (int)s.size(); }
        char &operator[](int i) { return s[i - 1]; }
        char operator[](int i) const { return s[i - 1]; }
        AnsiString SubString(int start, int count) const
        {
                if (start < 1) start = 1;
                if (start > (int)s.size() || count <= 0) return AnsiString();
                return AnsiString(s.substr(start - 1, count));
        }
        int AnsiPos(const char *sub) const
        {
                size_t p = s.find(sub);
                return p == std::string::npos ? 0 : (int)p + 1;
        }
        AnsiString UpperCase() const
        {
                std::string r = s;
                for (auto &c : r) c = (char)toupper((unsigned char)c);
                return AnsiString(r);
        }
        AnsiString &operator+=(const AnsiString &o) { s += o.s; return *this; }
        AnsiString &operator+=(const char *o) { s += o; return *this; }
        AnsiString &operator+=(char c) { s += c; return *this; }
        AnsiString operator+(const AnsiString &o) const { return AnsiString(s + o.s); }
        AnsiString operator+(const char *o) const { return AnsiString(s + o); }
        bool operator!=(const char *o) const { return s != o; }
        bool operator==(const char *o) const { return s == o; }
        bool operator==(const AnsiString &o) const { return s == o.s; }
        bool operator!=(const AnsiString &o) const { return s != o.s; }
};
inline AnsiString operator+(const char *a, const AnsiString &b) { return AnsiString(std::string(a) + b.s); }
AnsiString LowerCase(const AnsiString &a);
bool FileExists(const AnsiString &a);
AnsiString IntToStr(int v);
AnsiString IntToHex(int v, int digits);
typedef int BOOL;
typedef unsigned long ULONG;
#define MB_OK 0
#define MB_ICONERROR 0x10
#define MB_ICONWARNING 0x30
struct TApplication { int MessageBox(const char *, const char *, int) { return 0; } };
extern TApplication *Application;

/* ---- Tape ------------------------------------------------------------------ */
struct TWavLoad { bool IgnoreZX81 = false; };
extern TWavLoad *WavLoad;
int ZX81Strlen(unsigned char *str);
AnsiString FileNameGetExt(AnsiString Fname);
void ConvertASCIIZX81(AnsiString Str, unsigned char *Output);
AnsiString RemoveExt(AnsiString Fname);
AnsiString GetExt(AnsiString Fname);
AnsiString RemovePath(AnsiString Str);
void WavStart();
void WavStop();
void WavStartRec();
void WavRecordByte(int);
int IsFlashSaveable();

/* ---- UI objects --------------------------------------------------------- */
struct TLiveMemoryWindow { void Read(unsigned short) {} void Write(unsigned short) {} };
extern TLiveMemoryWindow *LiveMemoryWindow;
struct TBasicLister { bool Visible; void Refresh(bool) {} };
extern TBasicLister *BasicLister;
struct TForm1 { void BuildMenuJoystickSelection() {} };
extern TForm1 *Form1;

/* ---- Sound / speech ----------------------------------------------------- */
#define AY_TYPE_NONE 0
#define AY_TYPE_QUICKSILVA 1
#define AY_TYPE_ZONX 2
#define SPEECH_TYPE_NONE 0
#define SPEECH_TYPE_PARROT 2
#define SPEECH_TYPE_SWEETTALKER 3
#define SPEECH_TYPE_MAGECO 4
#define SPEECH_TYPE_TALKBACK 5
#define SPEECH_TYPE_DIGITALKER 8
struct TSound {
        int m_SampleRate, m_Channels, m_FPS, FrameSize, VolumeLevel3;
        int OldVal, OldValOrig, OldPos, FillPos, BeeperLastSubpos, BeeperTick, BeeperTickIncr;
        short *Buffer;
        void Init(int fps);
        void Frame(bool pause);
        void Beeper(int on, int frametstates);
        void AYWrite(int, int, int) {}
};
extern TSound Sound;
struct TSP0256 { void Write(unsigned char) {} bool Busy() { return false; } void Reset() {} };
extern TSP0256 sp0256_AL2;
struct TDigitalker { void Write1(unsigned char) {} void Write2(unsigned char) {} };
extern TDigitalker Digitalker;

/* ---- Tape --------------------------------------------------------------- */
int GetEarState();
void WavClockTick(int TStates, int MicState);

/* ---- 8255 / 8251 / printer / IEC / floppy / IDE --------------------------- */
#define D8255PRTA 1
#define D8255PRTB 2
#define D8255PRTC 3
unsigned char d8255_read(int);
void d8255_write(int, unsigned char);
void d8255_reset();
void d8251reset();
void d8251writeDATA(unsigned char);
void d8251writeCTRL(unsigned char);
unsigned char d8251readDATA();
unsigned char d8251readCTRL();
void ZXPrinterReset();
void ZXPrinterWritePort(unsigned char);
unsigned char ZXPrinterReadPort(unsigned char);
void ZXPrinterClockTick(int);
void IECClockTick(int);
void IECAssertReset(int); void IECReleaseReset(int);
void IECAssertATN(int);   void IECReleaseATN(int);
void IECAssertClock(int); void IECReleaseClock(int);
void IECAssertData(int);  void IECReleaseData(int);
int IECIsReset(); int IECIsATN(); int IECIsClock(); int IECIsData();
void floppy_init();
#define ATA_MODE_16BIT 1
void ATA_Reset(); void ATA_SetMode(int);
void ATA_WriteRegister(int, int); int ATA_ReadRegister(int);

/* ---- ZXpand --------------------------------------------------------------- */
#define CFG_BIT_LOW 0
#define CFG_BIT_DISABLED 1
class ZXpand {
public:
        static void SetSDCardPath() {}
        void GetConfig(int &c) { c = 0; }
        void IO_Write(int, int) {}
        int IO_Read(int) { return 0xff; }
        int IO_ReadStatus() { return 0; }
        void Update(int) {}
};

/* ---- ROM cartridge / Chroma / joystick ------------------------------------ */
extern int RomCartridgeCapacity;
void InitialiseRomCartridge();
void ResetRomCartridge();
bool ReadRomCartridge(int, unsigned char *);
bool WriteRomCartridge(int, unsigned char *);
void InitialiseChroma();
void DisableChroma();
bool ChromaRAMWrite(int, unsigned char, unsigned char *, unsigned char *);
bool ChromaRAMRead(int, unsigned char *, unsigned char *);
bool ChromaIOWrite(int, unsigned char);
bool ChromaIORead(int, unsigned char *);
void FetchChromaColour(int, unsigned char, int, unsigned char *);
void SetChromaColours();
void GetChromaColours(int *, int *);
int GetChromaBorderColour();
#define JOYSTICK_NONE 0
#define JOYSTICK_KEMPSTON 1
#define JOYSTICK_CURSOR 3
#define JOYSTICK_PROGRAMMABLE 6
struct TJoystickBinding { int AddressMask; };
extern TJoystickBinding JoystickLeft1, JoystickRight1, JoystickUp1, JoystickDown1, JoystickFire1;
void InitialiseJoysticks();
unsigned char ReadJoystick1();
unsigned char ReadJoystick1_Left();
unsigned char ReadJoystick1_Right();
unsigned char ReadJoystick1_Up();
unsigned char ReadJoystick1_Down();
unsigned char ReadJoystick1_Fire();
unsigned char ReadJoystick1_RightUpDownFire();

#endif
