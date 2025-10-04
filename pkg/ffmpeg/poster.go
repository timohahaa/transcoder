package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
)

func PosterThumbCPU(
	ctx context.Context,
	cpuIdx []int,
	src, dst string,
) (string, error) {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return "", err
	}

	var (
		output = filepath.Join(dst, "poster.jpg")
		args   = []string{
			"-xerror",
			"-hide_banner",
			"-y",
			"-t", "60", // <- !!!
			"-i", src,
			"-vf", "thumbnail",
			"-frames:v", "1",
			"-update", "1",
			"-qscale:v", "2",
			output,
		}
	)

	_, err := scope(ctx, cpuIdx, src, nil, args, DiscardProgress)
	return output, err
}
