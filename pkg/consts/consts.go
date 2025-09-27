package consts

const (
	CPU = "cpu"
	GPU = "gpu"
)

// preset/ffmpeg stuff
const (
	FPS30 = 30
	FPS60 = 60

	Q360p  = "360"
	Q480p  = "480"
	Q720p  = "720"
	Q1080p = "1080"
	Q1440p = "1440" // 2k
	Q2160p = "2160" // 4k

	TuneFilm       = "film"
	TuneAnimation  = "animation"
	TuneStillImage = "stillimage" // slideshow-like content

	CodecTypeVideo = "video"
	CodecTypeAudio = "audio"

	// H264 profiles
	ProfileMain     = "main"
	ProfileHigh     = "high"
	ProfileHigh10   = "high10"  // 10 bit compatible profile
	ProfileHigh422  = "high422" // supports yuv420p, yuv422p, yuv420p10le and yuv422p10le
	ProfileHigh444  = "high444" // supports as above as well as yuv444p and yuv444p10le
	ProfileBaseline = "baseline"

	// H264 levels
	Level_3_0 = "3.0"
	Level_3_1 = "3.1"
	Level_3_2 = "3.2"
	Level_4_0 = "4.0"
	Level_4_1 = "4.1"
	Level_4_2 = "4.2"
	Level_5_0 = "5.0"
	Level_5_1 = "5.1"
	Level_5_2 = "5.2"

	// codec names
	CodecAV1    = "av1"
	CodecVP9    = "vp9"
	CodecH264   = "h264"
	CodecPNG    = "png"
	CodecHEVC   = "hevc"
	CodecVP8    = "vp8"
	CodecProres = "prores"
	CodecMJPEG  = "mjpeg"
	CodecMPEG_1 = "mpeg1video"
	CodecMPEG_2 = "mpeg2video"
	CodecMPEG_4 = "mpeg4"
	CodecVC1    = "vc1"
	CodecAAC    = "aac"
	CodecWMV3   = "wmv3"

	TransposeCclock   = "cclock"
	TransposeClock    = "clock"
	TransposeReversal = "reversal"

	KBit = 1000
	Mbit = 1000 * KBit
)
