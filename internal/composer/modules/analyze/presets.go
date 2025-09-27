package analyze

import (
	"context"
	"math"

	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type ChunkPresets struct {
	Ffprobe *ffprobe.Info
	Presets []*pb.Preset
}

func CalcChunkPresets(info *ffprobe.Info, chunks []ffmpeg.Chunk) (map[string]ChunkPresets, error) {
	if IsSmallBitrate(info) {
		return calcChunkPresetsSmall(info, chunks)
	}
	return calcChunkPresets(info, chunks)
}

// optimize bitrate for every chunk to minimize output bitrate while preserving quality
func calcChunkPresets(info *ffprobe.Info, chunks []ffmpeg.Chunk) (map[string]ChunkPresets, error) {
	var (
		high              = info.GetHighestVideo()
		encodeQualities   = lessOrEqQualities(high.GetQuality())
		basePresets       = calcBasePresets(info, encodeQualities)
		bitrateLadder     = bitrateLadderMap[high.GetQuality()]
		bitrateMultiplier = calcBitrateMultiplier(info)
		chunkPresetsMap   = make(map[string]ChunkPresets, len(chunks))
	)

	for _, chunk := range chunks {
		chunkInfo, err := ffprobe.GetInfo(context.Background(), chunk.Path)
		if err != nil {
			return nil, err
		}

		var (
			chunkBitrate = chunkInfo.GetHighestVideo().BitRate
			chunkPresets = make([]*pb.Preset, 0, len(basePresets))
		)

		// for some codecs like prores
		if chunkBitrate == 0 {
			chunkBitrate = chunkInfo.Format.BitRate
		}

		for _, vp := range basePresets {
			preset := vp.Copy()
			bitrare := float64(chunkBitrate) * bitrateLadder[preset.Quality]

			if bitrare <= float64(preset.MaxBitRate) {
				preset.MaxBitRate = int64(math.Ceil(bitrare))
			}

			preset.MaxBitRate = int64(math.Round(float64(preset.MaxBitRate) * bitrateMultiplier))
			preset.Bufsize = preset.MaxBitRate * 2

			chunkPresets = append(chunkPresets, preset)
		}
		chunkPresetsMap[chunk.Name] = ChunkPresets{
			Ffprobe: chunkInfo,
			Presets: chunkPresets,
		}
	}

	return chunkPresetsMap, nil
}

func calcBasePresets(info *ffprobe.Info, encodeQualities []string) []*pb.Preset {
	var (
		highVideo  = info.GetHighestVideo()
		fps, _     = highVideo.GetFrameRate()
		fpsStr     = highVideo.RFrameRate
		presets    = map[string]preset{}
		gopSeconds = calcGOPSize(info)
	)

	// cap fps
	switch {
	case fps <= 40:
		fps = 30
		fpsStr = "30/1"
	default: // > 45 => cap at 60
		fps = 60
		fpsStr = "60/1"
	}

	// base presets
	var originalPresets = map[string]preset{}
	if fps <= 40 {
		originalPresets = presetFpsMap[consts.FPS30]
	} else {
		originalPresets = presetFpsMap[consts.FPS60]
	}

	for _, qual := range encodeQualities {
		presets[qual] = originalPresets[qual].copy()
	}

	var (
		isVertical      bool
		transposeFilter string
		origW, origH    int
	)
	origW, origH = highVideo.GetRenderResolution()
	isVertical = origH > origW
	// CPU handles rotation by default
	// if GPU decoding is used - should also calculate transpose filter
	transposeFilter = ""

	var res []*pb.Preset
	for _, preset := range presets {
		preset.FPS = fpsStr
		preset.IsVertical = isVertical
		preset.Transpose = transposeFilter
		preset.GOPSeconds = gopSeconds

		preset.setResolution(origW, origH)

		res = append(res, preset.toProto())
	}

	return res
}

// gop size in seconds
func calcGOPSize(_ *ffprobe.Info) int32 {
	// for now just return 4 seconds
	return 4
}

// because videos are encoded to H264
// bitrates need to be readjusted depending on source codec
// for example, h265 generally gives 30-50% smaller file sizes compared to H264
// so in that case we need to up bitrate for H265 -> H264 by 30-50%
func calcBitrateMultiplier(info *ffprobe.Info) float64 {
	switch info.GetHighestVideo().CodecName {
	case consts.CodecAV1, consts.CodecVP9:
		return 1.5
	case consts.CodecProres:
		return 2
	default:
		return 1
	}
}

func lessOrEqQualities(quality string) []string {
	switch quality {
	case consts.Q2160p:
		return []string{
			consts.Q2160p,
			consts.Q1440p,
			consts.Q1080p,
			consts.Q720p,
			consts.Q480p,
			consts.Q360p,
		}
	case consts.Q1440p:
		return []string{
			consts.Q1440p,
			consts.Q1080p,
			consts.Q720p,
			consts.Q480p,
			consts.Q360p,
		}
	case consts.Q1080p:
		return []string{
			consts.Q1080p,
			consts.Q720p,
			consts.Q480p,
			consts.Q360p,
		}
	case consts.Q720p:
		return []string{
			consts.Q720p,
			consts.Q480p,
			consts.Q360p,
		}
	case consts.Q480p:
		return []string{
			consts.Q480p,
			consts.Q360p,
		}
	case consts.Q360p:
		return []string{
			consts.Q360p,
		}
	default:
		return nil
	}
}
