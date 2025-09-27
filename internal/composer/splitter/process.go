package splitter

import (
	"context"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/analyze"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/internal/composer/modules/validate"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
)

func (s *Splitter) process(t task.Task) (task.Task, error) {
	var (
		lg          = s.l.WithFields(log.Fields{"task_id": t.ID})
		ctx, cancel = context.WithCancel(context.Background())
		taskDir     = filepath.Join(s.cfg.WorkDir, t.ID.String())
		err         error
	)
	defer cancel()

	var cleanFull, skipTask bool
	defer func() {
		if cleanFull {
			_ = s.cleanFull(t, taskDir)
		} else {
			// clean only necessary files
		}

		if skipTask {
			// mark task as skip
		}
	}()

	// check cancel
	{
		done := make(chan struct{})
		defer close(done)
		go func() {
			tic := time.NewTicker(5 * time.Second)
			defer tic.Stop()

			for range tic.C {
				select {
				case <-done:
					return
				default:
					if taskCanceled, _ := s.mod.task.TaskCanceled(t.ID); taskCanceled {
						lg.Warn("task canceled")
						cancel()
						cleanFull = true
						skipTask = true
					}
				}
			}
		}()
	}

	// download source
	var sourcePath string
	if sourcePath, err = s.downloadSource(ctx, t, filepath.Join(taskDir, "source")); err != nil {
		cleanFull = true
		return t, err
	}

	var sourceInfo *ffprobe.Info
	if sourceInfo, err = ffprobe.GetInfo(ctx, sourcePath); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}
	println(sourceInfo)

	// unmux audio/video
	var (
		videoFile  string
		audioFiles []string
	)
	if videoFile, err = ffmpeg.UnmuxVideo(
		ctx,
		sourceInfo,
		sourcePath,
		filepath.Join(taskDir, "videos"),
	); err != nil {
		cleanFull = true
		return t, errors.Unmux(err)
	}

	if audioFiles, err = ffmpeg.UnmuxAudios(
		ctx,
		sourceInfo,
		sourcePath,
		filepath.Join(taskDir, "audios"),
	); err != nil {
		cleanFull = true
		return t, errors.Unmux(err)
	}

	if err := validate.Pre(ctx, videoFile, audioFiles); err != nil {
		cleanFull = true
		return t, err
	}

	var chunks []ffmpeg.Chunk
	if chunks, err = s.split(ctx, sourceInfo, videoFile, filepath.Join(taskDir, "chunks")); err != nil {
		cleanFull = true
		return t, err
	}

	if err := validate.Chunks(ctx, chunks, sourceInfo); err != nil {
		cleanFull = true
		return t, err
	}

	// presets
	var chunkPresets map[string]analyze.ChunkPresets
	if chunkPresets, err = analyze.CalcChunkPresets(sourceInfo, chunks); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}

	_ = chunkPresets

	// write to redis

	// update db

	return t, nil
}
