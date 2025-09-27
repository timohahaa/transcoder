package analyze

import (
	"context"
	"math"
	"slices"
	"strconv"

	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

// for small bitrate videos do not optimize individual chunk bitrate
func calcChunkPresetsSmall(info *ffprobe.Info, chunks []ffmpeg.Chunk) (map[string]ChunkPresets, error) {
	var (
		high                = info.GetHighestVideo()
		baseEncodeQualities = lessOrEqQualities(high.GetQuality())
		encodeQualities     = smallBitrareEncodeQualities(baseEncodeQualities)
		basePresets         = calcBasePresets(info, encodeQualities)
		bitrateMultiplier   = calcBitrateMultiplier(info)
		chunkPresetsMap     = make(map[string]ChunkPresets, len(chunks))
		bitrate             = high.BitRate
	)

	// for some codecs like prores
	if bitrate == 0 {
		bitrate = info.Format.BitRate
	}

	for _, chunk := range chunks {
		chunkInfo, err := ffprobe.GetInfo(context.Background(), chunk.Path)
		if err != nil {
			return nil, err
		}

		var chunkPresets = make([]*pb.Preset, 0, len(basePresets))
		for _, vp := range basePresets {
			preset := vp.Copy()

			if bitrate <= preset.MaxBitRate {
				preset.MaxBitRate = bitrate
			}

			preset.MaxBitRate = int64(math.Round(float64(preset.MaxBitRate) * bitrateMultiplier))

			// + for small-bitrate up maxrate by N times
			preset.MaxBitRate = preset.MaxBitRate * 2
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

// eсли у исходника маленький битрейт, то нет смысла ему кодить все качества
// достаточно закодить то качество, которое уже есть + одно из low-res качеств
// if source has small bitrate, there is no need to encode all qualities
// we only need to encode original quality + one of "low-res" (360p-1080p) qualities
func smallBitrareEncodeQualities(encodeQualities []string) []string {
	var res []string = make([]string, 0, 2)

	slices.SortFunc(encodeQualities, func(a, b string) int {
		aNum, _ := strconv.Atoi(a)
		bNum, _ := strconv.Atoi(b)
		return aNum - bNum
	})

	highestQual := encodeQualities[len(encodeQualities)-1]
	res = append(res, highestQual)

	switch highestQual {
	case consts.Q1440p, consts.Q2160p:

		lowResQuals := lowResQualities(encodeQualities)
		if len(lowResQuals) != 0 {
			slices.SortFunc(lowResQuals, func(a, b string) int {
				aNum, _ := strconv.Atoi(a)
				bNum, _ := strconv.Atoi(b)
				return aNum - bNum
			})

			res = append(res, lowResQuals[len(lowResQuals)-1])
		}
	}

	return res
}

func lowResQualities(qs []string) []string {
	var res []string
	for _, q := range qs {
		if q == consts.Q1440p || q == consts.Q2160p {
			continue
		}
		res = append(res, q)
	}
	return res
}
