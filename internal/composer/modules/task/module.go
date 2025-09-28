package task

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	conn *pgxpool.Pool
}

var (
	ErrNoTasks = errors.New("no tasks")
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

	if errors.Is(err, pgx.ErrNoRows) {
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

	if errors.Is(err, pgx.ErrNoRows) {
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
