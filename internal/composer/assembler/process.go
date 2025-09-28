package assembler

import (
	"context"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

func (a *Assembler) process(t task.Task) (task.Task, error) {
	var (
		lg          = a.l.WithFields(log.Fields{"task_id": t.ID})
		ctx, cancel = context.WithCancel(context.Background())
		taskDir     = filepath.Join(a.cfg.WorkDir, t.ID.String())
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

	_, _ = ctx, err

	// find quality chunks, audios, poster

	// stitch videos

	// fragment videos, audios

	// maybe encrypt videos, audios if needed???

	// generate manifests

	// upload assets

	// update db

	return t, nil
}
