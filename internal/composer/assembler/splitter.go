package assembler

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

type (
	Assembler struct {
		l     *log.Entry
		redis redis.UniversalClient
		mod   mod
		tasks chan task.Task

		watcherWG   *sync.WaitGroup
		workerWG    *sync.WaitGroup
		workerDone  chan struct{}
		watcherDone chan struct{}
		once        sync.Once
	}

	mod struct {
		task *task.Module
	}
)

func New(
	conn *pgxpool.Pool,
	redis redis.UniversalClient,
) *Assembler {
	return &Assembler{
		l:     log.WithFields(log.Fields{"mod": "assembler"}),
		redis: redis,
		mod: mod{
			task: task.New(conn),
		},
		tasks: make(chan task.Task),

		watcherWG:   new(sync.WaitGroup),
		workerWG:    new(sync.WaitGroup),
		workerDone:  make(chan struct{}),
		watcherDone: make(chan struct{}),
		once:        sync.Once{},
	}
}
func (a *Assembler) Run(workers, watchers int) {
	for i := range watchers {
		a.watcherWG.Add(1)
		go a.watcher(i)
	}

	for i := range workers {
		a.workerWG.Add(1)
		go a.worker(i)
	}
}

func (a *Assembler) Shutdown() {
	// panic if channel is closed twice
	a.once.Do(func() {
		a.l.Info("shutting down watchers...")
		close(a.watcherDone)
		a.watcherWG.Wait()

		close(a.tasks)

		a.l.Info("shutting down workers...")
		close(a.workerDone)
		a.workerWG.Wait()
	})
}

func (a *Assembler) worker(idx int) {
	lg := a.l.WithFields(log.Fields{"worker": idx})
	lg.Info("started")
	defer func() {
		lg.Info("stopped")
		a.workerWG.Done()
	}()

	for {
		select {
		case <-a.workerDone:
			for task := range a.tasks {
				st := time.Now()
				t, err := a.process(task)
				a.finishTask(t, err, time.Since(st))
			}
			return
		case task, ok := <-a.tasks:
			if !ok {
				return
			}
			st := time.Now()
			t, err := a.process(task)
			a.finishTask(t, err, time.Since(st))
		}
	}
}

func (a *Assembler) finishTask(t task.Task, err error, duration time.Duration) {}
