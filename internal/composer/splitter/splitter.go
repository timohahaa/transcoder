package splitter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/queue"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type (
	Splitter struct {
		l     *log.Entry
		cfg   Config
		mod   mod
		tasks chan task.Task

		watcherWG   *sync.WaitGroup
		workerWG    *sync.WaitGroup
		workerDone  chan struct{}
		watcherDone chan struct{}
		once        sync.Once
	}

	mod struct {
		task  *task.Module
		queue *queue.Module
	}

	Config struct {
		HttpAddr string
		WorkDir  string
	}
)

func New(
	conn *pgxpool.Pool,
	redis redis.UniversalClient,
	cfg Config,
) *Splitter {
	return &Splitter{
		l:   log.WithFields(log.Fields{"mod": "splitter"}),
		cfg: cfg,
		mod: mod{
			task:  task.New(conn, redis),
			queue: queue.New(redis),
		},
		tasks: make(chan task.Task),

		watcherWG:   new(sync.WaitGroup),
		workerWG:    new(sync.WaitGroup),
		workerDone:  make(chan struct{}),
		watcherDone: make(chan struct{}),
		once:        sync.Once{},
	}
}
func (s *Splitter) Run(workers, watchers int) {
	for i := range watchers {
		s.watcherWG.Add(1)
		go s.watcher(i)
	}

	for i := range workers {
		s.workerWG.Add(1)
		go s.worker(i)
	}
}

func (s *Splitter) Shutdown() {
	// panic if channel is closed twice
	s.once.Do(func() {
		s.l.Info("shutting down watchers...")
		close(s.watcherDone)
		s.watcherWG.Wait()

		close(s.tasks)

		s.l.Info("shutting down workers...")
		close(s.workerDone)
		s.workerWG.Wait()
	})
}

func (s *Splitter) worker(idx int) {
	lg := s.l.WithFields(log.Fields{"worker": idx})
	lg.Info("started")
	defer func() {
		lg.Info("stopped")
		s.workerWG.Done()
	}()

	for {
		select {
		case <-s.workerDone:
			for task := range s.tasks { // finish whats left before exit
				st := time.Now()
				t, err := s.process(task)
				s.finishTask(t, err, time.Since(st))
			}
			return
		case task, ok := <-s.tasks:
			if !ok {
				continue
			}
			fmt.Println(task)
			st := time.Now()
			t, err := s.process(task)
			s.finishTask(t, err, time.Since(st))
		}
	}
}

func (s *Splitter) finishTask(t task.Task, err error, duration time.Duration) {
	if err == nil {
		return
	}

	var (
		pbErr *pb.Error
		lg    = log.WithFields(log.Fields{"task_id": t.ID})
	)
	switch e := err.(type) {
	case *pb.Error:
		pbErr = e
	default:
		pbErr = errors.Unknown(e)
	}

	if err := s.mod.task.UpdateStatus(context.Background(), t.ID, task.StatusError, pbErr); err != nil {
		lg.Errorf("update task status: %v", err)
	}
}
