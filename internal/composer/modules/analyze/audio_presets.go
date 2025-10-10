package analyze

import (
	"context"
	"strconv"

	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type AudioPreset struct {
	Ffprobe *ffprobe.Info
	Preset  *pb.AudioPreset
}

var baseAudioPreset = pb.AudioPreset{
	Channels:     2,
	Bitrate:      192 * consts.KBit,
	SampleRate:   44100,
	PadBefore:    0,
	PadAfter:     0,
	TrimBefore:   0,
	TrimDuration: 0,
}

func CalcAudioPresets(ctx context.Context, info *ffprobe.Info, files []string) (map[string]AudioPreset, error) {
	var (
		audioPresetsMap = make(map[string]AudioPreset, len(files))
		videoDur        = info.GetDuration()
	)

	for _, file := range files {
		aInfo, err := ffprobe.GetInfo(ctx, file)
		if err != nil {
			return nil, err
		}

		preset := baseAudioPreset.Copy()

		if streams := aInfo.GetAllAudios(); len(streams) > 0 {
			aStream := streams[0]

			if sampleRate, err := strconv.ParseInt(aStream.SampleRate, 10, 64); err == nil {
				preset.SampleRate = sampleRate
			}

			if aInfo.Format.BitRate != 0 {
				preset.Bitrate = min(preset.Bitrate, aInfo.Format.BitRate)
			} else if aStream.BitRate != 0 {
				preset.Bitrate = min(preset.Bitrate, aStream.BitRate)
			}

			startTime, _ := strconv.ParseFloat(aStream.StartTime, 64)

			if startTime < 0 {
				preset.TrimBefore = -1 * float32(startTime)
			} else {
				preset.PadBefore = float32(startTime)
			}

			switch dur := startTime + aInfo.GetDuration(); {
			case dur < videoDur:
				preset.PadAfter = float32(videoDur - dur)
			case dur > videoDur:
				preset.TrimDuration = float32(videoDur)
			}
		}

		audioPresetsMap[file] = AudioPreset{
			Ffprobe: aInfo,
			Preset:  preset,
		}
	}

	return audioPresetsMap, nil
}
