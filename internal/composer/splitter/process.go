package splitter

import (
	"context"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

func (s *Splitter) process(t task.Task) (task.Task, error) {
	var (
		lg          = s.l.WithFields(log.Fields{"task_id": t.ID})
		ctx, cancel = context.WithCancel(context.Background())
		taskDir     = filepath.Join(s.cfg.WorkDir, t.ID.String())
		sourcePath  string

		cleanFull, skipTask bool
		err                 error
	)
	defer cancel()

	defer func() {
		if cleanFull {
			// clean all files
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
	if sourcePath, err = s.downloadSource(ctx, t, filepath.Join(taskDir, "source")); err != nil {
		cleanFull = true
		return t, err
	}
	println(sourcePath)

	// get info

	// unmux audio/video

	// validate

	// split source

	// presets

	// write to redis

	// update db

	return t, nil
}
