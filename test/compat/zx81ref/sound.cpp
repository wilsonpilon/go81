/* Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * The beeper part of EightyOne's CSound (sound/sound.cpp, Copyright (C)
 * 2003-2025 Michael D Wynne; beeper code from Fuse by Russell Marks; GPL v2
 * or later), copied verbatim except for the DirectSound output, so the
 * harness produces the same audio buffer as EightyOne.
 */
#include <string.h>
#include "go81stubs.h"
#include "zx81config.h"

#define AMPL_BEEPER 31
#define BEEPER_FADEOUT (((1<<24)/150)/AMPL_BEEPER)

#define BEEPER_OLDVAL_ADJUST \
  BeeperTick+=BeeperTickIncr;	\
  if(BeeperTick>=BEEPER_FADEOUT)	\
    {					\
    BeeperTick-=BEEPER_FADEOUT;	\
    if(OldVal>0)		\
      OldVal--;			\
    else				\
      if(OldVal<0)		\
        OldVal++;			\
    }

void TSound::Init(int fps)
{
        m_SampleRate = 44100;
        m_Channels = 2;
        m_FPS = fps;
        VolumeLevel3 = 31 - 16;          /* BeeperVol slider default */
        BeeperLastSubpos = 0;
        FrameSize = m_SampleRate / m_FPS;
        Buffer = new short[FrameSize * m_Channels];
        memset(Buffer, 0, FrameSize * m_Channels * sizeof(short));
        OldVal = OldValOrig = 0;
        OldPos = -1;
        FillPos = 0;
        BeeperTick = 0;
        BeeperTickIncr = (1 << 24) / m_SampleRate;
}

void TSound::Frame(bool pause)
{
        int f;
        for(f=FillPos;f<FrameSize;f++)
        {
                BEEPER_OLDVAL_ADJUST;
                int tempval=(OldVal*256*VolumeLevel3)/AMPL_BEEPER;
                Buffer[f*m_Channels]=tempval;
                if(m_Channels == 2) Buffer[f*m_Channels+1]=tempval;
        }
        OldPos=-1;
        FillPos=0;
}

void TSound::Beeper(int on, int frametstates)
{
        int newpos,subpos;
        int val,subval;
        int f;

        val=AMPL_BEEPER*(on?1:-1);
        if(val==OldValOrig) return;

        newpos=(frametstates*FrameSize)/machine.tperframe;
        subpos=(frametstates*FrameSize*AMPL_BEEPER)/machine.tperframe-AMPL_BEEPER*newpos;

        if(newpos==OldPos)
        {
                if(on) BeeperLastSubpos+=AMPL_BEEPER-subpos;
                else BeeperLastSubpos-=AMPL_BEEPER-subpos;
        }
        else
                BeeperLastSubpos=(on?AMPL_BEEPER-subpos:subpos);

        subval=0-AMPL_BEEPER+BeeperLastSubpos;

        if(newpos>=0)
        {
                for(f=FillPos;f<newpos && f<FrameSize;f++)
                {
                        BEEPER_OLDVAL_ADJUST;
                        int tempval=(OldVal*256*VolumeLevel3)/AMPL_BEEPER;
                        Buffer[f*m_Channels]=tempval;
                        if(m_Channels==2) Buffer[f*m_Channels+1]=tempval;
                }
                if(newpos<FrameSize)
                {
                        if((OldVal<0 && subval<OldVal) || (OldVal>=0 && subval>OldVal))
                                subval=OldVal;
                        int tempval=(subval*256*VolumeLevel3)/AMPL_BEEPER;
                        Buffer[newpos*m_Channels]=tempval;
                        if(m_Channels==2) Buffer[newpos*m_Channels+1]=tempval;
                }
        }

        OldPos=newpos;
        FillPos=newpos+1;
        OldVal=OldValOrig=val;
}
