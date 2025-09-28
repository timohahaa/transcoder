package ffmpeg

import (
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
)

func Fragment(ctx context.Context, src, dstDir, fname string, fragSeconds int) (string, error) {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", err
	}

	var (
		out  = filepath.Join(dstDir, fname)
		args = []string{
			"-xerror",
			"-hide_banner",
			"-i", src,
			"-c", "copy",
			"-movflags", "+faststart+frag_keyframe+empty_moov+default_base_moof",
			"-frag_duration", strconv.FormatInt(int64(fragSeconds)*1_000_000, 10),
			"-f", "mp4",
			out,
		}
	)

	if err := execute(ctx, src, args); err != nil {
		return "", err
	}

	return out, nil
}
