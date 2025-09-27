package analyze

import (
	"math"
	"strconv"

	"github.com/timohahaa/transcoder/pkg/consts"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type preset struct {
	Quality        string
	MaxBitRate     int64
	MinBitRate     int64
	FPS            string
	Codec          string
	Bufsize        int64
	GOPSeconds     int32
	Profile        string
	Level          string
	CRF            int32
	ColorTrc       string
	ColorSpace     string
	ColorPrimaries string
	Tune           string
	Transpose      string
	IsVertical     bool
	Width          int32
	WidthMax       int32
	Height         int32
}

// recalculates the width and height of the preset based on the original width and height
// usually we calculate the video width based on the height
// but this doesn't work for non-standard aspect ratios
// for example:
// original is 4K, but its resolution is not 16:9, but 2:1 (3840x1920)
// then for 4K with a height of 2160, the width becomes 4320 => the video gets upscaled
// in this case we should do the opposite - calculate the video height based on the width
//
// so for each quality we have a certain `WidthMax`
// if when calculating the width based on the height we exceed this `WidthMax` threshold
// then we do the opposite - calculate the height based on the width
func (p *preset) setResolution(origW, origH int) {
	if origH > origW { // if video is vertical
		origW, origH = origH, origW
	}

	// width from height
	{
		var calcW int32

		calcW = int32(math.Round(float64(p.Height*int32(origW)) / float64(origH)))
		if calcW%2 != 0 {
			calcW += 1
		}

		if calcW <= p.WidthMax {
			p.Width = calcW

			if calcW > int32(origW) {
				p.Width = int32(origW)
				p.Height = int32(origH)
			}
			return
		}

	}

	// height from width
	{
		var calcH int32
		p.Width = p.WidthMax

		calcH = int32(math.Round(float64(p.Width*int32(origH)) / float64(origW)))
		if calcH%2 != 0 {
			calcH += 1
		}

		p.Height = calcH

		if calcH > int32(origH) {
			p.Width = int32(origW)
			p.Height = int32(origH)
		}
	}
}

func (p preset) copy() preset {
	return preset{
		Quality:        p.Quality,
		MaxBitRate:     p.MaxBitRate,
		MinBitRate:     p.MinBitRate,
		FPS:            p.FPS,
		Codec:          p.Codec,
		Bufsize:        p.Bufsize,
		GOPSeconds:     p.GOPSeconds,
		Profile:        p.Profile,
		Level:          p.Level,
		CRF:            p.CRF,
		ColorTrc:       p.ColorTrc,
		ColorSpace:     p.ColorSpace,
		ColorPrimaries: p.ColorPrimaries,
		Tune:           p.Tune,
		Transpose:      p.Transpose,
		IsVertical:     p.IsVertical,
		Width:          p.Width,
		WidthMax:       p.WidthMax,
		Height:         p.Height,
	}
}

func (p preset) toProto() *pb.Preset {
	return &pb.Preset{
		Quality:        p.Quality,
		MaxBitRate:     p.MaxBitRate,
		MinBitRate:     p.MinBitRate,
		FPS:            p.FPS,
		Codec:          p.Codec,
		Bufsize:        p.Bufsize,
		GOPSeconds:     p.GOPSeconds,
		Profile:        p.Profile,
		Level:          p.Level,
		CRF:            p.CRF,
		ColorTrc:       p.ColorTrc,
		ColorSpace:     p.ColorSpace,
		ColorPrimaries: p.ColorPrimaries,
		Tune:           p.Tune,
		Transpose:      p.Transpose,
		IsVertical:     p.IsVertical,
		Width:          p.Width,
		Height:         p.Height,
	}
}

var presetFpsMap = map[int]map[string]preset{
	consts.FPS30: {
		consts.Q360p: {
			Quality:        consts.Q360p,
			MaxBitRate:     1000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        2000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_3_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2, // -2 means ffmpeg decides automaticaly
			WidthMax:       640,
			Height:         360,
		},
		consts.Q480p: {
			Quality:        consts.Q480p,
			MaxBitRate:     1500 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        3000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_3_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       854,
			Height:         480,
		},
		consts.Q720p: {
			Quality:        consts.Q720p,
			MaxBitRate:     3000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        6000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_4_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       1280,
			Height:         720,
		},
		consts.Q1080p: {
			Quality:        consts.Q1080p,
			MaxBitRate:     6000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        12000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_4_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       1920,
			Height:         1080,
		},
		consts.Q1440p: {
			Quality:        consts.Q1440p,
			MaxBitRate:     9000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        18000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_5_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       2560,
			Height:         1440,
		},
		consts.Q2160p: {
			Quality:        consts.Q2160p,
			MaxBitRate:     12000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS30),
			Codec:          consts.CodecH264,
			Bufsize:        24000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_5_1,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       3840,
			Height:         2160,
		},
	},
	consts.FPS60: {
		consts.Q360p: {
			Quality:        consts.Q360p,
			MaxBitRate:     1500 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        3000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_3_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2, // -2 means ffmpeg decides automaticaly
			WidthMax:       640,
			Height:         360,
		},
		consts.Q480p: {
			Quality:        consts.Q480p,
			MaxBitRate:     2250 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        4500 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_3_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       854,
			Height:         480,
		},
		consts.Q720p: {
			Quality:        consts.Q720p,
			MaxBitRate:     4500 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        9000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileMain,
			Level:          consts.Level_4_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       1280,
			Height:         720,
		},
		consts.Q1080p: {
			Quality:        consts.Q1080p,
			MaxBitRate:     9000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        18000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_4_0,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       1920,
			Height:         1080,
		},
		consts.Q1440p: {
			Quality:        consts.Q1440p,
			MaxBitRate:     13500 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        27000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_5_1,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       2560,
			Height:         1440,
		},
		consts.Q2160p: {
			Quality:        consts.Q2160p,
			MaxBitRate:     18000 * consts.KBit,
			MinBitRate:     0,
			FPS:            strconv.Itoa(consts.FPS60),
			Codec:          consts.CodecH264,
			Bufsize:        36000 * consts.KBit,
			GOPSeconds:     2,
			Profile:        consts.ProfileHigh,
			Level:          consts.Level_5_2,
			CRF:            23,
			ColorTrc:       "",
			ColorSpace:     "",
			ColorPrimaries: "",
			Tune:           "",
			Transpose:      "",
			IsVertical:     false,
			Width:          -2,
			WidthMax:       3840,
			Height:         2160,
		},
	},
}
