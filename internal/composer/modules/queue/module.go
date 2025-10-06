package queue

import (
	"context"
	stdErrors "errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/task/key"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

const (
	queueLen     = 10
	taskTreshold = 20
)

var (
	ErrSkip    = stdErrors.New("skip task")
	ErrNoTasks = stdErrors.New("no tasks available")
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

func (m *Module) FinishSubtask(ctx context.Context, taskID uuid.UUID) (currCount int64, err error) {
	return m.redis.Incr(ctx, key.Counter(taskID)).Result()
}

func (m *Module) SkipTask(ctx context.Context, taskID uuid.UUID) error {
	var key = key.Skip(taskID)
	if err := m.redis.Set(context.Background(), key, true, 24*time.Hour).Err(); err != nil {
		return errors.Redis(err)
	}
	return nil
}

func (m *Module) GetSubtask(ctx context.Context, req *pb.GetTaskRequest) (task *pb.Task, err error) {
	var (
		inQueue int64
		routing = routing(req)
	)

	task, inQueue, err = m.next(ctx, routing)
	if err != nil {
		return
	}

	if inQueue < taskTreshold {
		if err := m.create(req, routing); err != nil {
			log.WithFields(log.Fields{
				"mod":     "queue",
				"routing": routing,
			}).Errorf("create queue: %s", err)
		}
	}

	if m.redis.Exists(ctx, key.Skip(uuid.UUID(task.ID))).Val() == 1 {
		return nil, ErrSkip
	}

	return task, nil
}

func (m *Module) next(ctx context.Context, routing string) (t *pb.Task, total int64, err error) {
	var (
		tx  = m.redis.Pipeline()
		cmd = make([]*redis.IntCmd, 0, queueLen)
	)

	for i := range queueLen {
		cmd = append(cmd, tx.LLen(ctx, routing+":"+strconv.Itoa(i)))
	}

	if _, err = tx.Exec(ctx); err != nil {
		return
	}

	for _, c := range cmd {
		total += c.Val()
	}

	if total == 0 {
		return nil, 0, ErrNoTasks
	}

	keys := make([]string, 0, queueLen)

	for i := range queueLen {
		keys = append(keys, routing+":"+strconv.Itoa(i))
	}

	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	if t, err = taskFromRedisCmd(m.redis.BLPop(ctx, time.Second, keys...)); err == nil {
		return
	}

	return nil, 0, ErrNoTasks
}

func taskFromRedisCmd(res *redis.StringSliceCmd) (*pb.Task, error) {
	switch val := res.Val(); len(val) {
	case 2:
		var task pb.Task
		if err := task.Unmarshal([]byte(val[1])); err != nil {
			return nil, ErrSkip
		}
		return &task, nil
	case 0:
		return nil, ErrNoTasks
	default:
		// should never :)
		panic(fmt.Sprintf("get task ln val: %d, %v", len(val), val))
	}
}

func routing(_ *pb.GetTaskRequest) string {
	//	@todo:	can implement custom routing logic based on hostname/hardware/etc
	return "transcoder:queue"
}

// @todo
func (m *Module) create(req *pb.GetTaskRequest, routing string) error {
	return nil
}
