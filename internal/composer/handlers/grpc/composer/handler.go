package composer

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/queue"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	return nil, nil
}
func (h *Handler) FinishTask(ctx context.Context, req *pb.FinishTaskRequest) (*emptypb.Empty, error) {
	return nil, nil
}
func (h *Handler) UpdateProgress(ctx context.Context, req *pb.UpdateProgressReq) (*emptypb.Empty, error) {
	var (
		taskID, err = uuid.FromBytes(req.ID)
		lg          = h.l.WithFields(log.Fields{"task_id": taskID})
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
