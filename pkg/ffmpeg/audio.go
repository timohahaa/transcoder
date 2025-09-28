package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

var ErrNoAudios = errors.New("no audios")

func UnmuxAudios(ctx context.Context, info *ffprobe.Info, srcFile, dstDir string) ([]string, error) {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return nil, err
	}

	if info == nil {
		var err error
		info, err = ffprobe.GetInfo(ctx, srcFile)
		if err != nil {
			return nil, err
		}
	}

	var audios = info.GetAllAudios()
	if len(audios) == 0 {
		return nil, ErrNoAudios
	}

	var (
		audioFiles = make([]string, 0, len(audios))
		audioMap   = make([]string, 0, len(audios))
		audioIdx   = 0
	)
	for _, s := range audios {
		var audioPath = filepath.Join(dstDir, fmt.Sprintf("orig_audio_%d.%s", audioIdx, info.GetFileExt()))
		audioMap = append(
			audioMap,
			"-map", fmt.Sprintf("0:%d", s.Index),
			"-c:a", "copy",
			"-vn",
			audioPath,
		)
		audioFiles = append(audioFiles, audioPath)
		audioIdx++
	}

	var args = append([]string{
		"-xerror",
		"-hide_banner",
		"-y",
		"-i", srcFile,
	}, audioMap...)
	if err := execute(ctx, srcFile, args); err != nil {
		return nil, err
	}
	return audioFiles, nil
}
