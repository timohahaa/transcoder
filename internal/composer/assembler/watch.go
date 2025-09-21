package assembler

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

var hostname, _ = os.Hostname()

func (a *Assembler) watcher(idx int) {
	var (
		tick = time.NewTicker(3 * time.Second)
		lg   = a.l.WithFields(log.Fields{"watcher": idx})
	)

	lg.Info("started")
	defer func() {
		lg.Info("stopped")
		a.watcherWG.Done()
	}()

	for {
		select {
		case <-tick.C:
			t, err := a.mod.task.GetForSplitting(hostname)
			switch err {
			case nil:
				a.tasks <- t
			case task.ErrNoTasks:
				time.Sleep(5 * time.Second)
			default:
				lg.Errorf("get task: %v", err)
				time.Sleep(5 * time.Second)
			}
		case <-a.watcherDone:
			return
		}
	}
}
