package composer

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/timohahaa/transcoder/internal/composer/modules/queue"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	Handler struct {
		pb.UnimplementedComposerServer
		mod mod
	}

	mod struct {
		task  *task.Module
		queue *queue.Module
	}
)

func New(conn *pgxpool.Pool, redis redis.UniversalClient) *Handler {
	return &Handler{
		mod: mod{
			task:  task.New(conn),
			queue: queue.New(redis),
		},
	}
}

func (h *Handler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	return nil, nil
}
func (h *Handler) FinishTask(ctx context.Context, req *pb.FinishTaskRequest) (*emptypb.Empty, error) {
	return nil, nil
}
func (h *Handler) UpdateProgress(ctx context.Context, req *pb.UpdateProgressReq) (*emptypb.Empty, error) {
	return nil, nil
}
