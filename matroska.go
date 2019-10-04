// Package matroska implements a wrapper for Haali's Matroska Parser in Go
//
// This was born out of the need for a simple way to get info and, more importantly,
// packets and codec private data easily out of Matroska files via standard Go I/O
// interfaces. All of the existing packages seemed far too low level (EBML-level)
// or just bad.
package matroska

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/pborman/uuid"
)

// #include <stdlib.h>
// #include "MatroskaParser.h"
// #include "io.h"
//
// /*
//  * This only exists because CGO is too stupid to cast between
//  * structs that are subsets of eachother, and this is required
//  * for MatroskaParser to work.
//  */
// InputStream *convert(IO* io) { return (InputStream *)io; }
import "C"

// Demuxer is a Matroska demuxer.
type Demuxer struct {
	m      *C.MatroskaFile
	errbuf *C.char
	io     *C.IO
	key    string
}

func newDemuxerWithFlag(r io.ReadSeeker, flag C.unsigned) (*Demuxer, error) {
	ret := new(Demuxer)

	ret.key = uuid.New()

	addReader(ret.key, r)

	ret.errbuf = (*C.char)(C.calloc(1, 1024))
	if ret.errbuf == nil {
		return nil, fmt.Errorf("could not allocate errbuf")
	}

	ret.io = C.io_alloc()
	if ret.io == nil {
		C.free(unsafe.Pointer(ret.errbuf))
		return nil, fmt.Errorf("could not allocate io struct")
	}

	ckey := C.CString(ret.key)
	defer C.free(unsafe.Pointer(ckey))

	C.io_set_callbacks(ret.io, ckey)

	ret.m = C.mkv_OpenEx(C.convert(ret.io), 0, flag, ret.errbuf, 1024)
	if ret.m == nil {
		reason := C.GoString(ret.errbuf)
		C.free(unsafe.Pointer(ret.errbuf))
		C.io_free(ret.io)
		return nil, fmt.Errorf("couldn't open matroska file: %s", reason)
	}

	return ret, nil
}

// NewDemuxer creates a new Matroska demuxer from r.
func NewDemuxer(r io.ReadSeeker) (*Demuxer, error) {
	return newDemuxerWithFlag(r, 0)
}

// NewStreamingDemuxer creates a new Matroska demuxer from an
// io.Reader that has no ability to seek on the input stream.
func NewStreamingDemuxer(r io.Reader) (*Demuxer, error) {
	fs := &fakeSeeker{r: r}
	return newDemuxerWithFlag(fs, C.MKVF_AVOID_SEEKS)
}

// Close closes a demuxer.
func (d *Demuxer) Close() {
	C.mkv_Close(d.m)
	C.io_free(d.io)
	C.free(unsafe.Pointer(d.errbuf))
	delReader(d.key)
}

// GetNumTracks gets the number of tracks available to a given demuxer.
func (d *Demuxer) GetNumTracks() (uint, error) {
	ret := uint(C.mkv_GetNumTracks(d.m))
	if ret <= 0 {
		reason := C.GoString(d.errbuf)
		return 0, fmt.Errorf("couldn't get number of tracks: %s", reason)
	}
	return ret, nil
}

// GetTrackInfo returns all track-level information available for a given track,
// where track is less than what is returned by GetNumTracks.
func (d *Demuxer) GetTrackInfo(track uint) (*TrackInfo, error) {
	ti := C.mkv_GetTrackInfo(d.m, C.unsigned(track))
	if ti == nil {
		reason := C.GoString(d.errbuf)
		return nil, fmt.Errorf("could not get track info: %s", reason)
	}

	return convertTrackInfo(ti), nil
}

// GetFileInfo gets all top-level (whole file) info available for a given
// demuxer.
func (d *Demuxer) GetFileInfo() (*SegmentInfo, error) {
	si := C.mkv_GetFileInfo(d.m)
	if si == nil {
		reason := C.GoString(d.errbuf)
		return nil, fmt.Errorf("could not get file info: %s", reason)
	}

	return convertSegmentInfo(si), nil
}

// GetAttachments returns information on all available attachments
// for a given demuxer. The returned slice may be of length 0.
func (d *Demuxer) GetAttachments() []*Attachment {
	var count C.unsigned
	var attachments *C.Attachment

	C.mkv_GetAttachments(d.m, &attachments, &count)

	if count == 0 {
		return []*Attachment{}
	}

	var attachmentsSlice []C.Attachment

	// This pattern is used enough that I should factor it out into its
	// own function... but I am lazy.
	sliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&attachmentsSlice))
	sliceConv.Data = uintptr(unsafe.Pointer(attachments))
	sliceConv.Len = int(count)
	sliceConv.Cap = int(count)

	ret := make([]*Attachment, int(count))
	for i := 0; i < int(count); i++ {
		ret[i] = convertAttachment(&attachmentsSlice[i])
	}

	return ret
}

