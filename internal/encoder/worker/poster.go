package worker

import (
	"context"
	"slices"
	"strconv"

	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

func (w *Worker) poster(task *pb.Task, dstDir string, out *ffmpeg.Output) (string, error) {
	if len(out.Qualities) == 0 {
		return "", nil
	}

	slices.SortFunc(out.Qualities, func(a, b ffmpeg.Quality) int {
		aNum, _ := strconv.Atoi(a.Name)
		bNum, _ := strconv.Atoi(b.Name)
		return aNum - bNum
	})

	var posterQuality = out.Qualities[len(out.Qualities)-1]

	return ffmpeg.PosterThumbCPU(
		context.Background(),
		[]int{w.opts.CpuIdx},
		posterQuality.Path,
		dstDir,
	)
}
