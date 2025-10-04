package encoder

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/encoder/worker"
	"github.com/timohahaa/transcoder/pkg/composer"
	"github.com/timohahaa/transcoder/pkg/consts"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *Service) watch(idx int) {
	var (
		l = log.WithFields(log.Fields{
			"mod": "watcher",
			"idx": strconv.Itoa(idx),
		})
	)

	l.Info("started")

	for {
		t, err := srv.composer.GetTask(context.Background(), &pb.GetTaskRequest{
			Encoder:       consts.CPU,
			Hostname:      hostname,
			FFmpegVersion: srv.ffmpegVersion,
		})

		if err != nil {
			switch {
			case composer.IsSkipErr(err):
				//l.Debugf("task skipped")
			case composer.IsNoTasksErr(err):
				time.Sleep(3 * time.Second)
			case composer.IsUnavailableErr(err):
				l.Warnf("composer unavailable: %v", err)
				time.Sleep(10 * time.Second)
			default:
				l.Errorf("get task: %v", err)
				time.Sleep(time.Second)
			}
			continue
		}

		taskID, err := uuid.FromBytes(t.ID)
		if err != nil {
			srv.finishTask(t, taskID, err)
			continue
		}

		if t.Source, err = srv.prefetch(t, taskID); err != nil {
			l.WithFields(log.Fields{
				"task_id": taskID,
			}).Errorf("prefetch: %s", err)
			srv.finishTask(t, taskID, err)
			continue
		}

		srv.backlog <- task{
			t:  t,
			id: taskID,
		}
	}
}

func (srv *Service) schedule() {
	for task := range srv.backlog {
		finish := func(err error) { srv.finishTask(task.t, task.id, err) }

	LOOP:
		for {

			slices.SortFunc(srv.workers, func(a, b *worker.Worker) int {
				return a.Stat().InProgress - b.Stat().InProgress
			})

			for _, w := range srv.workers {
				if w.Handle(task.t, task.id, finish) {
					break LOOP
				}
			}

			time.Sleep(15 * time.Millisecond)
		}
	}
}

func (srv *Service) finishTask(task *pb.Task, taskID uuid.UUID, err error) {
	var tErr *pb.Error
	if err != nil {
		switch e := err.(type) {
		case *pb.Error:
			tErr = e
		default:
			tErr = errors.Unknown(e)
		}
	}

	if err := srv.composer.FinishTask(context.Background(), &pb.FinishTaskRequest{
		Task:       task,
		Error:      tErr,
		FinishedAt: timestamppb.Now(),
	}); err != nil {
		log.WithFields(log.Fields{
			"task_id": taskID,
		}).Errorf("finish task: %v", err)
	}
}
