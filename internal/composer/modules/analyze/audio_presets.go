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
	Channels:   2,
	Bitrate:    192 * consts.KBit,
	SampleRate: 44000,
	PadBefore:  0,
	PadAfter:   0,
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

			startTime, err := strconv.ParseFloat(aStream.StartTime, 64)
			if err == nil {
				preset.PadBefore = float32(startTime)
			}

			if startTime+aInfo.GetDuration() < videoDur {
				preset.PadAfter = float32(videoDur - aInfo.GetDuration() - startTime)
			}
		}

		audioPresetsMap[file] = AudioPreset{
			Ffprobe: aInfo,
			Preset:  preset,
		}
	}

	return audioPresetsMap, nil
}
