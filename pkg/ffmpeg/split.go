package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Chunk struct {
	Name string
	Size int64
	Path string
	Num  int
}

func Split(
	ctx context.Context,
	srcFile, dstDir string,
	duration int,
	segmentFormat string,
) ([]Chunk, error) {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return nil, err
	}

	var args = []string{
		"-xerror",
		"-hide_banner",
		"-i", srcFile,
		"-c", "copy",
		"-an",
		"-f", "segment", // https://ffmpeg.org/ffmpeg-formats.html#Options-31
		"-segment_time", strconv.Itoa(duration),
		"-segment_format", segmentFormat,
		"-reset_timestamps", "1",
		filepath.Join(dstDir, "chunk_%03d."+segmentFormat),
	}

	if err := execute(ctx, srcFile, args); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dstDir)
	if err != nil {
		return nil, err
	}

	chunks := make([]Chunk, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}

		chunkNum, err := getChunkNum(e.Name())
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, Chunk{
			Name: e.Name(),
			Size: info.Size(),
			Path: filepath.Join(dstDir, e.Name()),
			Num:  chunkNum,
		})
	}
	return chunks, nil
}

func getChunkNum(filename string) (int, error) {
	parts := strings.Split(filename, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid chunk name format: %v", filename)
	}
	numStr := strings.TrimSuffix(parts[1], filepath.Ext(parts[1]))
	return strconv.Atoi(numStr)
}
