package splitter

import (
	"os"
	"path/filepath"

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

// need to delete only video sources
// keep chunks and audios
func (s *Splitter) clean(t task.Task, taskDir string) error {
	for _, dir := range []string{
		filepath.Join(taskDir, "video"),
		filepath.Join(taskDir, "original"),
	} {
		if err := os.RemoveAll(dir); err != nil {
			s.l.WithFields(log.Fields{"task_id": t.ID}).Errorf("clean: %v", err)
			return err
		}
	}
	return nil
}
