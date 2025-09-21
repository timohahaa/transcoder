package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

var ErrNoVideos = errors.New("no videos")

func UnmuxVideo(ctx context.Context, info *ffprobe.Info, srcFile string, dstDir string) (string, error) {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", err
	}

	if info == nil {
		var err error
		info, err = ffprobe.GetInfo(ctx, srcFile)
		if err != nil {
			return "", err
		}
	}

	video := info.GetHighestVideo()
	if video.CodecType != consts.CodecTypeVideo {
		return "", ErrNoVideos
	}

	var (
		out  = filepath.Join(dstDir, fmt.Sprintf("video.%s", info.GetFileExt()))
		args = []string{
			"-xerror",
			"-hide_banner",
			"-y",
			"-i", srcFile,
			"-map", fmt.Sprintf("0:%d", video.Index),
			"-c:v", "copy",
			"-an",
			"-sn",
			out,
		}
	)
	if err := execute(ctx, srcFile, args); err != nil {
		return "", err
	}
	return out, nil
}
