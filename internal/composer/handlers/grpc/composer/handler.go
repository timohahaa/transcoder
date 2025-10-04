package composer

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/queue"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	skipTask = "SKIP_TASK"
	noTasks  = "NO_TASKS"
)

type (
	Handler struct {
		pb.UnimplementedComposerServer
		mod mod
		l   *log.Entry
	}

	mod struct {
		task  *task.Module
		queue *queue.Module
	}
)

func New(conn *pgxpool.Pool, redis redis.UniversalClient) *Handler {
	return &Handler{
		mod: mod{
			task:  task.New(conn, redis),
			queue: queue.New(redis),
		},
		l: log.WithFields(log.Fields{
			"mod": "gRPC",
		}),
	}
}

func (h *Handler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	var t, err = h.mod.queue.GetSubtask(ctx, req)

	if err != nil {
		switch err {
		case queue.ErrNoTasks:
			return nil, status.Error(codes.Unavailable, noTasks)
		case queue.ErrSkip:
			return nil, status.Error(codes.NotFound, skipTask)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return t, nil
}

func (h *Handler) UpdateProgress(ctx context.Context, req *pb.UpdateProgressReq) (*emptypb.Empty, error) {
	var (
		taskID, err = uuid.FromBytes(req.ID)
		lg          = h.l.WithFields(log.Fields{
			"task_id": taskID,
			"method":  "UpdateProgress",
		})
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid format")
	}

	if err := h.mod.task.UpdateEncodingProgress(ctx, taskID, req.Delta.AsDuration().Milliseconds()); err != nil {
		lg.Errorf("update encoding progress: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) FinishTask(ctx context.Context, req *pb.FinishTaskRequest) (*emptypb.Empty, error) {
	if req.Task == nil || (req.Task.Video == nil && req.Task.Audio == nil) {
		return nil, status.Error(codes.InvalidArgument, "no task provided in request")
	}

	var (
		taskID, err = uuid.FromBytes(req.Task.ID)
		lg          = h.l.WithFields(log.Fields{
			"task_id": taskID,
			"method":  "FinishTask",
		})
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid format")
	}

	if req.Error != nil {
		if err := h.mod.queue.SkipTask(ctx, taskID); err != nil {
			lg.Errorf("skip task: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := h.mod.task.UpdateStatus(
			ctx,
			taskID,
			task.StatusError,
			req.Error,
		); err != nil {
			lg.Errorf("update task status: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &emptypb.Empty{}, nil
	}

	currSubtaskCount, err := h.mod.queue.FinishSubtask(ctx, taskID)
	if err != nil {
		lg.Errorf("finish subtask: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch {
	case currSubtaskCount > int64(req.Task.PartsTotal):
		if err := h.mod.task.UpdateStatus(
			ctx,
			taskID,
			task.StatusError,
			errors.ChunkOverflow(req.Task.PartsTotal, int32(currSubtaskCount)),
		); err != nil {
			lg.Errorf("update task status: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	case currSubtaskCount == int64(req.Task.PartsTotal):
		if err := h.mod.task.UpdateStatus(
			ctx,
			taskID,
			task.StatusWaitingAssembling,
			nil,
		); err != nil {
			lg.Errorf("update task status: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := h.mod.task.UpdateProgress(ctx, taskID, task.ProgressAfterEncoding); err != nil {
			lg.Warnf("update task progress: %v", err)
		}
	}

	return &emptypb.Empty{}, nil
}
