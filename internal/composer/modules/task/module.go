package task

import (
	"context"
	stdErrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/timohahaa/transcoder/pkg/errors"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

type Module struct {
	conn *pgxpool.Pool
}

var (
	ErrNoTasks = stdErrors.New("no tasks")
)

func New(conn *pgxpool.Pool) *Module {
	return &Module{
		conn: conn,
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

	if stdErrors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNoTasks
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

	if stdErrors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNoTasks
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
