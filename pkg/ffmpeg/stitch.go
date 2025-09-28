package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func Stitch(ctx context.Context, chunks []string, dstDir, fname string) (string, error) {
	var err error
	if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", err
	}

	var chunkList string
	if chunkList, err = makeChunkList(dstDir, fname, chunks); err != nil {
		return "", err
	}

	var (
		out  = filepath.Join(dstDir, fname)
		args = []string{
			"-xerror",
			"-hide_banner",
			"-f", "concat",
			"-safe", "0",
			"-i", chunkList,
			"-c", "copy",
			"-movflags", "+faststart",
			"-fflags", "+genpts",
			out,
		}
	)

	if err := execute(ctx, chunkList, args); err != nil {
		return "", err
	}

	return out, nil
}

// structure: https://trac.ffmpeg.org/wiki/Concatenate
func makeChunkList(dstFolder, fname string, chunkFiles []string) (string, error) {
	var (
		out    = filepath.Join(dstFolder, fmt.Sprintf("%v_chunk_list.txt", fname))
		chunks = []Chunk{}
	)

	for _, file := range chunkFiles {
		chunkNum, err := getChunkNum(filepath.Base(file))
		if err != nil {
			return "", err
		}

		chunks = append(chunks, Chunk{
			Num:  chunkNum,
			Path: file,
		})
	}

	slices.SortFunc(chunks, func(a, b Chunk) int {
		return a.Num - b.Num
	})

	chunkList, err := os.Create(out)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = chunkList.Sync()
		_ = chunkList.Close()
	}()

	for _, c := range chunks {
		_, _ = fmt.Fprintf(chunkList, "file %s\n", c.Path)
	}

	return out, nil
}
