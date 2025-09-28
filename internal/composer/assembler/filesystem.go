package assembler

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

func (a *Assembler) clean(t task.Task, taskDir string) error {
	if err := os.RemoveAll(taskDir); err != nil {
		a.l.WithFields(log.Fields{"task_id": t.ID}).Errorf("clean: %v", err)
		return err
	}
	return nil
}
