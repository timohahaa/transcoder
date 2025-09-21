package splitter

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

func (s *Splitter) cleanFull(t task.Task, taskDir string) error {
	if err := os.RemoveAll(taskDir); err != nil {
		s.l.WithFields(log.Fields{"task_id": t.ID}).Errorf("clean full: %v", err)
		return err
	}
	return nil
}
