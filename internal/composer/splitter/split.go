package splitter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/timohahaa/transcoder/internal/composer/modules/analyze"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

func (s *Splitter) split(
	ctx context.Context,
	info *ffprobe.Info,
	srcFile, dstDir string,
) ([]ffmpeg.Chunk, error) {
	var (
		chunkDuration, needSplit = analyze.CalcChunkSize(info)
		chunkContainerFormat     = info.GetFileExt()
	)

	if needSplit {
		chunks, err := ffmpeg.Split(
			ctx,
			srcFile,
			dstDir,
			chunkDuration,
			chunkContainerFormat,
		)
		if err != nil {
			return nil, errors.SplitSources(err)
		}
		return chunks, nil
	}

	// just copy file as first chunk
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return nil, errors.Splitter(err)
	}

	source, err := os.Open(srcFile)
	if err != nil {
		return nil, errors.Splitter(err)
	}
	defer source.Close()

	var (
		name = fmt.Sprintf("chunk_000.%s", chunkContainerFormat)
		path = filepath.Join(dstDir, name)
	)

	destination, err := os.Create(path)
	if err != nil {
		return nil, errors.Splitter(err)
	}
	defer destination.Close()

	size, err := io.Copy(destination, source)
	if err != nil {
		return nil, errors.Splitter(err)
	}

	return []ffmpeg.Chunk{{
		Name: name,
		Size: size,
		Path: path,
		Num:  0,
	}}, nil
}
