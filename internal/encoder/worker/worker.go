package worker

import (
	"sync/atomic"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/pkg/composer"
	"github.com/timohahaa/transcoder/pkg/consts"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type (
	Worker struct {
		opts     Opts
		l        *log.Entry
		composer *composer.Client
		stat     struct {
			inProgress  atomic.Int32
			totalWeight atomic.Int32
		}
	}

	Opts struct {
		CpuIdx   int
		MaxTasks int32
		WorkDir  string
	}
)

func New(composer *composer.Client, o Opts) *Worker {
	return &Worker{
		opts: o,
		l: log.WithFields(log.Fields{
			"mod":     "worker",
			"cpu_idx": o.CpuIdx,
		}),
		composer: composer,
	}
}

type Stat struct {
	InProgress int
	Weight     int
	Opts
}

func (w *Worker) Stat() Stat {
	return Stat{
		InProgress: int(w.stat.inProgress.Load()),
		Weight:     int(w.stat.totalWeight.Load()),
		Opts:       w.opts,
	}
}

func (w *Worker) Handle(task *pb.Task, taskID uuid.UUID, finish func(err error)) bool {
	var weight = 100 / w.opts.MaxTasks

	if task.Video != nil {
		switch task.Video.Codec {
		case consts.CodecH264:
			weight = 60
		}
	}

	var (
		inProgress  = w.stat.inProgress.Add(1)
		totalWeight = w.stat.totalWeight.Add(weight)
		done        = func() {
			w.stat.inProgress.Add(-1)
			w.stat.totalWeight.Add(-weight)
		}
	)

	switch {
	case totalWeight > 100, inProgress > w.opts.MaxTasks:
		done()
		return false
	}

	go func() { finish(w.handle(task, taskID, done)) }()
	w.l.WithFields(log.Fields{
		"max_tasks":    w.opts.MaxTasks,
		"in_progress":  inProgress,
		"curr_weight":  weight,
		"total_weight": totalWeight,
	}).Debug("stat")
	return true
}

func (w *Worker) handle(task *pb.Task, taskID uuid.UUID, done func()) (err error) {
	defer done()

	var lg = w.l.WithFields(log.Fields{"task_id": taskID})
	switch {
	case task.Audio != nil:
		if err = w.audio(task, taskID); err != nil {
			lg.Errorf("handle audio: %v", err)
		}
	case task.Video != nil:
		if err = w.video(task, taskID); err != nil {
			lg.Errorf("handle video: %v", err)
		}
	}

	return err
}