func processChapters(chapters *C.Chapter, count int) []*Chapter {
	var chapterSlice []C.Chapter

	sliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&chapterSlice))
	sliceConv.Data = uintptr(unsafe.Pointer(chapters))
	sliceConv.Len = count
	sliceConv.Cap = count

	ret := make([]*Chapter, count)
	for i := 0; i < count; i++ {
		ret[i] = convertPartialChapter(&chapterSlice[i])
		if chapterSlice[i].nChildren > 0 {
			ret[i].Children = processChapters(chapterSlice[i].Children, int(chapterSlice[i].nChildren))
		}

		var trackSlice []C.ulonglong

		tSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&trackSlice))
		tSliceConv.Data = uintptr(unsafe.Pointer(chapterSlice[i].Tracks))
		tSliceConv.Len = int(chapterSlice[i].nTracks)
		tSliceConv.Cap = int(chapterSlice[i].nTracks)

		ret[i].Tracks = make([]uint64, int(chapterSlice[i].nTracks))
		for j := 0; j < int(chapterSlice[i].nTracks); j++ {
			ret[i].Tracks[j] = uint64(trackSlice[j])
		}

		var displaySlice []C.struct_ChapterDisplay

		dSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&displaySlice))
		dSliceConv.Data = uintptr(unsafe.Pointer(chapterSlice[i].Display))
		dSliceConv.Len = int(chapterSlice[i].nDisplay)
		dSliceConv.Cap = int(chapterSlice[i].nDisplay)

		ret[i].Display = make([]ChapterDisplay, int(chapterSlice[i].nDisplay))
		for j := 0; j < int(chapterSlice[i].nDisplay); j++ {
			ret[i].Display[j] = convertChapterDisplay(&displaySlice[j])
		}

		var processSlice []C.struct_ChapterProcess

		pSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&processSlice))
		pSliceConv.Data = uintptr(unsafe.Pointer(chapterSlice[i].Process))
		pSliceConv.Len = int(chapterSlice[i].nProcess)
		pSliceConv.Cap = int(chapterSlice[i].nProcess)

		ret[i].Process = make([]ChapterProcess, int(chapterSlice[i].nProcess))
		for j := 0; j < int(chapterSlice[i].nProcess); j++ {
			ret[i].Process[j] = convertPartialChapterProcess(&processSlice[j])

			var commandSlice []C.struct_ChapterCommand

			cSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&commandSlice))
			cSliceConv.Data = uintptr(unsafe.Pointer(processSlice[j].Commands))
			cSliceConv.Len = int(processSlice[j].nCommands)
			cSliceConv.Cap = int(processSlice[j].nCommands)

			ret[i].Process[j].Commands = make([]ChapterCommand, int(processSlice[j].nCommands))
			for k := 0; k < int(processSlice[j].nCommands); k++ {
				ret[i].Process[j].Commands[k] = convertChapterCommand(&commandSlice[k])
			}
		}
	}

	return ret
}

// GetChapters returns all chapters for a given demuxer. The returned slice may
// be of length 0.
func (d *Demuxer) GetChapters() []*Chapter {
	var count C.unsigned
	var chapters *C.Chapter

	C.mkv_GetChapters(d.m, &chapters, &count)

	if count == 0 {
		return []*Chapter{}
	}

	return processChapters(chapters, int(count))
}

// GetTags returns all tags for a given demuxer. The returned slice may be of
// length 0.
func (d *Demuxer) GetTags() []*Tag {
	var count C.unsigned
	var tags *C.Tag

	C.mkv_GetTags(d.m, &tags, &count)

	if count == 0 {
		return []*Tag{}
	}

	var tagSlice []C.Tag

	sliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&tagSlice))
	sliceConv.Data = uintptr(unsafe.Pointer(tags))
	sliceConv.Len = int(count)
	sliceConv.Cap = int(count)

	ret := make([]*Tag, int(count))
	for i := 0; i < int(count); i++ {
		ret[i] = new(Tag)

		var targetSlice []C.struct_Target

		tSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&targetSlice))
		tSliceConv.Data = uintptr(unsafe.Pointer(tagSlice[i].Targets))
		tSliceConv.Len = int(tagSlice[i].nTargets)
		tSliceConv.Cap = int(tagSlice[i].nTargets)

		ret[i].Targets = make([]Target, int(tagSlice[i].nTargets))
		for j := 0; j < int(tagSlice[i].nTargets); j++ {
			ret[i].Targets[j] = convertTarget(&targetSlice[j])
		}

		var simpleTagSlice []C.struct_SimpleTag

		sSliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&simpleTagSlice))
		sSliceConv.Data = uintptr(unsafe.Pointer(tagSlice[i].SimpleTags))
		sSliceConv.Len = int(tagSlice[i].nSimpleTags)
		sSliceConv.Cap = int(tagSlice[i].nSimpleTags)

		ret[i].SimpleTags = make([]SimpleTag, int(tagSlice[i].nSimpleTags))
		for j := 0; j < int(tagSlice[i].nSimpleTags); j++ {
			ret[i].SimpleTags[j] = convertSimpleTag(&simpleTagSlice[j])
		}
	}

	return ret
}

