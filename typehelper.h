/*
 * Thise file contains helper getter functions for various members of
 * MatroskaParser's structs.
 *
 * In theory, these's *SHOULD NOT BE NEEDED*, but apparently CGO is
 * too stupid to handle basic C features like bitfields an unions.
 *
 * This does involve a C FFI call (and full stack switch) per get,
 * but this is only done once in most cases, when a user asks for
 * file-wide information, and never per packet.
 */

#ifndef _TYPE_HELPER_H
#define _TYPE_HELPER_H

#include "MatroskaParser.h"

#define TIH(type,member) type t##member(TrackInfo *t) { return t->member; }

TIH(unsigned int, Enabled)
TIH(unsigned int, Default);
TIH(unsigned int, Forced);
TIH(unsigned int, Lacing);
TIH(unsigned int, DecodeAll);
TIH(unsigned int, CompEnabled);

#define TAH(type,member) type ta##member(TrackInfo *t) { return t->AV.Audio.member; }

TAH(double, SamplingFreq)
TAH(double, OutputSamplingFreq)
TAH(unsigned char, Channels);
TAH(unsigned char, BitDepth);

#define TVH(type,member) type tv##member(TrackInfo *t) { return t->AV.Video.member; }

TVH(unsigned char, StereoMode);
TVH(unsigned char, DisplayUnit);
TVH(unsigned char, AspectRatioType);
TVH(unsigned int, PixelWidth);
TVH(unsigned int, PixelHeight);
TVH(unsigned int, DisplayWidth);
TVH(unsigned int, DisplayHeight);
TVH(unsigned int, CropL);
TVH(unsigned int, CropT);
TVH(unsigned int, CropR);
TVH(unsigned int, CropB);
TVH(unsigned int, ColourSpace);
TVH(double, GammaValue);
TVH(unsigned int, Interlaced);

#define TCH(type,member) type tc##member(TrackInfo *t) { return t->AV.Video.Colour.member; }

TCH(unsigned int, MatrixCoefficients);
TCH(unsigned int, BitsPerChannel);
TCH(unsigned int, ChromaSubsamplingHorz);
TCH(unsigned int, ChromaSubsamplingVert);
TCH(unsigned int, CbSubsamplingHorz);
TCH(unsigned int, CbSubsamplingVert);
TCH(unsigned int, ChromaSitingHorz);
TCH(unsigned int, ChromaSitingVert);
TCH(unsigned int, Range);
TCH(unsigned int, TransferCharacteristics);
TCH(unsigned int, Primaries);
TCH(unsigned int, MaxCLL);
TCH(unsigned int, MaxFALL);

#define TMMH(type,member) type tmm##member(TrackInfo *t) { return t->AV.Video.Colour.MasteringMetadata.member; }

TMMH(float, PrimaryRChromaticityX);
TMMH(float, PrimaryRChromaticityY);
TMMH(float, PrimaryGChromaticityX);
TMMH(float, PrimaryGChromaticityY);
TMMH(float, PrimaryBChromaticityX);
TMMH(float, PrimaryBChromaticityY);
TMMH(float, WhitePointChromaticityX);
TMMH(float, WhitePointChromaticityY);
TMMH(float, LuminanceMax);
TMMH(float, LuminanceMin);

#define CH(type,member) type ch##member(Chapter *c) { return c->member; }
CH(unsigned int, Hidden);
CH(unsigned int, Enabled);
CH(unsigned int, Default);
CH(unsigned int, Ordered);

// Just one for SimpleTag
int stDefault(struct SimpleTag *s) { return s->Default; }

#endif
