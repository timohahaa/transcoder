package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/timohahaa/transcoder/internal/composer/modules/task/key"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type Module struct {
	redis redis.UniversalClient
}

func New(redis redis.UniversalClient) *Module {
	return &Module{
		redis: redis,
	}
}

func (m *Module) PrepareTaskMeta(ctx context.Context, taskID uuid.UUID) error {
	taskCounterKey := key.Counter(taskID)
	if err := m.redis.Set(ctx, taskCounterKey, "0", 24*time.Hour).Err(); err != nil {
		return errors.Redis(err)
	}

	taskSkipKey := key.Skip(taskID)
	if err := m.redis.Del(ctx, taskSkipKey).Err(); err != nil {
		return errors.Redis(err)
	}

	return nil
}

func (m *Module) AddSubtask(ctx context.Context, queueKey string, subtask *pb.Task) error {
	data, err := subtask.Marshal()
	if err != nil {
		return errors.Generic(err)
	}

	if err := m.redis.RPush(ctx, queueKey, data).Err(); err != nil {
		return errors.Redis(err)
	}

	return nil
}

func (m *Module) Skip(ctx context.Context, taskID uuid.UUID) error {
	var key = key.Skip(taskID)
	if err := m.redis.Set(context.Background(), key, true, 24*time.Hour).Err(); err != nil {
		return errors.Redis(err)
	}
	return nil
}