// GetCues returns all cues for a given demuxer. The returned slice may be
// of length 0.
func (d *Demuxer) GetCues() []*Cue {
	var count C.unsigned
	var cues *C.Cue

	C.mkv_GetCues(d.m, &cues, &count)

	if count == 0 {
		return []*Cue{}
	}

	var cuesSlice []C.Cue

	sliceConv := (*reflect.SliceHeader)(unsafe.Pointer(&cuesSlice))
	sliceConv.Data = uintptr(unsafe.Pointer(cues))
	sliceConv.Len = int(count)
	sliceConv.Cap = int(count)

	ret := make([]*Cue, int(count))
	for i := 0; i < int(count); i++ {
		ret[i] = convertCue(&cuesSlice[i])
	}

	return ret
}

// GetSegment returns the position of the segment.
func (d *Demuxer) GetSegment() uint64 {
	return uint64(C.mkv_GetSegment(d.m))
}

// GetSegmentTop returns the position of the next byte after the segment.
func (d *Demuxer) GetSegmentTop() uint64 {
	return uint64(C.mkv_GetSegmentTop(d.m))
}

// GetCuesPos returna the position of the cues in the stream.
func (d *Demuxer) GetCuesPos() uint64 {
	return uint64(C.mkv_GetCuesPos(d.m))
}

// GetCuesTopPos returns the position of the byte after the end of the cues.
func (d *Demuxer) GetCuesTopPos() uint64 {
	return uint64(C.mkv_GetCuesTopPos(d.m))
}

// Seek seeks to a given timecode.
//
// Flags here may be: 0 (normal seek), matroska.SeekToPrevKeyFrame,
// or matoska.SeekToPrevKeyFrameStrict
func (d *Demuxer) Seek(timecode uint64, flags uint32) {
	C.mkv_Seek(d.m, C.ulonglong(timecode), C.unsigned(flags))
}

// SeekCueAware seeks to a given timecode while taking cues into account
//
// Flags here may be: 0 (normal seek), matroska.SeekToPrevKeyFrame,
// or matoska.SeekToPrevKeyFrameStrict
//
// fuzzy defines whether a fuzzy seek will be used or not.
func (d *Demuxer) SeekCueAware(timecode uint64, flags uint32, fuzzy bool) {
	f := 0
	if fuzzy {
		f = 1
	}

	C.mkv_Seek_CueAware(d.m, C.ulonglong(timecode), C.unsigned(flags), C.unsigned(f))
}

// SkipToKeyframe skips to the next keyframe in a stream.
func (d *Demuxer) SkipToKeyframe() {
	C.mkv_SkipToKeyframe(d.m)
}

// GetLowestQTimecode returns the lowest queued timecode in the demuxer.
func (d *Demuxer) GetLowestQTimecode() uint64 {
	return uint64(C.mkv_GetLowestQTimecode(d.m))
}

// SetTrackMask sets the demuxer's track mask; that is, it tells the demuxer
// which tracks to skip, and which to use. Any tracks with ones in their bit
// positions will be ignored.
//
// Calling this withh cause all parsed and queued frames to be discarded.
func (d *Demuxer) SetTrackMask(mask uint64) {
	C.mkv_SetTrackMask(d.m, C.ulonglong(mask))
}

// ReadPacketMask is the same as ReadPacket except with a track mask.
func (d *Demuxer) ReadPacketMask(mask uint64) (*Packet, error) {
	var track C.unsigned
	var startTime C.ulonglong
	var endTime C.ulonglong
	var filePos C.ulonglong
	var frameSize C.unsigned
	var frameData *C.char
	var frameFlags C.unsigned
	var discard C.longlong

	cret := C.mkv_ReadFrame(d.m, C.ulonglong(mask), &track, &startTime, &endTime, &filePos, &frameSize, &frameData, &frameFlags, &discard)
	if cret == -1 {
		return nil, io.EOF
	} else if cret != 0 {
		reason := C.GoString(d.errbuf)
		return nil, fmt.Errorf("could not read packet: %s", reason)
	}

	ret := &Packet{
		Track:     uint8(track),
		StartTime: uint64(startTime),
		EndTime:   uint64(endTime),
		FilePos:   uint64(filePos),
		Flags:     uint32(frameFlags),
		Discard:   int64(discard),
	}

	ret.Data = C.GoBytes(unsafe.Pointer(frameData), C.int(frameSize))

	return ret, nil
}

// ReadPacket returns the next packet from a demuxer.
func (d *Demuxer) ReadPacket() (*Packet, error) {
	return d.ReadPacketMask(0)
}
