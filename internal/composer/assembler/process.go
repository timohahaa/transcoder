package assembler

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
)

const fragmentSizeSeconds = 4

func (a *Assembler) process(t task.Task) (task.Task, error) {
	var (
		lg          = a.l.WithFields(log.Fields{"task_id": t.ID})
		ctx, cancel = context.WithCancel(context.Background())
		taskDir     = filepath.Join(a.cfg.WorkDir, t.ID.String())
		assetsDir   = filepath.Join(taskDir, "assets")
		err         error
	)
	defer cancel()

	defer func() {
		_ = a.clean(t, taskDir)
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
					if taskCanceled, _ := a.mod.task.TaskCanceled(t.ID); taskCanceled {
						lg.Warn("task canceled")
						cancel()
					}
				}
			}
		}()
	}

	var progress = func(point int64) { _ = a.mod.task.UpdateProgress(ctx, t.ID, point) }

	// find quality chunks, audios, poster
	var (
		videoChunks map[string][]string
		audios      map[string]string
		poster      string
	)
	_ = poster
	{
		if videoChunks, err = a.findQualityChunks(taskDir); err != nil {
			return t, errors.Assembler(err)
		}
		if audios, err = a.findAudios(taskDir); err != nil {
			return t, errors.Assembler(err)
		}
		if poster, err = a.findPoster(taskDir); err != nil {
			return t, errors.Assembler(err)
		}
	}

	// stitch videos
	var stitchedVideos = make(map[string]string, len(videoChunks))
	for quality, chunks := range videoChunks {
		outFile, err := ffmpeg.Stitch(
			ctx,
			chunks,
			assetsDir, fmt.Sprintf("%v_stitched.mp4", quality),
		)
		if err != nil {
			return t, errors.StitchSources(err)
		}

		stitchedVideos[quality] = outFile
	}

	progress(task.ProgressAfterStitch)

	// fragment videos, audios
	var (
		fragVideos = make(map[string]string, len(stitchedVideos))
		fragAudios = make(map[string]string, len(audios))
	)

	for quality, file := range stitchedVideos {
		outFile, err := ffmpeg.Fragment(
			ctx,
			file,
			assetsDir,
			fmt.Sprintf("%v_frag.mp4", quality),
			fragmentSizeSeconds,
		)
		if err != nil {
			return t, errors.FragmentSources(err)
		}

		fragVideos[quality] = outFile
	}
	progress(task.ProgressAfterFragmentVideo)

	for quality, file := range audios {
		outFile, err := ffmpeg.Fragment(
			ctx,
			file,
			assetsDir,
			fmt.Sprintf("%v_frag.mp4", quality),
			fragmentSizeSeconds,
		)
		if err != nil {
			return t, errors.FragmentSources(err)
		}

		fragAudios[quality] = outFile
	}
	progress(task.ProgressAfterFragmentAudio)

	//	@todo:
	// maybe encrypt videos, audios if needed???
	// generate manifests
	// upload assets (where???)

	// update db
	if err := a.mod.task.UpdateStatus(ctx, t.ID, task.StatusDone, nil); err != nil {
		return t, errors.DB(err)
	}

	return t, nil
}
