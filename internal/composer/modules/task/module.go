package task

import (
	"context"
	stdErrors "errors"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/timohahaa/transcoder/internal/composer/modules/task/key"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type Module struct {
	conn  *pgxpool.Pool
	redis redis.UniversalClient
}

var (
	ErrNoTasks = stdErrors.New("no tasks")
)

func New(conn *pgxpool.Pool, redis redis.UniversalClient) *Module {
	return &Module{
		conn:  conn,
		redis: redis,
	}
}

func (m *Module) GetForSplitting(hostname string) (Task, error) {
	var (
		t   Task
		err error
	)

	err = m.conn.QueryRow(context.Background(), getForSplittingQuery, hostname).Scan(
		&t.ID,
		&t.Source,
		&t.Encoder,
		&t.Routing,
		&t.Duration,
		&t.FileSize,
		&t.Settings,
	)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNoTasks
		}
		return Task{}, err
	}

	return t, nil
}

func (m *Module) GetForAssembling(hostname string) (Task, error) {
	var (
		t   Task
		err error
	)

	err = m.conn.QueryRow(context.Background(), getForAssemblingQuery, hostname).Scan(
		&t.ID,
		&t.Source,
		&t.Encoder,
		&t.Routing,
		&t.Duration,
		&t.FileSize,
		&t.Settings,
	)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNoTasks
		}
		return Task{}, err
	}

	return t, nil
}

func (m *Module) TaskCanceled(id uuid.UUID) (bool, error) {
	var (
		isCanceled bool
		err        error
	)

	err = m.conn.QueryRow(
		context.Background(),
		checkCancellationQuery,
		id,
	).Scan(&isCanceled)

	return isCanceled, err
}

func (m *Module) UpdateStatus(ctx context.Context, taskID uuid.UUID, status string, extErr error) error {
	var pbErr *pb.Error
	if extErr != nil {
		status = StatusError
		switch e := extErr.(type) {
		case *pb.Error:
			if e != nil {
				pbErr = e
			}
		default:
			pbErr = errors.Unknown(e)
		}
	}

	_, err := m.conn.Exec(ctx, updateTaskStatusQuery, taskID, status, pbErr)
	return err
}

func (m *Module) UpdateEncodingProgress(ctx context.Context, taskID uuid.UUID, delta int64) error {
	var (
		videoDur float64
	)

	if err := m.conn.QueryRow(ctx, getTaskDurationQuery, taskID).Scan(&videoDur); err != nil {
		return err
	}

	videoDur = math.Round(videoDur * 1000) // to milliseconds

	currEncProg, err := m.redis.IncrBy(ctx, key.EncodingProgress(taskID), int64(videoDur)).Result()
	if err != nil {
		return err
	}

	currEncProg = min(currEncProg, int64(videoDur))

	// transcodingPercentRange - how much an encoded delta contributed to percent progress
	// + progressAfterSplitting cause we started at 0
	percentProg := math.Round(
		float64(currEncProg)*encodingPercentRange/float64(videoDur),
	) + progressAfterSplitting

	return m.redis.Set(ctx, key.Progress(taskID), percentProg, 12*time.Hour).Err()
}

func (m *Module) UpdateProgress(ctx context.Context, taskID uuid.UUID, point int64) error {
	return m.redis.Set(ctx, key.Progress(taskID), point, 12*time.Hour).Err()
}

func (m *Module) GetProgress(ctx context.Context, taskID uuid.UUID) (int64, error) {
	strVal, err := m.redis.Get(ctx, key.Progress(taskID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(strVal, 10, 64)
}

func (m *Module) Create(ctx context.Context, form CreateForm) (Task, error) {
	var (
		t   Task
		err error
	)
	err = m.conn.QueryRow(ctx, createQuery,
		form.Source,
		form.Duration,
		form.FileSize,
		form.Settings,
	).Scan(
		&t.ID,
		&t.Source,
		&t.Status,
		&t.Encoder,
		&t.Routing,
		&t.Duration,
		&t.FileSize,
		&t.Settings,
		&t.Error,
	)
	return t, err
}

func (m *Module) Get(ctx context.Context, id uuid.UUID) (Task, error) {
	var (
		t   Task
		err error
	)
	err = m.conn.QueryRow(ctx, getQuery, id).Scan(
		&t.ID,
		&t.Source,
		&t.Status,
		&t.Encoder,
		&t.Routing,
		&t.Duration,
		&t.FileSize,
		&t.Settings,
		&t.Error,
	)
	return t, err
}

func (m *Module) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := m.conn.Exec(ctx, deleteQuery, id)
	return err
}
