package matroska

// Matroska compression types
const (
	CompZlib    = 0
	CompBzip    = 1
	CompLZO1X   = 2
	CompPrepend = 3
)

// Track types
const (
	TypeVideo    = 1
	TypeAudio    = 2
	TypeSubtitle = 17
)

// Tag target types
const (
	TargetTrack      = 0
	TargetChapter    = 1
	TargetAttachment = 2
	TargetEdition    = 3
)

// Seek types
//
// They do what they say on the tin.
const (
	SeekToPrevKeyFrame       = 1
	SeekToPrevKeyFrameStrict = 2
)

// Possible flags returned with a matroksa.Packet in its
// Flags member.
const (
	UnknownStart = 0x00000001
	UnknownEnd   = 0x00000002
	KF           = 0x00000004
	GAP          = 0x00800000
	StreamMask   = 0xff000000
	StreamShift  = 24
)

// Packet contains a demuxed packet
type Packet struct {
	// The track this packet belongs to.
	Track uint8
	// The start time of this packet.
	StartTime uint64
	// The end time of this packet.
	EndTime uint64
	// The position in the input stream where this packet
	// is located.
	FilePos uint64
	// The packet data.
	Data []byte
	// Any packet flags. See constants.
	Flags uint32
	// Whether this packet can be discarded.
	Discard int64
}

// TrackInfo contains information about a track.
type TrackInfo struct {
	// The track number.
	Number uint8
	// the track type. See constants.
	Type uint8
	// Whether or not to overlay this track.
	TrackOverlay uint8
	// The UID.
	UID uint64
	// The minimum amount of frames a player should keep around to
	// be able to play back this file properly, e.g. the min DPB
	// size.
	MinCache uint64
	// The largest possible size a player could need for its cache
	// in order to play back this file.
	MaxCache uint64
	// The track's default duration, which can be used, for example
	// to calculate the duration of the last packet.
	DefaultDuration uint64
	// Any inherent delay required by the codec.
	CodecDelay uint64
	// Any pre-roll that must be applied after seeking for
	// this codec.
	SeekPreRoll uint64
	// The timescale for this track's timecodes.
	TimecodeScale float64
	// Codec private data, to be passed to decoders.
	CodecPrivate []byte
	// Track compression method. See constants.
	CompMethod uint32
	// Any private data that should be passed to the decompressor
	// used to decompress the track.
	CompMethodPrivate []byte
	// Not useful
	MaxBlockAdditionID uint32

	// Whether or not this track is enabled.
	Enabled bool
	// Whether or not this track is on by default.
	Default bool
	// Whether or not this track is forced on.
	Forced bool
	// Whether this track is laced.
	Lacing bool
	// Whether or not this track as Error Resiliance capabilities.
	DecodeAll bool
	// Whether or not this track as compression enabled.
	CompEnabled bool

	// Video information. Only valid if the track is a video track.
	Video struct {
		// The Stereo3D mode used, of any.
		StereoMode uint8
		// Display unit used for DisplayWidth and DisplayHeight.
		DisplayUnit uint8
		// What type of resizing is needed for the aspect ratio:
		//     0 = free resizing
		//     1 = keep aspect ratio
		//     2 = fixed
		AspectRatioType uint8
		// The width in pixels.
		PixelWidth uint32
		// The height in pixels.
		PixelHeight uint32
		// The width to be displayed at.
		DisplayWidth uint32
		// The height to be displayed at.
		DisplayHeight uint32
		// How many pixels to crop from the left.
		CropL uint32
		// How many pixels to crop from the top.
		CropT uint32
		// How many pixels to crop from the right.
		CropR uint32
		// How many pixels to crop from the bottom.
		CropB uint32
		// The colouspace. Like biCompression from BITMAPINFOHEADER.
		ColourSpace uint32
		// Gamma value to use for adjustment.
		GammaValue float64
		// Colour information.
		Colour struct {
			// Matrix coefficients. See: ISO/IEC 23091-4/ITU-T H.273.
			MatrixCoefficients uint32
			// Bits per colour channel.
			BitsPerChannel uint32
			// Base 2 logarithm of horizontal chroma subsampling.
			ChromaSubsamplingHorz uint32
			// Base 2 logarithm of vertical chroma subsampling.
			ChromaSubsamplingVert uint32
			// The amount of pixels to remove in the Cb channel for every pixel
			// not removed horizontally. This is additive with ChromaSubsamplingHorz.
			CbSubsamplingHorz uint32
			// The amount of pixels to remove in the Cb channel for every pixel
			// not removed vertically. This is additive with ChromaSubsamplingHorz.
			CbSubsamplingVert uint32
			// Horizontal hroma position:
			//     0 = unspecified,
			//     1 = left collocated
			//     2 = half
			ChromaSitingHorz uint32
			// Vertical hroma position:
			//     0 = unspecified
			//     1 = left collocated
			//     2 = half
			ChromaSitingVert uint32
			// Colour range:
			//     0 = unspecified
			//     1 = broadcast range (16-235)
			//     2 = full range (0-255)
			//     3 = defined by MatrixCoefficients / TransferCharacteristics
			Range uint32
			// Transfer characteristics. See: ISO/IEC 23091-4/ITU-T H.273.
			TransferCharacteristics uint32
			// Colour rimaries. See: ISO/IEC 23091-4/ITU-T H.273.
			Primaries uint32
			// Max content light level.
			MaxCLL uint32
			// Max frame-average light level.
			MaxFALL uint32
			// Mastering metadata.
			MasteringMetadata struct {
				PrimaryRChromaticityX   float32
				PrimaryRChromaticityY   float32
				PrimaryGChromaticityX   float32
				PrimaryGChromaticityY   float32
				PrimaryBChromaticityX   float32
				PrimaryBChromaticityY   float32
				WhitePointChromaticityX float32
				WhitePointChromaticityY float32
				LuminanceMax            float32
				LuminanceMin            float32
			}
		}
		// Whether or not the track is interlaced.
		Interlaced bool
	}
	// Audio information. Only valid if the track is an audiotrack.
	Audio struct {
		// Sampling frequency.
		SamplingFreq float64
		// The samplign frequency to output during play back.
		OutputSamplingFreq float64
		// Number of channels.
		Channels uint8
		// The bit depth.
		BitDepth uint8
	}

	// Name of the track.
	Name string
	// Language of the track.
	Language string
	// The track's codec.
	CodecID string
}

