package validate

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

func Pre(ctx context.Context, videoFile string, audioFiles []string) error {
	var vInfo, err = ffprobe.GetInfo(ctx, videoFile)
	if err != nil {
		return err
	}

	var vDur = vInfo.GetDuration()

	for _, a := range audioFiles {

		aInfo, err := ffprobe.GetInfo(context.Background(), a)
		if err != nil {
			return err
		}

		if math.Abs(aInfo.GetDuration()-vDur) >= 0.5 {
			return errors.PreValidation(fmt.Errorf(
				"audio and video streams have different length: video = %v, audio %v = %v",
				vDur,
				filepath.Base(a),
				aInfo.GetDuration(),
			))
		}

		if len(aInfo.Streams) > 0 {
			if dur, err := time.ParseDuration(aInfo.Streams[0].StartTime + "s"); err == nil {
				if math.Abs(dur.Seconds()) >= 1 {
					return errors.PreValidation(fmt.Errorf(
						"audio stream %v has too big start time: %v",
						filepath.Base(a),
						aInfo.Streams[0].StartTime,
					))
				}
			}
		}
	}

	return nil
}

func Chunks(ctx context.Context, chunks []ffmpeg.Chunk, sourceInfo *ffprobe.Info) error {
	var accumDur float64

	for _, c := range chunks {
		cInfo, err := ffprobe.GetInfo(ctx, c.Path)
		if err != nil {
			return err
		}
		accumDur += cInfo.GetDuration()
	}

	if math.Abs(accumDur-sourceInfo.GetDuration()) >= 0.5 {
		return errors.PreValidation(fmt.Errorf(
			"got wrong duration after split: original = %v, chunks duration sum = %v",
			sourceInfo.GetDuration(),
			accumDur,
		))
	}
	return nil
}
