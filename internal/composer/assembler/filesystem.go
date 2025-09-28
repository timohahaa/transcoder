package assembler

import (
	"os"
	"path/filepath"
	"strings"

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

func (a *Assembler) findQualityChunks(taskDir string) (map[string][]string, error) {
	var (
		m            = map[string][]string{}
		path         = filepath.Join(taskDir, "chunks")
		entries, err = os.ReadDir(path)
	)

	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		q := e.Name()
		chunkPath := filepath.Join(path, q)
		m[q] = append(m[q], chunkPath)
	}

	return m, nil
}

func (a *Assembler) findAudios(taskDir string) (map[string]string, error) {
	var (
		m            = map[string]string{}
		path         = filepath.Join(taskDir, "audios")
		entries, err = os.ReadDir(path)
	)

	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if !strings.HasPrefix(e.Name(), "audio_") {
			continue
		}

		q := strings.TrimSuffix(e.Name(), ".mp4")
		audioPath := filepath.Join(path, e.Name())
		m[q] = audioPath
	}

	return m, nil
}

func (a *Assembler) findPoster(taskDir string) (string, error) {
	var path = filepath.Join(taskDir, "poster", "poster.jpg")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}
