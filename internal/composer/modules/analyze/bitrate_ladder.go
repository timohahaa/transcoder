package analyze

import (
	"math"

	"github.com/timohahaa/transcoder/pkg/consts"
)

// maps highest encode quality to its bitrate ladder
var bitrateLadderMap = map[string]map[string]float64{
	consts.Q2160p: {
		consts.Q2160p: 1,
		consts.Q1440p: getRelativeBitrateCoefficient(2560, 1440, 3840, 2160),
		consts.Q1080p: getRelativeBitrateCoefficient(1920, 1080, 3840, 2160),
		consts.Q720p:  getRelativeBitrateCoefficient(1280, 720, 3840, 2160),
		consts.Q480p:  getRelativeBitrateCoefficient(854, 480, 3840, 2160),
		consts.Q360p:  getRelativeBitrateCoefficient(640, 360, 3840, 2160),
	},
	consts.Q1440p: {
		consts.Q1440p: 1,
		consts.Q1080p: getRelativeBitrateCoefficient(1920, 1080, 2560, 1440),
		consts.Q720p:  getRelativeBitrateCoefficient(1280, 720, 2560, 1440),
		consts.Q480p:  getRelativeBitrateCoefficient(854, 480, 2560, 1440),
		consts.Q360p:  getRelativeBitrateCoefficient(640, 360, 2560, 1440),
	},
	consts.Q1080p: {
		consts.Q1080p: 1,
		consts.Q720p:  getRelativeBitrateCoefficient(1280, 720, 1920, 1080),
		consts.Q480p:  getRelativeBitrateCoefficient(854, 480, 1920, 1080),
		consts.Q360p:  getRelativeBitrateCoefficient(640, 360, 1920, 1080),
	},
	consts.Q720p: {
		consts.Q720p: 1,
		consts.Q480p: getRelativeBitrateCoefficient(854, 480, 1280, 720),
		consts.Q360p: getRelativeBitrateCoefficient(640, 360, 1280, 720),
	},
	consts.Q480p: {
		consts.Q480p: 1,
		consts.Q360p: getRelativeBitrateCoefficient(640, 360, 854, 480),
	},
	consts.Q360p: {
		consts.Q360p: 1,
	},
}

const k = 0.5

// Returns the coefficient by which the original bitrate should be multiplied
// to get the bitrate for the given quality.
// The essence of the formula:
// let W0, H0 be the original resolution, and W, H be the target video resolution.
// Then the coefficient is:
//
//	  (W * H)
//	( --------- ) ^ k
//	 (W0 * H0)
//
// Or in one line: = ((W * H) / (W0 * H0)) ^ k
//
// We change k to adjust the entire bitrate ladder.
//
// Alternatives: ratio of logarithms, ratio of square roots (k = 0.5)
func getRelativeBitrateCoefficient(w, h, w0, h0 int) float64 {
	var frac = float64(w*h) / float64(w0*h0)
	return math.Pow(frac, k)
}
