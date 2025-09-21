package splitter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/request"
)

const retryAttempts = 3

func (s *Splitter) downloadSource(ctx context.Context, t task.Task, dstDir string) (string, error) {
	var lg = s.l.WithFields(log.Fields{"task_id": t.ID})

	if t.Source.FS == nil && t.Source.HTTP == nil {
		return "", errors.Splitter(fmt.Errorf("no source available"))
	}

	if t.Source.FS != nil {
		lg.Debugf("source is on local fsys: %v", t.Source.FS.Path)
		return t.Source.FS.Path, nil
	}

	var saveToPath = filepath.Join(dstDir, "original")
	switch _, err := os.Stat(saveToPath); {
	case os.IsNotExist(err), err != nil:
		if err := os.MkdirAll(filepath.Dir(saveToPath), os.ModePerm); err != nil {
			return "", errors.Splitter(err)
		}

		lg.Debugf("download source from: %v", t.Source.HTTP.URL)
		if err := request.Download(ctx, t.Source.HTTP.URL, saveToPath, retryAttempts); err != nil {
			return "", err
		}
	default:
		lg.Debugf("source already downloaded: %v", saveToPath)
	}

	return saveToPath, nil
}
