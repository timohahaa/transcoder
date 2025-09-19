package splitter

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

var hostname, _ = os.Hostname()

func (s *Splitter) watcher(idx int) {
	var (
		tick = time.NewTicker(3 * time.Second)
		lg   = s.l.WithFields(log.Fields{"watcher": idx})
	)

	lg.Info("started")
	defer func() {
		lg.Info("stopped")
		s.watcherWG.Done()
	}()

	for {
		select {
		case <-tick.C:
			t, err := s.mod.task.GetForSplitting(hostname)
			switch err {
			case nil:
				s.tasks <- t
			case task.ErrNoTasks:
				time.Sleep(5 * time.Second)
			default:
				lg.Errorf("get task: %v", err)
				time.Sleep(5 * time.Second)
			}
		case <-s.watcherDone:
			return
		}
	}
}
