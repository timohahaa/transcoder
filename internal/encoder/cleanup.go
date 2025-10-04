package encoder

import (
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

const cleanupPeriod = time.Minute
const cleanupMaxAge = 2 * time.Hour

func (srv *Service) cleanup() {
	tick := time.NewTicker(cleanupPeriod)
	for range tick.C {
		if err := clearDir(srv.cfg.WorkDir); err != nil {
			log.Errorf("cleanup: %v", err)
		}
		tick.Reset(cleanupPeriod)
	}
}

func clearDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, e := range entries {
		i, err := e.Info()
		if err != nil {
			return err
		}
		if time.Since(i.ModTime()) > cleanupMaxAge {
			os.RemoveAll(filepath.Join(path, e.Name()))
		}
	}
	return nil
}
