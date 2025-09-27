package analyze

import (
	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

func IsSmallBitrate(info *ffprobe.Info) bool {
	var (
		high    = info.GetHighestVideo()
		bitrate = high.BitRate
	)

	if bitrate == 0 {
		bitrate = info.Format.BitRate
	}

	if bitrate == 0 { // sry, we dunno :(
		return false
	}

	var fps, err = info.GetHighestVideo().GetFrameRate()
	if err != nil {
		return false
	}

	var ladder map[string]int64
	if fps <= 40 {
		ladder = smallBitrateMap[consts.FPS30]
	} else {
		ladder = smallBitrateMap[consts.FPS60]
	}

	return bitrate <= ladder[high.GetQuality()]
}

var (
	// MaxBitrate in presets / 3
	smallBitrateMap = map[int]map[string]int64{
		consts.FPS30: {
			consts.Q360p:  1000 * consts.KBit / 3,
			consts.Q480p:  1500 * consts.KBit / 3,
			consts.Q720p:  3000 * consts.KBit / 3,
			consts.Q1080p: 6000 * consts.KBit / 3,
			consts.Q1440p: 9000 * consts.KBit / 3,
			consts.Q2160p: 12000 * consts.KBit / 3,
		},
		consts.FPS60: {
			consts.Q360p:  1500 * consts.KBit / 3,
			consts.Q480p:  2250 * consts.KBit / 3,
			consts.Q720p:  4500 * consts.KBit / 3,
			consts.Q1080p: 9000 * consts.KBit / 3,
			consts.Q1440p: 13500 * consts.KBit / 3,
			consts.Q2160p: 18000 * consts.KBit / 3,
		},
	}
)
