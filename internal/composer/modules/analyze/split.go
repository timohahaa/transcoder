package analyze

import (
	"math"

	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

const (
	oneChunkDuration    = 60  // in seconds
	splitInHalfDuration = 120 // in seconds
)

func CalcChunkSize(info *ffprobe.Info) (_ int, needSplit bool) {
	var duration = info.GetDuration()

	if duration <= float64(oneChunkDuration) {
		return 0, false
	}

	if duration <= float64(splitInHalfDuration) {
		return int(math.Ceil(duration / 2)), true
	}

	return oneChunkDuration, true
}
