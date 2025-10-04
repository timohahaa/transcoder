package encoder

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/encoder/worker"
	"github.com/timohahaa/transcoder/pkg/composer"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

var hostname, _ = os.Hostname()

type (
	Service struct {
		cfg           Config
		signal        chan os.Signal
		ffmpegVersion string
		composer      *composer.Client
		workers       []*worker.Worker
		backlog       chan task
	}

	task struct {
		t  *pb.Task
		id uuid.UUID
	}
)

func New(cfg Config) (*Service, error) {
	cfg.setDefaults()

	var (
		s = &Service{
			cfg:    cfg,
			signal: make(chan os.Signal),
		}
		err error
	)

	if s.ffmpegVersion, err = ffmpeg.Version(); err != nil {
		return nil, err
	}

	if s.composer, err = composer.NewClient(cfg.ComposerAddrs); err != nil {
		return nil, err
	}

	if err = os.MkdirAll(cfg.WorkDir, os.ModePerm); err != nil {
		return nil, err
	}

	if err = s.resetTasks(); err != nil {
		return nil, err
	}

	return s, nil
}

func (srv *Service) Run() error {
	var (
		signals = []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
		}
	)

	go srv.cleanup()

	// workers
	for i := 0; i < srv.cfg.CPUQuota; i++ {
		w := worker.New(srv.composer, worker.Opts{
			CpuIdx:   i,
			MaxTasks: 1,
			WorkDir:  srv.cfg.WorkDir,
		})

		srv.workers = append(srv.workers, w)
	}

	// watcher + scheduler
	srv.backlog = make(chan task, len(srv.workers))
	for i := range srv.cfg.CPUQuota {
		go srv.watch(i)
	}

	go srv.schedule()

	signal.Notify(srv.signal, signals...)
	signal := <-srv.signal
	log.Infof("got signal: %s", signal)

	return nil
}

func (srv *Service) resetTasks() error {
	entries, err := os.ReadDir(srv.cfg.WorkDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		taskID, err := uuid.Parse(e.Name())
		if err != nil {
			continue
		}

		os.RemoveAll(filepath.Join(srv.cfg.WorkDir, taskID.String()))

		if err := srv.composer.FinishTask(context.Background(), &pb.FinishTaskRequest{
			Task: &pb.Task{
				ID:    taskID[:],
				Video: &pb.Video{},
			},
			Error: errors.TaskReset(),
		}); err != nil {
			return err
		}
	}
	return nil
}