// SegmentInfo contains file-level (segment) information about a stream.
type SegmentInfo struct {
	// The top-level UID
	UID [16]byte
	// The UID of any files which should be played back before
	// this one.
	PrevUID [16]byte
	// The UID of any files which should be played back after
	// this one.
	NextUID [16]byte
	// The filename
	Filename string
	// The filename of any files which should be played back
	// before this one.
	PrevFilename string
	// The filename of any files which should be played back
	// after this one.
	NextFilename string
	// The title
	Title string
	// What program muxed this file
	MuxingApp string
	/// What library muxed this file
	WritingApp string
	// The timescale of any timecodes
	TimecodeScale uint64
	// The file's duration. May be 0.
	Duration uint64
	// The date the file was created on.
	DateUTC int64
	// Whether or not DateUTC can be considered valid.
	DateUTCValid bool
}

// Attachment contains info about a matroska attachment.
type Attachment struct {
	// Attachment's position within the stream.
	Position uint64
	// Attachment's length.
	Length uint64
	// Attachment's UID.
	UID uint64
	// Name of the attachment.
	Name string
	// A description of the attachment.
	Description string
	// The attachment's mime-type.
	MimeType string
}

// ChapterDisplay conatins display information for a given Chapter
type ChapterDisplay struct {
	// String. Usually chapter name.
	String string
	// What language a chapter is.
	Language string
	// The country this chapter is associated with (for when there may
	// be language dialects that vary).
	Country string
}

type ChapterCommand struct {
	// The command time
	Time uint32
	// The command
	Command []byte
}

type ChapterProcess struct {
	// The CodecID of this process
	CodecID uint32
	// Any private data for this process
	CodecPrivate []byte
	// All associated commands
	Commands []ChapterCommand
}

// Chapter contains all information about a Matroska chapter.
type Chapter struct {
	// The chapter's UID.
	UID uint64
	// Start time for the chapter.
	Start uint64
	// End time for the chapter.
	End uint64

	// Tracks this chapter pertains to.
	Tracks []uint64
	// Display information for this chapter.
	Display []ChapterDisplay
	// Any child chapters for this chapter.
	Children []*Chapter
	// Set of processes for this chapter.
	Process []ChapterProcess

	// The segment UID this chapter relates to.
	SegmentUID [16]byte

	// Whether or not this chpater is hidden.
	Hidden bool
	// Whether or not this chapter is enabled.
	Enabled bool

	// Whether or not this Edition is the default.
	Default bool
	// Whether or not this chapter is ordered.
	Ordered bool
}

// Cue contains all information about a matroska cue.
type Cue struct {
	// The cue's start time.
	Time uint64
	// The cue's duration.
	Duration uint64
	// The cue's position in the stream.
	Position uint64
	// The cue's position relative to the cluster.
	RelativePosition uint64
	// The block number.
	Block uint64
	// The track which this cue covers.
	Track uint8
}

// Target contains a information about a tag's target.
type Target struct {
	// The target's UID.
	UID uint64
	// The target type. See constants for types.
	Type uint32
}

// SimpleTag contains a simple Matroska tag.
type SimpleTag struct {
	// Tag name
	Name string
	// Tag value
	Value string
	// Tag language
	Language string
	// Whether or not this tag is applied by default.
	Default bool
}

// Tag contains all information relating to a Matroska tag.
type Tag struct {
	// A list of associated targets.
	Targets []Target
	// A list of associated simple tags.
	SimpleTags []SimpleTag
}
